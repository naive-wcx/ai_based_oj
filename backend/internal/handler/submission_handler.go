package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"oj-system/internal/judge"
	"oj-system/internal/middleware"
	"oj-system/internal/model"
	"oj-system/internal/service"
)

type SubmissionHandler struct {
	service *service.SubmissionService
}

func NewSubmissionHandler() *SubmissionHandler {
	return &SubmissionHandler{
		service: service.NewSubmissionService(),
	}
}

// Submit 提交代码
// POST /api/v1/submission
func (h *SubmissionHandler) Submit(c *gin.Context) {
	var req model.SubmissionCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.BadRequest("参数错误: "+err.Error()))
		return
	}

	userID := middleware.GetUserID(c)
	submission, err := h.service.Submit(&req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.BadRequest(err.Error()))
		return
	}

	// 提交到判题队列
	go func() {
		if err := judge.SubmitToQueue(submission); err != nil {
			// 记录错误但不影响响应
			submission.Status = model.StatusSystemError
			submission.FinalMessage = "提交到判题队列失败: " + err.Error()
			_ = h.service.UpdateResult(submission)
		}
	}()

	c.JSON(http.StatusOK, model.Success(submission))
}

// GetByID 获取提交详情
// GET /api/v1/submission/:id
func (h *SubmissionHandler) GetByID(c *gin.Context) {
	id := getUintParam(c, "id")
	if id == 0 {
		c.JSON(http.StatusBadRequest, model.BadRequest("提交 ID 无效"))
		return
	}

	userID := middleware.GetUserID(c)
	isAdmin := middleware.IsAdmin(c)

	submission, err := h.service.GetByID(id, userID, isAdmin)
	if err != nil {
		if err == service.ErrSubmissionForbidden {
			c.JSON(http.StatusForbidden, model.Forbidden(err.Error()))
			return
		}
		c.JSON(http.StatusNotFound, model.NotFound(err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.Success(submission))
}

// List 获取提交列表
// GET /api/v1/submission/list
func (h *SubmissionHandler) List(c *gin.Context) {
	page := getIntQuery(c, "page", 1)
	size := getIntQuery(c, "size", 20)
	problemID := getUintQuery(c, "problem_id")
	status := c.Query("status")

	currentUserID := middleware.GetUserID(c)
	isAdmin := middleware.IsAdmin(c)

	userID := currentUserID
	if isAdmin {
		userID = getUintQuery(c, "user_id")
	}

	data, err := h.service.List(page, size, problemID, userID, status, currentUserID, isAdmin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ServerError("获取提交列表失败"))
		return
	}

	c.JSON(http.StatusOK, model.Success(data))
}

// GetMySubmissions 获取我的提交
// GET /api/v1/submission/my
func (h *SubmissionHandler) GetMySubmissions(c *gin.Context) {
	page := getIntQuery(c, "page", 1)
	size := getIntQuery(c, "size", 20)
	problemID := getUintQuery(c, "problem_id")
	status := c.Query("status")

	userID := middleware.GetUserID(c)

	data, err := h.service.List(page, size, problemID, userID, status, userID, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ServerError("获取提交列表失败"))
		return
	}

	c.JSON(http.StatusOK, model.Success(data))
}
