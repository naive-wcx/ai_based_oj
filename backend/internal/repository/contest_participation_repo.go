package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"oj-system/internal/model"
)

type ContestParticipationRepository struct {
	db *gorm.DB
}

func NewContestParticipationRepository() *ContestParticipationRepository {
	return &ContestParticipationRepository{db: DB}
}

func (r *ContestParticipationRepository) GetByContestAndUser(contestID, userID uint) (*model.ContestParticipation, error) {
	var participation model.ContestParticipation
	err := r.db.Where("contest_id = ? AND user_id = ?", contestID, userID).First(&participation).Error
	if err != nil {
		return nil, err
	}
	return &participation, nil
}

func (r *ContestParticipationRepository) Create(participation *model.ContestParticipation) error {
	return r.db.Create(participation).Error
}

func (r *ContestParticipationRepository) GetOrCreate(participation *model.ContestParticipation) (*model.ContestParticipation, error) {
	if participation == nil {
		return nil, errors.New("invalid participation")
	}
	if err := r.db.Where("contest_id = ? AND user_id = ?", participation.ContestID, participation.UserID).
		FirstOrCreate(participation, participation).Error; err != nil {
		return nil, err
	}
	return participation, nil
}

func (r *ContestParticipationRepository) ListByContest(contestID uint) ([]model.ContestParticipation, error) {
	var participations []model.ContestParticipation
	err := r.db.Where("contest_id = ?", contestID).Find(&participations).Error
	if err != nil {
		return nil, err
	}
	return participations, nil
}

func (r *ContestParticipationRepository) DeleteByContestAndUser(contestID, userID uint) (bool, error) {
	result := r.db.Where("contest_id = ? AND user_id = ?", contestID, userID).Delete(&model.ContestParticipation{})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *ContestParticipationRepository) Save(participation *model.ContestParticipation) error {
	return r.db.Save(participation).Error
}

func (r *ContestParticipationRepository) SetSessionEndAt(contestID, userID uint, startAt, endAt time.Time) (bool, *model.ContestParticipation, error) {
	existing, err := r.GetByContestAndUser(contestID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			created := &model.ContestParticipation{
				ContestID: contestID,
				UserID:    userID,
				StartAt:   startAt,
				EndAt:     endAt,
			}
			if err := r.Create(created); err != nil {
				return false, nil, err
			}
			return true, created, nil
		}
		return false, nil, err
	}

	// 已经结束到更早/相同时间时，视为无需更新（幂等）
	if !existing.EndAt.After(endAt) {
		return false, existing, nil
	}

	if existing.StartAt.IsZero() || existing.StartAt.After(startAt) {
		existing.StartAt = startAt
	}
	existing.EndAt = endAt
	if err := r.Save(existing); err != nil {
		return false, nil, err
	}
	return true, existing, nil
}
