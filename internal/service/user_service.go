package service

import (
	"comfy-cloud/internal/models"
	"comfy-cloud/internal/repository"
	"errors"
	"time"

	"gorm.io/gorm"
)

type UserService struct {
	userRepo  *repository.UserRepository
	usageRepo *repository.UsageRepository
	db        *gorm.DB
}

func NewUserService(userRepo *repository.UserRepository, usageRepo *repository.UsageRepository, db *gorm.DB) *UserService {
	return &UserService{
		userRepo:  userRepo,
		usageRepo: usageRepo,
		db:        db,
	}
}

// GetUserInfo 获取用户完整信息（包含订阅）
func (s *UserService) GetUserInfo(userID uint) (map[string]interface{}, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	// 查询订阅信息
	var subscription models.Subscription
	err = s.db.Where("user_id = ? AND status = ?", userID, "active").First(&subscription).Error

	var subscriptionData map[string]interface{}
	if err == nil {
		subscriptionData = map[string]interface{}{
			"tier":       subscription.Plan,
			"status":     subscription.Status,
			"started_at": subscription.StartedAt,
			"expires_at": subscription.ExpiresAt,
		}
	}

	// 查询存储使用情况（从配置或计算）
	storageUsed := int64(0)
	storageLimit := int64(107374182400) // 默认 100GB

	return map[string]interface{}{
		"id":            user.ID,
		"username":      user.Username,
		"email":         user.Email,
		"tier":          user.Tier,
		"balance":       user.Balance,
		"storage_used":  storageUsed,
		"storage_limit": storageLimit,
		"created_at":    user.CreatedAt,
		"subscription":  subscriptionData,
	}, nil
}

// GetUserBalance 获取用户余额
func (s *UserService) GetUserBalance(userID uint) (map[string]interface{}, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"balance":      user.Balance,
		"currency":     "CNY",
		"last_updated": time.Now(),
	}, nil
}

// GetUserUsage 获取用户使用统计
func (s *UserService) GetUserUsage(userID uint, period string) (map[string]interface{}, error) {
	var startDate, endDate time.Time
	now := time.Now()

	switch period {
	case "day":
		startDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		endDate = startDate.AddDate(0, 0, 1)
	case "week":
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		startDate = now.AddDate(0, 0, -weekday+1)
		startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)
		endDate = startDate.AddDate(0, 0, 7)
	case "month":
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		endDate = startDate.AddDate(0, 1, 0)
	case "year":
		startDate = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
		endDate = startDate.AddDate(1, 0, 0)
	default:
		return nil, errors.New("invalid period")
	}

	records, err := s.usageRepo.FindByUserAndDateRange(userID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// 统计数据
	var gpuSeconds int64
	var storageGBHours float64
	var totalCost float64
	taskCount := len(records)

	for _, record := range records {
		gpuSeconds += int64(record.ExecutionTime)
		storageGBHours += record.StorageUsageGB
		totalCost += record.TotalCost
	}

	return map[string]interface{}{
		"period":          period,
		"start_date":      startDate,
		"end_date":        endDate,
		"gpu_seconds":     gpuSeconds,
		"storage_gb_hours": storageGBHours,
		"total_cost":      totalCost,
		"task_count":      taskCount,
	}, nil
}

// UpdateUserBalance 更新用户余额
func (s *UserService) UpdateUserBalance(userID uint, amount float64) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}

	user.Balance += amount
	return s.userRepo.Update(user)
}

// DeductBalance 扣除余额
func (s *UserService) DeductBalance(userID uint, amount float64) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}

	if user.Balance < amount {
		return errors.New("insufficient balance")
	}

	user.Balance -= amount
	return s.userRepo.Update(user)
}
