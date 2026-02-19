package service

import (
	"comfy-cloud/internal/models"
	"comfy-cloud/internal/repository"
	"time"

	"gorm.io/gorm"
)

type UsageService struct {
	usageRepo *repository.UsageRepository
	db        *gorm.DB
}

func NewUsageService(usageRepo *repository.UsageRepository, db *gorm.DB) *UsageService {
	return &UsageService{
		usageRepo: usageRepo,
		db:        db,
	}
}

// GetUsageRecords 获取使用记录列表
func (s *UsageService) GetUsageRecords(userID uint, limit, offset int, startDate, endDate string) ([]map[string]interface{}, int64, error) {
	var records []models.UsageRecord
	var total int64
	var err error

	query := s.db.Model(&models.UsageRecord{}).Where("user_id = ?", userID)

	// 日期范围过滤
	if startDate != "" {
		if start, err := time.Parse("2006-01-02", startDate); err == nil {
			query = query.Where("started_at >= ?", start)
		}
	}
	if endDate != "" {
		if end, err := time.Parse("2006-01-02", endDate); err == nil {
			query = query.Where("started_at <= ?", end.Add(24*time.Hour))
		}
	}

	// 获取总数
	if err = query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取记录
	err = query.Order("started_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&records).Error

	if err != nil {
		return nil, 0, err
	}

	// 转换为前端格式
	result := make([]map[string]interface{}, len(records))
	for i, record := range records {
		recordType := "gpu_usage"
		if record.StorageCost > 0 {
			recordType = "storage"
		}

		result[i] = map[string]interface{}{
			"id":               record.ID,
			"task_id":          record.TaskID,
			"type":             recordType,
			"started_at":       record.StartedAt,
			"ended_at":         record.CompletedAt,
			"duration_seconds": record.ExecutionTime,
			"cost":             record.TotalCost,
			"details": map[string]interface{}{
				"gpu_type":   "RTX 4090",
				"model":      record.ModelName,
				"vram_used":  record.VRAMUsageGB,
				"resolution": record.Resolution,
			},
		}
	}

	return result, total, nil
}

// GetUsageStats 获取使用统计
func (s *UsageService) GetUsageStats(userID uint, period string) (map[string]interface{}, error) {
	var startDate time.Time
	now := time.Now()

	switch period {
	case "day":
		startDate = now.AddDate(0, 0, -1)
	case "week":
		startDate = now.AddDate(0, 0, -7)
	case "month":
		startDate = now.AddDate(0, -1, 0)
	case "year":
		startDate = now.AddDate(-1, 0, 0)
	default:
		startDate = now.AddDate(0, -1, 0)
	}

	var totalCost float64
	var totalSeconds float64
	var taskCount int64

	s.db.Model(&models.UsageRecord{}).
		Where("user_id = ? AND started_at >= ?", userID, startDate).
		Select("COALESCE(SUM(total_cost), 0)").Scan(&totalCost)

	s.db.Model(&models.UsageRecord{}).
		Where("user_id = ? AND started_at >= ?", userID, startDate).
		Select("COALESCE(SUM(execution_time), 0)").Scan(&totalSeconds)

	s.db.Model(&models.UsageRecord{}).
		Where("user_id = ? AND started_at >= ?", userID, startDate).
		Count(&taskCount)

	return map[string]interface{}{
		"period":           period,
		"start_date":       startDate.Format("2006-01-02"),
		"end_date":         now.Format("2006-01-02"),
		"gpu_seconds":      totalSeconds,
		"storage_gb_hours": 0,
		"total_cost":       totalCost,
		"task_count":       taskCount,
	}, nil
}
