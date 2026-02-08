package model

import (
	"time"
)

// User 用户模型
type User struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Username     string    `json:"username" gorm:"uniqueIndex;size:50;not null"`
	Email        string    `json:"email" gorm:"size:100"`
	PasswordHash string    `json:"-" gorm:"size:255;not null"`
	StudentID    string    `json:"student_id" gorm:"size:50"`
	Role         string    `json:"role" gorm:"size:20;default:user"` // user, admin
	Group        string    `json:"group" gorm:"size:50"`
	SolvedCount  int       `json:"solved_count" gorm:"default:0"`   // AC 的题目数量
	AcceptedCount int      `json:"accepted_count" gorm:"default:0"` // AC 的提交总数
	SubmitCount  int       `json:"submit_count" gorm:"default:0"`   // 总提交次数
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// AdminCreateUserRequest 管理员创建用户请求
type AdminCreateUserRequest struct {
	Username  string `json:"username" binding:"required,min=3,max=20"`
	Email     string `json:"email"`
	Password  string `json:"password" binding:"required,min=6,max=20"`
	StudentID string `json:"student_id"`
	Role      string `json:"role"`
	Group     string `json:"group"`
}

// AdminCreateUsersRequest 批量创建用户请求
type AdminCreateUsersRequest struct {
	Users []AdminCreateUserRequest `json:"users" binding:"required"`
}

// AdminUpdateUserRequest 管理员更新用户请求
type AdminUpdateUserRequest struct {
	Username  *string `json:"username"`
	Email     *string `json:"email"`
	StudentID *string `json:"student_id"`
	Role      *string `json:"role"`
	Group     *string `json:"group"`
	Password  *string `json:"password"`
}

// UserLoginRequest 用户登录请求
type UserLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=20"`
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
	Group       string `json:"group"`
	SolvedCount int    `json:"solved_count"`
	AcceptedCount int  `json:"accepted_count"`
	SubmitCount int    `json:"submit_count"`
}

// ToUserInfo 将 User 转换为 UserInfo
func (u *User) ToUserInfo() *UserInfo {
	return &UserInfo{
		ID:            u.ID,
		Username:      u.Username,
		Email:         u.Email,
		StudentID:     u.StudentID,
		Role:          u.Role,
		Group:         u.Group,
		SolvedCount:   u.SolvedCount,
		AcceptedCount: u.AcceptedCount,
		SubmitCount:   u.SubmitCount,
	}
}
