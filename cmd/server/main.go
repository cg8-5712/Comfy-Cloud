package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"

	"comfy-cloud/internal/auth"
	"comfy-cloud/internal/config"
	"comfy-cloud/internal/database"
	"comfy-cloud/internal/handler"
	"comfy-cloud/internal/proxy"
	"comfy-cloud/internal/repository"
	"comfy-cloud/internal/service"
	"comfy-cloud/pkg/logger"
)

func main() {
	// 加载配置
	cfg := config.Load()
	log.Println("Configuration loaded successfully")

	// 初始化日志
	if err := logger.Init(cfg.Logging.Level, cfg.Logging.Output, cfg.Logging.File); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()
	log.Println("Logger initialized successfully")

	// 初始化 JWT
	auth.InitJWT(cfg.JWT.Secret)
	log.Println("JWT initialized successfully")

	// 连接数据库
	dbConfig := database.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.DBName,
		SSLMode:  cfg.Database.SSLMode,
	}

	if err := database.Connect(dbConfig); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()
	log.Println("Database connected successfully")

	// 运行数据库迁移
	if err := database.AutoMigrate(); err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}
	log.Println("Database migrations completed successfully")

	// 初始化默认数据
	if err := database.SeedAll(); err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}
	log.Println("Database seeding completed successfully")

	// 初始化 Repository 层
	userRepo := repository.NewUserRepository(database.DB)

	// 初始化 Service 层
	authService := service.NewAuthService(userRepo, cfg)

	// 初始化 Handler 层
	authHandler := handler.NewAuthHandler(authService)

	// 初始化 ComfyUI 实例池
	sharedURLs := make([]string, 0)
	for _, inst := range cfg.ComfyInstances.Shared {
		sharedURLs = append(sharedURLs, inst.URL)
	}
	instancePool := proxy.NewInstancePool(sharedURLs)
	log.Printf("Initialized instance pool with %d instances", len(sharedURLs))

	// 启动健康检查
	instancePool.StartHealthCheck(cfg.LoadBalancer.HealthCheckInterval)
	log.Println("Health check started")

	// 初始化代理处理器
	proxyHandler := proxy.NewProxyHandler(instancePool, database.DB)
	log.Println("Proxy handler initialized")

	// 设置 Gin 模式
	gin.SetMode(cfg.Server.Mode)

	// 创建 Gin 引擎
	r := gin.Default()

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "comfy-cloud-proxy",
		})
	})

	// 设置路由
	authHandler.SetupRoutes(r)

	// 代理路由（所有 /comfy/* 请求）
	r.Any("/comfy/*path", proxyHandler.Route)
	r.Any("/comfy", proxyHandler.Route)

	// 实例状态查询
	r.GET("/api/instances", func(c *gin.Context) {
		instances := instancePool.GetAllInstances()
		c.JSON(200, gin.H{
			"instances": instances,
		})
	})

	// 启动服务器
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Starting server on %s", addr)

	// 优雅关闭
	go func() {
		if err := r.Run(addr); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
}
