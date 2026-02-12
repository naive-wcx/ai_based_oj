package handler

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"strings"
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

	now := time.Now()
	sessionState, _ := h.service.GetSessionState(contest, userID, now)
	hasStarted := sessionState != nil && sessionState.Started
	inLive := sessionState != nil && sessionState.InLive

	acceptedSet := map[uint]struct{}{}
	submittedSet := map[uint]struct{}{}
	showAccepted := isAdmin || strings.ToLower(contest.Type) == "ioi" || (hasStarted && !inLive)
	if userID > 0 && hasStarted {
		rangeStart := contest.StartAt
		if sessionState.StartAt != nil {
			rangeStart = *sessionState.StartAt
		}
		submittedIDs, err := h.submissionRepo.GetSubmittedProblemIDsInRange(userID, []uint(contest.ProblemIDs), rangeStart, now)
		if err == nil {
			for _, pid := range submittedIDs {
				submittedSet[pid] = struct{}{}
			}
		}
		if showAccepted {
			acceptedIDs, err := h.submissionRepo.GetAcceptedProblemIDsInRange(userID, []uint(contest.ProblemIDs), rangeStart, now)
			if err == nil {
				for _, pid := range acceptedIDs {
					acceptedSet[pid] = struct{}{}
				}
			}
		}
	}

	hideHiddenTags := !isAdmin && !now.Before(contest.StartAt) && now.Before(contest.EndAt)
	ordered := buildContestProblemList(contest.ProblemIDs, problems, acceptedSet, submittedSet, showAccepted, hideHiddenTags)

	var myTotal *int
	var myLiveTotal *int
	var myPostTotal *int
	if userID > 0 && hasStarted {
		showScore := strings.ToLower(contest.Type) == "ioi" || !inLive
		if showScore {
			liveStart := contest.StartAt
			liveEnd := contest.EndAt
			if sessionState.StartAt != nil {
				liveStart = *sessionState.StartAt
			}
			if sessionState.EndAt != nil {
				liveEnd = *sessionState.EndAt
			}

			if liveEnd.After(now) {
				liveEnd = now
			}
			liveMap, err := h.submissionRepo.GetUserLastScoresInRange(userID, []uint(contest.ProblemIDs), liveStart, liveEnd)
			if err == nil {
				liveTotal := 0
				for _, pid := range contest.ProblemIDs {
					liveTotal += liveMap[pid]
				}
				myLiveTotal = &liveTotal
			}

			postStart := liveEnd.Add(time.Millisecond)
			if !postStart.After(now) {
				postMap, err := h.submissionRepo.GetUserLastScoresInRange(userID, []uint(contest.ProblemIDs), postStart, now)
				if err == nil {
					postTotal := 0
					for _, pid := range contest.ProblemIDs {
						postTotal += postMap[pid]
					}
					myPostTotal = &postTotal
				}
			}

			if myLiveTotal != nil || myPostTotal != nil {
				total := 0
				if myLiveTotal != nil {
					total += *myLiveTotal
				}
				if myPostTotal != nil {
					total += *myPostTotal
				}
				myTotal = &total
			}
		}
	}

	c.JSON(http.StatusOK, model.Success(gin.H{
		"contest":  contest,
		"problems": ordered,
		"session":  sessionState,
		"my_live_total": myLiveTotal,
		"my_post_total": myPostTotal,
		"my_total": myTotal,
	}))
}

// StartContest 开始窗口期比赛会话
// POST /api/v1/contest/:id/start
func (h *ContestHandler) StartContest(c *gin.Context) {
	contestID := getUintParam(c, "id")
	if contestID == 0 {
		c.JSON(http.StatusBadRequest, model.BadRequest("比赛 ID 无效"))
		return
	}
	userID := middleware.GetUserID(c)
	isAdmin := middleware.IsAdmin(c)

	participation, contest, err := h.service.StartWindowContest(contestID, userID, isAdmin)
	if err != nil {
		if err.Error() == "无权限访问该比赛" {
			c.JSON(http.StatusForbidden, model.Forbidden(err.Error()))
			return
		}
		c.JSON(http.StatusBadRequest, model.BadRequest(err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.Success(gin.H{
		"contest_id": contest.ID,
		"user_id": userID,
		"start_at": participation.StartAt,
		"end_at": participation.EndAt,
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

// RefreshStats 刷新比赛相关的统计数据
// POST /api/v1/admin/contests/:id/refresh
func (h *ContestHandler) RefreshStats(c *gin.Context) {
	contestID := getUintParam(c, "id")
	if contestID == 0 {
		c.JSON(http.StatusBadRequest, model.BadRequest("比赛 ID 无效"))
		return
	}

	if err := h.service.RefreshStats(contestID); err != nil {
		c.JSON(http.StatusInternalServerError, model.ServerError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.SuccessMessage("刷新统计数据成功", nil))
}

// GetLeaderboard 获取比赛排行榜（管理员）
// GET /api/v1/admin/contests/:id/leaderboard
func (h *ContestHandler) GetLeaderboard(c *gin.Context) {
	contestID := getUintParam(c, "id")
	if contestID == 0 {
		c.JSON(http.StatusBadRequest, model.BadRequest("比赛 ID 无效"))
		return
	}

	mode := c.DefaultQuery("board_mode", "combined")
	contest, problemIDs, entries, boardMode, err := h.service.GetLeaderboard(contestID, mode)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.BadRequest(err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.Success(gin.H{
		"contest":     contest,
		"problem_ids": problemIDs,
		"entries":     entries,
		"board_mode":  boardMode,
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

	mode := c.DefaultQuery("board_mode", "combined")
	_, problemIDs, entries, boardMode, err := h.service.GetLeaderboard(contestID, mode)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.BadRequest(err.Error()))
		return
	}

	filename := fmt.Sprintf("contest_%d_leaderboard_%s.csv", contestID, boardMode)
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	header := []string{"user_id", "username", "group"}
	if boardMode == "combined" {
		header = append(header, "live_total", "post_total", "combined_total")
		for _, pid := range problemIDs {
			header = append(header, fmt.Sprintf("P%d_live", pid), fmt.Sprintf("P%d_post", pid), fmt.Sprintf("P%d_combined", pid))
		}
	} else {
		header = append(header, "total")
		for _, pid := range problemIDs {
			header = append(header, fmt.Sprintf("P%d", pid))
		}
	}
	_ = writer.Write(header)

	for _, entry := range entries {
		row := []string{
			strconv.FormatUint(uint64(entry.UserID), 10),
			entry.Username,
			entry.Group,
		}
		if boardMode == "combined" {
			row = append(row, strconv.Itoa(entry.LiveTotal), strconv.Itoa(entry.PostTotal), strconv.Itoa(entry.LiveTotal+entry.PostTotal))
			for i := range problemIDs {
				liveScore := 0
				postScore := 0
				if i < len(entry.LiveScores) {
					liveScore = entry.LiveScores[i]
				}
				if i < len(entry.PostScores) {
					postScore = entry.PostScores[i]
				}
				row = append(row, strconv.Itoa(liveScore), strconv.Itoa(postScore), strconv.Itoa(liveScore+postScore))
			}
		} else {
			row = append(row, strconv.Itoa(entry.Total))
			for _, score := range entry.Scores {
				row = append(row, strconv.Itoa(score))
			}
		}
		_ = writer.Write(row)
	}
}

func buildContestProblemList(
	ids model.UintList,
	problems []model.Problem,
	acceptedSet, submittedSet map[uint]struct{},
	showAccepted bool,
	hideHiddenTags bool,
) []model.ProblemListItem {
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
		_, hasSubmitted := submittedSet[id]
		tags := []string(problem.Tags)
		if hideHiddenTags && (problem.IsPublic == nil || !*problem.IsPublic) {
			tags = []string{}
		}
		result = append(result, model.ProblemListItem{
			ID:            problem.ID,
			Title:         problem.Title,
			Difficulty:    problem.Difficulty,
			Tags:          tags,
			SubmitCount:   problem.SubmitCount,
			AcceptedCount: problem.AcceptedCount,
			HasAIJudge:    problem.AIJudgeConfig != nil && problem.AIJudgeConfig.Enabled,
			HasFileIO:     problem.FileIOEnabled,
			HasAccepted:   showAccepted && hasAccepted,
			HasSubmitted:  hasSubmitted,
			IsPublic:      problem.IsPublic,
		})
	}
	return result
}
