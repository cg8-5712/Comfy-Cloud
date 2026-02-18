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
