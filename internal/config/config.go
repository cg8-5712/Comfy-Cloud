package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server       ServerConfig       `mapstructure:"server"`
	Database     DatabaseConfig     `mapstructure:"database"`
	JWT          JWTConfig          `mapstructure:"jwt"`
	ComfyInstances []InstanceConfig `mapstructure:"comfy_instances"`
	LoadBalancer LoadBalancerConfig `mapstructure:"loadbalancer"`
	Logging      LoggingConfig      `mapstructure:"logging"`
	Redis        RedisConfig        `mapstructure:"redis"`
	RateLimit    RateLimitConfig    `mapstructure:"rate_limit"`
	Billing      BillingConfig      `mapstructure:"billing"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

type JWTConfig struct {
	Secret     string        `mapstructure:"secret"`
	Expiration time.Duration `mapstructure:"expiration"`
}

type InstanceConfig struct {
	URL string `mapstructure:"url"`
}

type LoadBalancerConfig struct {
	HealthCheckInterval time.Duration `mapstructure:"health_check_interval"`
	QueueUpdateInterval time.Duration `mapstructure:"queue_update_interval"`
	MaxQueueSize        int           `mapstructure:"max_queue_size"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Output string `mapstructure:"output"`
	File   string `mapstructure:"file"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type RateLimitConfig struct {
	Enabled            bool `mapstructure:"enabled"`
	RequestsPerMinute  int  `mapstructure:"requests_per_minute"`
}

type BillingConfig struct {
	CostPerPrompt float64 `mapstructure:"cost_per_prompt"`
	CostPerMinute float64 `mapstructure:"cost_per_minute"`
}

var AppConfig *Config

// Load 加载配置文件
func Load() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	// 设置默认值
	setDefaults()

	// 读取环境变量
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: Config file not found, using defaults: %v", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("Failed to unmarshal config: %v", err)
	}

	AppConfig = &cfg
	return &cfg
}

func setDefaults() {
	// Server
	viper.SetDefault("server.port", 3000)
	viper.SetDefault("server.mode", "release")

	// Database
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "comfy")
	viper.SetDefault("database.password", "comfy123")
	viper.SetDefault("database.dbname", "comfy_cloud")
	viper.SetDefault("database.sslmode", "disable")

	// JWT
	viper.SetDefault("jwt.secret", "your-secret-key-change-in-production")
	viper.SetDefault("jwt.expiration", "24h")

	// LoadBalancer
	viper.SetDefault("loadbalancer.health_check_interval", "5s")
	viper.SetDefault("loadbalancer.queue_update_interval", "2s")
	viper.SetDefault("loadbalancer.max_queue_size", 10)

	// Logging
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.output", "stdout")
	viper.SetDefault("logging.file", "logs/proxy.log")

	// Redis
	viper.SetDefault("redis.addr", "localhost:6379")
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)

	// RateLimit
	viper.SetDefault("rate_limit.enabled", true)
	viper.SetDefault("rate_limit.requests_per_minute", 60)

	// Billing
	viper.SetDefault("billing.cost_per_prompt", 0.10)
	viper.SetDefault("billing.cost_per_minute", 0.05)
}
