package sandbox

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"oj-system/internal/model"
)

// ExecuteResult 执行结果
type ExecuteResult struct {
	Status   string
	Time     int // ms
	Memory   int // KB
	Output   string
	Error    string
	ExitCode int
}

// LanguageConfig 语言配置
type LanguageConfig struct {
	SourceFile  string
	CompileCmd  []string
	ExecuteCmd  []string
	NeedCompile bool
}

// 语言配置表
var languageConfigs = map[string]LanguageConfig{
	"c": {
		SourceFile:  "main.c",
		CompileCmd:  []string{"gcc", "-o", "main", "main.c", "-O2", "-Wall", "-lm", "-std=c11"},
		ExecuteCmd:  []string{"./main"},
		NeedCompile: true,
	},
	"cpp": {
		SourceFile:  "main.cpp",
		CompileCmd:  []string{"g++", "-o", "main", "main.cpp", "-O2", "-Wall", "-std=c++17"},
		ExecuteCmd:  []string{"./main"},
		NeedCompile: true,
	},
	"python": {
		SourceFile:  "main.py",
		CompileCmd:  nil,
		ExecuteCmd:  []string{"python3", "main.py"},
		NeedCompile: false,
	},
	"java": {
		SourceFile:  "Main.java",
		CompileCmd:  []string{"javac", "Main.java"},
		ExecuteCmd:  []string{"java", "Main"},
		NeedCompile: true,
	},
	"go": {
		SourceFile:  "main.go",
		CompileCmd:  []string{"go", "build", "-o", "main", "main.go"},
		ExecuteCmd:  []string{"./main"},
		NeedCompile: true,
	},
}

// Sandbox 沙箱接口
type Sandbox interface {
	Execute(workDir string, language string, code string, input string, timeLimit int, memoryLimit int, submissionID uint) (*ExecuteResult, error)
}

// SimpleSandbox 简单沙箱（开发测试用）
type SimpleSandbox struct{}

type processMemorySample struct {
	vmPeakKB       int
	vmCurrentKB    int
	residentPeakKB int
	exceeded       bool
}

var submissionControl = struct {
	mu       sync.Mutex
	process  map[uint]*os.Process
	abortSet map[uint]struct{}
}{
	process:  make(map[uint]*os.Process),
	abortSet: make(map[uint]struct{}),
}

// NewSimpleSandbox 创建简单沙箱
func NewSimpleSandbox() *SimpleSandbox {
	return &SimpleSandbox{}
}

// Execute 执行代码
func (s *SimpleSandbox) Execute(workDir string, language string, code string, input string, timeLimit int, memoryLimit int, submissionID uint) (*ExecuteResult, error) {
	config, ok := languageConfigs[language]
	if !ok {
		return &ExecuteResult{
			Status: model.StatusSystemError,
			Error:  "不支持的编程语言",
		}, nil
	}

	// 创建工作目录
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return &ExecuteResult{
			Status: model.StatusSystemError,
			Error:  "创建工作目录失败",
		}, err
	}

	// 写入源代码
	sourceFile := filepath.Join(workDir, config.SourceFile)
	if err := os.WriteFile(sourceFile, []byte(code), 0644); err != nil {
		return &ExecuteResult{
			Status: model.StatusSystemError,
			Error:  "写入源代码失败",
		}, err
	}

	// 编译（如果需要）
	if config.NeedCompile {
		compileResult := s.compile(workDir, config.CompileCmd)
		if compileResult.Status != "" {
			return compileResult, nil
		}
	}

	// 执行
	return s.run(workDir, config.ExecuteCmd, input, timeLimit, memoryLimit, submissionID)
}

// compile 编译代码
func (s *SimpleSandbox) compile(workDir string, cmd []string) *ExecuteResult {
	if len(cmd) == 0 {
		return &ExecuteResult{}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	execCmd := exec.CommandContext(ctx, cmd[0], cmd[1:]...)
	execCmd.Dir = workDir

	var stderr bytes.Buffer
	execCmd.Stderr = &stderr

	if err := execCmd.Run(); err != nil {
		return &ExecuteResult{
			Status: model.StatusCompileError,
			Error:  stderr.String(),
		}
	}

	return &ExecuteResult{}
}

// run 运行程序
func (s *SimpleSandbox) run(workDir string, cmd []string, input string, timeLimit int, memoryLimit int, submissionID uint) (*ExecuteResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeLimit+1000)*time.Millisecond)
	defer cancel()

	execCmd := buildRunCommand(ctx, cmd, memoryLimit)
	execCmd.Dir = workDir

	// 设置输入
	execCmd.Stdin = strings.NewReader(input)

	var stdout, stderr bytes.Buffer
	execCmd.Stdout = &stdout
	execCmd.Stderr = &stderr

	if err := execCmd.Start(); err != nil {
		return &ExecuteResult{
			Status: model.StatusSystemError,
			Error:  err.Error(),
		}, nil
	}
	registerSubmissionProcess(submissionID, execCmd.Process)
	defer unregisterSubmissionProcess(submissionID, execCmd.Process)

	memoryLimitKB := 0
	if memoryLimit > 0 {
		memoryLimitKB = memoryLimit * 1024
	}
	monitorStop, monitorResult := startProcessMemoryMonitor(execCmd.Process, memoryLimitKB)

	startTime := time.Now()
	err := execCmd.Wait()
	monitorStop()
	sample := <-monitorResult
	elapsed := time.Since(startTime)
	timeUsed := int(elapsed.Milliseconds())
	memoryUsed := sample.vmPeakKB
	if memoryUsed <= 0 {
		memoryUsed = sample.vmCurrentKB
	}
	if memoryUsed <= 0 {
		memoryUsed = getProcessMaxRSSKB(execCmd.ProcessState)
	}
	memoryExceeded := sample.exceeded || (memoryLimitKB > 0 && memoryUsed > memoryLimitKB)

	result := &ExecuteResult{
		Time:   timeUsed,
		Memory: memoryUsed,
		Output: stdout.String(),
	}

	// 检查超时
	if ctx.Err() == context.DeadlineExceeded {
		result.Status = model.StatusTimeLimitExceeded
		return result, nil
	}

	if IsSubmissionAbortRequested(submissionID) {
		result.Status = model.StatusSystemError
		result.Error = "管理员已终止评测"
		return result, nil
	}

	// 检查运行时错误
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
			if memoryExceeded || looksLikeMemoryLimitError(stderr.String()) {
				result.Status = model.StatusMemoryLimitExceeded
				return result, nil
			}
			result.Status = model.StatusRuntimeError
			result.Error = stderr.String()
			return result, nil
		}
		result.Status = model.StatusSystemError
		result.Error = err.Error()
		return result, nil
	}

	// 检查时间限制
	if timeUsed > timeLimit {
		result.Status = model.StatusTimeLimitExceeded
		return result, nil
	}

	if memoryExceeded {
		result.Status = model.StatusMemoryLimitExceeded
		return result, nil
	}

	result.Status = "OK"
	return result, nil
}

// buildRunCommand 构建运行命令。
// Linux 下将进程栈上限与虚拟内存上限都设置为 memoryLimit（MB）对应的 KB。
func buildRunCommand(ctx context.Context, cmd []string, memoryLimit int) *exec.Cmd {
	if len(cmd) == 0 {
		return exec.CommandContext(ctx, "")
	}

	if runtime.GOOS != "linux" {
		return exec.CommandContext(ctx, cmd[0], cmd[1:]...)
	}
	if _, err := exec.LookPath("bash"); err != nil {
		return exec.CommandContext(ctx, cmd[0], cmd[1:]...)
	}

	limitKB := 0
	if memoryLimit > 0 {
		limitKB = memoryLimit * 1024
	}
	args := []string{
		"-c",
		"if [ \"$1\" -gt 0 ]; then ulimit -s \"$1\"; ulimit -v \"$1\"; fi; shift; exec \"$@\"",
		"sandbox",
		strconv.Itoa(limitKB),
	}
	args = append(args, cmd...)

	return exec.CommandContext(ctx, "bash", args...)
}

func startProcessMemoryMonitor(process *os.Process, limitKB int) (func(), <-chan processMemorySample) {
	done := make(chan struct{})
	result := make(chan processMemorySample, 1)

	if process == nil || process.Pid <= 0 || runtime.GOOS != "linux" {
		result <- processMemorySample{}
		close(result)
		return func() {}, result
	}

	go func(pid int) {
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()

		sample := processMemorySample{}
		for {
			vmPeak, vmSize, rss := readLinuxProcessMemoryKB(pid)
			if vmPeak > sample.vmPeakKB {
				sample.vmPeakKB = vmPeak
			}
			if vmSize > sample.vmCurrentKB {
				sample.vmCurrentKB = vmSize
			}
			if rss > sample.residentPeakKB {
				sample.residentPeakKB = rss
			}
			if limitKB > 0 && sample.vmPeakKB > limitKB {
				sample.exceeded = true
				_ = process.Kill()
			}

			select {
			case <-done:
				result <- sample
				close(result)
				return
			case <-ticker.C:
			}
		}
	}(process.Pid)

	stop := func() {
		close(done)
	}
	return stop, result
}

func readLinuxProcessMemoryKB(pid int) (vmPeak, vmSize, vmRSS int) {
	statusPath := filepath.Join("/proc", strconv.Itoa(pid), "status")
	data, err := os.ReadFile(statusPath)
	if err != nil {
		return 0, 0, 0
	}

	vmPeak = parseProcStatusKB(data, "VmPeak:")
	vmSize = parseProcStatusKB(data, "VmSize:")
	vmRSS = parseProcStatusKB(data, "VmRSS:")
	return vmPeak, vmSize, vmRSS
}

func parseProcStatusKB(data []byte, key string) int {
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if !strings.HasPrefix(line, key) {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			return 0
		}
		value, err := strconv.Atoi(fields[1])
		if err == nil && value > 0 {
			return value
		}
		return 0
	}

	return 0
}

func looksLikeMemoryLimitError(stderr string) bool {
	if stderr == "" {
		return false
	}
	lower := strings.ToLower(stderr)
	keywords := []string{
		"bad_alloc",
		"cannot allocate memory",
		"memoryerror",
		"out of memory",
		"killed",
	}
	for _, kw := range keywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

// CompareOutput 比较输出
func CompareOutput(expected, actual string) bool {
	// 规范化处理：去除首尾空白，统一换行符
	expected = normalizeOutput(expected)
	actual = normalizeOutput(actual)
	return expected == actual
}

// normalizeOutput 规范化输出
func normalizeOutput(s string) string {
	// 替换 \r\n 为 \n
	s = strings.ReplaceAll(s, "\r\n", "\n")
	// 去除首尾空白
	s = strings.TrimSpace(s)
	// 去除每行末尾的空白
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	return strings.Join(lines, "\n")
}

// GetWorkDir 获取工作目录
func GetWorkDir(submissionID uint) string {
	return fmt.Sprintf("./data/sandbox/%d", submissionID)
}

// CleanWorkDir 清理工作目录
func CleanWorkDir(workDir string) {
	os.RemoveAll(workDir)
}

func registerSubmissionProcess(submissionID uint, process *os.Process) {
	if submissionID == 0 || process == nil {
		return
	}
	submissionControl.mu.Lock()
	submissionControl.process[submissionID] = process
	submissionControl.mu.Unlock()
}

func unregisterSubmissionProcess(submissionID uint, process *os.Process) {
	if submissionID == 0 {
		return
	}
	submissionControl.mu.Lock()
	current := submissionControl.process[submissionID]
	if process == nil || current == process {
		delete(submissionControl.process, submissionID)
	}
	submissionControl.mu.Unlock()
}

// RequestAbortSubmission 请求终止指定提交的评测。
func RequestAbortSubmission(submissionID uint) bool {
	if submissionID == 0 {
		return false
	}

	var process *os.Process
	submissionControl.mu.Lock()
	submissionControl.abortSet[submissionID] = struct{}{}
	process = submissionControl.process[submissionID]
	submissionControl.mu.Unlock()

	if process != nil {
		_ = process.Kill()
		return true
	}
	return false
}

// IsSubmissionAbortRequested 判断指定提交是否已收到终止请求。
func IsSubmissionAbortRequested(submissionID uint) bool {
	if submissionID == 0 {
		return false
	}
	submissionControl.mu.Lock()
	_, ok := submissionControl.abortSet[submissionID]
	submissionControl.mu.Unlock()
	return ok
}

// ClearSubmissionAbortRequest 清理指定提交的终止标记。
func ClearSubmissionAbortRequest(submissionID uint) {
	if submissionID == 0 {
		return
	}
	submissionControl.mu.Lock()
	delete(submissionControl.abortSet, submissionID)
	delete(submissionControl.process, submissionID)
	submissionControl.mu.Unlock()
}
