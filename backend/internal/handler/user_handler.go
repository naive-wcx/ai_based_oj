package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"oj-system/internal/middleware"
	"oj-system/internal/model"
	"oj-system/internal/service"
)

type UserHandler struct {
	service *service.UserService
}

func NewUserHandler() *UserHandler {
	return &UserHandler{
		service: service.NewUserService(),
	}
}

// Login 用户登录
// POST /api/v1/user/login
func (h *UserHandler) Login(c *gin.Context) {
	var req model.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.BadRequest("参数错误"))
		return
	}

	resp, err := h.service.Login(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.BadRequest(err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.Success(resp))
}

// GetProfile 获取个人信息
// GET /api/v1/user/profile
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	
	info, err := h.service.GetProfile(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, model.NotFound(err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.Success(info))
}

// UpdateProfile 更新个人信息
// PUT /api/v1/user/profile
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req struct {
		Email     string `json:"email"`
		StudentID string `json:"student_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.BadRequest("参数错误"))
		return
	}

	if err := h.service.UpdateProfile(userID, req.Email, req.StudentID); err != nil {
		c.JSON(http.StatusBadRequest, model.BadRequest(err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.SuccessMessage("更新成功", nil))
}

// GetRankList 获取排行榜
// GET /api/v1/rank
func (h *UserHandler) GetRankList(c *gin.Context) {
	page := getIntQuery(c, "page", 1)
	size := getIntQuery(c, "size", 20)

	users, total, err := h.service.GetRankList(page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ServerError("获取排行榜失败"))
		return
	}

	c.JSON(http.StatusOK, model.Success(&model.PageData{
		Total: total,
		Page:  page,
		Size:  size,
		List:  users,
	}))
}

// GetUserList 获取用户列表（管理员）
// GET /api/v1/admin/users
func (h *UserHandler) GetUserList(c *gin.Context) {
	page := getIntQuery(c, "page", 1)
	size := getIntQuery(c, "size", 20)

	users, total, err := h.service.GetUserList(page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ServerError("获取用户列表失败"))
		return
	}

	c.JSON(http.StatusOK, model.Success(&model.PageData{
		Total: total,
		Page:  page,
		Size:  size,
		List:  users,
	}))
}

// SetUserRole 设置用户角色（管理员）
// PUT /api/v1/admin/users/:id/role
func (h *UserHandler) SetUserRole(c *gin.Context) {
	userID := getUintParam(c, "id")
	if userID == 0 {
		c.JSON(http.StatusBadRequest, model.BadRequest("用户 ID 无效"))
		return
	}

	var req struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.BadRequest("参数错误"))
		return
	}

	if err := h.service.SetUserRole(userID, req.Role); err != nil {
		c.JSON(http.StatusBadRequest, model.BadRequest(err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.SuccessMessage("设置成功", nil))
}

// CreateUser 管理员创建用户
// POST /api/v1/admin/users
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req model.AdminCreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.BadRequest("参数错误: "+err.Error()))
		return
	}

	user, err := h.service.CreateUserByAdmin(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.BadRequest(err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.Success(user.ToUserInfo()))
}

// CreateUsersBatch 管理员批量创建用户
// POST /api/v1/admin/users/batch
func (h *UserHandler) CreateUsersBatch(c *gin.Context) {
	var req model.AdminCreateUsersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.BadRequest("参数错误: "+err.Error()))
		return
	}

	created, errorsList := h.service.CreateUsersBatch(&req)
	c.JSON(http.StatusOK, model.Success(gin.H{
		"total":   len(req.Users),
		"created": created,
		"failed":  len(errorsList),
		"errors":  errorsList,
	}))
}

// UpdateUser 管理员更新用户信息
// PUT /api/v1/admin/users/:id
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userID := getUintParam(c, "id")
	if userID == 0 {
		c.JSON(http.StatusBadRequest, model.BadRequest("用户 ID 无效"))
		return
	}

	var req model.AdminUpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.BadRequest("参数错误: "+err.Error()))
		return
	}

	user, err := h.service.UpdateUserByAdmin(userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.BadRequest(err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.Success(user.ToUserInfo()))
}
