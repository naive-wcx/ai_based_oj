package judge

import (
	"log"
	"os"
	"path/filepath"

	"oj-system/internal/config"
	"oj-system/internal/judge/ai"
	"oj-system/internal/judge/queue"
	"oj-system/internal/judge/sandbox"
	"oj-system/internal/model"
	"oj-system/internal/repository"
	"oj-system/internal/service"
)

// Judger 判题器
type Judger struct {
	sandbox           sandbox.Sandbox
	aiClient          *ai.DeepSeekClient
	submissionService *service.SubmissionService
	problemRepo       *repository.ProblemRepository
}

// NewJudger 创建判题器
func NewJudger(cfg *config.Config) *Judger {
	return &Judger{
		sandbox:           sandbox.NewSimpleSandbox(),
		aiClient:          ai.NewDeepSeekClient(),
		submissionService: service.NewSubmissionService(),
		problemRepo:       repository.NewProblemRepository(),
	}
}

// Start 启动判题服务
func Start(cfg *config.Config) {
	judger := NewJudger(cfg)

	// 初始化队列
	queue.Init(100)
	q := queue.GetQueue()

	// 注册判题处理器
	q.RegisterHandler(judger.Handle)

	// 启动 worker
	q.Start(cfg.Judge.Workers)

	log.Printf("[Judger] 判题服务已启动")
}

// Handle 处理判题任务
func (j *Judger) Handle(task *queue.JudgeTask) {
	submission := task.Submission
	problem := task.Problem
	testcases := task.Testcases

	log.Printf("[Judger] 开始判题: submission_id=%d, problem_id=%d", submission.ID, problem.ID)

	// 更新状态为 Judging
	submission.Status = model.StatusJudging
	j.submissionService.UpdateResult(submission)

	// 1. 传统评测
	testcaseResults := j.runTestcases(submission, problem, testcases)
	submission.TestcaseResults = testcaseResults

	// 计算传统评测结果
	traditionalStatus := j.calculateTraditionalStatus(testcaseResults)
	submission.Status = traditionalStatus

	// 计算最大时间和内存
	var maxTime, maxMemory int
	for _, r := range testcaseResults {
		if r.Time > maxTime {
			maxTime = r.Time
		}
		if r.Memory > maxMemory {
			maxMemory = r.Memory
		}
	}
	submission.TimeUsed = maxTime
	submission.MemoryUsed = maxMemory

	// 2. AI 评测（如果启用）
	if problem.AIJudgeConfig != nil && problem.AIJudgeConfig.Enabled {
		log.Printf("[Judger] 执行 AI 判题: submission_id=%d", submission.ID)
		aiResult, err := j.aiClient.AnalyzeCode(problem, submission.Code, submission.Language)
		if err != nil {
			log.Printf("[Judger] AI 判题出错: %v", err)
		}
		submission.AIJudgeResult = aiResult

		// 如果传统评测通过但 AI 判定不通过
		if aiResult != nil && !aiResult.Passed && traditionalStatus == model.StatusAccepted {
			if problem.AIJudgeConfig.StrictMode {
				// 严格模式：AI 不通过则 WA
				submission.Status = model.StatusWrongAnswer
				submission.FinalMessage = "测试点全部通过，但未满足题目的算法/语言要求，判定为 Wrong Answer"
			} else {
				// 非严格模式：仅记录 AI 结果，不影响最终状态
				submission.FinalMessage = "测试点全部通过。AI 分析提示：" + aiResult.Reason
			}
		}
	}

	// 计算得分
	baseScore := j.calculateScore(testcaseResults, submission.Status == model.StatusAccepted)
	if submission.AIJudgeResult != nil && !submission.AIJudgeResult.Passed {
		if baseScore > 50 {
			baseScore = 50
		}
	}
	submission.Score = baseScore

	// 保存结果
	if err := j.submissionService.UpdateResult(submission); err != nil {
		log.Printf("[Judger] 保存结果失败: %v", err)
	}

	// 清理工作目录
	workDir := sandbox.GetWorkDir(submission.ID)
	sandbox.CleanWorkDir(workDir)

	log.Printf("[Judger] 判题完成: submission_id=%d, status=%s, score=%d",
		submission.ID, submission.Status, submission.Score)
}

// runTestcases 运行所有测试点
func (j *Judger) runTestcases(submission *model.Submission, problem *model.Problem, testcases []model.Testcase) []model.TestcaseResult {
	var results []model.TestcaseResult
	workDir := sandbox.GetWorkDir(submission.ID)
	fileIOEnabled := problem.FileIOEnabled && problem.FileInputName != "" && problem.FileOutputName != ""
	inputName := filepath.Base(problem.FileInputName)
	outputName := filepath.Base(problem.FileOutputName)

	for i, tc := range testcases {
		// 读取输入输出
		input, err := os.ReadFile(tc.InputFile)
		if err != nil {
			results = append(results, model.TestcaseResult{
				ID:      i + 1,
				Status:  model.StatusSystemError,
				Message: "读取测试输入失败",
			})
			continue
		}

		expectedOutput, err := os.ReadFile(tc.OutputFile)
		if err != nil {
			results = append(results, model.TestcaseResult{
				ID:      i + 1,
				Status:  model.StatusSystemError,
				Message: "读取测试输出失败",
			})
			continue
		}

		// 执行代码
		if fileIOEnabled {
			inputPath := filepath.Join(workDir, inputName)
			outputPath := filepath.Join(workDir, outputName)
			if err := os.WriteFile(inputPath, input, 0644); err != nil {
				results = append(results, model.TestcaseResult{
					ID:      i + 1,
					Status:  model.StatusSystemError,
					Message: "写入输入文件失败",
				})
				continue
			}
			_ = os.Remove(outputPath)
		}

		execInput := string(input)
		if fileIOEnabled {
			execInput = ""
		}
		execResult, err := j.sandbox.Execute(
			workDir,
			submission.Language,
			submission.Code,
			execInput,
			problem.TimeLimit,
			problem.MemoryLimit,
		)

		if err != nil {
			results = append(results, model.TestcaseResult{
				ID:      i + 1,
				Status:  model.StatusSystemError,
				Message: err.Error(),
			})
			continue
		}

		// 处理编译错误
		if execResult.Status == model.StatusCompileError {
			submission.CompileError = execResult.Error
			results = append(results, model.TestcaseResult{
				ID:      i + 1,
				Status:  model.StatusCompileError,
				Message: "编译错误",
			})
			// 编译错误，后续测试点跳过
			for k := i + 1; k < len(testcases); k++ {
				results = append(results, model.TestcaseResult{
					ID:     k + 1,
					Status: model.StatusCompileError,
				})
			}
			break
		}

		result := model.TestcaseResult{
			ID:     i + 1,
			Status: execResult.Status,
			Time:   execResult.Time,
			Memory: execResult.Memory,
		}

		// 如果运行成功，比较输出
		if execResult.Status == "OK" {
			actualOutput := execResult.Output
			if fileIOEnabled {
				outputPath := filepath.Join(workDir, outputName)
				outData, err := os.ReadFile(outputPath)
				if err != nil {
					result.Status = model.StatusWrongAnswer
					result.Message = "未生成输出文件"
					results = append(results, result)
					continue
				}
				actualOutput = string(outData)
			}

			if sandbox.CompareOutput(string(expectedOutput), actualOutput) {
				result.Status = model.StatusAccepted
			} else {
				result.Status = model.StatusWrongAnswer
			}
		}

		results = append(results, result)
	}

	return results
}

// calculateTraditionalStatus 计算传统评测状态
func (j *Judger) calculateTraditionalStatus(results []model.TestcaseResult) string {
	if len(results) == 0 {
		return model.StatusSystemError
	}

	allAccepted := true
	for _, r := range results {
		if r.Status != model.StatusAccepted {
			allAccepted = false
			// 返回第一个非 AC 的状态
			if r.Status == model.StatusCompileError {
				return model.StatusCompileError
			}
		}
	}

	if allAccepted {
		return model.StatusAccepted
	}

	// 找出最严重的错误状态
	statusPriority := map[string]int{
		model.StatusCompileError:        1,
		model.StatusRuntimeError:        2,
		model.StatusTimeLimitExceeded:   3,
		model.StatusMemoryLimitExceeded: 4,
		model.StatusWrongAnswer:         5,
		model.StatusSystemError:         6,
	}

	worstStatus := model.StatusAccepted
	worstPriority := 100

	for _, r := range results {
		if p, ok := statusPriority[r.Status]; ok && p < worstPriority {
			worstPriority = p
			worstStatus = r.Status
		}
	}

	return worstStatus
}

// calculateScore 计算得分
func (j *Judger) calculateScore(results []model.TestcaseResult, allPassed bool) int {
	if allPassed {
		return 100
	}

	// 按通过的测试点数计算得分
	if len(results) == 0 {
		return 0
	}

	passed := 0
	for _, r := range results {
		if r.Status == model.StatusAccepted {
			passed++
		}
	}

	return passed * 100 / len(results)
}

// SubmitToQueue 提交到判题队列
func SubmitToQueue(submission *model.Submission) error {
	problemRepo := repository.NewProblemRepository()

	// 获取题目
	problem, err := problemRepo.GetByID(submission.ProblemID)
	if err != nil {
		return err
	}

	// 获取测试用例
	testcases, err := problemRepo.GetTestcases(submission.ProblemID)
	if err != nil {
		return err
	}

	// 创建任务
	task := &queue.JudgeTask{
		Submission: submission,
		Problem:    problem,
		Testcases:  testcases,
	}

	// 加入队列
	return queue.GetQueue().Push(task)
}
