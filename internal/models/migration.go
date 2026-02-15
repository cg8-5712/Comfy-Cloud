package models

import "time"

// Migration 数据库迁移版本表
type Migration struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Version     string    `gorm:"size:50;not null;uniqueIndex" json:"version"` // 版本号（如 v1, v2）
	Description string    `gorm:"type:text" json:"description"`                // 迁移描述
	AppliedAt   time.Time `gorm:"not null" json:"applied_at"`                  // 应用时间
}

func (Migration) TableName() string {
	return "migrations"
}
