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

	// API v1
	v1 := r.Group("/api/v1")
	{
		// 用户模块
		user := v1.Group("/user")
		{
			user.POST("/register", userHandler.Register)
			user.POST("/login", userHandler.Login)
			user.GET("/profile", middleware.AuthMiddleware(), userHandler.GetProfile)
			user.PUT("/profile", middleware.AuthMiddleware(), userHandler.UpdateProfile)
		}

		// 题目模块
		problem := v1.Group("/problem")
		{
			problem.GET("/list", problemHandler.List)
			problem.GET("/:id", problemHandler.GetByID)
			
			// 管理员操作
			problem.POST("", middleware.AuthMiddleware(), middleware.AdminMiddleware(), problemHandler.Create)
			problem.PUT("/:id", middleware.AuthMiddleware(), middleware.AdminMiddleware(), problemHandler.Update)
			problem.DELETE("/:id", middleware.AuthMiddleware(), middleware.AdminMiddleware(), problemHandler.Delete)
			problem.POST("/:id/testcase", middleware.AuthMiddleware(), middleware.AdminMiddleware(), problemHandler.UploadTestcase)
			problem.GET("/:id/testcases", middleware.AuthMiddleware(), middleware.AdminMiddleware(), problemHandler.GetTestcases)
			problem.DELETE("/:id/testcases", middleware.AuthMiddleware(), middleware.AdminMiddleware(), problemHandler.DeleteTestcases)
		}

		// 提交模块
		submission := v1.Group("/submission")
		{
			submission.GET("/list", submissionHandler.List)
			submission.GET("/:id", middleware.OptionalAuthMiddleware(), submissionHandler.GetByID)
			submission.POST("", middleware.AuthMiddleware(), middleware.SubmitRateLimitMiddleware(), submissionHandler.Submit)
			submission.GET("/my", middleware.AuthMiddleware(), submissionHandler.GetMySubmissions)
		}

		// 排行榜
		v1.GET("/rank", userHandler.GetRankList)

		// 管理员模块
		admin := v1.Group("/admin")
		admin.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
		{
			admin.GET("/users", userHandler.GetUserList)
			admin.PUT("/users/:id/role", userHandler.SetUserRole)
			
			// 系统设置
			admin.GET("/settings/ai", settingHandler.GetAISettings)
			admin.PUT("/settings/ai", settingHandler.UpdateAISettings)
			admin.POST("/settings/ai/test", settingHandler.TestAIConnection)
		}
	}

	return r
}
