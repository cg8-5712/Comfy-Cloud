package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type SystemLog struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Level     string         `gorm:"size:20;not null;index" json:"level"` // info/warn/error
	Source    string         `gorm:"size:50;not null;index" json:"source"` // auth/proxy/billing/system/admin
	Message   string         `gorm:"type:text;not null" json:"message"`
	UserID    *uint          `gorm:"index" json:"user_id,omitempty"`
	User      *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Details   datatypes.JSON `json:"details,omitempty"`
	CreatedAt time.Time      `gorm:"index" json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SystemLog) TableName() string {
	return "system_logs"
}
