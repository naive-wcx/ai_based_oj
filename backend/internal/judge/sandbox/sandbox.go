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
	"time"

	"oj-system/internal/model"
)

// ExecuteResult 执行结果
type ExecuteResult struct {
	Status     string
	Time       int    // ms
	Memory     int    // KB
	Output     string
	Error      string
	ExitCode   int
}

// LanguageConfig 语言配置
type LanguageConfig struct {
	SourceFile   string
	CompileCmd   []string
	ExecuteCmd   []string
	NeedCompile  bool
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
	Execute(workDir string, language string, code string, input string, timeLimit int, memoryLimit int) (*ExecuteResult, error)
}

// SimpleSandbox 简单沙箱（开发测试用）
type SimpleSandbox struct{}

// NewSimpleSandbox 创建简单沙箱
func NewSimpleSandbox() *SimpleSandbox {
	return &SimpleSandbox{}
}

// Execute 执行代码
func (s *SimpleSandbox) Execute(workDir string, language string, code string, input string, timeLimit int, memoryLimit int) (*ExecuteResult, error) {
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
	return s.run(workDir, config.ExecuteCmd, input, timeLimit, memoryLimit)
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
func (s *SimpleSandbox) run(workDir string, cmd []string, input string, timeLimit int, memoryLimit int) (*ExecuteResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeLimit+1000)*time.Millisecond)
	defer cancel()

	execCmd := buildRunCommand(ctx, cmd, memoryLimit)
	execCmd.Dir = workDir

	// 设置输入
	execCmd.Stdin = strings.NewReader(input)

	var stdout, stderr bytes.Buffer
	execCmd.Stdout = &stdout
	execCmd.Stderr = &stderr

	startTime := time.Now()
	err := execCmd.Run()
	elapsed := time.Since(startTime)
	timeUsed := int(elapsed.Milliseconds())

	result := &ExecuteResult{
		Time:   timeUsed,
		Memory: 0, // 简单沙箱不统计内存
		Output: stdout.String(),
	}

	// 检查超时
	if ctx.Err() == context.DeadlineExceeded {
		result.Status = model.StatusTimeLimitExceeded
		return result, nil
	}

	// 检查运行时错误
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
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

	result.Status = "OK"
	return result, nil
}

// buildRunCommand 构建运行命令。
// Linux 下将进程栈上限设置为 memoryLimit（MB）对应的 KB，使栈空间与题目空间限制同量级。
func buildRunCommand(ctx context.Context, cmd []string, memoryLimit int) *exec.Cmd {
	if len(cmd) == 0 {
		return exec.CommandContext(ctx, "")
	}
	if runtime.GOOS != "linux" || memoryLimit <= 0 {
		return exec.CommandContext(ctx, cmd[0], cmd[1:]...)
	}
	if _, err := exec.LookPath("bash"); err != nil {
		return exec.CommandContext(ctx, cmd[0], cmd[1:]...)
	}

	stackKB := memoryLimit * 1024
	args := []string{
		"-c",
		"ulimit -s \"$1\" && shift && exec \"$@\"",
		"sandbox",
		strconv.Itoa(stackKB),
	}
	args = append(args, cmd...)

	return exec.CommandContext(ctx, "bash", args...)
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
