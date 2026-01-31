package service

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"oj-system/internal/config"
	"oj-system/internal/model"
	"oj-system/internal/repository"
)

var ErrSubmissionForbidden = errors.New("无权限")

type SubmissionService struct {
	repo        *repository.SubmissionRepository
	problemRepo *repository.ProblemRepository
	userRepo    *repository.UserRepository
}

func NewSubmissionService() *SubmissionService {
	return &SubmissionService{
		repo:        repository.NewSubmissionRepository(),
		problemRepo: repository.NewProblemRepository(),
		userRepo:    repository.NewUserRepository(),
	}
}

// Submit 提交代码
func (s *SubmissionService) Submit(req *model.SubmissionCreateRequest, userID uint) (*model.Submission, error) {
	// 检查题目是否存在
	problem, err := s.problemRepo.GetByID(req.ProblemID)
	if err != nil {
		return nil, errors.New("题目不存在")
	}

	// 验证语言
	if !isValidLanguage(req.Language) {
		return nil, errors.New("不支持的编程语言")
	}

	// 检查 AI 判题的语言要求
	if problem.AIJudgeConfig != nil && problem.AIJudgeConfig.Enabled {
		if problem.AIJudgeConfig.RequiredLanguage != "" {
			// 只是记录，不在提交时拒绝
		}
	}

	// 创建提交记录
	submission := &model.Submission{
		ProblemID: req.ProblemID,
		UserID:    userID,
		Language:  req.Language,
		Code:      req.Code,
		Status:    model.StatusPending,
	}

	if err := s.repo.Create(submission); err != nil {
		return nil, errors.New("创建提交失败")
	}

	// 保存代码文件
	s.saveCodeFile(submission)

	// 增加提交计数
	s.problemRepo.IncrementSubmitCount(req.ProblemID)
	s.userRepo.IncrementSubmitCount(userID)

	return submission, nil
}

// saveCodeFile 保存代码到文件
func (s *SubmissionService) saveCodeFile(submission *model.Submission) error {
	dir := filepath.Join(config.GlobalConfig.Paths.Submissions, fmt.Sprintf("%d", submission.ID))
	os.MkdirAll(dir, 0755)

	ext := getLanguageExtension(submission.Language)
	filePath := filepath.Join(dir, fmt.Sprintf("main%s", ext))

	return os.WriteFile(filePath, []byte(submission.Code), 0644)
}

// GetByID 获取提交详情
func (s *SubmissionService) GetByID(id uint, userID uint, isAdmin bool) (*model.Submission, error) {
	submission, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("提交不存在")
	}

	// 非管理员只能查看自己的提交
	if !isAdmin && submission.UserID != userID {
		return nil, ErrSubmissionForbidden
	}

	return submission, nil
}

// List 获取提交列表
func (s *SubmissionService) List(page, size int, problemID, userID uint, status string) (*model.PageData, error) {
	items, total, err := s.repo.List(page, size, problemID, userID, status)
	if err != nil {
		return nil, err
	}

	return &model.PageData{
		Total: total,
		Page:  page,
		Size:  size,
		List:  items,
	}, nil
}

// UpdateResult 更新判题结果
func (s *SubmissionService) UpdateResult(submission *model.Submission) error {
	// 如果是第一次 AC，增加用户解题数和题目通过数
	if submission.Status == model.StatusAccepted {
		if !s.repo.HasAccepted(submission.UserID, submission.ProblemID) {
			s.userRepo.IncrementSolvedCount(submission.UserID)
		}
		s.problemRepo.IncrementAcceptedCount(submission.ProblemID)
	}

	return s.repo.Update(submission)
}

// GetPendingSubmissions 获取待判题的提交
func (s *SubmissionService) GetPendingSubmissions(limit int) ([]model.Submission, error) {
	return s.repo.GetPendingSubmissions(limit)
}

// isValidLanguage 验证语言是否支持
func isValidLanguage(lang string) bool {
	supported := map[string]bool{
		"c":      true,
		"cpp":    true,
		"python": true,
		"java":   true,
		"go":     true,
	}
	return supported[lang]
}

// getLanguageExtension 获取语言对应的文件扩展名
func getLanguageExtension(lang string) string {
	extensions := map[string]string{
		"c":      ".c",
		"cpp":    ".cpp",
		"python": ".py",
		"java":   ".java",
		"go":     ".go",
	}
	if ext, ok := extensions[lang]; ok {
		return ext
	}
	return ".txt"
}
