package service

import (
	"comfy-cloud/internal/models"
	"fmt"
	"strconv"
	"sync"
	"time"

	"gorm.io/gorm"
)

// ConfigService 配置管理服务（支持热加载）
type ConfigService struct {
	db     *gorm.DB
	cache  map[string]string
	mu     sync.RWMutex
	stopCh chan struct{}
}

// NewConfigService 创建配置服务
func NewConfigService(db *gorm.DB) *ConfigService {
	service := &ConfigService{
		db:     db,
		cache:  make(map[string]string),
		stopCh: make(chan struct{}),
	}

	// 初始加载配置
	service.Reload()

	return service
}

// Reload 重新加载所有配置
func (s *ConfigService) Reload() error {
	var configs []models.SystemConfig
	if err := s.db.Where("is_active = ?", true).Find(&configs).Error; err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 清空缓存
	s.cache = make(map[string]string)

	// 重新加载
	for _, cfg := range configs {
		key := fmt.Sprintf("%s.%s", cfg.Category, cfg.Key)
		s.cache[key] = cfg.Value
	}

	return nil
}

// StartAutoReload 启动自动重载（定期从数据库刷新配置）
func (s *ConfigService) StartAutoReload(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				s.Reload()
			case <-s.stopCh:
				ticker.Stop()
				return
			}
		}
	}()
}

// Stop 停止自动重载
func (s *ConfigService) Stop() {
	close(s.stopCh)
}

// Get 获取配置值（字符串）
func (s *ConfigService) Get(category, key string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	fullKey := fmt.Sprintf("%s.%s", category, key)
	return s.cache[fullKey]
}

// GetInt 获取配置值（整数）
func (s *ConfigService) GetInt(category, key string) int {
	value := s.Get(category, key)
	if value == "" {
		return 0
	}
	result, _ := strconv.Atoi(value)
	return result
}

// GetFloat 获取配置值（浮点数）
func (s *ConfigService) GetFloat(category, key string) float64 {
	value := s.Get(category, key)
	if value == "" {
		return 0
	}
	result, _ := strconv.ParseFloat(value, 64)
	return result
}

// GetBool 获取配置值（布尔）
func (s *ConfigService) GetBool(category, key string) bool {
	value := s.Get(category, key)
	return value == "true" || value == "1"
}

// Set 设置配置值
func (s *ConfigService) Set(category, key, value string) error {
	// 更新数据库
	var config models.SystemConfig
	err := s.db.Where("category = ? AND key = ?", category, key).First(&config).Error

	if err == gorm.ErrRecordNotFound {
		// 创建新配置
		config = models.SystemConfig{
			Category: category,
			Key:      key,
			Value:    value,
			IsActive: true,
		}
		if err := s.db.Create(&config).Error; err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		// 更新现有配置
		config.Value = value
		if err := s.db.Save(&config).Error; err != nil {
			return err
		}
	}

	// 更新缓存
	s.mu.Lock()
	fullKey := fmt.Sprintf("%s.%s", category, key)
	s.cache[fullKey] = value
	s.mu.Unlock()

	return nil
}

// GetAll 获取所有配置
func (s *ConfigService) GetAll() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]string)
	for k, v := range s.cache {
		result[k] = v
	}
	return result
}

// GetByCategory 获取指定分类的所有配置
func (s *ConfigService) GetByCategory(category string) map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]string)
	prefix := category + "."
	for k, v := range s.cache {
		if len(k) > len(prefix) && k[:len(prefix)] == prefix {
			result[k[len(prefix):]] = v
		}
	}
	return result
}
