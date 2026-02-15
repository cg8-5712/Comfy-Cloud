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
	if err := SeedSystemConfig(); err != nil {
		return err
	}
	if err := SeedComfyInstances(); err != nil {
		return err
	}
	return nil
}

// SeedSystemConfig 初始化系统配置
func SeedSystemConfig() error {
	var count int64
	DB.Model(&models.SystemConfig{}).Count(&count)
	if count > 0 {
		log.Println("System configs already exist, skipping seed")
		return nil
	}

	configs := []models.SystemConfig{
		// 负载均衡配置
		{Category: "loadbalancer", Key: "health_check_interval", Value: "10", ValueType: "int", Description: "健康检查间隔（秒）"},
		{Category: "loadbalancer", Key: "queue_update_interval", Value: "2", ValueType: "int", Description: "队列状态更新间隔（秒）"},
		{Category: "loadbalancer", Key: "max_queue_size", Value: "10", ValueType: "int", Description: "最大队列长度"},

		// 存储配置
		{Category: "storage", Key: "user_data_dir", Value: "./data/users", ValueType: "string", Description: "用户数据根目录"},
		{Category: "storage", Key: "shared_models_dir", Value: "./data/models", ValueType: "string", Description: "共享模型目录"},
		{Category: "storage", Key: "max_user_storage_gb", Value: "100", ValueType: "float", Description: "每个用户最大存储空间（GB）"},

		// 限流配置
		{Category: "rate_limit", Key: "enabled", Value: "true", ValueType: "bool", Description: "是否启用限流"},
		{Category: "rate_limit", Key: "requests_per_minute", Value: "60", ValueType: "int", Description: "每分钟最大请求数"},

		// 日志配置
		{Category: "logging", Key: "level", Value: "info", ValueType: "string", Description: "日志级别（debug/info/warn/error）"},
	}

	for _, cfg := range configs {
		if err := DB.Create(&cfg).Error; err != nil {
			return err
		}
	}

	log.Println("System configs created successfully")
	return nil
}

// SeedComfyInstances 初始化 ComfyUI 实例配置
func SeedComfyInstances() error {
	var count int64
	DB.Model(&models.ComfyInstance{}).Count(&count)
	if count > 0 {
		log.Println("ComfyUI instances already exist, skipping seed")
		return nil
	}

	instances := []models.ComfyInstance{
		{Name: "comfyui-1", URL: "http://localhost:8188", GPUID: 0, Pool: "shared", Status: "active", MaxQueue: 10},
		{Name: "comfyui-2", URL: "http://localhost:8189", GPUID: 1, Pool: "shared", Status: "active", MaxQueue: 10},
		{Name: "comfyui-3", URL: "http://localhost:8190", GPUID: 2, Pool: "shared", Status: "active", MaxQueue: 10},
		{Name: "comfyui-4", URL: "http://localhost:8191", GPUID: 3, Pool: "shared", Status: "active", MaxQueue: 10},
		{Name: "comfyui-5", URL: "http://localhost:8192", GPUID: 4, Pool: "shared", Status: "active", MaxQueue: 10},
		{Name: "comfyui-6", URL: "http://localhost:8193", GPUID: 5, Pool: "shared", Status: "active", MaxQueue: 10},
	}

	for _, inst := range instances {
		if err := DB.Create(&inst).Error; err != nil {
			return err
		}
	}

	log.Println("ComfyUI instances created successfully")
	return nil
}
