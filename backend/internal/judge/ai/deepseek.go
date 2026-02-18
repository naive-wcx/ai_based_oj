package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"oj-system/internal/model"
	"oj-system/internal/service"
)

// DeepSeekClient DeepSeek API 客户端
type DeepSeekClient struct{}

// NewDeepSeekClient 创建 DeepSeek 客户端
func NewDeepSeekClient() *DeepSeekClient {
	return &DeepSeekClient{}
}

// getSettings 获取当前 AI 设置
func (c *DeepSeekClient) getSettings() *model.AISettings {
	return service.GetSettingService().GetAISettings()
}

// ChatRequest API 请求结构
type ChatRequest struct {
	Model          string        `json:"model"`
	Messages       []ChatMessage `json:"messages"`
	Temperature    float64       `json:"temperature"`
	ResponseFormat *ResponseFormat `json:"response_format,omitempty"`
}

// ChatMessage 消息结构
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ResponseFormat 响应格式
type ResponseFormat struct {
	Type string `json:"type"`
}

// ChatResponse API 响应结构
type ChatResponse struct {
	ID      string `json:"id"`
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// AIAnalysisResult AI 分析结果
type AIAnalysisResult struct {
	AlgorithmAnalysis struct {
		DetectedAlgorithms []string `json:"detected_algorithms"`
		PrimaryAlgorithm   string   `json:"primary_algorithm"`
		Confidence         float64  `json:"confidence"`
		Evidence           string   `json:"evidence"`
	} `json:"algorithm_analysis"`
	LanguageFeatures struct {
		Language              string   `json:"language"`
		UsedFeatures          []string `json:"used_features"`
		ForbiddenFeaturesUsed []string `json:"forbidden_features_used"`
	} `json:"language_features"`
	RequirementCheck struct {
		AlgorithmMatch     bool `json:"algorithm_match"`
		LanguageMatch      bool `json:"language_match"`
		AllRequirementsMet bool `json:"all_requirements_met"`
	} `json:"requirement_check"`
	Summary string `json:"summary"`
}

// AnalyzeCode 分析代码
func (c *DeepSeekClient) AnalyzeCode(problem *model.Problem, code string, language string) (*model.AIJudgeResult, error) {
	settings := c.getSettings()
	
	// 检查是否启用 AI 判题
	if !settings.Enabled {
		return &model.AIJudgeResult{
			Enabled: true,
			Passed:  true,
			Reason:  "AI 判题功能未启用",
			Summary: "AI 判题功能未启用",
		}, nil
	}
	
	// 检查 API Key
	if settings.APIKey == "" {
		return &model.AIJudgeResult{
			Enabled: true,
			Passed:  true,
			Reason:  "AI 判题未配置 API Key，已跳过",
			Summary: "AI 判题未配置 API Key，已跳过",
		}, nil
	}

	aiConfig := problem.AIJudgeConfig
	if aiConfig == nil || !aiConfig.Enabled {
		return nil, nil
	}

	// 构建 prompt
	prompt := buildPrompt(problem, code, language, aiConfig)

	// 调用 API
	messages := []ChatMessage{
		{
			Role:    "system",
			Content: "你是一个专业的代码分析专家，擅长识别代码中使用的算法和编程技术。请严格按照用户要求的 JSON 格式输出分析结果，不要输出其他内容。",
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	response, err := c.chat(messages, settings)
	if err != nil {
		return &model.AIJudgeResult{
			Enabled:     true,
			Passed:      true, // 出错时默认通过，不影响正常判题
			Reason:      fmt.Sprintf("AI 分析出错: %v", err),
			Summary:     fmt.Sprintf("AI 分析出错: %v", err),
		}, nil
	}

	// 解析响应
	return parseAIResponse(response, aiConfig)
}

// chat 发送聊天请求
func (c *DeepSeekClient) chat(messages []ChatMessage, settings *model.AISettings) (string, error) {
	reqBody := ChatRequest{
		Model:       settings.Model,
		Messages:    messages,
		Temperature: 0.1,
		ResponseFormat: &ResponseFormat{
			Type: "json_object",
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %v", err)
	}

	client := &http.Client{
		Timeout: time.Duration(settings.Timeout) * time.Second,
	}

	req, err := http.NewRequest("POST", settings.APIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+settings.APIKey)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}

	if chatResp.Error != nil {
		return "", fmt.Errorf("API 错误: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("无响应内容")
	}

	return chatResp.Choices[0].Message.Content, nil
}

// buildPrompt 构建分析提示词
func buildPrompt(problem *model.Problem, code string, language string, aiConfig *model.AIJudgeConfig) string {
	var sb strings.Builder

	sb.WriteString("# 任务\n")
	sb.WriteString("分析用户提交的代码，判断是否符合题目的特定要求。\n\n")

	sb.WriteString("# 题目信息\n")
	sb.WriteString(fmt.Sprintf("- 题目标题：%s\n", problem.Title))
	sb.WriteString(fmt.Sprintf("- 题目描述：%s\n\n", problem.Description))

	sb.WriteString("# 题目要求\n")
	if aiConfig.RequiredAlgorithm != "" {
		sb.WriteString(fmt.Sprintf("- 必须使用的算法：%s\n", aiConfig.RequiredAlgorithm))
	}
	if len(aiConfig.RequiredLanguage) > 0 {
		sb.WriteString(fmt.Sprintf("- 必须使用的编程语言：%s\n", strings.Join(aiConfig.RequiredLanguage, "、")))
	}
	if len(aiConfig.ForbiddenFeatures) > 0 {
		sb.WriteString(fmt.Sprintf("- 禁止使用的特性：%s\n", strings.Join(aiConfig.ForbiddenFeatures, ", ")))
	}
	if aiConfig.CustomPrompt != "" {
		sb.WriteString(fmt.Sprintf("- 额外要求：%s\n", aiConfig.CustomPrompt))
	}
	sb.WriteString("\n")

	sb.WriteString(fmt.Sprintf("# 用户提交的代码（%s）\n", language))
	sb.WriteString("```" + language + "\n")
	sb.WriteString(code)
	sb.WriteString("\n```\n\n")

	sb.WriteString(`# 输出要求
请严格按照以下 JSON 格式输出分析结果：

{
    "algorithm_analysis": {
        "detected_algorithms": ["检测到的算法列表"],
        "primary_algorithm": "主要使用的算法",
        "confidence": 0.0到1.0的置信度,
        "evidence": "判断依据说明"
    },
    "language_features": {
        "language": "编程语言",
        "used_features": ["使用的语言特性"],
        "forbidden_features_used": ["使用了的禁止特性"]
    },
    "requirement_check": {
        "algorithm_match": true或false,
        "language_match": true或false,
        "all_requirements_met": true或false
    },
    "summary": "一句话总结"
}`)

	return sb.String()
}

// parseAIResponse 解析 AI 响应
func parseAIResponse(response string, aiConfig *model.AIJudgeConfig) (*model.AIJudgeResult, error) {
	var analysis AIAnalysisResult
	if err := json.Unmarshal([]byte(response), &analysis); err != nil {
		return &model.AIJudgeResult{
			Enabled:     true,
			Passed:      true, // 解析失败时默认通过
			Reason:      fmt.Sprintf("AI 响应解析失败: %v", err),
			Summary:     fmt.Sprintf("AI 响应解析失败: %v", err),
		}, nil
	}

	result := &model.AIJudgeResult{
		Enabled:           true,
		Passed:            analysis.RequirementCheck.AllRequirementsMet,
		AlgorithmDetected: analysis.AlgorithmAnalysis.PrimaryAlgorithm,
		LanguageCheck:     "passed",
		Reason:            analysis.Summary,
		Summary:           analysis.Summary,
	}

	if !analysis.RequirementCheck.LanguageMatch {
		result.LanguageCheck = "failed"
	}

	if !result.Passed {
		result.Details = &model.AIJudgeDetails{
			Required:   aiConfig.RequiredAlgorithm,
			Detected:   analysis.AlgorithmAnalysis.PrimaryAlgorithm,
			Confidence: analysis.AlgorithmAnalysis.Confidence,
		}
	}

	return result, nil
}
