package repository

import (
	"time"

	"oj-system/internal/model"

	"gorm.io/gorm"
)

type ContestRepository struct {
	db *gorm.DB
}

func NewContestRepository() *ContestRepository {
	return &ContestRepository{db: DB}
}

func (r *ContestRepository) Create(contest *model.Contest) error {
	return r.db.Create(contest).Error
}

func (r *ContestRepository) Update(contest *model.Contest) error {
	return r.db.Save(contest).Error
}

func (r *ContestRepository) Delete(id uint) error {
	return r.db.Delete(&model.Contest{}, id).Error
}

func (r *ContestRepository) GetByID(id uint) (*model.Contest, error) {
	var contest model.Contest
	if err := r.db.First(&contest, id).Error; err != nil {
		return nil, err
	}
	return &contest, nil
}

func (r *ContestRepository) List(page, size int) ([]model.Contest, int64, error) {
	var contests []model.Contest
	var total int64

	r.db.Model(&model.Contest{}).Count(&total)

	offset := (page - 1) * size
	if err := r.db.Offset(offset).Limit(size).Order("start_at DESC").Find(&contests).Error; err != nil {
		return nil, 0, err
	}

	return contests, total, nil
}

func (r *ContestRepository) ListAll() ([]model.Contest, error) {
	var contests []model.Contest
	if err := r.db.Order("start_at DESC").Find(&contests).Error; err != nil {
		return nil, err
	}
	return contests, nil
}

// GetPendingSyncContests 获取已结束但未同步统计的比赛
func (r *ContestRepository) GetPendingSyncContests() ([]model.Contest, error) {
	var contests []model.Contest
	// 查找 EndAt < Now 且 IsStatsSynced = false 的比赛
	// 使用 Go 的 time.Now() 避免 SQLite 时区问题
	if err := r.db.Where("end_at < ? AND is_stats_synced = ?", time.Now(), false).Find(&contests).Error; err != nil {
		return nil, err
	}
	return contests, nil
}

// MarkStatsSynced 标记比赛统计已同步
func (r *ContestRepository) MarkStatsSynced(id uint) error {
	return r.db.Model(&model.Contest{}).Where("id = ?", id).Update("is_stats_synced", true).Error
}
