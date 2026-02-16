package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"comfy-cloud/internal/auth"
	"comfy-cloud/internal/config"
	"comfy-cloud/internal/database"
	"comfy-cloud/internal/handler"
	"comfy-cloud/internal/middleware"
	"comfy-cloud/internal/models"
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

	// 运行数据库迁移（版本化）
	if err := database.RunMigrations(); err != nil {
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

	// 初始化配置服务（支持热加载）
	configService := service.NewConfigService(database.DB)
	configService.StartAutoReload(30 * time.Second) // 每 30 秒重新加载配置
	defer configService.Stop()
	log.Println("Config service initialized with auto-reload")

	// 初始化 Handler 层
	authHandler := handler.NewAuthHandler(authService)

	// 初始化 ComfyUI 实例池（从数据库加载）
	instanceService := service.NewInstanceService(database.DB, nil)
	instances, err := instanceService.LoadInstances()
	if err != nil {
		log.Fatalf("Failed to load instances: %v", err)
	}
	instancePool := proxy.NewInstancePoolFromInstances(instances)
	instanceService.SetPool(instancePool) // 设置实例池引用
	log.Printf("Initialized instance pool with %d instances from database", len(instances))

	// 启动健康检查
	healthCheckInterval := time.Duration(configService.GetInt("loadbalancer", "health_check_interval")) * time.Second
	instancePool.StartHealthCheck(healthCheckInterval)
	log.Printf("Health check started (interval: %v)", healthCheckInterval)

	// 初始化代理处理器
	proxyHandler := proxy.NewProxyHandler(instancePool, database.DB)
	log.Println("Proxy handler initialized")

	// 初始化用户目录服务
	userDataDir := configService.Get("storage", "user_data_dir")
	userDirService := service.NewUserDirectoryService(userDataDir)
	log.Printf("User directory service initialized (base: %s)", userDataDir)

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
	// 需要认证 + 路径重写
	comfyGroup := r.Group("/comfy")
	comfyGroup.Use(middleware.AuthMiddleware())
	comfyGroup.Use(middleware.PathRewriteMiddleware())
	{
		comfyGroup.Any("/*path", proxyHandler.Route)
		comfyGroup.Any("", proxyHandler.Route)
	}

	// 用户目录管理接口
	r.POST("/api/user/init-directory", middleware.AuthMiddleware(), func(c *gin.Context) {
		userID := c.GetUint("user_id")
		if err := userDirService.InitializeUserDirectory(userID); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"message": "User directory initialized"})
	})

	// 查询用户存储使用情况
	r.GET("/api/user/storage", middleware.AuthMiddleware(), func(c *gin.Context) {
		userID := c.GetUint("user_id")
		usage, err := userDirService.GetUserStorageUsage(userID)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		maxStorage := configService.GetFloat("storage", "max_user_storage_gb")
		c.JSON(200, gin.H{
			"storage_used_gb":  usage,
			"storage_limit_gb": maxStorage,
		})
	})

	// 配置管理接口
	r.GET("/api/admin/config", func(c *gin.Context) {
		configs := configService.GetAll()
		c.JSON(200, gin.H{"configs": configs})
	})

	r.POST("/api/admin/config", func(c *gin.Context) {
		var req struct {
			Category string `json:"category" binding:"required"`
			Key      string `json:"key" binding:"required"`
			Value    string `json:"value" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		if err := configService.Set(req.Category, req.Key, req.Value); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"message": "Config updated successfully"})
	})

	// 实例管理接口
	r.GET("/api/admin/instances", func(c *gin.Context) {
		instances, err := instanceService.GetAllInstances()
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"instances": instances})
	})

	r.POST("/api/admin/instances", func(c *gin.Context) {
		var inst models.ComfyInstance
		if err := c.ShouldBindJSON(&inst); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		if err := instanceService.AddInstance(&inst); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"message": "Instance added successfully", "instance": inst})
	})

	r.PUT("/api/admin/instances/:id", func(c *gin.Context) {
		var inst models.ComfyInstance
		if err := c.ShouldBindJSON(&inst); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		if err := instanceService.UpdateInstance(&inst); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"message": "Instance updated successfully"})
	})

	r.DELETE("/api/admin/instances/:id", func(c *gin.Context) {
		id := c.Param("id")
		var idUint uint
		fmt.Sscanf(id, "%d", &idUint)

		if err := instanceService.DeleteInstance(idUint); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"message": "Instance deleted successfully"})
	})

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
