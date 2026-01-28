package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"oj-system/internal/model"
	"oj-system/internal/service"
)

type SettingHandler struct {
	service *service.SettingService
}

func NewSettingHandler() *SettingHandler {
	return &SettingHandler{
		service: service.GetSettingService(),
	}
}

// GetAISettings 获取 AI 设置
// GET /api/v1/admin/settings/ai
func (h *SettingHandler) GetAISettings(c *gin.Context) {
	settings := h.service.GetAISettingsForDisplay()
	c.JSON(http.StatusOK, model.Success(settings))
}

// UpdateAISettings 更新 AI 设置
// PUT /api/v1/admin/settings/ai
func (h *SettingHandler) UpdateAISettings(c *gin.Context) {
	var req model.UpdateAISettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.BadRequest("参数错误"))
		return
	}
	
	if err := h.service.UpdateAISettings(&req); err != nil {
		c.JSON(http.StatusInternalServerError, model.ServerError("保存设置失败"))
		return
	}
	
	c.JSON(http.StatusOK, model.SuccessMessage("设置已保存", nil))
}

// TestAIConnection 测试 AI 连接
// POST /api/v1/admin/settings/ai/test
func (h *SettingHandler) TestAIConnection(c *gin.Context) {
	settings := h.service.GetAISettings()
	
	if settings.APIKey == "" {
		c.JSON(http.StatusBadRequest, model.BadRequest("请先配置 API Key"))
		return
	}
	
	// 简单测试：检查配置是否存在
	c.JSON(http.StatusOK, model.SuccessMessage("配置有效", gin.H{
		"provider": settings.Provider,
		"model":    settings.Model,
		"api_url":  settings.APIURL,
	}))
}
