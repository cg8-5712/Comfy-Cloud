package repository

import (
	"comfy-cloud/internal/models"

	"gorm.io/gorm"
)

type SettingsRepository struct {
	db *gorm.DB
}

func NewSettingsRepository(db *gorm.DB) *SettingsRepository {
	return &SettingsRepository{db: db}
}

// FindByUserID 根据用户 ID 查找设置
func (r *SettingsRepository) FindByUserID(userID uint) (*models.UserSettings, error) {
	var settings models.UserSettings
	err := r.db.Where("user_id = ?", userID).First(&settings).Error
	if err != nil {
		return nil, err
	}
	return &settings, nil
}

// Create 创建设置
func (r *SettingsRepository) Create(settings *models.UserSettings) error {
	return r.db.Create(settings).Error
}

// Update 更新设置
func (r *SettingsRepository) Update(settings *models.UserSettings) error {
	return r.db.Save(settings).Error
}

// CreateOrUpdate 创建或更新设置
func (r *SettingsRepository) CreateOrUpdate(settings *models.UserSettings) error {
	var existing models.UserSettings
	err := r.db.Where("user_id = ?", settings.UserID).First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		return r.Create(settings)
	}

	if err != nil {
		return err
	}

	settings.ID = existing.ID
	return r.Update(settings)
}
