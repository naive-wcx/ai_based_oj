package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// Contest 比赛模型
type Contest struct {
	ID              uint       `json:"id" gorm:"primaryKey"`
	Title           string     `json:"title" gorm:"size:200;not null"`
	Description     string     `json:"description" gorm:"type:text"`
	Type            string     `json:"type" gorm:"size:10;not null"`             // oi | ioi
	TimingMode      string     `json:"timing_mode" gorm:"size:20;default:fixed"` // fixed | window
	DurationMinutes int        `json:"duration_minutes"`                         // 仅 timing_mode=window 时生效
	SubmissionLimit int        `json:"submission_limit" gorm:"default:99"`       // 比赛提交总次数上限（固定 99）
	StartAt         time.Time  `json:"start_at"`
	EndAt           time.Time  `json:"end_at"`
	ProblemIDs      UintList   `json:"problem_ids" gorm:"type:text"`
	AllowedUsers    UintList   `json:"allowed_users" gorm:"type:text"`
	AllowedGroups   StringList `json:"allowed_groups" gorm:"type:text"`
	IsStatsSynced   bool       `json:"is_stats_synced" gorm:"default:false"`
	CreatedBy       uint       `json:"created_by"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// ContestCreateRequest 创建比赛请求
type ContestCreateRequest struct {
	Title           string    `json:"title" binding:"required,max=200"`
	Description     string    `json:"description"`
	Type            string    `json:"type" binding:"required"`
	TimingMode      string    `json:"timing_mode"`
	DurationMinutes int       `json:"duration_minutes"`
	StartAt         time.Time `json:"start_at" binding:"required"`
	EndAt           time.Time `json:"end_at" binding:"required"`
	ProblemIDs      []uint    `json:"problem_ids"`
	AllowedUsers    []uint    `json:"allowed_users"`
	AllowedGroups   []string  `json:"allowed_groups"`
}

// ContestUpdateRequest 更新比赛请求
type ContestUpdateRequest struct {
	Title           string    `json:"title" binding:"required,max=200"`
	Description     string    `json:"description"`
	Type            string    `json:"type" binding:"required"`
	TimingMode      string    `json:"timing_mode"`
	DurationMinutes int       `json:"duration_minutes"`
	StartAt         time.Time `json:"start_at" binding:"required"`
	EndAt           time.Time `json:"end_at" binding:"required"`
	ProblemIDs      []uint    `json:"problem_ids"`
	AllowedUsers    []uint    `json:"allowed_users"`
	AllowedGroups   []string  `json:"allowed_groups"`
}

// ContestListItem 比赛列表项
type ContestListItem struct {
	ID              uint      `json:"id"`
	Title           string    `json:"title"`
	Type            string    `json:"type"`
	TimingMode      string    `json:"timing_mode"`
	DurationMinutes int       `json:"duration_minutes"`
	SubmissionLimit int       `json:"submission_limit"`
	StartAt         time.Time `json:"start_at"`
	EndAt           time.Time `json:"end_at"`
	ProblemCount    int       `json:"problem_count"`
}

// ContestLeaderboardEntry 比赛排行榜项
type ContestLeaderboardEntry struct {
	UserID         uint       `json:"user_id"`
	Username       string     `json:"username"`
	Group          string     `json:"group"`
	Total          int        `json:"total"`
	Scores         []int      `json:"scores"`
	LiveTotal      int        `json:"live_total"`
	PostTotal      int        `json:"post_total"`
	LiveScores     []int      `json:"live_scores,omitempty"`
	PostScores     []int      `json:"post_scores,omitempty"`
	StartedAt      *time.Time `json:"started_at,omitempty"`
	ElapsedSeconds int64      `json:"elapsed_seconds,omitempty"`
}

type ContestSessionState struct {
	Started          bool       `json:"started"`
	CanStart         bool       `json:"can_start"`
	InLive           bool       `json:"in_live"`
	StartAt          *time.Time `json:"start_at,omitempty"`
	EndAt            *time.Time `json:"end_at,omitempty"`
	RemainingSeconds int64      `json:"remaining_seconds"`
}

// UintList 无符号整数列表（用于 GORM 序列化）
type UintList []uint

func (s UintList) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *UintList) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		str, ok := value.(string)
		if !ok {
			*s = nil
			return nil
		}
		bytes = []byte(str)
	}
	return json.Unmarshal(bytes, s)
}
