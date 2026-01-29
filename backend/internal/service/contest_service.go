package service

import (
	"errors"
	"sort"
	"strings"
	"time"

	"oj-system/internal/model"
	"oj-system/internal/repository"
)

type ContestService struct {
	contestRepo *repository.ContestRepository
	problemRepo *repository.ProblemRepository
	userRepo    *repository.UserRepository
	submissionRepo *repository.SubmissionRepository
}

func NewContestService() *ContestService {
	return &ContestService{
		contestRepo: repository.NewContestRepository(),
		problemRepo: repository.NewProblemRepository(),
		userRepo:    repository.NewUserRepository(),
		submissionRepo: repository.NewSubmissionRepository(),
	}
}

func (s *ContestService) Create(req *model.ContestCreateRequest, createdBy uint) (*model.Contest, error) {
	if err := validateContestRequest(req.Title, req.Type, req.StartAt, req.EndAt); err != nil {
		return nil, err
	}

	problemIDs := uniqueUintList(req.ProblemIDs)
	if err := s.validateProblemIDs(problemIDs); err != nil {
		return nil, err
	}

	contest := &model.Contest{
		Title:         strings.TrimSpace(req.Title),
		Description:   strings.TrimSpace(req.Description),
		Type:          strings.ToLower(strings.TrimSpace(req.Type)),
		StartAt:       req.StartAt,
		EndAt:         req.EndAt,
		ProblemIDs:    model.UintList(problemIDs),
		AllowedUsers:  model.UintList(uniqueUintList(req.AllowedUsers)),
		AllowedGroups: model.StringList(uniqueStringList(req.AllowedGroups)),
		CreatedBy:     createdBy,
	}

	if err := s.contestRepo.Create(contest); err != nil {
		return nil, errors.New("创建比赛失败")
	}

	return contest, nil
}

func (s *ContestService) Update(id uint, req *model.ContestUpdateRequest) (*model.Contest, error) {
	if err := validateContestRequest(req.Title, req.Type, req.StartAt, req.EndAt); err != nil {
		return nil, err
	}

	contest, err := s.contestRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("比赛不存在")
	}

	problemIDs := uniqueUintList(req.ProblemIDs)
	if err := s.validateProblemIDs(problemIDs); err != nil {
		return nil, err
	}

	contest.Title = strings.TrimSpace(req.Title)
	contest.Description = strings.TrimSpace(req.Description)
	contest.Type = strings.ToLower(strings.TrimSpace(req.Type))
	contest.StartAt = req.StartAt
	contest.EndAt = req.EndAt
	contest.ProblemIDs = model.UintList(problemIDs)
	contest.AllowedUsers = model.UintList(uniqueUintList(req.AllowedUsers))
	contest.AllowedGroups = model.StringList(uniqueStringList(req.AllowedGroups))

	if err := s.contestRepo.Update(contest); err != nil {
		return nil, errors.New("更新比赛失败")
	}

	return contest, nil
}

func (s *ContestService) Delete(id uint) error {
	if err := s.contestRepo.Delete(id); err != nil {
		return errors.New("删除比赛失败")
	}
	return nil
}

func (s *ContestService) GetByIDForUser(id uint, userID uint, isAdmin bool) (*model.Contest, error) {
	contest, err := s.contestRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("比赛不存在")
	}

	if isAdmin {
		return contest, nil
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	if !canAccessContest(contest, userID, user.Group) {
		return nil, errors.New("无权限访问该比赛")
	}

	if time.Now().Before(contest.StartAt) {
		return nil, errors.New("比赛未开始")
	}

	return contest, nil
}

func (s *ContestService) ListForUser(page, size int, userID uint, isAdmin bool) ([]model.ContestListItem, int64, error) {
	if isAdmin {
		contests, total, err := s.contestRepo.List(page, size)
		if err != nil {
			return nil, 0, err
		}
		return buildContestListItems(contests), total, nil
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, 0, errors.New("用户不存在")
	}

	allContests, err := s.contestRepo.ListAll()
	if err != nil {
		return nil, 0, err
	}

	var filtered []model.Contest
	for _, contest := range allContests {
		if canAccessContest(&contest, userID, user.Group) {
			filtered = append(filtered, contest)
		}
	}

	total := int64(len(filtered))
	start := (page - 1) * size
	if start < 0 {
		start = 0
	}
	end := start + size
	if start > len(filtered) {
		return []model.ContestListItem{}, total, nil
	}
	if end > len(filtered) {
		end = len(filtered)
	}

	return buildContestListItems(filtered[start:end]), total, nil
}

func (s *ContestService) GetProblemsByIDs(ids []uint) ([]model.Problem, error) {
	return s.problemRepo.GetByIDs(ids)
}

func (s *ContestService) GetLeaderboard(contestID uint) (*model.Contest, []uint, []model.ContestLeaderboardEntry, error) {
	contest, err := s.contestRepo.GetByID(contestID)
	if err != nil {
		return nil, nil, nil, errors.New("比赛不存在")
	}

	problemIDs := []uint(contest.ProblemIDs)
	submissions, err := s.submissionRepo.ListForContest(problemIDs, contest.StartAt, contest.EndAt)
	if err != nil {
		return nil, nil, nil, errors.New("获取提交记录失败")
	}

	allowedAll := len(contest.AllowedUsers) == 0 && len(contest.AllowedGroups) == 0

	type userEntry struct {
		userID   uint
		username string
		group    string
		scores   map[uint]int
	}

	userMap := make(map[uint]*userEntry)
	for _, sub := range submissions {
		if !allowedAll && !canAccessContest(contest, sub.UserID, sub.Group) {
			continue
		}

		entry, ok := userMap[sub.UserID]
		if !ok {
			entry = &userEntry{
				userID:   sub.UserID,
				username: sub.Username,
				group:    sub.Group,
				scores:   make(map[uint]int),
			}
			userMap[sub.UserID] = entry
		}

		current := entry.scores[sub.ProblemID]
		if sub.Score > current {
			entry.scores[sub.ProblemID] = sub.Score
		}
	}

	entries := make([]model.ContestLeaderboardEntry, 0, len(userMap))
	for _, entry := range userMap {
		scores := make([]int, 0, len(problemIDs))
		total := 0
		for _, pid := range problemIDs {
			score := entry.scores[pid]
			scores = append(scores, score)
			total += score
		}
		entries = append(entries, model.ContestLeaderboardEntry{
			UserID:   entry.userID,
			Username: entry.username,
			Group:    entry.group,
			Total:    total,
			Scores:   scores,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Total == entries[j].Total {
			return entries[i].UserID < entries[j].UserID
		}
		return entries[i].Total > entries[j].Total
	})

	return contest, problemIDs, entries, nil
}

func (s *ContestService) validateProblemIDs(ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	problems, err := s.problemRepo.GetByIDs(ids)
	if err != nil {
		return errors.New("题目不存在")
	}
	if len(problems) != len(ids) {
		return errors.New("题目列表包含无效 ID")
	}
	return nil
}

func validateContestRequest(title, contestType string, startAt, endAt time.Time) error {
	if strings.TrimSpace(title) == "" {
		return errors.New("标题不能为空")
	}
	contestType = strings.ToLower(strings.TrimSpace(contestType))
	if contestType != "oi" && contestType != "ioi" {
		return errors.New("无效的赛制类型")
	}
	if endAt.Before(startAt) || endAt.Equal(startAt) {
		return errors.New("结束时间必须晚于开始时间")
	}
	return nil
}

func canAccessContest(contest *model.Contest, userID uint, group string) bool {
	if contest == nil {
		return false
	}
	allowedUsers := []uint(contest.AllowedUsers)
	allowedGroups := []string(contest.AllowedGroups)

	if len(allowedUsers) == 0 && len(allowedGroups) == 0 {
		return true
	}
	if containsUint(allowedUsers, userID) {
		return true
	}
	if group != "" && containsString(allowedGroups, group) {
		return true
	}
	return false
}

func buildContestListItems(contests []model.Contest) []model.ContestListItem {
	items := make([]model.ContestListItem, 0, len(contests))
	for _, contest := range contests {
		items = append(items, model.ContestListItem{
			ID:           contest.ID,
			Title:        contest.Title,
			Type:         contest.Type,
			StartAt:      contest.StartAt,
			EndAt:        contest.EndAt,
			ProblemCount: len(contest.ProblemIDs),
		})
	}
	return items
}

func uniqueUintList(list []uint) []uint {
	if len(list) == 0 {
		return []uint{}
	}
	seen := make(map[uint]struct{}, len(list))
	result := make([]uint, 0, len(list))
	for _, val := range list {
		if _, ok := seen[val]; ok {
			continue
		}
		seen[val] = struct{}{}
		result = append(result, val)
	}
	return result
}

func uniqueStringList(list []string) []string {
	if len(list) == 0 {
		return []string{}
	}
	seen := make(map[string]struct{}, len(list))
	result := make([]string, 0, len(list))
	for _, val := range list {
		trimmed := strings.TrimSpace(val)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}

func containsUint(list []uint, value uint) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}

func containsString(list []string, value string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}
