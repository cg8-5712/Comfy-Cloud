package models

import (
	"time"

	"gorm.io/gorm"
)

type PrivateModel struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	UserID           uint           `gorm:"not null;index" json:"user_id"`
	User             User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Name             string         `gorm:"size:255;not null" json:"name"`
	Type             string         `gorm:"size:50;not null" json:"type"` // checkpoint/lora/vae/embedding
	SizeBytes        int64          `gorm:"not null" json:"size_bytes"`
	FilePath         string         `gorm:"size:500;not null" json:"file_path"`
	Visibility       string         `gorm:"size:20;default:private" json:"visibility"` // base/vip/private
	Status           string         `gorm:"size:20;default:active" json:"status"` // active/pending/disabled
	StorageCostPerDay float64       `gorm:"type:decimal(10,4)" json:"storage_cost_per_day"`
	UploadedAt       time.Time      `json:"uploaded_at"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

func (PrivateModel) TableName() string {
	return "private_models"
}
