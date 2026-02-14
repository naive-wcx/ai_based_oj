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
	contestRepo       *repository.ContestRepository
	problemRepo       *repository.ProblemRepository
	userRepo          *repository.UserRepository
	submissionRepo    *repository.SubmissionRepository
	participationRepo *repository.ContestParticipationRepository
}

func NewContestService() *ContestService {
	return &ContestService{
		contestRepo:       repository.NewContestRepository(),
		problemRepo:       repository.NewProblemRepository(),
		userRepo:          repository.NewUserRepository(),
		submissionRepo:    repository.NewSubmissionRepository(),
		participationRepo: repository.NewContestParticipationRepository(),
	}
}

func (s *ContestService) Create(req *model.ContestCreateRequest, createdBy uint) (*model.Contest, error) {
	timingMode := normalizeContestTimingMode(req.TimingMode)
	durationMinutes := req.DurationMinutes
	submissionLimit := contestSubmissionLimitMax
	if err := validateContestRequest(req.Title, req.Type, req.StartAt, req.EndAt, timingMode, durationMinutes); err != nil {
		return nil, err
	}
	if timingMode != contestTimingWindow {
		durationMinutes = 0
	}

	problemIDs := uniqueUintList(req.ProblemIDs)
	if err := s.validateProblemIDs(problemIDs); err != nil {
		return nil, err
	}

	allowedUsers := uniqueUintList(req.AllowedUsers)
	allowedGroups := uniqueStringList(req.AllowedGroups)
	if len(allowedUsers) == 0 && len(allowedGroups) == 0 {
		return nil, errors.New("请至少选择一个参赛用户或分组")
	}

	contest := &model.Contest{
		Title:           strings.TrimSpace(req.Title),
		Description:     strings.TrimSpace(req.Description),
		Type:            strings.ToLower(strings.TrimSpace(req.Type)),
		TimingMode:      timingMode,
		DurationMinutes: durationMinutes,
		SubmissionLimit: submissionLimit,
		StartAt:         req.StartAt,
		EndAt:           req.EndAt,
		ProblemIDs:      model.UintList(problemIDs),
		AllowedUsers:    model.UintList(allowedUsers),
		AllowedGroups:   model.StringList(allowedGroups),
		CreatedBy:       createdBy,
	}

	if err := s.contestRepo.Create(contest); err != nil {
		return nil, errors.New("创建比赛失败")
	}

	return contest, nil
}

func (s *ContestService) Update(id uint, req *model.ContestUpdateRequest) (*model.Contest, error) {
	timingMode := normalizeContestTimingMode(req.TimingMode)
	durationMinutes := req.DurationMinutes
	submissionLimit := contestSubmissionLimitMax
	if err := validateContestRequest(req.Title, req.Type, req.StartAt, req.EndAt, timingMode, durationMinutes); err != nil {
		return nil, err
	}
	if timingMode != contestTimingWindow {
		durationMinutes = 0
	}

	contest, err := s.contestRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("比赛不存在")
	}

	problemIDs := uniqueUintList(req.ProblemIDs)
	if err := s.validateProblemIDs(problemIDs); err != nil {
		return nil, err
	}

	allowedUsers := uniqueUintList(req.AllowedUsers)
	allowedGroups := uniqueStringList(req.AllowedGroups)
	if len(allowedUsers) == 0 && len(allowedGroups) == 0 {
		return nil, errors.New("请至少选择一个参赛用户或分组")
	}

	contest.Title = strings.TrimSpace(req.Title)
	contest.Description = strings.TrimSpace(req.Description)
	contest.Type = strings.ToLower(strings.TrimSpace(req.Type))
	contest.TimingMode = timingMode
	contest.DurationMinutes = durationMinutes
	contest.SubmissionLimit = submissionLimit
	contest.StartAt = req.StartAt
	contest.EndAt = req.EndAt
	contest.ProblemIDs = model.UintList(problemIDs)
	contest.AllowedUsers = model.UintList(allowedUsers)
	contest.AllowedGroups = model.StringList(allowedGroups)

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
	contest.SubmissionLimit = normalizeContestSubmissionLimit(contest.SubmissionLimit)

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

func (s *ContestService) StartWindowContest(contestID uint, userID uint, isAdmin bool) (*model.ContestParticipation, *model.Contest, error) {
	contest, err := s.contestRepo.GetByID(contestID)
	if err != nil {
		return nil, nil, errors.New("比赛不存在")
	}
	contest.SubmissionLimit = normalizeContestSubmissionLimit(contest.SubmissionLimit)
	if normalizeContestTimingMode(contest.TimingMode) != contestTimingWindow {
		return nil, nil, errors.New("当前比赛不是窗口期模式")
	}
	now := time.Now()
	if now.Before(contest.StartAt) {
		return nil, nil, errors.New("比赛尚未开始")
	}
	if !now.Before(contest.EndAt) {
		return nil, nil, errors.New("比赛窗口期已结束")
	}
	if !isAdmin {
		user, err := s.userRepo.GetByID(userID)
		if err != nil {
			return nil, nil, errors.New("用户不存在")
		}
		if !canAccessContest(contest, userID, user.Group) {
			return nil, nil, errors.New("无权限访问该比赛")
		}
	}

	if existing, err := s.participationRepo.GetByContestAndUser(contestID, userID); err == nil {
		return existing, contest, nil
	}

	duration := time.Duration(contest.DurationMinutes) * time.Minute
	endAt := now.Add(duration)
	if endAt.After(contest.EndAt) {
		endAt = contest.EndAt
	}
	participation := &model.ContestParticipation{
		ContestID: contestID,
		UserID:    userID,
		StartAt:   now,
		EndAt:     endAt,
	}

	created, err := s.participationRepo.GetOrCreate(participation)
	if err != nil {
		return nil, nil, errors.New("创建比赛会话失败")
	}
	return created, contest, nil
}

func (s *ContestService) ResetWindowContestStart(contestID uint, userID uint) (bool, error) {
	contest, err := s.contestRepo.GetByID(contestID)
	if err != nil {
		return false, errors.New("比赛不存在")
	}
	contest.SubmissionLimit = normalizeContestSubmissionLimit(contest.SubmissionLimit)
	if normalizeContestTimingMode(contest.TimingMode) != contestTimingWindow {
		return false, errors.New("当前比赛不是窗口期模式")
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return false, errors.New("用户不存在")
	}
	if !canAccessContest(contest, userID, user.Group) {
		return false, errors.New("该用户不在比赛参赛范围内")
	}

	deleted, err := s.participationRepo.DeleteByContestAndUser(contestID, userID)
	if err != nil {
		return false, errors.New("重置比赛会话失败")
	}
	return deleted, nil
}

func (s *ContestService) ForceFinishContest(contestID uint, userID uint) (bool, time.Time, error) {
	contest, err := s.contestRepo.GetByID(contestID)
	if err != nil {
		return false, time.Time{}, errors.New("比赛不存在")
	}
	contest.SubmissionLimit = normalizeContestSubmissionLimit(contest.SubmissionLimit)

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return false, time.Time{}, errors.New("用户不存在")
	}
	if !canAccessContest(contest, userID, user.Group) {
		return false, time.Time{}, errors.New("该用户不在比赛参赛范围内")
	}

	now := time.Now()
	if now.Before(contest.StartAt) {
		return false, time.Time{}, errors.New("比赛尚未开始")
	}
	if !now.Before(contest.EndAt) {
		return false, time.Time{}, errors.New("比赛已结束，无需终止")
	}

	mode := normalizeContestTimingMode(contest.TimingMode)
	if mode == contestTimingWindow {
		participation, err := s.participationRepo.GetByContestAndUser(contestID, userID)
		if err != nil {
			return false, time.Time{}, errors.New("用户尚未开始比赛")
		}
		if !participation.EndAt.After(now) {
			return false, participation.EndAt, nil
		}
		participation.EndAt = now
		if err := s.participationRepo.Save(participation); err != nil {
			return false, time.Time{}, errors.New("终止比赛失败")
		}
		return true, participation.EndAt, nil
	}

	updated, participation, err := s.participationRepo.SetSessionEndAt(contestID, userID, contest.StartAt, now)
	if err != nil {
		return false, time.Time{}, errors.New("终止比赛失败")
	}
	return updated, participation.EndAt, nil
}

func (s *ContestService) GetSessionState(contest *model.Contest, userID uint, now time.Time) (*model.ContestSessionState, error) {
	state := &model.ContestSessionState{}
	if contest == nil || userID == 0 {
		return state, nil
	}
	contest.SubmissionLimit = normalizeContestSubmissionLimit(contest.SubmissionLimit)

	mode := normalizeContestTimingMode(contest.TimingMode)
	if mode != contestTimingWindow {
		start := contest.StartAt
		end := contest.EndAt
		if participation, err := s.participationRepo.GetByContestAndUser(contest.ID, userID); err == nil && participation != nil {
			if participation.StartAt.After(start) {
				start = participation.StartAt
			}
			if participation.EndAt.Before(end) {
				end = participation.EndAt
			}
		}
		state.Started = !now.Before(start)
		state.InLive = state.Started && now.Before(end)
		state.CanStart = false
		if state.Started {
			state.StartAt = &start
			state.EndAt = &end
			if state.InLive {
				state.RemainingSeconds = int64(end.Sub(now).Seconds())
			}
		}
		return state, nil
	}

	if now.Before(contest.StartAt) {
		return state, nil
	}

	participation, err := s.participationRepo.GetByContestAndUser(contest.ID, userID)
	if err != nil {
		state.CanStart = now.Before(contest.EndAt)
		return state, nil
	}

	state.Started = true
	state.CanStart = false
	state.InLive = now.Before(participation.EndAt)
	start := participation.StartAt
	end := participation.EndAt
	state.StartAt = &start
	state.EndAt = &end
	if state.InLive {
		state.RemainingSeconds = int64(participation.EndAt.Sub(now).Seconds())
	}

	return state, nil
}

// CountUserLiveSubmissions 统计用户在该比赛赛时阶段的提交次数。
func (s *ContestService) CountUserLiveSubmissions(contest *model.Contest, userID uint, now time.Time) (int, error) {
	if contest == nil || userID == 0 {
		return 0, nil
	}

	participation := model.ContestParticipation{}
	if p, err := s.participationRepo.GetByContestAndUser(contest.ID, userID); err == nil && p != nil {
		participation = *p
	}

	times, err := s.submissionRepo.ListUserSubmissionTimesInRange(userID, []uint(contest.ProblemIDs), contest.StartAt, now)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, submittedAt := range times {
		if classifySubmissionPhase(contest, participation, submittedAt) == leaderboardPhaseLive {
			count++
		}
	}
	return count, nil
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

func (s *ContestService) GetLeaderboard(contestID uint, boardMode string) (*model.Contest, []uint, []model.ContestLeaderboardEntry, string, error) {
	contest, err := s.contestRepo.GetByID(contestID)
	if err != nil {
		return nil, nil, nil, "", errors.New("比赛不存在")
	}
	contest.SubmissionLimit = normalizeContestSubmissionLimit(contest.SubmissionLimit)
	boardMode = normalizeLeaderboardMode(boardMode)

	problemIDs := []uint(contest.ProblemIDs)
	submissions, err := s.submissionRepo.ListForContestSince(problemIDs, contest.StartAt)
	if err != nil {
		return nil, nil, nil, "", errors.New("获取提交记录失败")
	}
	now := time.Now()
	isWindowMode := normalizeContestTimingMode(contest.TimingMode) == contestTimingWindow

	participationMap := map[uint]model.ContestParticipation{}
	participationUserIDs := make([]uint, 0)
	participations, err := s.participationRepo.ListByContest(contestID)
	if err != nil {
		return nil, nil, nil, "", errors.New("获取比赛会话失败")
	}
	for _, participation := range participations {
		participationMap[participation.UserID] = participation
		participationUserIDs = append(participationUserIDs, participation.UserID)
	}

	type userEntry struct {
		userID        uint
		username      string
		group         string
		liveScores    map[uint]int
		postScores    map[uint]int
		participation *model.ContestParticipation
	}

	userMap := make(map[uint]*userEntry)
	if len(participationUserIDs) > 0 {
		users, err := s.userRepo.GetByIDs(uniqueUintList(participationUserIDs))
		if err != nil {
			return nil, nil, nil, "", errors.New("获取参赛用户失败")
		}
		for _, user := range users {
			participation := participationMap[user.ID]
			participationCopy := participation
			userMap[user.ID] = &userEntry{
				userID:        user.ID,
				username:      user.Username,
				group:         user.Group,
				liveScores:    make(map[uint]int),
				postScores:    make(map[uint]int),
				participation: &participationCopy,
			}
		}
	}

	for _, sub := range submissions {
		if sub.CreatedAt.After(now) {
			continue
		}
		if !canAccessContest(contest, sub.UserID, sub.Group) {
			continue
		}

		entry, ok := userMap[sub.UserID]
		if !ok {
			var participation *model.ContestParticipation
			if p, exists := participationMap[sub.UserID]; exists {
				pCopy := p
				participation = &pCopy
			}
			entry = &userEntry{
				userID:        sub.UserID,
				username:      sub.Username,
				group:         sub.Group,
				liveScores:    make(map[uint]int),
				postScores:    make(map[uint]int),
				participation: participation,
			}
			userMap[sub.UserID] = entry
		} else {
			if entry.username == "" {
				entry.username = sub.Username
			}
			if entry.group == "" {
				entry.group = sub.Group
			}
		}

		phase := classifySubmissionPhase(contest, participationMap[sub.UserID], sub.CreatedAt)
		if phase == leaderboardPhaseLive {
			entry.liveScores[sub.ProblemID] = sub.Score
		} else if phase == leaderboardPhasePost {
			entry.postScores[sub.ProblemID] = sub.Score
		}
	}

	entries := make([]model.ContestLeaderboardEntry, 0, len(userMap))
	for _, entry := range userMap {
		liveScores := make([]int, 0, len(problemIDs))
		postScores := make([]int, 0, len(problemIDs))
		displayScores := make([]int, 0, len(problemIDs))
		liveTotal := 0
		postTotal := 0
		displayTotal := 0
		for _, pid := range problemIDs {
			liveScore := entry.liveScores[pid]
			postScore, hasPost := entry.postScores[pid]
			correctedPostScore := liveScore
			if hasPost {
				correctedPostScore = postScore
			}
			liveScores = append(liveScores, liveScore)
			postScores = append(postScores, correctedPostScore)
			liveTotal += liveScore
			postTotal += correctedPostScore
			switch boardMode {
			case leaderboardModePost:
				displayScores = append(displayScores, correctedPostScore)
				displayTotal += correctedPostScore
			case leaderboardModeCombined:
				displayScores = append(displayScores, correctedPostScore)
				displayTotal += correctedPostScore
			default:
				displayScores = append(displayScores, liveScore)
				displayTotal += liveScore
			}
		}

		var startedAt *time.Time
		elapsedSeconds := int64(0)
		if entry.participation != nil && isWindowMode {
			start := entry.participation.StartAt
			startedAt = &start
			elapsedEnd := now
			if elapsedEnd.After(entry.participation.EndAt) {
				elapsedEnd = entry.participation.EndAt
			}
			if elapsedEnd.After(entry.participation.StartAt) {
				elapsedSeconds = int64(elapsedEnd.Sub(entry.participation.StartAt).Seconds())
			}
		}

		entries = append(entries, model.ContestLeaderboardEntry{
			UserID:         entry.userID,
			Username:       entry.username,
			Group:          entry.group,
			Total:          displayTotal,
			Scores:         displayScores,
			LiveTotal:      liveTotal,
			PostTotal:      postTotal,
			LiveScores:     liveScores,
			PostScores:     postScores,
			StartedAt:      startedAt,
			ElapsedSeconds: elapsedSeconds,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		if boardMode == leaderboardModeCombined {
			if entries[i].LiveTotal == entries[j].LiveTotal {
				if entries[i].PostTotal == entries[j].PostTotal {
					return entries[i].UserID < entries[j].UserID
				}
				return entries[i].PostTotal > entries[j].PostTotal
			}
			return entries[i].LiveTotal > entries[j].LiveTotal
		}
		if entries[i].Total == entries[j].Total {
			return entries[i].UserID < entries[j].UserID
		}
		return entries[i].Total > entries[j].Total
	})

	return contest, problemIDs, entries, boardMode, nil
}

// RefreshStats 刷新比赛相关的统计数据
func (s *ContestService) RefreshStats(contestID uint) error {
	// 1. 获取比赛信息
	contest, err := s.contestRepo.GetByID(contestID)
	if err != nil {
		return errors.New("比赛不存在")
	}

	// 2. 刷新题目统计
	problemIDs := []uint(contest.ProblemIDs)
	for _, pid := range problemIDs {
		if err := s.problemRepo.SyncStats(pid); err != nil {
			// 记录错误但继续
		}
	}

	// 3. 刷新用户统计
	// 获取该比赛期间的所有提交记录，提取用户 ID
	submissions, err := s.submissionRepo.ListForContest(problemIDs, contest.StartAt, contest.EndAt)
	if err != nil {
		return errors.New("获取提交记录失败")
	}

	userSet := make(map[uint]struct{})
	for _, sub := range submissions {
		userSet[sub.UserID] = struct{}{}
	}

	for uid := range userSet {
		if err := s.userRepo.SyncStats(uid); err != nil {
			// 记录错误但继续
		}
	}

	return nil
}

// SyncEndedContests 同步已结束比赛的统计数据（供定时任务调用）
func (s *ContestService) SyncEndedContests() {
	contests, err := s.contestRepo.GetPendingSyncContests()
	if err != nil {
		// 日志记录错误? 目前没有统一的日志服务，暂时忽略
		return
	}

	for _, contest := range contests {
		// 复用 RefreshStats 逻辑
		if err := s.RefreshStats(contest.ID); err == nil {
			// 只有同步成功才标记
			s.contestRepo.MarkStatsSynced(contest.ID)
		}
	}
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

const (
	contestTimingFixed        = "fixed"
	contestTimingWindow       = "window"
	contestSubmissionLimitMax = 99

	leaderboardModeLive     = "live"
	leaderboardModePost     = "post"
	leaderboardModeCombined = "combined"

	leaderboardPhaseLive   = "live"
	leaderboardPhasePost   = "post"
	leaderboardPhaseIgnore = "ignore"
)

func validateContestRequest(title, contestType string, startAt, endAt time.Time, timingMode string, durationMinutes int) error {
	if strings.TrimSpace(title) == "" {
		return errors.New("标题不能为空")
	}
	contestType = strings.ToLower(strings.TrimSpace(contestType))
	if contestType != "oi" && contestType != "ioi" {
		return errors.New("无效的赛制类型")
	}
	if timingMode != contestTimingFixed && timingMode != contestTimingWindow {
		return errors.New("无效的计时模式")
	}
	if timingMode == contestTimingWindow && durationMinutes <= 0 {
		return errors.New("窗口期模式下比赛时长必须大于 0 分钟")
	}
	if endAt.Before(startAt) || endAt.Equal(startAt) {
		return errors.New("结束时间必须晚于开始时间")
	}
	return nil
}

func normalizeContestTimingMode(mode string) string {
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode == contestTimingWindow {
		return contestTimingWindow
	}
	return contestTimingFixed
}

func normalizeContestSubmissionLimit(limit int) int {
	return contestSubmissionLimitMax
}

func normalizeLeaderboardMode(mode string) string {
	mode = strings.ToLower(strings.TrimSpace(mode))
	switch mode {
	case leaderboardModeLive, leaderboardModePost, leaderboardModeCombined:
		return mode
	default:
		return leaderboardModeCombined
	}
}

func canAccessContest(contest *model.Contest, userID uint, group string) bool {
	if contest == nil {
		return false
	}
	allowedUsers := []uint(contest.AllowedUsers)
	allowedGroups := []string(contest.AllowedGroups)

	if len(allowedUsers) == 0 && len(allowedGroups) == 0 {
		return false
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
			ID:              contest.ID,
			Title:           contest.Title,
			Type:            contest.Type,
			TimingMode:      normalizeContestTimingMode(contest.TimingMode),
			DurationMinutes: contest.DurationMinutes,
			SubmissionLimit: normalizeContestSubmissionLimit(contest.SubmissionLimit),
			StartAt:         contest.StartAt,
			EndAt:           contest.EndAt,
			ProblemCount:    len(contest.ProblemIDs),
		})
	}
	return items
}

func classifySubmissionPhase(contest *model.Contest, participation model.ContestParticipation, submittedAt time.Time) string {
	if contest == nil {
		return leaderboardPhaseIgnore
	}
	mode := normalizeContestTimingMode(contest.TimingMode)
	if mode == contestTimingWindow {
		if participation.ID == 0 {
			if submittedAt.After(contest.EndAt) {
				return leaderboardPhasePost
			}
			return leaderboardPhaseIgnore
		}
		if submittedAt.Before(participation.StartAt) {
			return leaderboardPhaseIgnore
		}
		if !submittedAt.After(participation.EndAt) {
			return leaderboardPhaseLive
		}
		return leaderboardPhasePost
	}

	liveStart := contest.StartAt
	liveEnd := contest.EndAt
	if participation.ID != 0 {
		if participation.StartAt.After(liveStart) {
			liveStart = participation.StartAt
		}
		if participation.EndAt.Before(liveEnd) {
			liveEnd = participation.EndAt
		}
	}

	if submittedAt.Before(liveStart) {
		return leaderboardPhaseIgnore
	}
	if !submittedAt.After(liveEnd) {
		return leaderboardPhaseLive
	}
	return leaderboardPhasePost
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
