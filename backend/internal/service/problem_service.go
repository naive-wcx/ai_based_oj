package service

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"oj-system/internal/config"
	"oj-system/internal/model"
	"oj-system/internal/repository"
)

type ProblemService struct {
	repo *repository.ProblemRepository
	submissionRepo *repository.SubmissionRepository
}

func NewProblemService() *ProblemService {
	return &ProblemService{
		repo: repository.NewProblemRepository(),
		submissionRepo: repository.NewSubmissionRepository(),
	}
}

// Create 创建题目
func (s *ProblemService) Create(req *model.ProblemCreateRequest, createdBy uint) (*model.Problem, error) {
	fileEnabled, inputName, outputName, err := normalizeFileIO(req)
	if err != nil {
		return nil, err
	}

	problem := &model.Problem{
		Title:         req.Title,
		Description:   req.Description,
		InputFormat:   req.InputFormat,
		OutputFormat:  req.OutputFormat,
		Samples:       req.Samples,
		TimeLimit:     req.TimeLimit,
		MemoryLimit:   req.MemoryLimit,
		Difficulty:    req.Difficulty,
		Tags:          req.Tags,
		AIJudgeConfig: req.AIJudgeConfig,
		FileIOEnabled: fileEnabled,
		FileInputName: inputName,
		FileOutputName: outputName,
		IsPublic:      req.IsPublic,
		CreatedBy:     createdBy,
	}

	// 设置默认值
	if problem.TimeLimit == 0 {
		problem.TimeLimit = 1000
	}
	if problem.MemoryLimit == 0 {
		problem.MemoryLimit = 256
	}

	if err := s.repo.Create(problem); err != nil {
		return nil, errors.New("创建题目失败")
	}

	// 创建题目数据目录
	problemDir := filepath.Join(config.GlobalConfig.Paths.Problems, fmt.Sprintf("%d", problem.ID))
	os.MkdirAll(problemDir, 0755)

	return problem, nil
}

// GetByID 获取题目详情
func (s *ProblemService) GetByID(id uint) (*model.Problem, error) {
	problem, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("题目不存在")
	}
	return problem, nil
}

// GetByIDWithUser 获取题目详情（含通过标识）
func (s *ProblemService) GetByIDWithUser(id uint, userID uint) (*model.Problem, error) {
	problem, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("题目不存在")
	}
	if userID > 0 {
		hasAccepted := s.submissionRepo.HasAccepted(userID, id)
		problem.HasAccepted = hasAccepted
	}
	return problem, nil
}

// Update 更新题目
func (s *ProblemService) Update(id uint, req *model.ProblemCreateRequest) (*model.Problem, error) {
	problem, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("题目不存在")
	}

	fileEnabled, inputName, outputName, err := normalizeFileIO(req)
	if err != nil {
		return nil, err
	}

	problem.Title = req.Title
	problem.Description = req.Description
	problem.InputFormat = req.InputFormat
	problem.OutputFormat = req.OutputFormat
	problem.Samples = req.Samples
	problem.TimeLimit = req.TimeLimit
	problem.MemoryLimit = req.MemoryLimit
	problem.Difficulty = req.Difficulty
	problem.Tags = req.Tags
	problem.AIJudgeConfig = req.AIJudgeConfig
	problem.FileIOEnabled = fileEnabled
	problem.FileInputName = inputName
	problem.FileOutputName = outputName
	problem.IsPublic = req.IsPublic

	if err := s.repo.Update(problem); err != nil {
		return nil, errors.New("更新题目失败")
	}

	return problem, nil
}

// Delete 删除题目
func (s *ProblemService) Delete(id uint) error {
	_, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New("题目不存在")
	}

	// 删除题目数据目录
	problemDir := filepath.Join(config.GlobalConfig.Paths.Problems, fmt.Sprintf("%d", id))
	os.RemoveAll(problemDir)

	return s.repo.Delete(id)
}

// List 获取题目列表
func (s *ProblemService) List(page, size int, difficulty, tag, keyword string) ([]model.ProblemListItem, int64, error) {
	problems, total, err := s.repo.List(page, size, difficulty, tag, keyword)
	if err != nil {
		return nil, 0, err
	}

	var items []model.ProblemListItem
	for _, p := range problems {
		hasAI := p.AIJudgeConfig != nil && p.AIJudgeConfig.Enabled
		items = append(items, model.ProblemListItem{
			ID:            p.ID,
			Title:         p.Title,
			Difficulty:    p.Difficulty,
			Tags:          p.Tags,
			SubmitCount:   p.SubmitCount,
			AcceptedCount: p.AcceptedCount,
			HasAIJudge:    hasAI,
			HasFileIO:     p.FileIOEnabled,
		})
	}

	return items, total, nil
}

// ListWithUser 获取题目列表（含通过标识）
func (s *ProblemService) ListWithUser(page, size int, difficulty, tag, keyword string, userID uint) ([]model.ProblemListItem, int64, error) {
	problems, total, err := s.repo.List(page, size, difficulty, tag, keyword)
	if err != nil {
		return nil, 0, err
	}

	problemIDs := make([]uint, 0, len(problems))
	for _, p := range problems {
		problemIDs = append(problemIDs, p.ID)
	}

	acceptedSet := map[uint]struct{}{}
	if userID > 0 {
		acceptedIDs, err := s.submissionRepo.GetAcceptedProblemIDs(userID, problemIDs)
		if err == nil {
			for _, id := range acceptedIDs {
				acceptedSet[id] = struct{}{}
			}
		}
	}

	var items []model.ProblemListItem
	for _, p := range problems {
		hasAI := p.AIJudgeConfig != nil && p.AIJudgeConfig.Enabled
		_, hasAccepted := acceptedSet[p.ID]
		items = append(items, model.ProblemListItem{
			ID:            p.ID,
			Title:         p.Title,
			Difficulty:    p.Difficulty,
			Tags:          p.Tags,
			SubmitCount:   p.SubmitCount,
			AcceptedCount: p.AcceptedCount,
			HasAIJudge:    hasAI,
			HasFileIO:     p.FileIOEnabled,
			HasAccepted:   hasAccepted,
		})
	}

	return items, total, nil
}

// GetTestcases 获取测试用例
func (s *ProblemService) GetTestcases(problemID uint) ([]model.Testcase, error) {
	return s.repo.GetTestcases(problemID)
}

// AddTestcase 添加测试用例
func (s *ProblemService) AddTestcase(problemID uint, inputReader, outputReader io.Reader, score int, isSample bool) error {
	// 确保题目存在
	_, err := s.repo.GetByID(problemID)
	if err != nil {
		return errors.New("题目不存在")
	}

	// 获取当前测试用例数量作为序号
	testcases, _ := s.repo.GetTestcases(problemID)
	orderNum := len(testcases) + 1

	// 创建测试数据目录
	problemDir := filepath.Join(config.GlobalConfig.Paths.Problems, fmt.Sprintf("%d", problemID))
	os.MkdirAll(problemDir, 0755)

	// 保存输入文件
	inputFile := filepath.Join(problemDir, fmt.Sprintf("%d.in", orderNum))
	inputData, err := io.ReadAll(inputReader)
	if err != nil {
		return errors.New("读取输入数据失败")
	}
	if err := os.WriteFile(inputFile, inputData, 0644); err != nil {
		return errors.New("保存输入文件失败")
	}

	// 保存输出文件
	outputFile := filepath.Join(problemDir, fmt.Sprintf("%d.out", orderNum))
	outputData, err := io.ReadAll(outputReader)
	if err != nil {
		return errors.New("读取输出数据失败")
	}
	if err := os.WriteFile(outputFile, outputData, 0644); err != nil {
		return errors.New("保存输出文件失败")
	}

	// 创建测试用例记录
	testcase := &model.Testcase{
		ProblemID:  problemID,
		InputFile:  inputFile,
		OutputFile: outputFile,
		Score:      score,
		IsSample:   isSample,
		OrderNum:   orderNum,
	}

	return s.repo.CreateTestcase(testcase)
}

// DeleteTestcases 删除所有测试用例
func (s *ProblemService) DeleteTestcases(problemID uint) error {
	// 删除文件
	problemDir := filepath.Join(config.GlobalConfig.Paths.Problems, fmt.Sprintf("%d", problemID))
	files, _ := filepath.Glob(filepath.Join(problemDir, "*.in"))
	for _, f := range files {
		os.Remove(f)
	}
	files, _ = filepath.Glob(filepath.Join(problemDir, "*.out"))
	for _, f := range files {
		os.Remove(f)
	}

	return s.repo.DeleteTestcases(problemID)
}

func normalizeFileIO(req *model.ProblemCreateRequest) (bool, string, string, error) {
	if req == nil || !req.FileIOEnabled {
		return false, "", "", nil
	}

	inputName := strings.TrimSpace(req.FileInputName)
	outputName := strings.TrimSpace(req.FileOutputName)

	if inputName == "" || outputName == "" {
		return false, "", "", errors.New("文件操作已启用，请填写输入/输出文件名")
	}
	if inputName == outputName {
		return false, "", "", errors.New("输入/输出文件名不能相同")
	}
	if err := validateFileName(inputName, ".in"); err != nil {
		return false, "", "", err
	}
	if err := validateFileName(outputName, ".out"); err != nil {
		return false, "", "", err
	}

	return true, inputName, outputName, nil
}

func validateFileName(name string, suffix string) error {
	if strings.Contains(name, "..") || strings.ContainsAny(name, `/\`) {
		return errors.New("文件名不合法")
	}
	if !strings.HasSuffix(name, suffix) {
		return errors.New("文件名需以 " + suffix + " 结尾")
	}
	return nil
}
