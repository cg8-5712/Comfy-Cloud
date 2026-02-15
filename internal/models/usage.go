package models

import (
	"time"

	"gorm.io/datatypes"
)

type UsageRecord struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	UserID     uint           `gorm:"not null;index" json:"user_id"`
	User       User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	TaskID     string         `gorm:"size:100" json:"task_id"`                // ComfyUI 任务 ID
	ModelName  string         `gorm:"size:100" json:"model_name"`             // 使用的模型（sd15/sdxl/flux）
	InstanceID string         `gorm:"size:50" json:"instance_id"`             // 使用的实例

	// 时间相关
	QueueTime     int       `json:"queue_time"`                               // 排队时长（秒）
	ExecutionTime int       `json:"execution_time"`                           // 执行时长（秒）
	TotalTime     int       `json:"total_time"`                               // 总时长（秒）
	StartedAt     time.Time `json:"started_at"`                               // 任务开始时间
	CompletedAt   time.Time `json:"completed_at"`                             // 任务完成时间

	// 资源使用
	GPUUsagePercent int     `json:"gpu_usage_percent"`                        // GPU 使用率（0-100）
	VRAMUsageGB     float64 `gorm:"type:decimal(10,2)" json:"vram_usage_gb"`  // 显存占用（GB）
	StorageUsageGB  float64 `gorm:"type:decimal(10,2)" json:"storage_usage_gb"` // 存储占用（GB）

	// 费用明细
	GPUCost      float64 `gorm:"type:decimal(10,4)" json:"gpu_cost"`         // GPU 费用
	VRAMCost     float64 `gorm:"type:decimal(10,4)" json:"vram_cost"`        // 显存费用
	StorageCost  float64 `gorm:"type:decimal(10,4)" json:"storage_cost"`     // 存储费用
	Subtotal     float64 `gorm:"type:decimal(10,4)" json:"subtotal"`         // 小计
	WaitDiscount float64 `gorm:"type:decimal(5,2)" json:"wait_discount"`     // 等待折扣系数（0.0-1.0）
	TotalCost    float64 `gorm:"type:decimal(10,4);not null" json:"total_cost"` // 总费用

	// 其他信息
	Resolution string         `gorm:"size:20" json:"resolution,omitempty"`    // 分辨率
	Metadata   datatypes.JSON `json:"metadata,omitempty"`                     // 额外信息
	CreatedAt  time.Time      `gorm:"index" json:"created_at"`
}

func (UsageRecord) TableName() string {
	return "usage_records"
}
