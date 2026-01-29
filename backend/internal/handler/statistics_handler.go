package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"oj-system/internal/model"
	"oj-system/internal/repository"
)

type StatisticsHandler struct {
	userRepo       *repository.UserRepository
	problemRepo    *repository.ProblemRepository
	submissionRepo *repository.SubmissionRepository
}

func NewStatisticsHandler() *StatisticsHandler {
	return &StatisticsHandler{
		userRepo:       repository.NewUserRepository(),
		problemRepo:    repository.NewProblemRepository(),
		submissionRepo: repository.NewSubmissionRepository(),
	}
}

// GetPublic 获取系统统计（公开）
// GET /api/v1/statistics
func (h *StatisticsHandler) GetPublic(c *gin.Context) {
	userCount, err := h.userRepo.CountAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ServerError("获取用户统计失败"))
		return
	}

	problemCount, err := h.problemRepo.CountPublic()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ServerError("获取题目统计失败"))
		return
	}

	submissionCount, err := h.submissionRepo.CountAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ServerError("获取提交统计失败"))
		return
	}

	c.JSON(http.StatusOK, model.Success(gin.H{
		"users":       userCount,
		"problems":    problemCount,
		"submissions": submissionCount,
	}))
}
