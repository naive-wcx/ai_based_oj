package handler

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"oj-system/internal/middleware"
	"oj-system/internal/model"
	"oj-system/internal/repository"
	"oj-system/internal/service"
)

type ContestHandler struct {
	service *service.ContestService
	submissionRepo *repository.SubmissionRepository
}

func NewContestHandler() *ContestHandler {
	return &ContestHandler{
		service: service.NewContestService(),
		submissionRepo: repository.NewSubmissionRepository(),
	}
}

// List 获取比赛列表
// GET /api/v1/contest/list
func (h *ContestHandler) List(c *gin.Context) {
	page := getIntQuery(c, "page", 1)
	size := getIntQuery(c, "size", 20)
	userID := middleware.GetUserID(c)
	isAdmin := middleware.IsAdmin(c)

	contests, total, err := h.service.ListForUser(page, size, userID, isAdmin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ServerError("获取比赛列表失败"))
		return
	}

	c.JSON(http.StatusOK, model.Success(&model.PageData{
		Total: total,
		Page:  page,
		Size:  size,
		List:  contests,
	}))
}

// GetByID 获取比赛详情
// GET /api/v1/contest/:id
func (h *ContestHandler) GetByID(c *gin.Context) {
	contestID := getUintParam(c, "id")
	if contestID == 0 {
		c.JSON(http.StatusBadRequest, model.BadRequest("比赛 ID 无效"))
		return
	}

	userID := middleware.GetUserID(c)
	isAdmin := middleware.IsAdmin(c)

	contest, err := h.service.GetByIDForUser(contestID, userID, isAdmin)
	if err != nil {
		if err.Error() == "比赛不存在" || err.Error() == "用户不存在" {
			c.JSON(http.StatusNotFound, model.NotFound(err.Error()))
		} else {
			c.JSON(http.StatusForbidden, model.Forbidden(err.Error()))
		}
		return
	}

	problems, err := h.service.GetProblemsByIDs([]uint(contest.ProblemIDs))
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ServerError("获取题目失败"))
		return
	}

	acceptedSet := map[uint]struct{}{}
	if userID > 0 {
		acceptedIDs, err := h.submissionRepo.GetAcceptedProblemIDsInRange(userID, []uint(contest.ProblemIDs), contest.StartAt, contest.EndAt)
		if err == nil {
			for _, pid := range acceptedIDs {
				acceptedSet[pid] = struct{}{}
			}
		}
	}

	ordered := buildContestProblemList(contest.ProblemIDs, problems, acceptedSet)

	var myTotal *int
	if userID > 0 {
		showScore := contest.Type == "ioi" || !time.Now().Before(contest.EndAt)
		if showScore {
			scoreMap, err := h.submissionRepo.GetUserBestScoresInRange(userID, []uint(contest.ProblemIDs), contest.StartAt, contest.EndAt)
			if err == nil {
				total := 0
				for _, pid := range contest.ProblemIDs {
					total += scoreMap[pid]
				}
				myTotal = &total
			}
		}
	}

	c.JSON(http.StatusOK, model.Success(gin.H{
		"contest":  contest,
		"problems": ordered,
		"my_total": myTotal,
	}))
}

// Create 创建比赛（管理员）
// POST /api/v1/admin/contests
func (h *ContestHandler) Create(c *gin.Context) {
	var req model.ContestCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.BadRequest("参数错误: "+err.Error()))
		return
	}

	userID := middleware.GetUserID(c)
	contest, err := h.service.Create(&req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.BadRequest(err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.Success(contest))
}

// Update 更新比赛（管理员）
// PUT /api/v1/admin/contests/:id
func (h *ContestHandler) Update(c *gin.Context) {
	contestID := getUintParam(c, "id")
	if contestID == 0 {
		c.JSON(http.StatusBadRequest, model.BadRequest("比赛 ID 无效"))
		return
	}

	var req model.ContestUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.BadRequest("参数错误: "+err.Error()))
		return
	}

	contest, err := h.service.Update(contestID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.BadRequest(err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.Success(contest))
}

// Delete 删除比赛（管理员）
// DELETE /api/v1/admin/contests/:id
func (h *ContestHandler) Delete(c *gin.Context) {
	contestID := getUintParam(c, "id")
	if contestID == 0 {
		c.JSON(http.StatusBadRequest, model.BadRequest("比赛 ID 无效"))
		return
	}

	if err := h.service.Delete(contestID); err != nil {
		c.JSON(http.StatusBadRequest, model.BadRequest(err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.SuccessMessage("删除成功", nil))
}

// GetLeaderboard 获取比赛排行榜（管理员）
// GET /api/v1/admin/contests/:id/leaderboard
func (h *ContestHandler) GetLeaderboard(c *gin.Context) {
	contestID := getUintParam(c, "id")
	if contestID == 0 {
		c.JSON(http.StatusBadRequest, model.BadRequest("比赛 ID 无效"))
		return
	}

	contest, problemIDs, entries, err := h.service.GetLeaderboard(contestID)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.BadRequest(err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.Success(gin.H{
		"contest":     contest,
		"problem_ids": problemIDs,
		"entries":     entries,
	}))
}

// ExportLeaderboard 导出比赛成绩（管理员）
// GET /api/v1/admin/contests/:id/export
func (h *ContestHandler) ExportLeaderboard(c *gin.Context) {
	contestID := getUintParam(c, "id")
	if contestID == 0 {
		c.JSON(http.StatusBadRequest, model.BadRequest("比赛 ID 无效"))
		return
	}

	_, problemIDs, entries, err := h.service.GetLeaderboard(contestID)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.BadRequest(err.Error()))
		return
	}

	filename := fmt.Sprintf("contest_%d_leaderboard.csv", contestID)
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	header := []string{"user_id", "username", "group", "total"}
	for _, pid := range problemIDs {
		header = append(header, fmt.Sprintf("P%d", pid))
	}
	_ = writer.Write(header)

	for _, entry := range entries {
		row := []string{
			strconv.FormatUint(uint64(entry.UserID), 10),
			entry.Username,
			entry.Group,
			strconv.Itoa(entry.Total),
		}
		for _, score := range entry.Scores {
			row = append(row, strconv.Itoa(score))
		}
		_ = writer.Write(row)
	}
}

func buildContestProblemList(ids model.UintList, problems []model.Problem, acceptedSet map[uint]struct{}) []model.ProblemListItem {
	result := make([]model.ProblemListItem, 0, len(ids))
	problemMap := make(map[uint]model.Problem, len(problems))
	for _, problem := range problems {
		problemMap[problem.ID] = problem
	}
	for _, id := range ids {
		problem, ok := problemMap[id]
		if !ok {
			continue
		}
		_, hasAccepted := acceptedSet[id]
		result = append(result, model.ProblemListItem{
			ID:            problem.ID,
			Title:         problem.Title,
			Difficulty:    problem.Difficulty,
			Tags:          problem.Tags,
			SubmitCount:   problem.SubmitCount,
			AcceptedCount: problem.AcceptedCount,
			HasAIJudge:    problem.AIJudgeConfig != nil && problem.AIJudgeConfig.Enabled,
			HasAccepted:   hasAccepted,
		})
	}
	return result
}
