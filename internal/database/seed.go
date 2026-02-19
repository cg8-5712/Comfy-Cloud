package database

import (
	"comfy-cloud/internal/models"
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/datatypes"
)

// SeedAll 初始化所有默认数据，仅在 dev 环境下执行
func SeedAll(env string) error {
	if env != "dev" {
		log.Println("Skipping database seed (env is not dev)")
		return nil
	}

	log.Println("Running database seed (dev mode)...")

	if err := SeedUsers(); err != nil {
		return err
	}
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
	if err := SeedSubscriptions(); err != nil {
		return err
	}
	if err := SeedUsageRecords(); err != nil {
		return err
	}
	if err := SeedRechargeRecords(); err != nil {
		return err
	}
	if err := SeedPrivateModels(); err != nil {
		return err
	}
	if err := SeedModelPermissions(); err != nil {
		return err
	}
	if err := SeedUserSettings(); err != nil {
		return err
	}
	if err := SeedSystemLogs(); err != nil {
		return err
	}
	if err := SeedDedicatedInstances(); err != nil {
		return err
	}
	return nil
}
// SeedUsers 创建测试用户
func SeedUsers() error {
	var count int64
	DB.Model(&models.User{}).Count(&count)
	if count > 0 {
		log.Println("Users already exist, skipping seed")
		return nil
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	users := []models.User{
		{Username: "admin", Email: "admin@comfy-cloud.com", PasswordHash: string(hashedPassword), Role: "admin", Tier: "enterprise", Balance: 9999.00, Status: "active"},
		{Username: "testuser", Email: "test@comfy-cloud.com", PasswordHash: string(hashedPassword), Role: "user", Tier: "basic", Balance: 100.00, Status: "active"},
		{Username: "prouser", Email: "pro@comfy-cloud.com", PasswordHash: string(hashedPassword), Role: "user", Tier: "pro", Balance: 500.00, Status: "active"},
		{Username: "vipuser", Email: "vip@comfy-cloud.com", PasswordHash: string(hashedPassword), Role: "user", Tier: "enterprise", Balance: 2000.00, Status: "active"},
		{Username: "suspended", Email: "suspended@comfy-cloud.com", PasswordHash: string(hashedPassword), Role: "user", Tier: "basic", Balance: 0, Status: "suspended"},
	}

	for _, u := range users {
		if err := DB.Create(&u).Error; err != nil {
			return err
		}
	}

	log.Println("Seed users created (password: admin123)")
	return nil
}
// SeedBillingConfig 初始化默认计费配置
func SeedBillingConfig() error {
	var count int64
	DB.Model(&models.BillingConfig{}).Where("name = ?", "default").Count(&count)
	if count > 0 {
		log.Println("Default billing config already exists, skipping seed")
		return nil
	}

	defaultConfig := models.BillingConfig{
		Name: "default", IsActive: true,
		GPUBase: 0.05, GPUPerPercent: 0.001,
		VRAMPerGB: 0.01, StoragePerGB: 0.01,
		WaitThreshold: 60, WaitRate: 0.01,
		PreChargeSD15: 2, PreChargeSDXL: 5, PreChargeFlux: 10, PreChargeDefault: 5,
	}

	if err := DB.Create(&defaultConfig).Error; err != nil {
		return err
	}

	log.Println("Default billing config created successfully")
	return nil
}

// SeedDedicatedPricing 初始化独占模式定价
func SeedDedicatedPricing() error {
	var count int64
	DB.Model(&models.DedicatedPricing{}).Count(&count)
	if count > 0 {
		log.Println("Dedicated pricing already exists, skipping seed")
		return nil
	}

	pricings := []models.DedicatedPricing{
		{GPUCount: 1, HourlyPrice: 10.00, DailyPrice: 200.00, MonthlyPrice: 5000.00, IsActive: true},
		{GPUCount: 2, HourlyPrice: 18.00, DailyPrice: 360.00, MonthlyPrice: 9000.00, IsActive: true},
		{GPUCount: 4, HourlyPrice: 32.00, DailyPrice: 640.00, MonthlyPrice: 16000.00, IsActive: true},
	}

	for _, p := range pricings {
		if err := DB.Create(&p).Error; err != nil {
			return err
		}
	}

	log.Println("Dedicated pricing created successfully")
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
		{Category: "loadbalancer", Key: "health_check_interval", Value: "10", ValueType: "int", Description: "健康检查间隔（秒）"},
		{Category: "loadbalancer", Key: "queue_update_interval", Value: "2", ValueType: "int", Description: "队列状态更新间隔（秒）"},
		{Category: "loadbalancer", Key: "max_queue_size", Value: "10", ValueType: "int", Description: "最大队列长度"},
		{Category: "storage", Key: "user_data_dir", Value: "./data/users", ValueType: "string", Description: "用户数据根目录"},
		{Category: "storage", Key: "shared_models_dir", Value: "./data/models", ValueType: "string", Description: "共享模型目录"},
		{Category: "storage", Key: "max_user_storage_gb", Value: "100", ValueType: "float", Description: "每个用户最大存储空间（GB）"},
		{Category: "rate_limit", Key: "enabled", Value: "true", ValueType: "bool", Description: "是否启用限流"},
		{Category: "rate_limit", Key: "requests_per_minute", Value: "60", ValueType: "int", Description: "每分钟最大请求数"},
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
// SeedSubscriptions 初始化订阅数据
func SeedSubscriptions() error {
	var count int64
	DB.Model(&models.Subscription{}).Count(&count)
	if count > 0 {
		log.Println("Subscriptions already exist, skipping seed")
		return nil
	}

	now := time.Now()
	expires := now.AddDate(0, 1, 0)
	vipSubdomain := "vip1"

	subs := []models.Subscription{
		{UserID: 1, Plan: "enterprise", ResourceMode: "queue", Status: "active", StartedAt: now, ExpiresAt: &expires},
		{UserID: 2, Plan: "basic", ResourceMode: "queue", Status: "active", StartedAt: now, ExpiresAt: &expires},
		{UserID: 3, Plan: "pro", ResourceMode: "queue", Status: "active", StartedAt: now, ExpiresAt: &expires},
		{UserID: 4, Plan: "enterprise", ResourceMode: "dedicated", DedicatedGPUs: 1, Subdomain: &vipSubdomain, Status: "active", StartedAt: now, ExpiresAt: &expires},
	}

	for _, s := range subs {
		if err := DB.Create(&s).Error; err != nil {
			return err
		}
	}

	log.Println("Subscriptions created successfully")
	return nil
}

// SeedUsageRecords 初始化使用记录
func SeedUsageRecords() error {
	var count int64
	DB.Model(&models.UsageRecord{}).Count(&count)
	if count > 0 {
		log.Println("Usage records already exist, skipping seed")
		return nil
	}

	now := time.Now()
	records := []models.UsageRecord{
		{
			UserID: 2, TaskID: "task-001", ModelName: "sd_v1.5", InstanceID: "comfyui-1",
			QueueTime: 5, ExecutionTime: 30, TotalTime: 35,
			StartedAt: now.Add(-1 * time.Hour), CompletedAt: now.Add(-1*time.Hour + 35*time.Second),
			GPUUsagePercent: 85, VRAMUsageGB: 4.5,
			GPUCost: 0.025, VRAMCost: 0.005, StorageCost: 0, Subtotal: 0.03, TotalCost: 0.03,
			Resolution: "512x512",
		},
		{
			UserID: 3, TaskID: "task-002", ModelName: "sdxl_base", InstanceID: "comfyui-2",
			QueueTime: 10, ExecutionTime: 120, TotalTime: 130,
			StartedAt: now.Add(-30 * time.Minute), CompletedAt: now.Add(-30*time.Minute + 130*time.Second),
			GPUUsagePercent: 95, VRAMUsageGB: 10.2,
			GPUCost: 0.10, VRAMCost: 0.02, StorageCost: 0.001, Subtotal: 0.121, TotalCost: 0.12,
			Resolution: "1024x1024",
		},
	}

	for _, r := range records {
		if err := DB.Create(&r).Error; err != nil {
			return err
		}
	}

	log.Println("Usage records created successfully")
	return nil
}
// SeedRechargeRecords 初始化充值记录
func SeedRechargeRecords() error {
	var count int64
	DB.Model(&models.RechargeRecord{}).Count(&count)
	if count > 0 {
		log.Println("Recharge records already exist, skipping seed")
		return nil
	}

	now := time.Now()
	completed := now.Add(-2 * time.Hour)

	records := []models.RechargeRecord{
		{UserID: 2, Amount: 100.00, Currency: "CNY", PaymentMethod: "alipay", Status: "completed", OrderID: "ORD-DEV-001", TransactionID: "TXN-DEV-001", CompletedAt: &completed},
		{UserID: 3, Amount: 500.00, Currency: "CNY", PaymentMethod: "wechat", Status: "completed", OrderID: "ORD-DEV-002", TransactionID: "TXN-DEV-002", CompletedAt: &completed},
		{UserID: 4, Amount: 2000.00, Currency: "CNY", PaymentMethod: "alipay", Status: "completed", OrderID: "ORD-DEV-003", TransactionID: "TXN-DEV-003", CompletedAt: &completed},
		{UserID: 2, Amount: 50.00, Currency: "CNY", PaymentMethod: "wechat", Status: "pending", OrderID: "ORD-DEV-004"},
	}

	for _, r := range records {
		if err := DB.Create(&r).Error; err != nil {
			return err
		}
	}

	log.Println("Recharge records created successfully")
	return nil
}

// SeedPrivateModels 初始化私有模型
func SeedPrivateModels() error {
	var count int64
	DB.Model(&models.PrivateModel{}).Count(&count)
	if count > 0 {
		log.Println("Private models already exist, skipping seed")
		return nil
	}

	now := time.Now()
	models_ := []models.PrivateModel{
		{UserID: 3, Name: "my-custom-lora.safetensors", Type: "lora", SizeBytes: 150_000_000, FilePath: "./data/users/3/models/my-custom-lora.safetensors", Visibility: "private", Status: "active", StorageCostPerDay: 0.0015, UploadedAt: now},
		{UserID: 4, Name: "portrait-checkpoint.safetensors", Type: "checkpoint", SizeBytes: 2_000_000_000, FilePath: "./data/users/4/models/portrait-checkpoint.safetensors", Visibility: "private", Status: "active", StorageCostPerDay: 0.02, UploadedAt: now},
	}

	for _, m := range models_ {
		if err := DB.Create(&m).Error; err != nil {
			return err
		}
	}

	log.Println("Private models created successfully")
	return nil
}
// SeedModelPermissions 初始化模型权限
func SeedModelPermissions() error {
	var count int64
	DB.Model(&models.ModelPermission{}).Count(&count)
	if count > 0 {
		log.Println("Model permissions already exist, skipping seed")
		return nil
	}

	perms := []models.ModelPermission{
		{UserID: 3, ModelPath: "vip/premium_model_1.safetensors", ModelName: "premium_model_1", ModelType: "checkpoint", FileSize: 3_000_000_000},
		{UserID: 4, ModelPath: "vip/premium_model_1.safetensors", ModelName: "premium_model_1", ModelType: "checkpoint", FileSize: 3_000_000_000},
		{UserID: 4, ModelPath: "vip/premium_model_2.safetensors", ModelName: "premium_model_2", ModelType: "checkpoint", FileSize: 4_000_000_000},
	}

	for _, p := range perms {
		if err := DB.Create(&p).Error; err != nil {
			return err
		}
	}

	log.Println("Model permissions created successfully")
	return nil
}

// SeedUserSettings 初始化用户设置
func SeedUserSettings() error {
	var count int64
	DB.Model(&models.UserSettings{}).Count(&count)
	if count > 0 {
		log.Println("User settings already exist, skipping seed")
		return nil
	}

	notifications := datatypes.JSON([]byte(`{"email":true,"browser":true}`))
	preferences := datatypes.JSON([]byte(`{"theme":"dark","language":"zh-CN"}`))

	settings := []models.UserSettings{
		{UserID: 1, Notifications: notifications, Preferences: preferences},
		{UserID: 2, Notifications: notifications, Preferences: preferences},
		{UserID: 3, Notifications: notifications, Preferences: preferences},
		{UserID: 4, Notifications: notifications, Preferences: preferences},
	}

	for _, s := range settings {
		if err := DB.Create(&s).Error; err != nil {
			return err
		}
	}

	log.Println("User settings created successfully")
	return nil
}

// SeedSystemLogs 初始化系统日志
func SeedSystemLogs() error {
	var count int64
	DB.Model(&models.SystemLog{}).Count(&count)
	if count > 0 {
		log.Println("System logs already exist, skipping seed")
		return nil
	}

	userID2 := uint(2)
	logs := []models.SystemLog{
		{Level: "info", Source: "system", Message: "System started successfully"},
		{Level: "info", Source: "auth", Message: fmt.Sprintf("User %d logged in", userID2), UserID: &userID2},
		{Level: "warn", Source: "proxy", Message: "ComfyUI instance comfyui-3 health check failed"},
	}

	for _, l := range logs {
		if err := DB.Create(&l).Error; err != nil {
			return err
		}
	}

	log.Println("System logs created successfully")
	return nil
}

// SeedDedicatedInstances 初始化独占实例分配
func SeedDedicatedInstances() error {
	var count int64
	DB.Model(&models.DedicatedInstance{}).Count(&count)
	if count > 0 {
		log.Println("Dedicated instances already exist, skipping seed")
		return nil
	}

	inst := models.DedicatedInstance{
		UserID:         4,
		SubscriptionID: 4,
		Subdomain:      "vip1",
		InstanceIDs:    "5",
		GPUIDs:         "4",
		Status:         "active",
	}

	if err := DB.Create(&inst).Error; err != nil {
		return err
	}

	log.Println("Dedicated instances created successfully")
	return nil
}
