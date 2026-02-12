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
			
			// 管理员操作
			problem.POST("", middleware.AuthMiddleware(), middleware.AdminMiddleware(), problemHandler.Create)
			problem.PUT("/:id", middleware.AuthMiddleware(), middleware.AdminMiddleware(), problemHandler.Update)
			problem.DELETE("/:id", middleware.AuthMiddleware(), middleware.AdminMiddleware(), problemHandler.Delete)
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

		// 管理员模块
		admin := v1.Group("/admin")
		admin.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
		{
			admin.GET("/users", userHandler.GetUserList)
			admin.POST("/users", userHandler.CreateUser)
			admin.POST("/users/batch", userHandler.CreateUsersBatch)
			admin.PUT("/users/:id", userHandler.UpdateUser)
			admin.PUT("/users/:id/role", userHandler.SetUserRole)
			admin.POST("/contests", contestHandler.Create)
			admin.PUT("/contests/:id", contestHandler.Update)
			admin.DELETE("/contests/:id", contestHandler.Delete)
			admin.POST("/contests/:id/refresh", contestHandler.RefreshStats)
			admin.GET("/contests/:id/leaderboard", contestHandler.GetLeaderboard)
			admin.GET("/contests/:id/export", contestHandler.ExportLeaderboard)
			
			// 系统设置
			admin.GET("/settings/ai", settingHandler.GetAISettings)
			admin.PUT("/settings/ai", settingHandler.UpdateAISettings)
			admin.POST("/settings/ai/test", settingHandler.TestAIConnection)
		}
	}

	return r
}
