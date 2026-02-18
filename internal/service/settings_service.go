package service

import (
	"comfy-cloud/internal/models"
	"comfy-cloud/internal/repository"
	"encoding/json"
	"errors"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type SettingsService struct {
	settingsRepo *repository.SettingsRepository
	userRepo     *repository.UserRepository
	db           *gorm.DB
}

func NewSettingsService(settingsRepo *repository.SettingsRepository, userRepo *repository.UserRepository, db *gorm.DB) *SettingsService {
	return &SettingsService{
		settingsRepo: settingsRepo,
		userRepo:     userRepo,
		db:           db,
	}
}

// GetSettings 获取用户设置
func (s *SettingsService) GetSettings(userID uint) (map[string]interface{}, error) {
	settings, err := s.settingsRepo.FindByUserID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 返回默认设置
			return map[string]interface{}{
				"notifications": map[string]interface{}{
					"email_on_task_complete": true,
					"email_on_low_balance":   true,
					"low_balance_threshold":  10.0,
				},
				"preferences": map[string]interface{}{
					"language": "zh-CN",
					"timezone": "Asia/Shanghai",
				},
			}, nil
		}
		return nil, err
	}

	// 解析 JSON 字段
	var notifications map[string]interface{}
	var preferences map[string]interface{}

	if err := json.Unmarshal(settings.Notifications, &notifications); err != nil {
		notifications = make(map[string]interface{})
	}

	if err := json.Unmarshal(settings.Preferences, &preferences); err != nil {
		preferences = make(map[string]interface{})
	}

	return map[string]interface{}{
		"notifications": notifications,
		"preferences":   preferences,
	}, nil
}

// UpdateSettings 更新用户设置
func (s *SettingsService) UpdateSettings(userID uint, updates map[string]interface{}) error {
	// 获取现有设置
	settings, err := s.settingsRepo.FindByUserID(userID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	// 如果不存在，创建新设置
	if err == gorm.ErrRecordNotFound {
		settings = &models.UserSettings{
			UserID: userID,
		}
	}

	// 解析现有设置
	var notifications map[string]interface{}
	var preferences map[string]interface{}

	if settings.Notifications != nil {
		json.Unmarshal(settings.Notifications, &notifications)
	} else {
		notifications = make(map[string]interface{})
	}

	if settings.Preferences != nil {
		json.Unmarshal(settings.Preferences, &preferences)
	} else {
		preferences = make(map[string]interface{})
	}

	// 更新字段
	if notif, ok := updates["notifications"].(map[string]interface{}); ok {
		for k, v := range notif {
			notifications[k] = v
		}
	}

	if pref, ok := updates["preferences"].(map[string]interface{}); ok {
		for k, v := range pref {
			preferences[k] = v
		}
	}

	// 序列化回 JSON
	notifJSON, _ := json.Marshal(notifications)
	prefJSON, _ := json.Marshal(preferences)

	settings.Notifications = notifJSON
	settings.Preferences = prefJSON

	return s.settingsRepo.CreateOrUpdate(settings)
}

// ChangePassword 修改密码
func (s *SettingsService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return errors.New("invalid old password")
	}

	// 生成新密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hashedPassword)
	return s.userRepo.Update(user)
}
