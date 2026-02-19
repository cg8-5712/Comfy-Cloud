package database

import (
	"comfy-cloud/internal/models"
	"fmt"
	"log"
	"time"
)

// RunMigrations 运行数据库迁移
func RunMigrations() error {
	// 首先确保 migrations 表存在
	if err := DB.AutoMigrate(&models.Migration{}); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// 定义所有迁移
	migrations := []struct {
		version     string
		description string
		up          func() error
	}{
		{
			version:     "v1",
			description: "Initial schema",
			up:          migrateV1,
		},
		{
			version:     "v2",
			description: "Add system configs and comfy instances tables",
			up:          migrateV2,
		},
		{
			version:     "v3",
			description: "Add recharge, private models, user settings, and system logs tables",
			up:          migrateV3,
		},
		{
			version:     "v4",
			description: "Add role field to users table",
			up:          migrateV4,
		},
	}

	// 执行迁移
	for _, m := range migrations {
		// 检查是否已应用
		var existing models.Migration
		err := DB.Where("version = ?", m.version).First(&existing).Error
		if err == nil {
			log.Printf("Migration %s already applied, skipping", m.version)
			continue
		}

		// 执行迁移
		log.Printf("Applying migration %s: %s", m.version, m.description)
		if err := m.up(); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", m.version, err)
		}

		// 记录迁移
		migration := models.Migration{
			Version:     m.version,
			Description: m.description,
			AppliedAt:   time.Now(),
		}
		if err := DB.Create(&migration).Error; err != nil {
			return fmt.Errorf("failed to record migration %s: %w", m.version, err)
		}

		log.Printf("Migration %s applied successfully", m.version)
	}

	return nil
}

// migrateV1 初始 schema
func migrateV1() error {
	return DB.AutoMigrate(
		&models.User{},
		&models.Subscription{},
		&models.UsageRecord{},
		&models.ModelPermission{},
		&models.BillingConfig{},
		&models.DedicatedPricing{},
		&models.DedicatedInstance{},
	)
}

// migrateV2 添加系统配置和实例表
func migrateV2() error {
	return DB.AutoMigrate(
		&models.SystemConfig{},
		&models.ComfyInstance{},
	)
}

// migrateV3 添加充值、私有模型、用户设置和系统日志表
func migrateV3() error {
	return DB.AutoMigrate(
		&models.RechargeRecord{},
		&models.PrivateModel{},
		&models.UserSettings{},
		&models.SystemLog{},
	)
}

// migrateV4 给 users 表添加 role 字段
func migrateV4() error {
	return DB.AutoMigrate(&models.User{})
}
