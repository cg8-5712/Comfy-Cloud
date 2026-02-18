package service

import (
	"comfy-cloud/internal/models"
	"comfy-cloud/internal/repository"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type RechargeService struct {
	rechargeRepo *repository.RechargeRepository
	userRepo     *repository.UserRepository
	db           *gorm.DB
}

func NewRechargeService(rechargeRepo *repository.RechargeRepository, userRepo *repository.UserRepository, db *gorm.DB) *RechargeService {
	return &RechargeService{
		rechargeRepo: rechargeRepo,
		userRepo:     userRepo,
		db:           db,
	}
}

// CreateRecharge 创建充值订单
func (s *RechargeService) CreateRecharge(userID uint, amount float64, paymentMethod string) (map[string]interface{}, error) {
	// 创建充值记录
	record := &models.RechargeRecord{
		UserID:        userID,
		Amount:        amount,
		Currency:      "CNY",
		PaymentMethod: paymentMethod,
		Status:        "pending",
	}

	if err := s.rechargeRepo.Create(record); err != nil {
		return nil, err
	}

	// 生成订单 ID
	orderID := fmt.Sprintf("ord_%d_%d", userID, record.ID)

	// 模拟支付 URL（实际应该调用支付网关）
	paymentURL := fmt.Sprintf("https://payment.example.com/checkout/%s", orderID)

	return map[string]interface{}{
		"order_id":    orderID,
		"amount":      amount,
		"currency":    "CNY",
		"payment_url": paymentURL,
		"status":      "pending",
		"created_at":  record.CreatedAt,
	}, nil
}

// GetRechargeHistory 获取充值记录
func (s *RechargeService) GetRechargeHistory(userID uint, limit, offset int) ([]map[string]interface{}, int64, error) {
	records, total, err := s.rechargeRepo.FindByUserID(userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	result := make([]map[string]interface{}, len(records))
	for i, record := range records {
		result[i] = map[string]interface{}{
			"id":             record.ID,
			"amount":         record.Amount,
			"currency":       record.Currency,
			"payment_method": record.PaymentMethod,
			"status":         record.Status,
			"created_at":     record.CreatedAt,
			"completed_at":   record.CompletedAt,
		}
	}

	return result, total, nil
}

// CompleteRecharge 完成充值（由支付回调触发）
func (s *RechargeService) CompleteRecharge(recordID uint) error {
	record, err := s.rechargeRepo.FindByID(recordID)
	if err != nil {
		return err
	}

	if record.Status != "pending" {
		return fmt.Errorf("recharge already processed")
	}

	// 开始事务
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 更新充值记录状态
		now := time.Now()
		record.Status = "completed"
		record.CompletedAt = &now
		if err := tx.Save(record).Error; err != nil {
			return err
		}

		// 增加用户余额
		user, err := s.userRepo.FindByID(record.UserID)
		if err != nil {
			return err
		}

		user.Balance += record.Amount
		if err := tx.Save(user).Error; err != nil {
			return err
		}

		return nil
	})
}
