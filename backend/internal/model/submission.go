package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// 提交状态常量
const (
	StatusPending            = "Pending"
	StatusJudging            = "Judging"
	StatusAccepted           = "Accepted"
	StatusWrongAnswer        = "Wrong Answer"
	StatusTimeLimitExceeded  = "Time Limit Exceeded"
	StatusMemoryLimitExceeded = "Memory Limit Exceeded"
	StatusRuntimeError       = "Runtime Error"
	StatusCompileError       = "Compile Error"
	StatusSystemError        = "System Error"
)

// Submission 提交记录
type Submission struct {
	ID              uint              `json:"id" gorm:"primaryKey"`
	ProblemID       uint              `json:"problem_id" gorm:"index;not null"`
	UserID          uint              `json:"user_id" gorm:"index;not null"`
	Language        string            `json:"language" gorm:"size:20;not null"`
	Code            string            `json:"code" gorm:"type:text;not null"`
	Status          string            `json:"status" gorm:"size:30;default:Pending"`
	TimeUsed        int               `json:"time_used"`   // ms
	MemoryUsed      int               `json:"memory_used"` // KB
	Score           int               `json:"score" gorm:"default:0"`
	TestcaseResults TestcaseResultList `json:"testcase_results" gorm:"type:text"`
	AIJudgeResult   *AIJudgeResult    `json:"ai_judge_result" gorm:"type:text"`
	CompileError    string            `json:"compile_error" gorm:"type:text"`
	FinalMessage    string            `json:"final_message" gorm:"type:text"`
	CreatedAt       time.Time         `json:"created_at"`

	// 关联字段（不存储）
	Problem  *Problem `json:"problem,omitempty" gorm:"-"`
	User     *User    `json:"user,omitempty" gorm:"-"`
}

// TestcaseResult 单个测试点结果
type TestcaseResult struct {
	ID       int    `json:"id"`
	Status   string `json:"status"`
	Time     int    `json:"time"`   // ms
	Memory   int    `json:"memory"` // KB
	Message  string `json:"message,omitempty"`
}

// TestcaseResultList 测试点结果列表
type TestcaseResultList []TestcaseResult

func (t TestcaseResultList) Value() (driver.Value, error) {
	return json.Marshal(t)
}

func (t *TestcaseResultList) Scan(value interface{}) error {
	if value == nil {
		*t = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		str, ok := value.(string)
		if !ok {
			*t = nil
			return nil
		}
		bytes = []byte(str)
	}
	return json.Unmarshal(bytes, t)
}

// AIJudgeResult AI 判题结果
type AIJudgeResult struct {
	Enabled           bool              `json:"enabled"`
	Passed            bool              `json:"passed"`
	AlgorithmDetected string            `json:"algorithm_detected,omitempty"`
	LanguageCheck     string            `json:"language_check,omitempty"`
	Reason            string            `json:"reason,omitempty"`
	Details           *AIJudgeDetails   `json:"details,omitempty"`
	RawResponse       string            `json:"raw_response,omitempty"`
}

// AIJudgeDetails AI 判题详细信息
type AIJudgeDetails struct {
	Required   string  `json:"required,omitempty"`
	Detected   string  `json:"detected,omitempty"`
	Confidence float64 `json:"confidence,omitempty"`
}

func (a *AIJudgeResult) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a)
}

func (a *AIJudgeResult) Scan(value interface{}) error {
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
	return json.Unmarshal(bytes, a)
}

// SubmissionCreateRequest 提交代码请求
type SubmissionCreateRequest struct {
	ProblemID uint   `json:"problem_id" binding:"required"`
	Language  string `json:"language" binding:"required"`
	Code      string `json:"code" binding:"required"`
}

// SubmissionListItem 提交列表项
type SubmissionListItem struct {
	ID         uint      `json:"id"`
	ProblemID  uint      `json:"problem_id"`
	ProblemTitle string  `json:"problem_title"`
	UserID     uint      `json:"user_id"`
	Username   string    `json:"username"`
	Language   string    `json:"language"`
	Status     string    `json:"status"`
	TimeUsed   int       `json:"time_used"`
	MemoryUsed int       `json:"memory_used"`
	Score      int       `json:"score"`
	CreatedAt  time.Time `json:"created_at"`
}
