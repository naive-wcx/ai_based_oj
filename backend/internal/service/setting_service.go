package service

import (
	"strconv"
	"sync"

	"oj-system/internal/model"
	"oj-system/internal/repository"
)

type SettingService struct {
	repo  *repository.SettingRepository
	cache map[string]string
	mu    sync.RWMutex
}

var settingService *SettingService
var settingOnce sync.Once

// GetSettingService 获取设置服务单例
func GetSettingService() *SettingService {
	settingOnce.Do(func() {
		settingService = &SettingService{
			repo:  repository.NewSettingRepository(),
			cache: make(map[string]string),
		}
		// 初始化时加载所有设置到缓存
		settingService.loadCache()
	})
	return settingService
}

// loadCache 加载设置到缓存
func (s *SettingService) loadCache() {
	settings, err := s.repo.GetAll()
	if err != nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache = settings
}

// Get 获取设置值
func (s *SettingService) Get(key string) string {
	s.mu.RLock()
	if value, ok := s.cache[key]; ok {
		s.mu.RUnlock()
		return value
	}
	s.mu.RUnlock()
	
	// 缓存未命中，从数据库获取
	value, err := s.repo.Get(key)
	if err != nil {
		return ""
	}
	
	s.mu.Lock()
	s.cache[key] = value
	s.mu.Unlock()
	
	return value
}

// Set 设置值
func (s *SettingService) Set(key, value string) error {
	if err := s.repo.Set(key, value); err != nil {
		return err
	}
	
	s.mu.Lock()
	s.cache[key] = value
	s.mu.Unlock()
	
	return nil
}

// GetAISettings 获取 AI 设置
func (s *SettingService) GetAISettings() *model.AISettings {
	settings := &model.AISettings{
		Enabled:  s.Get(model.SettingAIEnabled) == "true",
		Provider: s.Get(model.SettingAIProvider),
		APIKey:   s.Get(model.SettingAIAPIKey),
		APIURL:   s.Get(model.SettingAIAPIURL),
		Model:    s.Get(model.SettingAIModel),
	}
	
	if timeout, err := strconv.Atoi(s.Get(model.SettingAITimeout)); err == nil {
		settings.Timeout = timeout
	} else {
		settings.Timeout = 60
	}
	
	// 设置默认值
	if settings.Provider == "" {
		settings.Provider = "deepseek"
	}
	if settings.APIURL == "" {
		settings.APIURL = "https://api.deepseek.com/v1/chat/completions"
	}
	if settings.Model == "" {
		settings.Model = "deepseek-chat"
	}
	
	return settings
}

// UpdateAISettings 更新 AI 设置
func (s *SettingService) UpdateAISettings(req *model.UpdateAISettingsRequest) error {
	// 保存各项设置
	if err := s.Set(model.SettingAIEnabled, strconv.FormatBool(req.Enabled)); err != nil {
		return err
	}
	if err := s.Set(model.SettingAIProvider, req.Provider); err != nil {
		return err
	}
	if req.APIKey != "" && req.APIKey != "********" {
		if err := s.Set(model.SettingAIAPIKey, req.APIKey); err != nil {
			return err
		}
	}
	if err := s.Set(model.SettingAIAPIURL, req.APIURL); err != nil {
		return err
	}
	if err := s.Set(model.SettingAIModel, req.Model); err != nil {
		return err
	}
	if err := s.Set(model.SettingAITimeout, strconv.Itoa(req.Timeout)); err != nil {
		return err
	}
	
	return nil
}

// GetAISettingsForDisplay 获取用于显示的 AI 设置（隐藏 API Key）
func (s *SettingService) GetAISettingsForDisplay() *model.AISettings {
	settings := s.GetAISettings()
	if settings.APIKey != "" {
		settings.APIKey = "********"
	}
	return settings
}

// HasAIAPIKey 检查是否配置了 API Key
func (s *SettingService) HasAIAPIKey() bool {
	return s.Get(model.SettingAIAPIKey) != ""
}
