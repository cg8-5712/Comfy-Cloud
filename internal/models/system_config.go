package models

import (
	"time"

	"gorm.io/gorm"
)

// SystemConfig 系统配置表（支持热加载）
type SystemConfig struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Category    string         `gorm:"size:50;not null;index" json:"category"` // 配置分类：loadbalancer/storage/rate_limit/logging
	Key         string         `gorm:"size:100;not null;uniqueIndex:idx_category_key" json:"key"`
	Value       string         `gorm:"type:text;not null" json:"value"`
	ValueType   string         `gorm:"size:20;not null" json:"value_type"` // string/int/float/bool/json
	Description string         `gorm:"type:text" json:"description"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SystemConfig) TableName() string {
	return "system_configs"
}

// ComfyInstance ComfyUI 实例配置表
type ComfyInstance struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"size:50;not null;uniqueIndex" json:"name"` // 实例名称（如 comfyui-1）
	URL         string         `gorm:"size:255;not null" json:"url"`             // 实例 URL
	GPUID       int            `gorm:"not null" json:"gpu_id"`                   // GPU ID
	Pool        string         `gorm:"size:20;default:shared" json:"pool"`       // shared/dedicated
	Status      string         `gorm:"size:20;default:active" json:"status"`     // active/inactive/maintenance
	Priority    int            `gorm:"default:0" json:"priority"`                // 优先级（数字越大优先级越高）
	MaxQueue    int            `gorm:"default:10" json:"max_queue"`              // 最大队列长度
	Description string         `gorm:"type:text" json:"description,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ComfyInstance) TableName() string {
	return "comfy_instances"
}
