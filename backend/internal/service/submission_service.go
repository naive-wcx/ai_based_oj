package service

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"oj-system/internal/config"
	"oj-system/internal/model"
	"oj-system/internal/repository"
)

var ErrSubmissionForbidden = errors.New("无权限")

type SubmissionService struct {
	repo        *repository.SubmissionRepository
	problemRepo *repository.ProblemRepository
	userRepo    *repository.UserRepository
	contestRepo *repository.ContestRepository
	participationRepo *repository.ContestParticipationRepository
}

func NewSubmissionService() *SubmissionService {
	return &SubmissionService{
		repo:        repository.NewSubmissionRepository(),
		problemRepo: repository.NewProblemRepository(),
		userRepo:    repository.NewUserRepository(),
		contestRepo: repository.NewContestRepository(),
		participationRepo: repository.NewContestParticipationRepository(),
	}
}

// Submit 提交代码
func (s *SubmissionService) Submit(req *model.SubmissionCreateRequest, userID uint) (*model.Submission, error) {
	// 检查题目是否存在
	problem, err := s.problemRepo.GetByID(req.ProblemID)
	if err != nil {
		return nil, errors.New("题目不存在")
	}
	if (problem.IsPublic == nil || !*problem.IsPublic) {
		if ok, err := s.canAccessHiddenProblem(problem.ID, userID); err != nil {
			return nil, errors.New("校验题目权限失败")
		} else if !ok {
			return nil, errors.New("题目未开放")
		}
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

	// 增加提交计数（仅当不在活动比赛中时）
	if !s.isProblemInActiveContest(req.ProblemID) {
		s.problemRepo.IncrementSubmitCount(req.ProblemID)
		s.userRepo.IncrementSubmitCount(userID)
	}

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

	if !isAdmin {
		s.maskSubmissionForOngoingOI(submission, userID)
	}

	return submission, nil
}

// List 获取提交列表
func (s *SubmissionService) List(page, size int, problemID, filterUserID uint, status string, viewerID uint, isAdmin bool) (*model.PageData, error) {
	items, total, err := s.repo.List(page, size, problemID, filterUserID, status)
	if err != nil {
		return nil, err
	}

	if !isAdmin {
		items = s.maskListForOngoingOI(items, viewerID)
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
		// 只有非比赛提交才立即更新全局统计
		if !s.isProblemInActiveContest(submission.ProblemID) {
			if !s.repo.HasAccepted(submission.UserID, submission.ProblemID) {
				s.userRepo.IncrementSolvedCount(submission.UserID)
			}
			s.userRepo.IncrementAcceptedCount(submission.UserID)
			s.problemRepo.IncrementAcceptedCount(submission.ProblemID)
		}
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

// isProblemInActiveContest 检查题目是否属于正在进行的比赛
func (s *SubmissionService) isProblemInActiveContest(problemID uint) bool {
	contests, err := s.contestRepo.ListAll()
	if err != nil {
		// 如果获取比赛列表失败，为了安全起见（避免泄题），假设不在比赛中？
		// 或者假设在比赛中？
		// 这里选择假设不在，因为这主要影响统计数据。
		return false
	}
	now := time.Now()
	for _, contest := range contests {
		if now.After(contest.StartAt) && now.Before(contest.EndAt) {
			if containsUint([]uint(contest.ProblemIDs), problemID) {
				return true
			}
		}
	}
	return false
}

func (s *SubmissionService) maskSubmissionForOngoingOI(submission *model.Submission, viewerID uint) {
	if submission == nil || viewerID == 0 {
		return
	}
	user, err := s.userRepo.GetByID(viewerID)
	if err != nil {
		return
	}

	contests, err := s.contestRepo.ListAll()
	if err != nil {
		return
	}

	now := time.Now()
	for _, contest := range contests {
		if strings.ToLower(contest.Type) != "oi" {
			continue
		}
		if now.Before(contest.StartAt) || !now.Before(contest.EndAt) {
			continue
		}
		if !canAccessContest(&contest, viewerID, user.Group) {
			continue
		}
		if !containsUint([]uint(contest.ProblemIDs), submission.ProblemID) {
			continue
		}
		if submission.CreatedAt.Before(contest.StartAt) || submission.CreatedAt.After(contest.EndAt) {
			continue
		}

		if submission.Status != model.StatusPending && submission.Status != model.StatusJudging {
			submission.Status = "Submitted"
		}
		submission.Score = 0
		submission.TimeUsed = 0
		submission.MemoryUsed = 0
		submission.TestcaseResults = nil
		submission.AIJudgeResult = nil
		submission.CompileError = ""
		submission.FinalMessage = ""
		return
	}
}

func (s *SubmissionService) maskListForOngoingOI(items []model.SubmissionListItem, viewerID uint) []model.SubmissionListItem {
	if viewerID == 0 || len(items) == 0 {
		return items
	}
	user, err := s.userRepo.GetByID(viewerID)
	if err != nil {
		return items
	}

	contests, err := s.contestRepo.ListAll()
	if err != nil {
		return items
	}

	now := time.Now()
	for i := range items {
		for _, contest := range contests {
			if strings.ToLower(contest.Type) != "oi" {
				continue
			}
			if now.Before(contest.StartAt) || !now.Before(contest.EndAt) {
				continue
			}
			if !canAccessContest(&contest, viewerID, user.Group) {
				continue
			}
			if !containsUint([]uint(contest.ProblemIDs), items[i].ProblemID) {
				continue
			}
			if items[i].CreatedAt.Before(contest.StartAt) || items[i].CreatedAt.After(contest.EndAt) {
				continue
			}

			if items[i].Status != model.StatusPending && items[i].Status != model.StatusJudging {
				items[i].Status = "Submitted"
			}
			items[i].TimeUsed = 0
			items[i].MemoryUsed = 0
			items[i].Score = 0
			break
		}
	}

	return items
}

func (s *SubmissionService) canAccessHiddenProblem(problemID uint, userID uint) (bool, error) {
	if userID == 0 {
		return false, nil
	}
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return false, err
	}
	if strings.ToLower(user.Role) == "admin" {
		return true, nil
	}

	contests, err := s.contestRepo.ListAll()
	if err != nil {
		return false, err
	}
	now := time.Now()
	for _, contest := range contests {
		if now.Before(contest.StartAt) {
			continue
		}
		if !containsUint([]uint(contest.ProblemIDs), problemID) {
			continue
		}
		if !canAccessContest(&contest, userID, user.Group) {
			continue
		}
		if normalizeContestTimingMode(contest.TimingMode) != contestTimingWindow {
			return true, nil
		}
		if !now.Before(contest.EndAt) {
			return true, nil
		}
		participation, err := s.participationRepo.GetByContestAndUser(contest.ID, userID)
		if err == nil && participation != nil && !now.Before(participation.StartAt) {
			return true, nil
		}
	}

	return false, nil
}
