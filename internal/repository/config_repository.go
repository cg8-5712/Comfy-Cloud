package repository

import (
	"comfy-cloud/internal/models"

	"gorm.io/gorm"
)

type ConfigRepository struct {
	db *gorm.DB
}

func NewConfigRepository(db *gorm.DB) *ConfigRepository {
	return &ConfigRepository{db: db}
}

// Get 获取配置项
func (r *ConfigRepository) Get(category, key string) (*models.SystemConfig, error) {
	var config models.SystemConfig
	err := r.db.Where("category = ? AND key = ?", category, key).First(&config).Error
	return &config, err
}

// GetByCategory 获取某个分类的所有配置
func (r *ConfigRepository) GetByCategory(category string) ([]models.SystemConfig, error) {
	var configs []models.SystemConfig
	err := r.db.Where("category = ?", category).Find(&configs).Error
	return configs, err
}

// GetAll 获取所有配置
func (r *ConfigRepository) GetAll() ([]models.SystemConfig, error) {
	var configs []models.SystemConfig
	err := r.db.Find(&configs).Error
	return configs, err
}

// Set 设置配置项
func (r *ConfigRepository) Set(config *models.SystemConfig) error {
	var existing models.SystemConfig
	err := r.db.Where("category = ? AND key = ?", config.Category, config.Key).First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		return r.db.Create(config).Error
	}

	if err != nil {
		return err
	}

	config.ID = existing.ID
	return r.db.Save(config).Error
}

// Delete 删除配置项
func (r *ConfigRepository) Delete(category, key string) error {
	return r.db.Where("category = ? AND key = ?", category, key).Delete(&models.SystemConfig{}).Error
}
