package repository

import (
	"oj-system/internal/model"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository() *UserRepository {
	return &UserRepository{db: DB}
}

// Create 创建用户
func (r *UserRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

// GetByID 根据 ID 获取用户
func (r *UserRepository) GetByID(id uint) (*model.User, error) {
	var user model.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByUsername 根据用户名获取用户
func (r *UserRepository) GetByUsername(username string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail 根据邮箱获取用户
func (r *UserRepository) GetByEmail(email string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// Update 更新用户
func (r *UserRepository) Update(user *model.User) error {
	return r.db.Save(user).Error
}

// List 获取用户列表
func (r *UserRepository) List(page, size int) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	r.db.Model(&model.User{}).Count(&total)

	offset := (page - 1) * size
	if err := r.db.Offset(offset).Limit(size).Order("solved_count DESC").Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// IncrementSolvedCount 增加用户解题数
func (r *UserRepository) IncrementSolvedCount(userID uint) error {
	return r.db.Model(&model.User{}).Where("id = ?", userID).
		UpdateColumn("solved_count", gorm.Expr("solved_count + 1")).Error
}

// IncrementSubmitCount 增加用户提交数
func (r *UserRepository) IncrementSubmitCount(userID uint) error {
	return r.db.Model(&model.User{}).Where("id = ?", userID).
		UpdateColumn("submit_count", gorm.Expr("submit_count + 1")).Error
}

// ExistsByUsername 检查用户名是否存在
func (r *UserRepository) ExistsByUsername(username string) bool {
	var count int64
	r.db.Model(&model.User{}).Where("username = ?", username).Count(&count)
	return count > 0
}

// ExistsByEmail 检查邮箱是否存在
func (r *UserRepository) ExistsByEmail(email string) bool {
	var count int64
	r.db.Model(&model.User{}).Where("email = ?", email).Count(&count)
	return count > 0
}
