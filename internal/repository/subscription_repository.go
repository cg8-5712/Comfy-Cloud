package repository

import (
	"comfy-cloud/internal/models"
	"time"

	"gorm.io/gorm"
)

type SubscriptionRepository struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

// Create 创建订阅
func (r *SubscriptionRepository) Create(subscription *models.Subscription) error {
	return r.db.Create(subscription).Error
}

// FindByUserID 根据用户 ID 查找活跃订阅
func (r *SubscriptionRepository) FindByUserID(userID uint) (*models.Subscription, error) {
	var subscription models.Subscription
	err := r.db.Where("user_id = ? AND status = ?", userID, "active").First(&subscription).Error
	if err != nil {
		return nil, err
	}
	return &subscription, nil
}

// FindByID 根据 ID 查找订阅
func (r *SubscriptionRepository) FindByID(id uint) (*models.Subscription, error) {
	var subscription models.Subscription
	err := r.db.First(&subscription, id).Error
	return &subscription, err
}

// Update 更新订阅
func (r *SubscriptionRepository) Update(subscription *models.Subscription) error {
	return r.db.Save(subscription).Error
}

// CancelSubscription 取消订阅
func (r *SubscriptionRepository) CancelSubscription(userID uint) error {
	return r.db.Model(&models.Subscription{}).
		Where("user_id = ? AND status = ?", userID, "active").
		Update("status", "cancelled").Error
}

// CheckExpiredSubscriptions 检查过期订阅
func (r *SubscriptionRepository) CheckExpiredSubscriptions() error {
	now := time.Now()
	return r.db.Model(&models.Subscription{}).
		Where("status = ? AND expires_at < ?", "active", now).
		Update("status", "expired").Error
}
