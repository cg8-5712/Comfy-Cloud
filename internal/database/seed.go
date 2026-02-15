package database

import (
	"comfy-cloud/internal/models"
	"log"
)

// SeedBillingConfig 初始化默认计费配置
func SeedBillingConfig() error {
	// 检查是否已存在默认配置
	var count int64
	DB.Model(&models.BillingConfig{}).Where("name = ?", "default").Count(&count)
	if count > 0 {
		log.Println("Default billing config already exists, skipping seed")
		return nil
	}

	// 创建默认计费配置
	defaultConfig := models.BillingConfig{
		Name:      "default",
		IsActive:  true,

		// GPU 使用（按分钟）
		GPUBase:       0.05,   // ¥0.05/分钟
		GPUPerPercent: 0.001,  // 每 1% 使用率额外 ¥0.001/分钟

		// 显存使用（按分钟）
		VRAMPerGB: 0.01,       // ¥0.01/GB/分钟

		// 存储使用（按天）
		StoragePerGB: 0.01,    // ¥0.01/GB/天

		// 等待时间折扣
		WaitThreshold: 60,     // 等待超过 60 秒开始折扣
		WaitRate:      0.01,   // 每多等待 10 秒，折扣 1%

		// 预扣费配置
		PreChargeSD15:    2,   // SD 1.5 预估 2 分钟
		PreChargeSDXL:    5,   // SDXL 预估 5 分钟
		PreChargeFlux:    10,  // Flux 预估 10 分钟
		PreChargeDefault: 5,   // 默认预估 5 分钟
	}

	if err := DB.Create(&defaultConfig).Error; err != nil {
		return err
	}

	log.Println("Default billing config created successfully")
	return nil
}

// SeedDedicatedPricing 初始化独占模式定价
func SeedDedicatedPricing() error {
	// 检查是否已存在定价
	var count int64
	DB.Model(&models.DedicatedPricing{}).Count(&count)
	if count > 0 {
		log.Println("Dedicated pricing already exists, skipping seed")
		return nil
	}

	// 创建独占模式定价
	pricings := []models.DedicatedPricing{
		{
			GPUCount:     1,
			HourlyPrice:  10.00,
			DailyPrice:   200.00,
			MonthlyPrice: 5000.00,
			IsActive:     true,
		},
		{
			GPUCount:     2,
			HourlyPrice:  18.00,
			DailyPrice:   360.00,
			MonthlyPrice: 9000.00,
			IsActive:     true,
		},
		{
			GPUCount:     4,
			HourlyPrice:  32.00,
			DailyPrice:   640.00,
			MonthlyPrice: 16000.00,
			IsActive:     true,
		},
	}

	for _, pricing := range pricings {
		if err := DB.Create(&pricing).Error; err != nil {
			return err
		}
	}

	log.Println("Dedicated pricing created successfully")
	return nil
}

// SeedAll 初始化所有默认数据
func SeedAll() error {
	if err := SeedBillingConfig(); err != nil {
		return err
	}
	if err := SeedDedicatedPricing(); err != nil {
		return err
	}
	return nil
}
