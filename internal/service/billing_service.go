package service

import (
	"comfy-cloud/internal/models"
	"comfy-cloud/internal/repository"
	"errors"
	"math"
	"time"

	"gorm.io/gorm"
)

var (
	ErrInsufficientBalance = errors.New("insufficient balance")
)

// BillingConfig 计费配置
type BillingConfig struct {
	// GPU 使用（按分钟）
	GPUBase       float64 // 基础价格 ¥0.05/分钟
	GPUPerPercent float64 // 每 1% 使用率额外 ¥0.001/分钟

	// 显存使用（按分钟）
	VRAMPerGB float64 // ¥0.01/GB/分钟

	// 存储使用（按天）
	StoragePerGB float64 // ¥0.01/GB/天

	// 等待时间折扣
	WaitThreshold int     // 等待超过 60 秒开始折扣
	WaitRate      float64 // 每多等待 10 秒，折扣 1%
}

// DefaultBillingConfig 默认计费配置
var DefaultBillingConfig = BillingConfig{
	GPUBase:       0.05,
	GPUPerPercent: 0.001,
	VRAMPerGB:     0.01,
	StoragePerGB:  0.01,
	WaitThreshold: 60,
	WaitRate:      0.01,
}

type BillingService struct {
	usageRepo *repository.UsageRepository
	userRepo  *repository.UserRepository
	db        *gorm.DB
}

func NewBillingService(usageRepo *repository.UsageRepository, userRepo *repository.UserRepository, db *gorm.DB) *BillingService {
	return &BillingService{
		usageRepo: usageRepo,
		userRepo:  userRepo,
		db:        db,
	}
}

// GetActiveConfig 获取当前激活的计费配置
func (s *BillingService) GetActiveConfig() (*models.BillingConfig, error) {
	var config models.BillingConfig
	err := s.db.Where("is_active = ?", true).First(&config).Error
	if err != nil {
		// 如果没有找到，返回默认配置
		return &models.BillingConfig{
			GPUBase:       0.05,
			GPUPerPercent: 0.001,
			VRAMPerGB:     0.01,
			StoragePerGB:  0.01,
			WaitThreshold: 60,
			WaitRate:      0.01,
		}, nil
	}
	return &config, nil
}

// CalculateCost 计算任务费用
// 公式：总费用 = (GPU费用 + 显存费用 + 存储费用) × 等待时间折扣系数
func (s *BillingService) CalculateCost(
	gpuUsagePercent int,
	vramUsageGB float64,
	storageUsageGB float64,
	executionSeconds int,
	queueSeconds int,
) (gpuCost, vramCost, storageCost, subtotal, waitDiscount, totalCost float64) {
	// 获取当前激活的计费配置
	config, err := s.GetActiveConfig()
	if err != nil {
		// 使用默认配置
		config = &models.BillingConfig{
			GPUBase:       0.05,
			GPUPerPercent: 0.001,
			VRAMPerGB:     0.01,
			StoragePerGB:  0.01,
			WaitThreshold: 60,
			WaitRate:      0.01,
		}
	}

	// 转换为分钟
	executionMinutes := float64(executionSeconds) / 60.0

	// 1. GPU 费用 = (基础价格 + GPU使用率 × 单位价格) × 使用时长
	gpuCost = (config.GPUBase + float64(gpuUsagePercent)*config.GPUPerPercent) * executionMinutes

	// 2. 显存费用 = 显存占用(GB) × 使用时长 × 显存单价
	vramCost = config.VRAMPerGB * vramUsageGB * executionMinutes

	// 3. 存储费用（暂时设为 0，因为按天计费需要单独处理）
	storageCost = 0.0

	// 4. 小计
	subtotal = gpuCost + vramCost + storageCost

	// 5. 等待时间折扣
	// 等待折扣 = 1.0 - (max(0, 等待秒数 - 阈值) / 10) × 折扣率
	if queueSeconds > config.WaitThreshold {
		discountAmount := float64(queueSeconds-config.WaitThreshold) / 10.0 * config.WaitRate
		waitDiscount = 1.0 - discountAmount
		// 确保折扣不会小于 0
		if waitDiscount < 0 {
			waitDiscount = 0
		}
	} else {
		waitDiscount = 1.0 // 无折扣
	}

	// 6. 总费用
	totalCost = subtotal * waitDiscount

	return
}

// PreChargeEstimate 预估最大费用（用于预扣费）
func (s *BillingService) PreChargeEstimate(modelName string) float64 {
	// 获取当前激活的计费配置
	config, err := s.GetActiveConfig()
	if err != nil {
		config = &models.BillingConfig{
			GPUBase:          0.05,
			GPUPerPercent:    0.001,
			VRAMPerGB:        0.01,
			PreChargeSD15:    2,
			PreChargeSDXL:    5,
			PreChargeFlux:    10,
			PreChargeDefault: 5,
		}
	}

	// 预估最大执行时间（分钟）
	var estimatedMinutes float64
	switch modelName {
	case "sd15":
		estimatedMinutes = float64(config.PreChargeSD15)
	case "sdxl":
		estimatedMinutes = float64(config.PreChargeSDXL)
	case "flux":
		estimatedMinutes = float64(config.PreChargeFlux)
	default:
		estimatedMinutes = float64(config.PreChargeDefault)
	}

	// 按最大 GPU 使用率和显存占用预估
	maxGPUUsage := 100
	var maxVRAM float64
	switch modelName {
	case "sd15":
		maxVRAM = 6.0
	case "sdxl":
		maxVRAM = 12.0
	case "flux":
		maxVRAM = 20.0
	default:
		maxVRAM = 12.0
	}

	gpuCost := (config.GPUBase + float64(maxGPUUsage)*config.GPUPerPercent) * estimatedMinutes
	vramCost := config.VRAMPerGB * maxVRAM * estimatedMinutes

	return gpuCost + vramCost
}

// FreezeBalance 预扣费（冻结余额）
func (s *BillingService) FreezeBalance(userID uint, amount float64) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}

	// 检查余额是否充足
	if user.Balance < amount {
		return ErrInsufficientBalance
	}

	// 冻结余额
	user.Balance -= amount
	user.FrozenBalance += amount

	return s.userRepo.Update(user)
}

// UnfreezeAndCharge 解冻并实际扣费
func (s *BillingService) UnfreezeAndCharge(userID uint, frozenAmount, actualCost float64) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}

	// 解冻
	user.FrozenBalance -= frozenAmount

	// 如果实际费用小于预扣金额，退还差额
	if actualCost < frozenAmount {
		user.Balance += (frozenAmount - actualCost)
	} else if actualCost > frozenAmount {
		// 如果实际费用大于预扣金额，补扣差额
		diff := actualCost - frozenAmount
		if user.Balance < diff {
			return ErrInsufficientBalance
		}
		user.Balance -= diff
	}

	return s.userRepo.Update(user)
}

// RecordUsage 记录使用并扣费
func (s *BillingService) RecordUsage(record *models.UsageRecord) error {
	// 计算费用
	gpuCost, vramCost, storageCost, subtotal, waitDiscount, totalCost := s.CalculateCost(
		record.GPUUsagePercent,
		record.VRAMUsageGB,
		record.StorageUsageGB,
		record.ExecutionTime,
		record.QueueTime,
	)

	// 填充费用字段
	record.GPUCost = gpuCost
	record.VRAMCost = vramCost
	record.StorageCost = storageCost
	record.Subtotal = subtotal
	record.WaitDiscount = waitDiscount
	record.TotalCost = totalCost
	record.TotalTime = record.QueueTime + record.ExecutionTime
	record.CreatedAt = time.Now()

	// 保存记录
	return s.usageRepo.Create(record)
}

// GetUserUsageStats 获取用户使用统计
func (s *BillingService) GetUserUsageStats(userID uint, startDate, endDate time.Time) ([]models.UsageRecord, float64, error) {
	records, err := s.usageRepo.FindByUserAndDateRange(userID, startDate, endDate)
	if err != nil {
		return nil, 0, err
	}

	var totalCost float64
	for _, record := range records {
		totalCost += record.TotalCost
	}

	return records, totalCost, nil
}

// CalculateStorageCost 计算存储费用（按天）
func (s *BillingService) CalculateStorageCost(storageGB float64, days int) float64 {
	return s.config.StoragePerGB * storageGB * float64(days)
}

// 四舍五入到指定小数位
func round(val float64, precision int) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}
