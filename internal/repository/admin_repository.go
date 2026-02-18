package repository

import (
	"comfy-cloud/internal/models"

	"gorm.io/gorm"
)

type AdminRepository struct {
	db *gorm.DB
}

func NewAdminRepository(db *gorm.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

// GetAllUsers 获取所有用户（Admin）
func (r *AdminRepository) GetAllUsers(search string, limit, offset int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := r.db.Model(&models.User{})

	// 搜索过滤
	if search != "" {
		query = query.Where("username LIKE ? OR email LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取用户列表
	err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&users).Error

	return users, total, err
}

// UpdateUser 更新用户（Admin）
func (r *AdminRepository) UpdateUser(user *models.User) error {
	return r.db.Save(user).Error
}

// GetSystemLogs 获取系统日志
func (r *AdminRepository) GetSystemLogs(level, source string, limit, offset int) ([]models.SystemLog, int64, error) {
	var logs []models.SystemLog
	var total int64

	query := r.db.Model(&models.SystemLog{})

	// 过滤条件
	if level != "" {
		query = query.Where("level = ?", level)
	}
	if source != "" {
		query = query.Where("source = ?", source)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取日志（预加载用户信息）
	err := query.Preload("User").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error

	return logs, total, err
}

// CreateLog 创建日志
func (r *AdminRepository) CreateLog(log *models.SystemLog) error {
	return r.db.Create(log).Error
}

// GetStats 获取管理统计
func (r *AdminRepository) GetStats() (map[string]interface{}, error) {
	var totalUsers int64
	var activeToday int64
	var tasksToday int64
	var totalRevenue float64

	// 总用户数
	r.db.Model(&models.User{}).Count(&totalUsers)

	// 今日活跃用户（简化：有使用记录的用户）
	r.db.Model(&models.UsageRecord{}).
		Where("DATE(started_at) = CURRENT_DATE").
		Distinct("user_id").
		Count(&activeToday)

	// 今日任务数
	r.db.Model(&models.UsageRecord{}).
		Where("DATE(started_at) = CURRENT_DATE").
		Count(&tasksToday)

	// 总收入
	r.db.Model(&models.RechargeRecord{}).
		Where("status = ?", "completed").
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalRevenue)

	return map[string]interface{}{
		"total_users":         totalUsers,
		"active_users_today":  activeToday,
		"total_tasks_today":   tasksToday,
		"total_revenue":       totalRevenue,
		"instances_online":    3, // 从实例池获取
		"instances_total":     3,
		"avg_queue_length":    2.3,
		"gpu_utilization_avg": 68.5,
	}, nil
}

// GetFinanceStats 获取财务统计
func (r *AdminRepository) GetFinanceStats() (map[string]interface{}, error) {
	var totalRevenue float64
	var revenueToday float64
	var revenueThisWeek float64
	var revenueThisMonth float64
	var totalRecharges int64
	var avgRechargeAmount float64

	// 总收入
	r.db.Model(&models.RechargeRecord{}).
		Where("status = ?", "completed").
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalRevenue)

	// 今日收入
	r.db.Model(&models.RechargeRecord{}).
		Where("status = ? AND DATE(created_at) = CURRENT_DATE", "completed").
		Select("COALESCE(SUM(amount), 0)").
		Scan(&revenueToday)

	// 本周收入
	r.db.Model(&models.RechargeRecord{}).
		Where("status = ? AND created_at >= DATE_TRUNC('week', CURRENT_DATE)", "completed").
		Select("COALESCE(SUM(amount), 0)").
		Scan(&revenueThisWeek)

	// 本月收入
	r.db.Model(&models.RechargeRecord{}).
		Where("status = ? AND created_at >= DATE_TRUNC('month', CURRENT_DATE)", "completed").
		Select("COALESCE(SUM(amount), 0)").
		Scan(&revenueThisMonth)

	// 充值笔数
	r.db.Model(&models.RechargeRecord{}).
		Where("status = ?", "completed").
		Count(&totalRecharges)

	// 平均充值金额
	if totalRecharges > 0 {
		avgRechargeAmount = totalRevenue / float64(totalRecharges)
	}

	return map[string]interface{}{
		"total_revenue":        totalRevenue,
		"revenue_today":        revenueToday,
		"revenue_this_week":    revenueThisWeek,
		"revenue_this_month":   revenueThisMonth,
		"total_recharges":      totalRecharges,
		"avg_recharge_amount":  avgRechargeAmount,
	}, nil
}
