package service

import (
	"log"
	"oj-system/internal/model"
	"oj-system/internal/repository"
)

type MaintenanceService struct {
	userRepo    *repository.UserRepository
	problemRepo *repository.ProblemRepository
}

func NewMaintenanceService() *MaintenanceService {
	return &MaintenanceService{
		userRepo:    repository.NewUserRepository(),
		problemRepo: repository.NewProblemRepository(),
	}
}

// SyncAllStats 强制全量同步所有用户和题目的统计数据
func (s *MaintenanceService) SyncAllStats() {
	log.Println("[Maintenance] 开始全量同步统计数据...")

	// 1. 同步所有题目
	var problems []model.Problem
	// 使用 repository 的 DB 实例直接查询，避免循环依赖（虽然这里没有）
	// 这里为了简单直接用 GORM
	db := repository.GetDB()
	if err := db.Find(&problems).Error; err != nil {
		log.Printf("[Maintenance] 获取题目列表失败: %v", err)
	} else {
		for _, p := range problems {
			if err := s.problemRepo.SyncStats(p.ID); err != nil {
				log.Printf("[Maintenance] 同步题目 %d 失败: %v", p.ID, err)
			}
		}
		log.Printf("[Maintenance] 已同步 %d 个题目的统计数据", len(problems))
	}

	// 2. 同步所有用户
	var users []model.User
	if err := db.Find(&users).Error; err != nil {
		log.Printf("[Maintenance] 获取用户列表失败: %v", err)
	} else {
		for _, u := range users {
			if err := s.userRepo.SyncStats(u.ID); err != nil {
				log.Printf("[Maintenance] 同步用户 %d 失败: %v", u.ID, err)
			}
		}
		log.Printf("[Maintenance] 已同步 %d 个用户的统计数据", len(users))
	}

	log.Println("[Maintenance] 全量同步完成")
}
