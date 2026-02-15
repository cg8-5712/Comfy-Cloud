package repository

import (
	"comfy-cloud/internal/models"
	"time"

	"gorm.io/gorm"
)

type UsageRepository struct {
	db *gorm.DB
}

func NewUsageRepository(db *gorm.DB) *UsageRepository {
	return &UsageRepository{db: db}
}

// Create 创建使用记录
func (r *UsageRepository) Create(record *models.UsageRecord) error {
	return r.db.Create(record).Error
}

// FindByID 根据 ID 查询
func (r *UsageRepository) FindByID(id uint) (*models.UsageRecord, error) {
	var record models.UsageRecord
	err := r.db.Preload("User").First(&record, id).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// FindByUserID 查询用户的所有使用记录
func (r *UsageRepository) FindByUserID(userID uint, limit, offset int) ([]models.UsageRecord, error) {
	var records []models.UsageRecord
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&records).Error
	return records, err
}

// FindByUserAndDateRange 查询用户在指定日期范围内的使用记录
func (r *UsageRepository) FindByUserAndDateRange(userID uint, startDate, endDate time.Time) ([]models.UsageRecord, error) {
	var records []models.UsageRecord
	err := r.db.Where("user_id = ? AND created_at BETWEEN ? AND ?", userID, startDate, endDate).
		Order("created_at DESC").
		Find(&records).Error
	return records, err
}

// FindByTaskID 根据任务 ID 查询
func (r *UsageRepository) FindByTaskID(taskID string) (*models.UsageRecord, error) {
	var record models.UsageRecord
	err := r.db.Where("task_id = ?", taskID).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// Update 更新使用记录
func (r *UsageRepository) Update(record *models.UsageRecord) error {
	return r.db.Save(record).Error
}

// GetTotalCostByUser 获取用户的总消费
func (r *UsageRepository) GetTotalCostByUser(userID uint) (float64, error) {
	var totalCost float64
	err := r.db.Model(&models.UsageRecord{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(total_cost), 0)").
		Scan(&totalCost).Error
	return totalCost, err
}

// GetMonthlyStats 获取用户的月度统计
func (r *UsageRepository) GetMonthlyStats(userID uint, year, month int) (map[string]interface{}, error) {
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0)

	var stats struct {
		TotalCost     float64
		TotalTasks    int64
		TotalTime     int64
		AvgQueueTime  float64
		AvgExecTime   float64
	}

	err := r.db.Model(&models.UsageRecord{}).
		Where("user_id = ? AND created_at BETWEEN ? AND ?", userID, startDate, endDate).
		Select(`
			COALESCE(SUM(total_cost), 0) as total_cost,
			COUNT(*) as total_tasks,
			COALESCE(SUM(total_time), 0) as total_time,
			COALESCE(AVG(queue_time), 0) as avg_queue_time,
			COALESCE(AVG(execution_time), 0) as avg_exec_time
		`).
		Scan(&stats).Error

	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_cost":      stats.TotalCost,
		"total_tasks":     stats.TotalTasks,
		"total_time":      stats.TotalTime,
		"avg_queue_time":  stats.AvgQueueTime,
		"avg_exec_time":   stats.AvgExecTime,
	}, nil
}
