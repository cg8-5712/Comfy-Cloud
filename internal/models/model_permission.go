package models

import (
	"time"
)

type ModelPermission struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index;uniqueIndex:idx_user_model" json:"user_id"`
	User      User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	ModelPath string    `gorm:"size:255;not null;uniqueIndex:idx_user_model" json:"model_path"`
	ModelName string    `gorm:"size:100;not null" json:"model_name"`
	ModelType string    `gorm:"size:50" json:"model_type"` // checkpoint/lora/vae/embedding
	FileSize  int64     `json:"file_size,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

func (ModelPermission) TableName() string {
	return "model_permissions"
}
