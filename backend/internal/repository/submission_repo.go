package repository

import (
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
	if err := r.db.First(&submission, id).Error; err != nil {
		return nil, err
	}
	return &submission, nil
}

// Update 更新提交记录
func (r *SubmissionRepository) Update(submission *model.Submission) error {
	return r.db.Save(submission).Error
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
	rows, err := query.Select("submissions.*, problems.title as problem_title, users.username").
		Joins("LEFT JOIN problems ON submissions.problem_id = problems.id").
		Joins("LEFT JOIN users ON submissions.user_id = users.id").
		Offset(offset).Limit(size).Order("submissions.id DESC").Rows()
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var item model.SubmissionListItem
		var submission model.Submission
		var problemTitle, username string
		
		if err := rows.Scan(
			&submission.ID, &submission.ProblemID, &submission.UserID,
			&submission.Language, &submission.Code, &submission.Status,
			&submission.TimeUsed, &submission.MemoryUsed, &submission.Score,
			&submission.TestcaseResults, &submission.AIJudgeResult,
			&submission.CompileError, &submission.FinalMessage, &submission.CreatedAt,
			&problemTitle, &username,
		); err != nil {
			continue
		}
		
		item = model.SubmissionListItem{
			ID:           submission.ID,
			ProblemID:    submission.ProblemID,
			ProblemTitle: problemTitle,
			UserID:       submission.UserID,
			Username:     username,
			Language:     submission.Language,
			Status:       submission.Status,
			TimeUsed:     submission.TimeUsed,
			MemoryUsed:   submission.MemoryUsed,
			Score:        submission.Score,
			CreatedAt:    submission.CreatedAt,
		}
		items = append(items, item)
	}

	return items, total, nil
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
