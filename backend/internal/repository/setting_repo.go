package repository

import (
	"oj-system/internal/model"

	"gorm.io/gorm"
)

type SettingRepository struct {
	db *gorm.DB
}

func NewSettingRepository() *SettingRepository {
	return &SettingRepository{db: DB}
}

// Get 获取设置值
func (r *SettingRepository) Get(key string) (string, error) {
	var setting model.Setting
	if err := r.db.Where("key = ?", key).First(&setting).Error; err != nil {
		return "", err
	}
	return setting.Value, nil
}

// Set 设置值
func (r *SettingRepository) Set(key, value string) error {
	var setting model.Setting
	result := r.db.Where("key = ?", key).First(&setting)
	
	if result.Error == gorm.ErrRecordNotFound {
		// 创建新记录
		setting = model.Setting{
			Key:   key,
			Value: value,
		}
		return r.db.Create(&setting).Error
	} else if result.Error != nil {
		return result.Error
	}
	
	// 更新现有记录
	setting.Value = value
	return r.db.Save(&setting).Error
}

// GetAll 获取所有设置
func (r *SettingRepository) GetAll() (map[string]string, error) {
	var settings []model.Setting
	if err := r.db.Find(&settings).Error; err != nil {
		return nil, err
	}
	
	result := make(map[string]string)
	for _, s := range settings {
		result[s.Key] = s.Value
	}
	return result, nil
}

// GetMultiple 获取多个设置
func (r *SettingRepository) GetMultiple(keys []string) (map[string]string, error) {
	var settings []model.Setting
	if err := r.db.Where("key IN ?", keys).Find(&settings).Error; err != nil {
		return nil, err
	}
	
	result := make(map[string]string)
	for _, s := range settings {
		result[s.Key] = s.Value
	}
	return result, nil
}
