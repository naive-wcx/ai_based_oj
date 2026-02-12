package model

import "time"

// ContestParticipation 选手在窗口期比赛中的个人开赛会话
type ContestParticipation struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	ContestID uint      `json:"contest_id" gorm:"not null;index;uniqueIndex:idx_contest_user"`
	UserID    uint      `json:"user_id" gorm:"not null;index;uniqueIndex:idx_contest_user"`
	StartAt   time.Time `json:"start_at"`
	EndAt     time.Time `json:"end_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
