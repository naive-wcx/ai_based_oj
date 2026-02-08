package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"oj-system/internal/config"
	"oj-system/internal/judge"
	"oj-system/internal/model"
	"oj-system/internal/repository"
	"oj-system/internal/router"
	"oj-system/internal/service"
	"oj-system/internal/utils"
)

func main() {
	// 命令行参数
	configPath := flag.String("config", "./configs/config.yaml", "配置文件路径")
	flag.Parse()

	// 加载配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	log.Printf("配置加载成功")

	// 创建必要的目录
	createDirectories(cfg)

	// 初始化 JWT
	utils.InitJWT(cfg.JWT.Secret)

	// 初始化数据库
	if err := repository.InitDatabase(&cfg.Database); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	log.Printf("数据库初始化成功")

	// 创建默认管理员账号
	createDefaultAdmin()

	// 启动判题服务
	judge.Start(cfg)

	// 启动定时任务（如赛后统计同步）
	startCronTasks()

	// 设置路由
	r := router.SetupRouter(cfg.Server.Mode)

	// 启动服务器
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("服务器启动在 http://localhost%s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}

// createDirectories 创建必要的目录
func createDirectories(cfg *config.Config) {
	dirs := []string{
		cfg.Paths.Problems,
		cfg.Paths.Submissions,
		"./data/sandbox",
		"./data/db",
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Printf("创建目录失败 %s: %v", dir, err)
		}
	}
}

// createDefaultAdmin 创建默认管理员账号
func createDefaultAdmin() {
	userRepo := repository.NewUserRepository()
	
	// 检查是否已存在 admin 用户
	if userRepo.ExistsByUsername("admin") {
		return
	}

	// 创建默认管理员
	hashedPassword, err := utils.HashPassword("admin123")
	if err != nil {
		log.Printf("创建默认管理员失败: %v", err)
		return
	}

	admin := &model.User{
		Username:     "admin",
		Email:        "admin@oj.local",
		PasswordHash: hashedPassword,
		Role:         "admin",
	}

	if err := userRepo.Create(admin); err != nil {
		log.Printf("创建默认管理员失败: %v", err)
		return
	}

	log.Printf("已创建默认管理员账号: admin / admin123")
}

// startCronTasks 启动定时任务
func startCronTasks() {
	go func() {
		contestService := service.NewContestService()
		maintenanceService := service.NewMaintenanceService()

		// 启动时立即执行一次全量同步，修复所有历史数据不一致问题
		log.Println("正在执行启动时全量数据修复...")
		maintenanceService.SyncAllStats()

		// 启动时立即检查一次已结束的比赛
		contestService.SyncEndedContests()

		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			// 同步已结束比赛的统计数据
			contestService.SyncEndedContests()
		}
	}()
	log.Printf("定时任务已启动")
}
