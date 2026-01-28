package model

import (
	"time"
)

// Setting 系统设置
type Setting struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Key       string    `json:"key" gorm:"uniqueIndex;size:100;not null"`
	Value     string    `json:"value" gorm:"type:text"`
	UpdatedAt time.Time `json:"updated_at"`
}

// 设置键名常量
const (
	SettingAIEnabled    = "ai_enabled"
	SettingAIProvider   = "ai_provider"
	SettingAIAPIKey     = "ai_api_key"
	SettingAIAPIURL     = "ai_api_url"
	SettingAIModel      = "ai_model"
	SettingAITimeout    = "ai_timeout"
	SettingJWTSecret    = "jwt_secret"
)

// AISettings AI 相关设置
type AISettings struct {
	Enabled  bool   `json:"enabled"`
	Provider string `json:"provider"`
	APIKey   string `json:"api_key"`
	APIURL   string `json:"api_url"`
	Model    string `json:"model"`
	Timeout  int    `json:"timeout"`
}

// UpdateAISettingsRequest 更新 AI 设置请求
type UpdateAISettingsRequest struct {
	Enabled  bool   `json:"enabled"`
	Provider string `json:"provider"`
	APIKey   string `json:"api_key"`
	APIURL   string `json:"api_url"`
	Model    string `json:"model"`
	Timeout  int    `json:"timeout"`
}
