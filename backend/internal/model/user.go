package model

import (
	"time"
)

// User 用户模型
type User struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Username     string    `json:"username" gorm:"uniqueIndex;size:50;not null"`
	Email        string    `json:"email" gorm:"uniqueIndex;size:100;not null"`
	PasswordHash string    `json:"-" gorm:"size:255;not null"`
	StudentID    string    `json:"student_id" gorm:"size:50"`
	Role         string    `json:"role" gorm:"size:20;default:user"` // user, admin
	SolvedCount  int       `json:"solved_count" gorm:"default:0"`
	SubmitCount  int       `json:"submit_count" gorm:"default:0"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UserRegisterRequest 用户注册请求
type UserRegisterRequest struct {
	Username  string `json:"username" binding:"required,min=3,max=20"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6,max=20"`
	StudentID string `json:"student_id"`
}

// UserLoginRequest 用户登录请求
type UserLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UserLoginResponse 用户登录响应
type UserLoginResponse struct {
	Token string    `json:"token"`
	User  *UserInfo `json:"user"`
}

// UserInfo 用户信息（不含敏感信息）
type UserInfo struct {
	ID          uint   `json:"id"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	StudentID   string `json:"student_id"`
	Role        string `json:"role"`
	SolvedCount int    `json:"solved_count"`
	SubmitCount int    `json:"submit_count"`
}

// ToUserInfo 将 User 转换为 UserInfo
func (u *User) ToUserInfo() *UserInfo {
	return &UserInfo{
		ID:          u.ID,
		Username:    u.Username,
		Email:       u.Email,
		StudentID:   u.StudentID,
		Role:        u.Role,
		SolvedCount: u.SolvedCount,
		SubmitCount: u.SubmitCount,
	}
}
