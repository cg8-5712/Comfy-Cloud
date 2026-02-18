package service

import (
	"comfy-cloud/internal/models"
	"comfy-cloud/internal/repository"
	"errors"
	"time"

	"gorm.io/gorm"
)

type SubscriptionService struct {
	subscriptionRepo *repository.SubscriptionRepository
	userRepo         *repository.UserRepository
	db               *gorm.DB
}

func NewSubscriptionService(subscriptionRepo *repository.SubscriptionRepository, userRepo *repository.UserRepository, db *gorm.DB) *SubscriptionService {
	return &SubscriptionService{
		subscriptionRepo: subscriptionRepo,
		userRepo:         userRepo,
		db:               db,
	}
}

// GetTiers 获取所有订阅等级配置
func (s *SubscriptionService) GetTiers() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"key":      "basic",
			"label":    "基础版",
			"color":    "bg-muted text-muted-foreground",
			"price":    "免费",
			"features": []string{"每月 100 次任务", "5 GB 存储空间", "基础模型访问", "社区支持"},
			"popular":  false,
		},
		{
			"key":      "pro",
			"label":    "专业版",
			"color":    "bg-primary/10 text-primary",
			"price":    "¥99/月",
			"features": []string{"无限任务", "50 GB 存储空间", "VIP 模型访问", "优先队列", "邮件支持"},
			"popular":  true,
		},
		{
			"key":      "enterprise",
			"label":    "企业版",
			"color":    "bg-amber-500/10 text-amber-600",
			"price":    "¥299/月",
			"features": []string{"无限任务", "200 GB 存储空间", "全部模型访问", "最高优先级", "专属支持", "团队协作"},
			"popular":  false,
		},
	}
}

// GetSubscription 获取用户订阅信息
func (s *SubscriptionService) GetSubscription(userID uint) (map[string]interface{}, error) {
	subscription, err := s.subscriptionRepo.FindByUserID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 用户没有订阅，返回基础版信息
			return map[string]interface{}{
				"tier":       "basic",
				"status":     "active",
				"started_at": time.Now(),
				"expires_at": nil,
				"auto_renew": false,
			}, nil
		}
		return nil, err
	}

	return map[string]interface{}{
		"tier":       subscription.Plan,
		"status":     subscription.Status,
		"started_at": subscription.StartedAt,
		"expires_at": subscription.ExpiresAt,
		"auto_renew": true, // 可以从数据库字段读取
	}, nil
}

// UpgradeSubscription 升级订阅
func (s *SubscriptionService) UpgradeSubscription(userID uint, targetTier string) (map[string]interface{}, error) {
	// 验证目标等级
	validTiers := map[string]bool{
		"basic":      true,
		"pro":        true,
		"enterprise": true,
	}

	if !validTiers[targetTier] {
		return nil, errors.New("invalid target tier")
	}

	// 获取用户
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	// 更新用户等级
	user.Tier = targetTier
	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	// 查找或创建订阅记录
	subscription, err := s.subscriptionRepo.FindByUserID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 创建新订阅
			expiresAt := time.Now().AddDate(0, 1, 0) // 1个月后过期
			subscription = &models.Subscription{
				UserID:    userID,
				Plan:      targetTier,
				Status:    "active",
				StartedAt: time.Now(),
				ExpiresAt: &expiresAt,
			}
			if err := s.subscriptionRepo.Create(subscription); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		// 更新现有订阅
		subscription.Plan = targetTier
		subscription.Status = "active"
		expiresAt := time.Now().AddDate(0, 1, 0)
		subscription.ExpiresAt = &expiresAt
		if err := s.subscriptionRepo.Update(subscription); err != nil {
			return nil, err
		}
	}

	return map[string]interface{}{
		"tier":       subscription.Plan,
		"status":     subscription.Status,
		"started_at": subscription.StartedAt,
		"expires_at": subscription.ExpiresAt,
	}, nil
}
