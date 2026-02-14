package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"oj-system/internal/handler"
	"oj-system/internal/middleware"
)

// SetupRouter 配置路由
func SetupRouter(mode string) *gin.Engine {
	if mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	r.MaxMultipartMemory = 256 << 20

	// 全局中间件
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.RateLimitMiddleware(120, time.Minute))

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 创建 handler
	userHandler := handler.NewUserHandler()
	problemHandler := handler.NewProblemHandler()
	submissionHandler := handler.NewSubmissionHandler()
	settingHandler := handler.NewSettingHandler()
	contestHandler := handler.NewContestHandler()
	statsHandler := handler.NewStatisticsHandler()

	// API v1
	v1 := r.Group("/api/v1")
	{
		// 用户模块
		user := v1.Group("/user")
		{
			user.POST("/login", userHandler.Login)
			user.GET("/profile", middleware.AuthMiddleware(), userHandler.GetProfile)
			user.PUT("/profile", middleware.AuthMiddleware(), userHandler.UpdateProfile)
			user.PUT("/password", middleware.AuthMiddleware(), userHandler.ChangePassword)
		}

			// 题目模块
			problem := v1.Group("/problem")
			{
				problem.GET("/list", middleware.OptionalAuthMiddleware(), problemHandler.List)
				problem.GET("/:id", middleware.OptionalAuthMiddleware(), problemHandler.GetByID)
				problem.GET("/:id/image/:filename", problemHandler.GetProblemImage)

				// 管理员操作
				problem.POST("", middleware.AuthMiddleware(), middleware.AdminMiddleware(), problemHandler.Create)
				problem.PUT("/:id", middleware.AuthMiddleware(), middleware.AdminMiddleware(), problemHandler.Update)
				problem.DELETE("/:id", middleware.AuthMiddleware(), middleware.AdminMiddleware(), problemHandler.Delete)
				problem.POST("/:id/image", middleware.AuthMiddleware(), middleware.AdminMiddleware(), problemHandler.UploadProblemImage)
				problem.POST("/:id/testcase", middleware.AuthMiddleware(), middleware.AdminMiddleware(), problemHandler.UploadTestcase)
				problem.POST("/:id/testcase/zip", middleware.AuthMiddleware(), middleware.AdminMiddleware(), problemHandler.UploadTestcaseZip)
				problem.POST("/:id/rejudge", middleware.AuthMiddleware(), middleware.AdminMiddleware(), problemHandler.RejudgeProblem)
				problem.GET("/:id/testcases", middleware.AuthMiddleware(), middleware.AdminMiddleware(), problemHandler.GetTestcases)
				problem.DELETE("/:id/testcases", middleware.AuthMiddleware(), middleware.AdminMiddleware(), problemHandler.DeleteTestcases)
			}

		// 提交模块
		submission := v1.Group("/submission")
		submission.Use(middleware.AuthMiddleware())
		{
			submission.GET("/list", submissionHandler.List)
			submission.GET("/:id", submissionHandler.GetByID)
			submission.POST("", middleware.SubmitRateLimitMiddleware(), submissionHandler.Submit)
			submission.GET("/my", middleware.AuthMiddleware(), submissionHandler.GetMySubmissions)
		}

		// 排行榜
		v1.GET("/rank", userHandler.GetRankList)
		// 统计
		v1.GET("/statistics", statsHandler.GetPublic)

			// 比赛模块
			contest := v1.Group("/contest")
			contest.Use(middleware.AuthMiddleware())
			{
				contest.GET("/list", contestHandler.List)
				contest.GET("/:id", contestHandler.GetByID)
				contest.POST("/:id/start", contestHandler.StartContest)
			}

			// 管理模块
			admin := v1.Group("/admin")
			admin.Use(middleware.AuthMiddleware())
			{
				// 普通管理员与超级管理员都可访问（题目管理接口在 /problem 下）
				adminEditor := admin.Group("")
				adminEditor.Use(middleware.AdminMiddleware())
				{
					adminEditor.GET("/users", userHandler.GetUserList) // 比赛编辑时用于选择参赛用户
					adminEditor.POST("/users", userHandler.CreateUser)
					adminEditor.POST("/users/batch", userHandler.CreateUsersBatch)
					adminEditor.PUT("/users/:id", userHandler.UpdateUser)
					adminEditor.POST("/contests", contestHandler.Create)
					adminEditor.PUT("/contests/:id", contestHandler.Update)
					adminEditor.DELETE("/contests/:id", contestHandler.Delete)
					adminEditor.POST("/contests/:id/refresh", contestHandler.RefreshStats)
					adminEditor.GET("/contests/:id/leaderboard", contestHandler.GetLeaderboard)
					adminEditor.GET("/contests/:id/export", contestHandler.ExportLeaderboard)

					// 系统设置
					adminEditor.GET("/settings/ai", settingHandler.GetAISettings)
					adminEditor.PUT("/settings/ai", settingHandler.UpdateAISettings)
					adminEditor.POST("/settings/ai/test", settingHandler.TestAIConnection)
				}

				// 仅超级管理员可访问：管理员权限变更
				superAdmin := admin.Group("")
				superAdmin.Use(middleware.SuperAdminMiddleware())
				{
					superAdmin.PUT("/users/:id/role", userHandler.SetUserRole)
				}
			}
		}

	return r
}
