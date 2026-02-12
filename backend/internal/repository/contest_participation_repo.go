package repository

import (
	"errors"

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
