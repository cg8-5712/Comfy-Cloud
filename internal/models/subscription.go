package models

import (
	"time"

	"gorm.io/gorm"
)

type Subscription struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	UserID       uint           `gorm:"not null;index" json:"user_id"`
	User         User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Plan         string         `gorm:"size:20;not null" json:"plan"` // basic/pro/enterprise/dedicated
	ResourceMode string         `gorm:"size:20;default:queue" json:"resource_mode"` // queue/dedicated
	Subdomain    *string        `gorm:"size:50;uniqueIndex" json:"subdomain,omitempty"` // 独占模式的子域名
	DedicatedGPUs int           `gorm:"default:0" json:"dedicated_gpus"` // 独占 GPU 数量
	Status       string         `gorm:"size:20;default:active" json:"status"` // active/cancelled/expired
	StartedAt    time.Time      `gorm:"not null" json:"started_at"`
	ExpiresAt    *time.Time     `json:"expires_at,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Subscription) TableName() string {
	return "subscriptions"
}
