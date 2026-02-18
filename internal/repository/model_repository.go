package repository

import (
	"comfy-cloud/internal/models"

	"gorm.io/gorm"
)

type ModelRepository struct {
	db *gorm.DB
}

func NewModelRepository(db *gorm.DB) *ModelRepository {
	return &ModelRepository{db: db}
}

// Create 创建模型记录
func (r *ModelRepository) Create(model *models.PrivateModel) error {
	return r.db.Create(model).Error
}

// FindByID 根据 ID 查找模型
func (r *ModelRepository) FindByID(id uint) (*models.PrivateModel, error) {
	var model models.PrivateModel
	err := r.db.Preload("User").First(&model, id).Error
	return &model, err
}

// FindByUserID 根据用户 ID 查找私有模型
func (r *ModelRepository) FindByUserID(userID uint) ([]models.PrivateModel, error) {
	var models []models.PrivateModel
	err := r.db.Where("user_id = ? AND visibility = ?", userID, "private").
		Order("uploaded_at DESC").
		Find(&models).Error
	return models, err
}

// FindAll 查找所有模型（Admin）
func (r *ModelRepository) FindAll(visibility string, limit, offset int) ([]models.PrivateModel, int64, error) {
	var models []models.PrivateModel
	var total int64

	query := r.db.Model(&models)
	if visibility != "" {
		query = query.Where("visibility = ?", visibility)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取记录
	err := query.Preload("User").
		Order("uploaded_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error

	return models, total, err
}

// Update 更新模型
func (r *ModelRepository) Update(model *models.PrivateModel) error {
	return r.db.Save(model).Error
}

// Delete 删除模型
func (r *ModelRepository) Delete(id uint) error {
	return r.db.Delete(&models.PrivateModel{}, id).Error
}

// GetTotalStorageByUser 获取用户的总存储使用量
func (r *ModelRepository) GetTotalStorageByUser(userID uint) (int64, error) {
	var total int64
	err := r.db.Model(&models.PrivateModel{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(size_bytes), 0)").
		Scan(&total).Error
	return total, err
}
