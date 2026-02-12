package handler

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"oj-system/internal/judge"
	"oj-system/internal/middleware"
	"oj-system/internal/model"
	"oj-system/internal/service"
)

type ProblemHandler struct {
	service *service.ProblemService
}

func NewProblemHandler() *ProblemHandler {
	return &ProblemHandler{
		service: service.NewProblemService(),
	}
}

// List 获取题目列表
// GET /api/v1/problem/list
func (h *ProblemHandler) List(c *gin.Context) {
	page := getIntQuery(c, "page", 1)
	size := getIntQuery(c, "size", 20)
	difficulty := c.Query("difficulty")
	tag := c.Query("tag")
	keyword := c.Query("keyword")
	userID := middleware.GetUserID(c)
	isAdmin := middleware.IsAdmin(c)

	problems, total, err := h.service.ListWithUser(page, size, difficulty, tag, keyword, userID, isAdmin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ServerError("获取题目列表失败"))
		return
	}

	c.JSON(http.StatusOK, model.Success(&model.PageData{
		Total: total,
		Page:  page,
		Size:  size,
		List:  problems,
	}))
}

// GetByID 获取题目详情
// GET /api/v1/problem/:id
func (h *ProblemHandler) GetByID(c *gin.Context) {
	id := getUintParam(c, "id")
	if id == 0 {
		c.JSON(http.StatusBadRequest, model.BadRequest("题目 ID 无效"))
		return
	}

	userID := middleware.GetUserID(c)
	isAdmin := middleware.IsAdmin(c)
	problem, err := h.service.GetByIDWithUser(id, userID, isAdmin)
	if err != nil {
		if err.Error() == "无权限访问该题目" {
			c.JSON(http.StatusForbidden, model.Forbidden(err.Error()))
			return
		}
		c.JSON(http.StatusNotFound, model.NotFound(err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.Success(problem))
}

// Create 创建题目（管理员）
// POST /api/v1/problem
func (h *ProblemHandler) Create(c *gin.Context) {
	var req model.ProblemCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.BadRequest("参数错误: "+err.Error()))
		return
	}

	userID := middleware.GetUserID(c)
	problem, err := h.service.Create(&req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.BadRequest(err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.Success(problem))
}

// Update 更新题目（管理员）
// PUT /api/v1/problem/:id
func (h *ProblemHandler) Update(c *gin.Context) {
	id := getUintParam(c, "id")
	if id == 0 {
		c.JSON(http.StatusBadRequest, model.BadRequest("题目 ID 无效"))
		return
	}

	var req model.ProblemCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.BadRequest("参数错误: "+err.Error()))
		return
	}

	problem, err := h.service.Update(id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.BadRequest(err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.Success(problem))
}

// Delete 删除题目（管理员）
// DELETE /api/v1/problem/:id
func (h *ProblemHandler) Delete(c *gin.Context) {
	id := getUintParam(c, "id")
	if id == 0 {
		c.JSON(http.StatusBadRequest, model.BadRequest("题目 ID 无效"))
		return
	}

	if err := h.service.Delete(id); err != nil {
		c.JSON(http.StatusBadRequest, model.BadRequest(err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.SuccessMessage("删除成功", nil))
}

// UploadTestcase 上传测试用例（管理员）
// POST /api/v1/problem/:id/testcase
func (h *ProblemHandler) UploadTestcase(c *gin.Context) {
	id := getUintParam(c, "id")
	if id == 0 {
		c.JSON(http.StatusBadRequest, model.BadRequest("题目 ID 无效"))
		return
	}

	// 获取上传的文件
	inputFile, err := c.FormFile("input")
	if err != nil {
		c.JSON(http.StatusBadRequest, model.BadRequest("请上传输入文件"))
		return
	}

	outputFile, err := c.FormFile("output")
	if err != nil {
		c.JSON(http.StatusBadRequest, model.BadRequest("请上传输出文件"))
		return
	}

	// 打开文件
	inputReader, err := inputFile.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, model.BadRequest("无法读取输入文件"))
		return
	}
	defer inputReader.Close()

	outputReader, err := outputFile.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, model.BadRequest("无法读取输出文件"))
		return
	}
	defer outputReader.Close()

	// 获取分数
	score := getIntFormValue(c, "score", 10)
	isSample := c.PostForm("is_sample") == "true"

	// 添加测试用例
	if err := h.service.AddTestcase(id, inputReader, outputReader, score, isSample); err != nil {
		c.JSON(http.StatusBadRequest, model.BadRequest(err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.SuccessMessage("上传成功", nil))
}

// UploadTestcaseZip 批量上传测试用例 (Zip)
// POST /api/v1/problem/:id/testcase/zip
func (h *ProblemHandler) UploadTestcaseZip(c *gin.Context) {
	id := getUintParam(c, "id")
	if id == 0 {
		c.JSON(http.StatusBadRequest, model.BadRequest("题目 ID 无效"))
		return
	}

	// 获取上传的文件
	file, err := c.FormFile("zip_file")
	if err != nil {
		c.JSON(http.StatusBadRequest, model.BadRequest("请上传 zip 文件"))
		return
	}

	// 保存到临时文件
	tmpFile, err := os.CreateTemp("", "testcase-*.zip")
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ServerError("创建临时文件失败"))
		return
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	if err := c.SaveUploadedFile(file, tmpPath); err != nil {
		c.JSON(http.StatusInternalServerError, model.ServerError("保存文件失败"))
		return
	}

	// 处理 Zip
	if err := h.service.UploadTestcaseZip(id, tmpPath); err != nil {
		c.JSON(http.StatusBadRequest, model.BadRequest(err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.SuccessMessage("批量上传并处理成功", nil))
}

// UploadProblemImage 上传题面图片（管理员）
// POST /api/v1/problem/:id/image
func (h *ProblemHandler) UploadProblemImage(c *gin.Context) {
	id := getUintParam(c, "id")
	if id == 0 {
		c.JSON(http.StatusBadRequest, model.BadRequest("题目 ID 无效"))
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, model.BadRequest("请上传图片文件"))
		return
	}

	reader, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, model.BadRequest("无法读取图片文件"))
		return
	}
	defer reader.Close()

	url, markdown, savedName, err := h.service.UploadProblemImage(id, file.Filename, reader)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.BadRequest(err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.Success(gin.H{
		"url":      url,
		"markdown": markdown,
		"filename": savedName,
	}))
}

// GetProblemImage 获取题面图片
// GET /api/v1/problem/:id/image/:filename
func (h *ProblemHandler) GetProblemImage(c *gin.Context) {
	id := getUintParam(c, "id")
	if id == 0 {
		c.JSON(http.StatusBadRequest, model.BadRequest("题目 ID 无效"))
		return
	}
	filename := c.Param("filename")
	path, err := h.service.ResolveProblemImagePath(id, filename)
	if err != nil {
		c.JSON(http.StatusNotFound, model.NotFound(err.Error()))
		return
	}

	c.Header("Cache-Control", "public, max-age=86400")
	c.File(path)
}

// GetTestcases 获取测试用例列表（管理员）
// GET /api/v1/problem/:id/testcases
func (h *ProblemHandler) GetTestcases(c *gin.Context) {
	id := getUintParam(c, "id")
	if id == 0 {
		c.JSON(http.StatusBadRequest, model.BadRequest("题目 ID 无效"))
		return
	}

	testcases, err := h.service.GetTestcases(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ServerError("获取测试用例失败"))
		return
	}

	c.JSON(http.StatusOK, model.Success(testcases))
}

// DeleteTestcases 删除所有测试用例（管理员）
// DELETE /api/v1/problem/:id/testcases
func (h *ProblemHandler) DeleteTestcases(c *gin.Context) {
	id := getUintParam(c, "id")
	if id == 0 {
		c.JSON(http.StatusBadRequest, model.BadRequest("题目 ID 无效"))
		return
	}

	if err := h.service.DeleteTestcases(id); err != nil {
		c.JSON(http.StatusBadRequest, model.BadRequest(err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.SuccessMessage("删除成功", nil))
}

// RejudgeProblem 重新评测题目的所有历史提交（管理员）
// POST /api/v1/problem/:id/rejudge
func (h *ProblemHandler) RejudgeProblem(c *gin.Context) {
	id := getUintParam(c, "id")
	if id == 0 {
		c.JSON(http.StatusBadRequest, model.BadRequest("题目 ID 无效"))
		return
	}

	submissions, err := h.service.PrepareProblemRejudge(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.BadRequest(err.Error()))
		return
	}
	if len(submissions) == 0 {
		c.JSON(http.StatusOK, model.SuccessMessage("暂无可重测的历史提交", gin.H{
			"total":  0,
			"queued": 0,
			"failed": 0,
		}))
		return
	}

	failed := 0
	for i := range submissions {
		if err := judge.SubmitToQueue(&submissions[i]); err != nil {
			failed++
		}
	}

	queued := len(submissions) - failed
	message := "整题重测任务已提交"
	if failed > 0 {
		message = "部分提交进入重测队列失败"
	}

	c.JSON(http.StatusOK, model.SuccessMessage(message, gin.H{
		"total":  len(submissions),
		"queued": queued,
		"failed": failed,
	}))
}
