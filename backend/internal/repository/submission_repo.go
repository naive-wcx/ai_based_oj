package repository

import (
	"time"

	"oj-system/internal/model"

	"gorm.io/gorm"
)

type SubmissionRepository struct {
	db *gorm.DB
}

func NewSubmissionRepository() *SubmissionRepository {
	return &SubmissionRepository{db: DB}
}

// Create 创建提交记录
func (r *SubmissionRepository) Create(submission *model.Submission) error {
	return r.db.Create(submission).Error
}

// GetByID 根据 ID 获取提交记录
func (r *SubmissionRepository) GetByID(id uint) (*model.Submission, error) {
	var submission model.Submission
	var problemTitle, username string

	row := r.db.Table("submissions").
		Select(
			"submissions.id, submissions.problem_id, submissions.user_id, submissions.language, submissions.code, "+
				"submissions.status, submissions.time_used, submissions.memory_used, submissions.score, "+
				"submissions.testcase_results, submissions.ai_judge_result, submissions.compile_error, "+
				"submissions.final_message, submissions.created_at, problems.title as problem_title, users.username as username",
		).
		Joins("LEFT JOIN problems ON submissions.problem_id = problems.id").
		Joins("LEFT JOIN users ON submissions.user_id = users.id").
		Where("submissions.id = ?", id).
		Row()
	if err := row.Scan(
		&submission.ID, &submission.ProblemID, &submission.UserID,
		&submission.Language, &submission.Code, &submission.Status,
		&submission.TimeUsed, &submission.MemoryUsed, &submission.Score,
		&submission.TestcaseResults, &submission.AIJudgeResult,
		&submission.CompileError, &submission.FinalMessage, &submission.CreatedAt,
		&problemTitle, &username,
	); err != nil {
		return nil, err
	}

	submission.ProblemTitle = problemTitle
	submission.Username = username

	return &submission, nil
}

// Update 更新提交记录
func (r *SubmissionRepository) Update(submission *model.Submission) error {
	result := r.db.Model(&model.Submission{}).
		Where("id = ?", submission.ID).
		Updates(map[string]interface{}{
			"status":           submission.Status,
			"time_used":        submission.TimeUsed,
			"memory_used":      submission.MemoryUsed,
			"score":            submission.Score,
			"testcase_results": submission.TestcaseResults,
			"ai_judge_result":  submission.AIJudgeResult,
			"compile_error":    submission.CompileError,
			"final_message":    submission.FinalMessage,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// DeleteByID 删除提交记录
func (r *SubmissionRepository) DeleteByID(id uint) error {
	result := r.db.Delete(&model.Submission{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// List 获取提交列表
func (r *SubmissionRepository) List(page, size int, problemID, userID uint, status string) ([]model.SubmissionListItem, int64, error) {
	var total int64
	var items []model.SubmissionListItem

	query := r.db.Model(&model.Submission{})

	if problemID > 0 {
		query = query.Where("problem_id = ?", problemID)
	}
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)

	offset := (page - 1) * size
	rows, err := query.Select(
		"submissions.id, submissions.problem_id, submissions.user_id, submissions.language, submissions.status, " +
			"submissions.time_used, submissions.memory_used, submissions.score, submissions.created_at, " +
			"problems.title as problem_title, users.username as username",
	).
		Joins("LEFT JOIN problems ON submissions.problem_id = problems.id").
		Joins("LEFT JOIN users ON submissions.user_id = users.id").
		Offset(offset).Limit(size).Order("submissions.id DESC").Rows()
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var item model.SubmissionListItem
		if err := rows.Scan(
			&item.ID, &item.ProblemID, &item.UserID,
			&item.Language, &item.Status,
			&item.TimeUsed, &item.MemoryUsed, &item.Score,
			&item.CreatedAt, &item.ProblemTitle, &item.Username,
		); err != nil {
			continue
		}
		items = append(items, item)
	}

	return items, total, nil
}

func (r *SubmissionRepository) ListRejudgeCandidatesByProblem(problemID uint) ([]model.Submission, error) {
	var submissions []model.Submission
	err := r.db.Where("problem_id = ?", problemID).
		Where("status NOT IN ?", []string{model.StatusPending, model.StatusJudging}).
		Order("id ASC").
		Find(&submissions).Error
	if err != nil {
		return nil, err
	}
	return submissions, nil
}

func (r *SubmissionRepository) ResetForRejudge(ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.Model(&model.Submission{}).
		Where("id IN ?", ids).
		Updates(map[string]interface{}{
			"status":           model.StatusPending,
			"time_used":        0,
			"memory_used":      0,
			"score":            0,
			"testcase_results": model.TestcaseResultList{},
			"ai_judge_result":  nil,
			"compile_error":    "",
			"final_message":    "",
		}).Error
}

// GetPendingSubmissions 获取待判题的提交
func (r *SubmissionRepository) GetPendingSubmissions(limit int) ([]model.Submission, error) {
	var submissions []model.Submission
	if err := r.db.Where("status = ?", model.StatusPending).
		Order("id ASC").Limit(limit).Find(&submissions).Error; err != nil {
		return nil, err
	}
	return submissions, nil
}

// UpdateStatus 更新提交状态
func (r *SubmissionRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&model.Submission{}).Where("id = ?", id).
		Update("status", status).Error
}

// HasAccepted 检查用户是否已通过该题目
func (r *SubmissionRepository) HasAccepted(userID, problemID uint) bool {
	var count int64
	r.db.Model(&model.Submission{}).
		Where("user_id = ? AND problem_id = ? AND status = ?", userID, problemID, model.StatusAccepted).
		Count(&count)
	return count > 0
}

// GetUserSubmissionCount 获取用户在某题目的提交数
func (r *SubmissionRepository) GetUserSubmissionCount(userID, problemID uint) int64 {
	var count int64
	r.db.Model(&model.Submission{}).
		Where("user_id = ? AND problem_id = ?", userID, problemID).
		Count(&count)
	return count
}

// CountAll 获取提交总数
func (r *SubmissionRepository) CountAll() (int64, error) {
	var count int64
	if err := r.db.Model(&model.Submission{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// ListForContest 获取比赛期间相关提交
func (r *SubmissionRepository) ListForContest(problemIDs []uint, startAt, endAt time.Time) ([]model.ContestSubmission, error) {
	if len(problemIDs) == 0 {
		return []model.ContestSubmission{}, nil
	}
	startAt = startAt.In(time.Local)
	endAt = endAt.In(time.Local)

	var submissions []model.ContestSubmission
	err := r.db.Table("submissions").
		Select("submissions.user_id, submissions.problem_id, submissions.score, submissions.created_at, users.username, users.`group` as user_group").
		Joins("LEFT JOIN users ON submissions.user_id = users.id").
		Where("submissions.problem_id IN ?", problemIDs).
		Where("submissions.created_at >= ? AND submissions.created_at <= ?", startAt, endAt).
		Order("submissions.created_at ASC, submissions.id ASC").
		Scan(&submissions).Error

	if err != nil {
		return nil, err
	}

	return submissions, nil
}

// ListForContestSince 获取比赛起始时间后的相关提交
func (r *SubmissionRepository) ListForContestSince(problemIDs []uint, startAt time.Time) ([]model.ContestSubmission, error) {
	if len(problemIDs) == 0 {
		return []model.ContestSubmission{}, nil
	}
	startAt = startAt.In(time.Local)

	var submissions []model.ContestSubmission
	err := r.db.Table("submissions").
		Select("submissions.user_id, submissions.problem_id, submissions.score, submissions.created_at, users.username, users.`group` as user_group").
		Joins("LEFT JOIN users ON submissions.user_id = users.id").
		Where("submissions.problem_id IN ?", problemIDs).
		Where("submissions.created_at >= ?", startAt).
		Order("submissions.created_at ASC, submissions.id ASC").
		Scan(&submissions).Error
	if err != nil {
		return nil, err
	}

	return submissions, nil
}

// GetAcceptedProblemIDs 获取用户已通过的题目 ID 列表
func (r *SubmissionRepository) GetAcceptedProblemIDs(userID uint, problemIDs []uint) ([]uint, error) {
	if userID == 0 || len(problemIDs) == 0 {
		return []uint{}, nil
	}
	var ids []uint
	err := r.db.Model(&model.Submission{}).
		Where("user_id = ? AND status = ?", userID, model.StatusAccepted).
		Where("problem_id IN ?", problemIDs).
		Distinct().
		Pluck("problem_id", &ids).Error
	if err != nil {
		return nil, err
	}
	return ids, nil
}

// GetAcceptedProblemIDsInRange 获取用户在时间范围内通过的题目 ID 列表
func (r *SubmissionRepository) GetAcceptedProblemIDsInRange(userID uint, problemIDs []uint, startAt, endAt time.Time) ([]uint, error) {
	if userID == 0 || len(problemIDs) == 0 {
		return []uint{}, nil
	}
	startAt = startAt.In(time.Local)
	endAt = endAt.In(time.Local)
	var ids []uint
	err := r.db.Model(&model.Submission{}).
		Where("user_id = ? AND status = ?", userID, model.StatusAccepted).
		Where("problem_id IN ?", problemIDs).
		Where("created_at >= ? AND created_at <= ?", startAt, endAt).
		Distinct().
		Pluck("problem_id", &ids).Error
	if err != nil {
		return nil, err
	}
	return ids, nil
}

// GetSubmittedProblemIDsInRange 获取用户在时间范围内提交过的题目 ID 列表
func (r *SubmissionRepository) GetSubmittedProblemIDsInRange(userID uint, problemIDs []uint, startAt, endAt time.Time) ([]uint, error) {
	if userID == 0 || len(problemIDs) == 0 {
		return []uint{}, nil
	}
	startAt = startAt.In(time.Local)
	endAt = endAt.In(time.Local)
	var ids []uint
	err := r.db.Model(&model.Submission{}).
		Where("user_id = ?", userID).
		Where("problem_id IN ?", problemIDs).
		Where("created_at >= ? AND created_at <= ?", startAt, endAt).
		Distinct().
		Pluck("problem_id", &ids).Error
	if err != nil {
		return nil, err
	}
	return ids, nil
}

// GetUserLastScoresInRange 获取用户在时间范围内每题最后一次提交的分数
func (r *SubmissionRepository) GetUserLastScoresInRange(userID uint, problemIDs []uint, startAt, endAt time.Time) (map[uint]int, error) {
	if userID == 0 || len(problemIDs) == 0 {
		return map[uint]int{}, nil
	}
	startAt = startAt.In(time.Local)
	endAt = endAt.In(time.Local)

	type row struct {
		ProblemID uint      `gorm:"column:problem_id"`
		Score     int       `gorm:"column:score"`
		CreatedAt time.Time `gorm:"column:created_at"`
		ID        uint      `gorm:"column:id"`
	}

	var rows []row
	err := r.db.Model(&model.Submission{}).
		Select("problem_id, score, created_at, id").
		Where("user_id = ? AND problem_id IN ?", userID, problemIDs).
		Where("created_at >= ? AND created_at <= ?", startAt, endAt).
		Order("created_at ASC, id ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[uint]int, len(problemIDs))
	for _, r := range rows {
		result[r.ProblemID] = r.Score
	}
	return result, nil
}

// GetUserBestScoresInRange 获取用户在时间范围内每题最高分
func (r *SubmissionRepository) GetUserBestScoresInRange(userID uint, problemIDs []uint, startAt, endAt time.Time) (map[uint]int, error) {
	if userID == 0 || len(problemIDs) == 0 {
		return map[uint]int{}, nil
	}
	startAt = startAt.In(time.Local)
	endAt = endAt.In(time.Local)

	type row struct {
		ProblemID uint `gorm:"column:problem_id"`
		Score     int  `gorm:"column:score"`
	}

	var rows []row
	err := r.db.Model(&model.Submission{}).
		Select("problem_id, MAX(score) as score").
		Where("user_id = ? AND problem_id IN ?", userID, problemIDs).
		Where("created_at >= ? AND created_at <= ?", startAt, endAt).
		Group("problem_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[uint]int, len(rows))
	for _, r := range rows {
		result[r.ProblemID] = r.Score
	}
	return result, nil
}

// ListUserSubmissionTimesInRange 获取用户在题目集合中的提交时间列表
func (r *SubmissionRepository) ListUserSubmissionTimesInRange(userID uint, problemIDs []uint, startAt, endAt time.Time) ([]time.Time, error) {
	if userID == 0 || len(problemIDs) == 0 {
		return []time.Time{}, nil
	}
	startAt = startAt.In(time.Local)
	endAt = endAt.In(time.Local)

	type row struct {
		CreatedAt time.Time `gorm:"column:created_at"`
	}

	var rows []row
	err := r.db.Model(&model.Submission{}).
		Select("created_at").
		Where("user_id = ? AND problem_id IN ?", userID, problemIDs).
		Where("created_at >= ? AND created_at <= ?", startAt, endAt).
		Order("created_at ASC, id ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	times := make([]time.Time, 0, len(rows))
	for _, row := range rows {
		times = append(times, row.CreatedAt)
	}
	return times, nil
}
