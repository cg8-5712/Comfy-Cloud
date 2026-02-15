package models

import (
	"time"

	"gorm.io/gorm"
)

// DedicatedInstance 独占实例分配表
type DedicatedInstance struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	UserID         uint           `gorm:"not null;index" json:"user_id"`
	User           User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	SubscriptionID uint           `gorm:"not null" json:"subscription_id"`
	Subscription   Subscription   `gorm:"foreignKey:SubscriptionID" json:"subscription,omitempty"`
	Subdomain      string         `gorm:"size:50;uniqueIndex;not null" json:"subdomain"` // 子域名
	InstanceIDs    string         `gorm:"type:text" json:"instance_ids"`                 // 分配的实例 ID（逗号分隔）
	GPUIDs         string         `gorm:"type:text" json:"gpu_ids"`                      // 分配的 GPU ID（逗号分隔）
	Status         string         `gorm:"size:20;default:active" json:"status"`          // active/suspended
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (DedicatedInstance) TableName() string {
	return "dedicated_instances"
}
