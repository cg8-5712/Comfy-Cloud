package repository

import (
	"comfy-cloud/internal/models"

	"gorm.io/gorm"
)

type RechargeRepository struct {
	db *gorm.DB
}

func NewRechargeRepository(db *gorm.DB) *RechargeRepository {
	return &RechargeRepository{db: db}
}

// Create 创建充值记录
func (r *RechargeRepository) Create(record *models.RechargeRecord) error {
	return r.db.Create(record).Error
}

// FindByID 根据 ID 查找充值记录
func (r *RechargeRepository) FindByID(id uint) (*models.RechargeRecord, error) {
	var record models.RechargeRecord
	err := r.db.First(&record, id).Error
	return &record, err
}

// FindByUserID 根据用户 ID 查找充值记录
func (r *RechargeRepository) FindByUserID(userID uint, limit, offset int) ([]models.RechargeRecord, int64, error) {
	var records []models.RechargeRecord
	var total int64

	// 获取总数
	if err := r.db.Model(&models.RechargeRecord{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取记录
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&records).Error

	return records, total, err
}

// FindByOrderID 根据订单 ID 查找充值记录
func (r *RechargeRepository) FindByOrderID(orderID string) (*models.RechargeRecord, error) {
	var record models.RechargeRecord
	err := r.db.Where("order_id = ?", orderID).First(&record).Error
	return &record, err
}

// Update 更新充值记录
func (r *RechargeRepository) Update(record *models.RechargeRecord) error {
	return r.db.Save(record).Error
}

// GetTotalRechargeByUser 获取用户总充值金额
func (r *RechargeRepository) GetTotalRechargeByUser(userID uint) (float64, error) {
	var total float64
	err := r.db.Model(&models.RechargeRecord{}).
		Where("user_id = ? AND status = ?", userID, "completed").
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error
	return total, err
}

// GetAllRecharges 获取所有充值记录（Admin）
func (r *RechargeRepository) GetAllRecharges(limit, offset int) ([]models.RechargeRecord, int64, error) {
	var records []models.RechargeRecord
	var total int64

	// 获取总数
	if err := r.db.Model(&models.RechargeRecord{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取记录（预加载用户信息）
	err := r.db.Preload("User").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&records).Error

	return records, total, err
}
