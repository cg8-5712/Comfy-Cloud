package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type UserSettings struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	UserID        uint           `gorm:"uniqueIndex;not null" json:"user_id"`
	User          User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Notifications datatypes.JSON `json:"notifications"` // 通知设置
	Preferences   datatypes.JSON `json:"preferences"`   // 偏好设置
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (UserSettings) TableName() string {
	return "user_settings"
}
