package models

import (
	"time"

	"gorm.io/gorm"
)

// BillingConfig 计费配置表
type BillingConfig struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"size:50;not null;uniqueIndex" json:"name"` // 配置名称（如 "default", "vip"）
	IsActive  bool           `gorm:"default:true" json:"is_active"`            // 是否启用

	// GPU 使用（按分钟）
	GPUBase       float64 `gorm:"type:decimal(10,4);not null" json:"gpu_base"`        // 基础价格 ¥0.05/分钟
	GPUPerPercent float64 `gorm:"type:decimal(10,6);not null" json:"gpu_per_percent"` // 每 1% 使用率额外 ¥0.001/分钟

	// 显存使用（按分钟）
	VRAMPerGB float64 `gorm:"type:decimal(10,4);not null" json:"vram_per_gb"` // ¥0.01/GB/分钟

	// 存储使用（按天）
	StoragePerGB float64 `gorm:"type:decimal(10,4);not null" json:"storage_per_gb"` // ¥0.01/GB/天

	// 等待时间折扣
	WaitThreshold int     `gorm:"not null" json:"wait_threshold"`           // 等待超过 N 秒开始折扣（默认 60）
	WaitRate      float64 `gorm:"type:decimal(5,4);not null" json:"wait_rate"` // 每多等待 10 秒，折扣 N%（默认 0.01 = 1%）

	// 预扣费配置（预估最大执行时间，分钟）
	PreChargeSD15 int `gorm:"default:2" json:"pre_charge_sd15"`   // SD 1.5 预估 2 分钟
	PreChargeSDXL int `gorm:"default:5" json:"pre_charge_sdxl"`   // SDXL 预估 5 分钟
	PreChargeFlux int `gorm:"default:10" json:"pre_charge_flux"`  // Flux 预估 10 分钟
	PreChargeDefault int `gorm:"default:5" json:"pre_charge_default"` // 默认预估 5 分钟

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (BillingConfig) TableName() string {
	return "billing_configs"
}

// DedicatedPricing 独占模式定价表
type DedicatedPricing struct {
	ID       uint           `gorm:"primaryKey" json:"id"`
	GPUCount int            `gorm:"not null;uniqueIndex" json:"gpu_count"` // GPU 数量（1/2/4）

	// 定价（元）
	HourlyPrice  float64 `gorm:"type:decimal(10,2);not null" json:"hourly_price"`  // 每小时价格
	DailyPrice   float64 `gorm:"type:decimal(10,2);not null" json:"daily_price"`   // 每天价格
	MonthlyPrice float64 `gorm:"type:decimal(10,2);not null" json:"monthly_price"` // 每月价格

	IsActive  bool           `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (DedicatedPricing) TableName() string {
	return "dedicated_pricings"
}
