package models

import (
	"time"

	"gorm.io/gorm"
)

type RechargeRecord struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	UserID        uint           `gorm:"not null;index" json:"user_id"`
	User          User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Amount        float64        `gorm:"type:decimal(10,2);not null" json:"amount"`
	Currency      string         `gorm:"size:10;default:CNY" json:"currency"`
	PaymentMethod string         `gorm:"size:50" json:"payment_method"` // alipay/wechat/stripe
	Status        string         `gorm:"size:20;default:pending" json:"status"` // pending/completed/failed/refunded
	OrderID       string         `gorm:"size:100;uniqueIndex" json:"order_id,omitempty"`
	TransactionID string         `gorm:"size:100" json:"transaction_id,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	CompletedAt   *time.Time     `json:"completed_at,omitempty"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (RechargeRecord) TableName() string {
	return "recharge_records"
}
