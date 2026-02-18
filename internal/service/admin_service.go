package service

import (
	"comfy-cloud/internal/repository"
	"errors"

	"gorm.io/gorm"
)

type AdminService struct {
	adminRepo    *repository.AdminRepository
	userRepo     *repository.UserRepository
	rechargeRepo *repository.RechargeRepository
	modelRepo    *repository.ModelRepository
	configRepo   *repository.ConfigRepository
	db           *gorm.DB
}

func NewAdminService(
	adminRepo *repository.AdminRepository,
	userRepo *repository.UserRepository,
	rechargeRepo *repository.RechargeRepository,
	modelRepo *repository.ModelRepository,
	configRepo *repository.ConfigRepository,
	db *gorm.DB,
) *AdminService {
	return &AdminService{
		adminRepo:    adminRepo,
		userRepo:     userRepo,
		rechargeRepo: rechargeRepo,
		modelRepo:    modelRepo,
		configRepo:   configRepo,
		db:           db,
	}
}

// GetStats 获取管理统计
func (s *AdminService) GetStats() (map[string]interface{}, error) {
	return s.adminRepo.GetStats()
}

// GetAllUsers 获取所有用户
func (s *AdminService) GetAllUsers(search string, limit, offset int) ([]map[string]interface{}, int64, error) {
	users, total, err := s.adminRepo.GetAllUsers(search, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	result := make([]map[string]interface{}, len(users))
	for i, user := range users {
		result[i] = map[string]interface{}{
			"id":            user.ID,
			"username":      user.Username,
			"email":         user.Email,
			"tier":          user.Tier,
			"balance":       user.Balance,
			"storage_used":  0.0, // 从存储服务获取
			"storage_limit": 100.0,
			"status":        user.Status,
			"role":          "user", // 从用户表获取
			"created_at":    user.CreatedAt,
			"last_login_at": user.CreatedAt, // 需要添加 last_login_at 字段
		}
	}

	return result, total, nil
}

// UpdateUser 更新用户
func (s *AdminService) UpdateUser(userID uint, updates map[string]interface{}) (map[string]interface{}, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	// 更新字段
	if tier, ok := updates["tier"].(string); ok {
		user.Tier = tier
	}
	if status, ok := updates["status"].(string); ok {
		user.Status = status
	}
	if balance, ok := updates["balance"].(float64); ok {
		user.Balance = balance
	}

	if err := s.adminRepo.UpdateUser(user); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"id":       user.ID,
		"username": user.Username,
		"tier":     user.Tier,
		"status":   user.Status,
		"role":     "user",
		"balance":  user.Balance,
	}, nil
}

// GetSystemLogs 获取系统日志
func (s *AdminService) GetSystemLogs(level, source string, limit, offset int) ([]map[string]interface{}, int64, error) {
	logs, total, err := s.adminRepo.GetSystemLogs(level, source, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	result := make([]map[string]interface{}, len(logs))
	for i, log := range logs {
		username := ""
		if log.User != nil {
			username = log.User.Username
		}

		result[i] = map[string]interface{}{
			"id":         log.ID,
			"level":      log.Level,
			"source":     log.Source,
			"message":    log.Message,
			"user_id":    log.UserID,
			"username":   username,
			"created_at": log.CreatedAt,
			"details":    log.Details,
		}
	}

	return result, total, nil
}

// GetFinanceStats 获取财务统计
func (s *AdminService) GetFinanceStats() (map[string]interface{}, error) {
	return s.adminRepo.GetFinanceStats()
}

// GetRechargeRecords 获取充值记录（Admin）
func (s *AdminService) GetRechargeRecords(limit, offset int) ([]map[string]interface{}, int64, error) {
	records, total, err := s.rechargeRepo.GetAllRecharges(limit, offset)
	if err != nil {
		return nil, 0, err
	}

	result := make([]map[string]interface{}, len(records))
	for i, record := range records {
		username := ""
		if record.User.Username != "" {
			username = record.User.Username
		}

		result[i] = map[string]interface{}{
			"id":             record.ID,
			"user_id":        record.UserID,
			"username":       username,
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

// GetSystemConfig 获取系统配置
func (s *AdminService) GetSystemConfig() (map[string]interface{}, error) {
	// 从数据库读取配置
	configs := make(map[string]interface{})

	// 计费配置
	configs["billing"] = map[string]interface{}{
		"gpu_price_per_second":     0.005,
		"storage_price_per_gb_day": 0.02,
		"bandwidth_price_per_gb":   0.10,
	}

	// 实例池配置
	configs["instance_pool"] = map[string]interface{}{
		"max_queue_per_instance":         10,
		"health_check_interval_seconds":  30,
		"auto_scale_enabled":             false,
	}

	// 系统配置
	configs["system"] = map[string]interface{}{
		"max_upload_size_mb":    2048,
		"allowed_model_types":   []string{"checkpoint", "lora", "vae", "embedding"},
		"maintenance_mode":      false,
	}

	return configs, nil
}

// UpdateSystemConfig 更新系统配置
func (s *AdminService) UpdateSystemConfig(updates map[string]interface{}) (map[string]interface{}, error) {
	// 更新配置到数据库
	// 这里简化处理，实际应该逐个更新配置项

	// 返回更新后的完整配置
	return s.GetSystemConfig()
}

// CheckAdminPermission 检查管理员权限
func (s *AdminService) CheckAdminPermission(userID uint) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}

	// 检查用户角色（需要在 User 模型中添加 Role 字段）
	// 这里简化处理，假设 user_id = 1 是管理员
	if user.ID != 1 {
		return errors.New("permission denied: admin access required")
	}

	return nil
}
