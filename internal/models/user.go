package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Username     string         `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Email        string         `gorm:"uniqueIndex;size:100;not null" json:"email"`
	PasswordHash string         `gorm:"size:255;not null" json:"-"`
	Role          string         `gorm:"size:20;default:user" json:"role"` // admin/user
	Tier          string         `gorm:"size:20;default:basic" json:"tier"` // basic/pro/enterprise
	Balance       float64        `gorm:"type:decimal(10,2);default:0.00" json:"balance"`
	FrozenBalance float64        `gorm:"type:decimal(10,2);default:0.00" json:"frozen_balance"` // 预扣费冻结余额
	Status        string         `gorm:"size:20;default:active" json:"status"` // active/suspended/deleted
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (User) TableName() string {
	return "users"
}
