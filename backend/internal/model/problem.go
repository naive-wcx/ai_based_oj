package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// Problem 题目模型
type Problem struct {
	ID            uint          `json:"id" gorm:"primaryKey"`
	Title         string        `json:"title" gorm:"size:200;not null"`
	Description   string        `json:"description" gorm:"type:text;not null"`
	InputFormat   string        `json:"input_format" gorm:"type:text"`
	OutputFormat  string        `json:"output_format" gorm:"type:text"`
	Hint          string        `json:"hint" gorm:"type:text"`
	Samples       SampleList    `json:"samples" gorm:"type:text"`
	TimeLimit     int           `json:"time_limit" gorm:"default:1000"`  // ms
	MemoryLimit   int           `json:"memory_limit" gorm:"default:256"` // MB
	Difficulty    string        `json:"difficulty" gorm:"size:20"`       // easy, medium, hard
	Tags          StringList    `json:"tags" gorm:"type:text"`
	AIJudgeConfig *AIJudgeConfig `json:"ai_judge_config" gorm:"type:text"`
	FileIOEnabled bool          `json:"file_io_enabled" gorm:"default:false"`
	FileInputName string        `json:"file_input_name" gorm:"size:100"`
	FileOutputName string       `json:"file_output_name" gorm:"size:100"`
	IsPublic      *bool         `json:"is_public" gorm:"default:true"`
	CreatedBy     uint          `json:"created_by"`
	SubmitCount   int           `json:"submit_count" gorm:"default:0"`
	AcceptedCount int           `json:"accepted_count" gorm:"default:0"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
	HasAccepted   bool          `json:"has_accepted" gorm:"-"`
}

// Sample 样例
type Sample struct {
	Input  string `json:"input"`
	Output string `json:"output"`
}

// SampleList 样例列表（用于 GORM 序列化）
type SampleList []Sample

func (s SampleList) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *SampleList) Scan(value interface{}) error {
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

// StringList 字符串列表（用于 GORM 序列化）
type StringList []string

func (s StringList) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *StringList) Scan(value interface{}) error {
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

// AIJudgeConfig AI 判题配置
type AIJudgeConfig struct {
	Enabled            bool     `json:"enabled"`
	RequiredAlgorithm  string   `json:"required_algorithm,omitempty"`
	RequiredLanguage   string   `json:"required_language,omitempty"`
	ForbiddenFeatures  []string `json:"forbidden_features,omitempty"`
	CustomPrompt       string   `json:"custom_prompt,omitempty"`
	StrictMode         bool     `json:"strict_mode"`
}

func (c *AIJudgeConfig) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c)
}

func (c *AIJudgeConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		str, ok := value.(string)
		if !ok {
			return nil
		}
		bytes = []byte(str)
	}
	return json.Unmarshal(bytes, c)
}

// Testcase 测试用例
type Testcase struct {
	ID         uint   `json:"id" gorm:"primaryKey"`
	ProblemID  uint   `json:"problem_id" gorm:"index;not null"`
	InputFile  string `json:"input_file" gorm:"size:255;not null"`
	OutputFile string `json:"output_file" gorm:"size:255;not null"`
	Score      int    `json:"score" gorm:"default:0"`
	IsSample   bool   `json:"is_sample" gorm:"default:false"`
	OrderNum   int    `json:"order_num" gorm:"default:0"`
}

// ProblemCreateRequest 创建题目请求
type ProblemCreateRequest struct {
	Title         string         `json:"title" binding:"required,max=200"`
	Description   string         `json:"description" binding:"required"`
	InputFormat   string         `json:"input_format"`
	OutputFormat  string         `json:"output_format"`
	Hint          string         `json:"hint"`
	Samples       []Sample       `json:"samples"`
	TimeLimit     int            `json:"time_limit"`
	MemoryLimit   int            `json:"memory_limit"`
	Difficulty    string         `json:"difficulty"`
	Tags          []string       `json:"tags"`
	AIJudgeConfig *AIJudgeConfig `json:"ai_judge_config"`
	FileIOEnabled bool           `json:"file_io_enabled"`
	FileInputName string         `json:"file_input_name"`
	FileOutputName string        `json:"file_output_name"`
	IsPublic      *bool          `json:"is_public"`
}

// ProblemListItem 题目列表项
type ProblemListItem struct {
	ID            uint     `json:"id"`
	Title         string   `json:"title"`
	Difficulty    string   `json:"difficulty"`
	Tags          []string `json:"tags"`
	SubmitCount   int      `json:"submit_count"`
	AcceptedCount int      `json:"accepted_count"`
	HasAIJudge    bool     `json:"has_ai_judge"`
	HasFileIO     bool     `json:"has_file_io"`
	HasAccepted   bool     `json:"has_accepted"`
	HasSubmitted  bool     `json:"has_submitted"`
	IsPublic      *bool    `json:"is_public"`
}
