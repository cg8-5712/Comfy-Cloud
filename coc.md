# Comfy-Cloud 开发路线

## 已完成 ✅

### Phase 0: 数据库设计和基础架构
- [x] 标准 Go 项目结构（handler/service/repository）
- [x] 数据库模型（User, Subscription, UsageRecord, ModelPermission）
- [x] 用户认证（注册/登录/JWT）
- [x] 配置管理（Viper）
- [x] 日志系统（Zap）
- [x] 计费服务（按 billing.md 实现）
- [x] 使用记录仓储

## 待开发

### Phase 1: 反向代理基础
- [ ] 实现反向代理中间件（httputil.ReverseProxy）
- [ ] 配置 ComfyUI 实例池
- [ ] 简单的轮询负载均衡
- [ ] 健康检查

### Phase 2: 路径重写和数据隔离
- [ ] 实现路径重写中间件
  - `/output/` → `/users/{user_id}/output/`
  - `/workflows/` → `/users/{user_id}/workflows/`
  - `/upload/` → `/users/{user_id}/upload/`
- [ ] 文件系统布局设计
- [ ] 用户目录初始化

### Phase 3: 模型权限控制
- [ ] 模型访问权限检查中间件
  - 基础模型：所有用户可访问
  - VIP 模型：Pro/Enterprise 用户
  - 私有模型：检查所有权
- [ ] 模型上传接口
- [ ] 模型列表接口

### Phase 4: 智能调度（模型亲和性）
- [ ] 实例状态管理
  - 当前加载的模型
  - 队列长度
  - 显存使用情况
- [ ] 模型亲和性调度算法
  - 优先选择已加载相同模型的实例
  - 队列长度次优先
- [ ] 动态实例分配
  - 监控模型请求分布
  - 自动调整专用实例

### Phase 5: 计费集成
- [ ] 任务提交时预扣费
- [ ] 任务完成时实际扣费
- [ ] 余额管理接口
  - 充值
  - 查询余额
  - 消费记录
- [ ] 使用统计接口
  - 月度统计
  - 详细记录

### Phase 6: WebSocket 代理
- [ ] WebSocket 连接代理
- [ ] Token 验证（query 参数）
- [ ] 进度推送
- [ ] 连接管理

### Phase 7: 独占模式
- [ ] 独占实例分配
- [ ] 资源隔离
- [ ] 专属路由
- [ ] 独占模式计费

### Phase 8: 监控和运维
- [ ] 性能监控
  - GPU 利用率
  - 队列长度
  - 响应时间
- [ ] 日志聚合
- [ ] 告警系统
- [ ] 自动扩缩容（可选）

### Phase 9: 部署
- [ ] Docker Compose 配置（8 个 ComfyUI 实例）
- [ ] 环境变量配置
- [ ] 数据备份策略
- [ ] 生产环境优化

## 技术栈

- **Go 1.21+** - 主语言
- **Gin** - Web 框架
- **GORM** - ORM
- **PostgreSQL 15** - 数据库
- **JWT** - 认证
- **Viper** - 配置
- **Zap** - 日志
- **Docker** - 容器化

## 核心功能

1. **多租户共享** - 多用户共享 ComfyUI 实例池
2. **数据隔离** - 路径重写实现用户数据隔离
3. **智能调度** - 模型亲和性缓存，减少加载时间
4. **灵活计费** - 按量计费（GPU + VRAM + 存储）+ 等待折扣
5. **独占模式** - 支持独占 GPU 资源

## 开发优先级

1. **Phase 1-2** - 基础代理和数据隔离（核心功能）
2. **Phase 3** - 模型权限控制
3. **Phase 4-5** - 智能调度和计费集成
4. **Phase 6** - WebSocket 支持
5. **Phase 7-9** - 独占模式、监控、部署
│   ├── loadbalancer/
│   │   ├── selector.go               # 实例选择器
│   │   └── health.go                 # 健康检查
│   ├── permission/
│   │   ├── model.go                  # 模型权限检查
│   │   └── middleware.go             # 权限中间件
│   ├── billing/
│   │   ├── service.go                # 计费服务
│   │   └── recorder.go               # 使用记录器
│   ├── config/
│   │   └── config.go                 # 配置加载
│   └── logger/
│       └── logger.go                 # 日志初始化
├── api/
│   └── routes/
│       ├── auth.go                   # 认证路由
│       ├── user.go                   # 用户管理路由
│       └── admin.go                  # 管理员路由
├── configs/
│   ├── config.yaml                   # 配置文件
│   └── config.example.yaml           # 配置示例
├── migrations/
│   └── 001_init.sql                  # 初始化 SQL
├── scripts/
│   ├── build.sh                      # 编译脚本
│   └── migrate.sh                    # 数据库迁移脚本
├── Dockerfile
├── go.mod
├── go.sum
└── README.md
```

## 开发阶段

### Phase 0: 数据库设计和 GORM 集成（第 1 天）

**目标：** 设计数据库模型，集成 GORM，实现基础的数据库操作

#### 数据库模型设计

```sql
-- 用户表
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    tier VARCHAR(20) DEFAULT 'basic',  -- basic/pro/enterprise
    balance DECIMAL(10,2) DEFAULT 0.00,
    status VARCHAR(20) DEFAULT 'active', -- active/suspended/deleted
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 订阅表
CREATE TABLE subscriptions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    plan VARCHAR(20) NOT NULL,  -- basic/pro/enterprise
    status VARCHAR(20) DEFAULT 'active',  -- active/cancelled/expired
    started_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 使用记录表
CREATE TABLE usage_records (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    task_type VARCHAR(50) NOT NULL,  -- prompt/upload/download
    cost DECIMAL(10,4) NOT NULL,
    duration INTEGER,  -- 秒
    metadata JSONB,  -- 额外信息（模型名称、图片尺寸等）
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 模型权限表（用户私有模型）
CREATE TABLE model_permissions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    model_path VARCHAR(255) NOT NULL,
    model_name VARCHAR(100) NOT NULL,
    model_type VARCHAR(50),  -- checkpoint/lora/vae/embedding
    file_size BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, model_path)
);

-- 创建索引
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX idx_usage_records_user_id ON usage_records(user_id);
CREATE INDEX idx_usage_records_created_at ON usage_records(created_at);
CREATE INDEX idx_model_permissions_user_id ON model_permissions(user_id);
```

#### 任务清单
- [ ] 定义 GORM 模型结构
- [ ] 数据库连接配置
- [ ] 自动迁移（AutoMigrate）
- [ ] 基础 CRUD 操作
- [ ] 用户注册/登录服务
- [ ] 密码加密（bcrypt）

#### GORM 模型定义

```go
// internal/models/user.go
package models

import (
    "time"
    "gorm.io/gorm"
)

type User struct {
    ID           uint           `gorm:"primaryKey" json:"id"`
    Username     string         `gorm:"uniqueIndex;size:50;not null" json:"username"`
    Email        string         `gorm:"uniqueIndex;size:100;not null" json:"email"`
    PasswordHash string         `gorm:"size:255;not null" json:"-"`
    Tier         string         `gorm:"size:20;default:basic" json:"tier"` // basic/pro/enterprise
    Balance      float64        `gorm:"type:decimal(10,2);default:0.00" json:"balance"`
    Status       string         `gorm:"size:20;default:active" json:"status"` // active/suspended/deleted
    CreatedAt    time.Time      `json:"created_at"`
    UpdatedAt    time.Time      `json:"updated_at"`
    DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// internal/models/subscription.go
package models

import (
    "time"
    "gorm.io/gorm"
)

type Subscription struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    UserID    uint           `gorm:"not null;index" json:"user_id"`
    User      User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
    Plan      string         `gorm:"size:20;not null" json:"plan"` // basic/pro/enterprise
    Status    string         `gorm:"size:20;default:active" json:"status"` // active/cancelled/expired
    StartedAt time.Time      `gorm:"not null" json:"started_at"`
    ExpiresAt *time.Time     `json:"expires_at,omitempty"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// internal/models/usage.go
package models

import (
    "time"
    "gorm.io/datatypes"
)

type UsageRecord struct {
    ID        uint            `gorm:"primaryKey" json:"id"`
    UserID    uint            `gorm:"not null;index" json:"user_id"`
    User      User            `gorm:"foreignKey:UserID" json:"user,omitempty"`
    TaskType  string          `gorm:"size:50;not null" json:"task_type"` // prompt/upload/download
    Cost      float64         `gorm:"type:decimal(10,4);not null" json:"cost"`
    Duration  int             `json:"duration,omitempty"` // 秒
    Metadata  datatypes.JSON  `json:"metadata,omitempty"`
    CreatedAt time.Time       `gorm:"index" json:"created_at"`
}

// internal/models/model_permission.go
package models

import (
    "time"
)

type ModelPermission struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    UserID    uint      `gorm:"not null;index" json:"user_id"`
    User      User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
    ModelPath string    `gorm:"size:255;not null" json:"model_path"`
    ModelName string    `gorm:"size:100;not null" json:"model_name"`
    ModelType string    `gorm:"size:50" json:"model_type"` // checkpoint/lora/vae/embedding
    FileSize  int64     `json:"file_size,omitempty"`
    CreatedAt time.Time `json:"created_at"`
}
```

#### 数据库连接

```go
// internal/database/db.go
package database

import (
    "fmt"
    "log"
    "time"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"

    "proxy-layer/internal/models"
)

var DB *gorm.DB

type Config struct {
    Host     string
    Port     int
    User     string
    Password string
    DBName   string
    SSLMode  string
}

func Connect(cfg Config) error {
    dsn := fmt.Sprintf(
        "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
        cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
    )

    var err error
    DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
        NowFunc: func() time.Time {
            return time.Now().UTC()
        },
    })

    if err != nil {
        return fmt.Errorf("failed to connect to database: %w", err)
    }

    // 配置连接池
    sqlDB, err := DB.DB()
    if err != nil {
        return fmt.Errorf("failed to get database instance: %w", err)
    }

    sqlDB.SetMaxIdleConns(10)
    sqlDB.SetMaxOpenConns(100)
    sqlDB.SetConnMaxLifetime(time.Hour)

    log.Println("Database connected successfully")
    return nil
}

func AutoMigrate() error {
    return DB.AutoMigrate(
        &models.User{},
        &models.Subscription{},
        &models.UsageRecord{},
        &models.ModelPermission{},
    )
}

func Close() error {
    sqlDB, err := DB.DB()
    if err != nil {
        return err
    }
    return sqlDB.Close()
}
```

#### 认证服务

```go
// internal/auth/service.go
package auth

import (
    "errors"
    "time"

    "golang.org/x/crypto/bcrypt"
    "gorm.io/gorm"

    "proxy-layer/internal/database"
    "proxy-layer/internal/models"
)

type RegisterRequest struct {
    Username string `json:"username" binding:"required,min=3,max=50"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
    Username string `json:"username" binding:"required"`
    Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
    Token string       `json:"token"`
    User  *models.User `json:"user"`
}

// 注册用户
func Register(req RegisterRequest) (*models.User, error) {
    // 检查用户名是否存在
    var existingUser models.User
    if err := database.DB.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
        return nil, errors.New("username already exists")
    }

    // 检查邮箱是否存在
    if err := database.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
        return nil, errors.New("email already exists")
    }

    // 加密密码
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        return nil, err
    }

    // 创建用户
    user := &models.User{
        Username:     req.Username,
        Email:        req.Email,
        PasswordHash: string(hashedPassword),
        Tier:         "basic",
        Balance:      10.00, // 新用户赠送 10 元
        Status:       "active",
    }

    if err := database.DB.Create(user).Error; err != nil {
        return nil, err
    }

    // 创建默认订阅
    subscription := &models.Subscription{
        UserID:    user.ID,
        Plan:      "basic",
        Status:    "active",
        StartedAt: time.Now(),
    }
    database.DB.Create(subscription)

    return user, nil
}

// 登录验证
func Login(req LoginRequest) (*models.User, error) {
    var user models.User

    // 查找用户
    if err := database.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, errors.New("invalid username or password")
        }
        return nil, err
    }

    // 验证密码
    if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
        return nil, errors.New("invalid username or password")
    }

    // 检查账户状态
    if user.Status != "active" {
        return nil, errors.New("account is suspended or deleted")
    }

    return &user, nil
}

// 根据 ID 获取用户
func GetUserByID(userID uint) (*models.User, error) {
    var user models.User
    if err := database.DB.First(&user, userID).Error; err != nil {
        return nil, err
    }
    return &user, nil
}
```

#### 验收标准

```bash
# 1. 启动 PostgreSQL
docker run -d \
  --name comfy-postgres \
  -e POSTGRES_DB=comfy_cloud \
  -e POSTGRES_USER=comfy \
  -e POSTGRES_PASSWORD=comfy123 \
  -p 5432:5432 \
  postgres:15

# 2. 运行数据库迁移
go run cmd/proxy/main.go migrate

# 3. 测试注册
curl -X POST http://localhost:3000/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123"
  }'

# 应该返回：
# {
#   "token": "eyJhbGc...",
#   "user": {
#     "id": 1,
#     "username": "testuser",
#     "email": "test@example.com",
#     "tier": "basic",
#     "balance": 10.00
#   }
# }

# 4. 测试登录
curl -X POST http://localhost:3000/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "password123"
  }'

# 5. 验证数据库
docker exec -it comfy-postgres psql -U comfy -d comfy_cloud
# \dt  -- 查看表
# SELECT * FROM users;  -- 查看用户
```

---

### Phase 1: 基础框架搭建（第 2 天）

**目标：** 搭建项目骨架，实现基础的 HTTP 代理

#### 任务清单
- [x] 初始化 Go 模块
- [ ] 创建项目目录结构
- [ ] 配置管理（读取 YAML 配置）
- [ ] 日志系统（zap）
- [ ] 基础 HTTP 服务器（Gin）
- [ ] 简单的反向代理（不带认证）

#### 验收标准
```bash
# 启动代理服务
go run cmd/proxy/main.go

# 测试转发
curl http://localhost:3000/health
# 应该返回：{"status": "ok"}

# 测试代理（假设 ComfyUI 在 8188 端口）
curl http://localhost:3000/comfy/queue
# 应该返回 ComfyUI 的队列信息
```

#### 核心代码示例
```go
// cmd/proxy/main.go
package main

import (
    "github.com/gin-gonic/gin"
    "proxy-layer/internal/config"
    "proxy-layer/internal/logger"
    "proxy-layer/internal/proxy"
)

func main() {
    // 加载配置
    cfg := config.Load()

    // 初始化日志
    logger.Init(cfg.LogLevel)

    // 创建路由
    r := gin.Default()

    // 健康检查
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

    // 代理路由
    r.Any("/comfy/*path", proxy.Handler(cfg))

    // 启动服务
    r.Run(":3000")
}
```

---

### Phase 2: JWT 认证（第 3 天）

**目标：** 实现 Token 验证，拒绝未授权请求

#### 任务清单
- [ ] JWT 验证逻辑
- [ ] 认证中间件
- [ ] 从 Header/Query 提取 Token
- [ ] Token 解析和验证
- [ ] 将用户信息注入到 Context

#### 验收标准
```bash
# 无 Token 请求 - 应该返回 401
curl http://localhost:3000/comfy/queue
# {"error": "No token provided"}

# 无效 Token - 应该返回 401
curl -H "Authorization: Bearer invalid_token" http://localhost:3000/comfy/queue
# {"error": "Invalid token"}

# 有效 Token - 应该正常转发
curl -H "Authorization: Bearer <valid_jwt>" http://localhost:3000/comfy/queue
# 返回 ComfyUI 队列信息
```

#### 核心代码示例
```go
// internal/auth/middleware.go
package auth

import (
    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(secret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 提取 Token
        tokenString := extractToken(c)
        if tokenString == "" {
            c.JSON(401, gin.H{"error": "No token provided"})
            c.Abort()
            return
        }

        // 验证 Token
        claims, err := verifyToken(tokenString, secret)
        if err != nil {
            c.JSON(401, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }

        // 注入用户信息
        c.Set("userId", claims.UserID)
        c.Set("userTier", claims.Tier)
        c.Next()
    }
}
```

---

### Phase 3: 路径重写（第 4 天）

**目标：** 根据用户 ID 重写路径，实现数据隔离

#### 任务清单
- [ ] 路径重写逻辑
- [ ] 处理 `/output/` 路径
- [ ] 处理 `/workflows/` 路径
- [ ] 处理 `/upload/` 路径
- [ ] 处理 `/input/` 路径
- [ ] 保持 `/models/` 路径不变（后续 Phase 4 处理）

#### 验收标准
```bash
# 用户 123 请求
curl -H "Authorization: Bearer <user_123_token>" \
     http://localhost:3000/comfy/output/image.png

# 实际转发到 ComfyUI 的路径应该是：
# /users/123/output/image.png

# 验证方法：查看日志
# [INFO] Path rewrite: /output/image.png -> /users/123/output/image.png
```

#### 核心代码示例
```go
// internal/proxy/rewrite.go
package proxy

import (
    "strings"
    "github.com/gin-gonic/gin"
)

func PathRewriteMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        userId := c.GetString("userId")
        originalPath := c.Request.URL.Path

        // 移除 /comfy 前缀
        path := strings.TrimPrefix(originalPath, "/comfy")

        // 重写用户数据路径
        if strings.HasPrefix(path, "/output/") {
            path = "/users/" + userId + path
        } else if strings.HasPrefix(path, "/workflows/") {
            path = "/users/" + userId + path
        } else if strings.HasPrefix(path, "/upload/") {
            path = "/users/" + userId + path
        } else if strings.HasPrefix(path, "/input/") {
            path = "/users/" + userId + path
        }

        // 更新请求路径
        c.Request.URL.Path = path
        c.Next()
    }
}
```

---

### Phase 4: 模型权限控制（第 5 天）

**目标：** 检查用户对模型的访问权限

#### 任务清单
- [ ] 模型路径解析
- [ ] 基础模型权限（所有人可访问）
- [ ] VIP 模型权限（检查用户等级）
- [ ] 私有模型权限（检查所有权）
- [ ] 权限拒绝响应

#### 验收标准
```bash
# 基础用户访问基础模型 - 允许
curl -H "Authorization: Bearer <basic_user_token>" \
     http://localhost:3000/comfy/models/base/sd_v1.5.safetensors
# 200 OK

# 基础用户访问 VIP 模型 - 拒绝
curl -H "Authorization: Bearer <basic_user_token>" \
     http://localhost:3000/comfy/models/vip/premium.safetensors
# 403 {"error": "VIP models require Pro subscription"}

# Pro 用户访问 VIP 模型 - 允许
curl -H "Authorization: Bearer <pro_user_token>" \
     http://localhost:3000/comfy/models/vip/premium.safetensors
# 200 OK

# 用户访问自己的私有模型 - 允许
curl -H "Authorization: Bearer <user_123_token>" \
     http://localhost:3000/comfy/models/user_123/custom.safetensors
# 200 OK

# 用户访问别人的私有模型 - 拒绝
curl -H "Authorization: Bearer <user_123_token>" \
     http://localhost:3000/comfy/models/user_456/custom.safetensors
# 403 {"error": "No access to this model"}
```

#### 核心代码示例
```go
// internal/permission/middleware.go
package permission

import (
    "strings"
    "github.com/gin-gonic/gin"
)

func ModelAccessMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        path := c.Request.URL.Path

        // 只检查模型路径
        if !strings.Contains(path, "/models/") {
            c.Next()
            return
        }

        userId := c.GetString("userId")
        userTier := c.GetString("userTier")

        // 提取模型路径
        parts := strings.Split(path, "/models/")
        if len(parts) < 2 {
            c.Next()
            return
        }
        modelPath := parts[1]

        // 基础模型：所有人可访问
        if strings.HasPrefix(modelPath, "base/") {
            c.Next()
            return
        }

        // 私有模型：检查所有权
        if strings.HasPrefix(modelPath, "user_"+userId+"/") {
            c.Next()
            return
        }

        // VIP 模型：检查等级
        if strings.HasPrefix(modelPath, "vip/") {
            if userTier == "pro" || userTier == "enterprise" {
                c.Next()
                return
            }
            c.JSON(403, gin.H{"error": "VIP models require Pro subscription"})
            c.Abort()
            return
        }

        // 默认拒绝
        c.JSON(403, gin.H{"error": "No access to this model"})
        c.Abort()
    }
}
```

---

### Phase 5: 负载均衡（第 6-7 天）

**目标：** 实现多实例负载均衡，选择队列最短的实例

#### 任务清单
- [ ] ComfyUI 实例配置
- [ ] 实例健康检查
- [ ] 队列长度查询（定期轮询）
- [ ] 实例选择器（最短队列算法）
- [ ] 动态目标切换
- [ ] 实例故障转移

#### 验收标准
```bash
# 配置 3 个 ComfyUI 实例
# config.yaml:
# comfy_instances:
#   - url: http://comfyui-1:8188
#   - url: http://comfyui-2:8188
#   - url: http://comfyui-3:8188

# 启动代理，查看日志
# [INFO] Instance comfyui-1: queue_size=2
# [INFO] Instance comfyui-2: queue_size=0  <- 选择这个
# [INFO] Instance comfyui-3: queue_size=5

# 发送请求
curl -H "Authorization: Bearer <token>" \
     http://localhost:3000/comfy/prompt
# 应该转发到 comfyui-2

# 模拟实例故障
# 停止 comfyui-2
# 再次发送请求，应该自动转发到其他实例
```

#### 核心代码示例
```go
// internal/loadbalancer/selector.go
package loadbalancer

import (
    "sync"
    "time"
)

type Instance struct {
    URL       string
    QueueSize int
    Healthy   bool
    mu        sync.RWMutex
}

type LoadBalancer struct {
    instances []*Instance
    mu        sync.RWMutex
}

func NewLoadBalancer(urls []string) *LoadBalancer {
    lb := &LoadBalancer{
        instances: make([]*Instance, len(urls)),
    }

    for i, url := range urls {
        lb.instances[i] = &Instance{
            URL:     url,
            Healthy: true,
        }
    }

    // 启动定期更新
    go lb.updateQueueSizes()

    return lb
}

func (lb *LoadBalancer) updateQueueSizes() {
    ticker := time.NewTicker(2 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        for _, instance := range lb.instances {
            go instance.fetchQueueSize()
        }
    }
}

func (lb *LoadBalancer) SelectInstance() *Instance {
    lb.mu.RLock()
    defer lb.mu.RUnlock()

    var best *Instance
    minQueue := int(^uint(0) >> 1) // max int

    for _, instance := range lb.instances {
        if !instance.Healthy {
            continue
        }

        instance.mu.RLock()
        queueSize := instance.QueueSize
        instance.mu.RUnlock()

        if queueSize < minQueue {
            minQueue = queueSize
            best = instance
        }
    }

    return best
}
```

---

### Phase 6: WebSocket 支持（第 8 天）

**目标：** 支持 WebSocket 连接，实现实时进度推送

#### 任务清单
- [ ] WebSocket 升级处理
- [ ] Token 验证（从 query 参数）
- [ ] WebSocket 代理转发
- [ ] 双向消息转发
- [ ] 连接关闭处理

#### 验收标准
```bash
# 使用 wscat 测试
npm install -g wscat

# 连接 WebSocket（带 Token）
wscat -c "ws://localhost:3000/comfy/ws?token=<valid_jwt>"

# 应该能收到 ComfyUI 的实时消息
# {"type": "status", "data": {...}}
# {"type": "progress", "data": {"value": 50, "max": 100}}
```

#### 核心代码示例
```go
// internal/proxy/websocket.go
package proxy

import (
    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
    "net/http"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

func WebSocketHandler(lb *loadbalancer.LoadBalancer) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 选择实例
        instance := lb.SelectInstance()
        if instance == nil {
            c.JSON(503, gin.H{"error": "No available instance"})
            return
        }

        // 升级到 WebSocket
        clientConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
        if err != nil {
            return
        }
        defer clientConn.Close()

        // 连接到 ComfyUI
        backendURL := "ws://" + instance.URL + c.Request.URL.Path
        backendConn, _, err := websocket.DefaultDialer.Dial(backendURL, nil)
        if err != nil {
            return
        }
        defer backendConn.Close()

        // 双向转发
        go copyWebSocket(clientConn, backendConn)
        copyWebSocket(backendConn, clientConn)
    }
}
```

---

### Phase 7: 监控和日志（第 9 天）

**目标：** 添加详细的日志和监控指标

#### 任务清单
- [ ] 请求日志（请求路径、用户 ID、耗时）
- [ ] 错误日志（认证失败、权限拒绝、代理错误）
- [ ] 性能指标（请求延迟、队列长度、实例健康）
- [ ] Prometheus metrics 接口（可选）

#### 验收标准
```bash
# 查看日志
tail -f logs/proxy.log

# 应该看到结构化日志
# {"level":"info","time":"2024-01-15T10:30:00Z","msg":"Request","userId":"123","path":"/comfy/prompt","duration":"45ms"}
# {"level":"error","time":"2024-01-15T10:30:05Z","msg":"Auth failed","reason":"Invalid token"}

# 访问 metrics 接口
curl http://localhost:3000/metrics
# comfy_proxy_requests_total{status="200"} 1234
# comfy_proxy_requests_duration_seconds{quantile="0.5"} 0.045
# comfy_instance_queue_size{instance="comfyui-1"} 2
```

---

### Phase 8: 配置和部署（第 10 天）

**目标：** 完善配置管理，准备生产部署

#### 任务清单
- [ ] 配置文件完善（所有参数可配置）
- [ ] 环境变量支持
- [ ] Docker 镜像构建
- [ ] Docker Compose 配置
- [ ] 健康检查接口
- [ ] 优雅关闭

#### 验收标准
```bash
# 构建 Docker 镜像
docker build -t comfy-proxy:latest .

# 运行容器
docker run -p 3000:3000 \
  -e JWT_SECRET=your-secret \
  -e COMFY_INSTANCES=http://comfyui-1:8188,http://comfyui-2:8188 \
  comfy-proxy:latest

# 健康检查
curl http://localhost:3000/health
# {"status": "ok", "instances": 2, "healthy": 2}
```

---

## 配置文件示例

```yaml
# configs/config.yaml
server:
  port: 3000
  mode: release  # debug/release

database:
  host: localhost
  port: 5432
  user: comfy
  password: comfy123
  dbname: comfy_cloud
  sslmode: disable  # disable/require

jwt:
  secret: "your-jwt-secret-key"
  expiration: 24h

comfy_instances:
  - url: http://comfyui-1:8188
  - url: http://comfyui-2:8188
  - url: http://comfyui-3:8188

loadbalancer:
  health_check_interval: 5s
  queue_update_interval: 2s
  max_queue_size: 10  # 超过则拒绝请求

logging:
  level: info  # debug/info/warn/error
  output: stdout  # stdout/file
  file: logs/proxy.log

redis:
  addr: localhost:6379
  password: ""
  db: 0

rate_limit:
  enabled: true
  requests_per_minute: 60

billing:
  cost_per_prompt: 0.10  # 每次生成 0.1 元
  cost_per_minute: 0.05  # 每分钟 0.05 元
```

---

## 测试策略

### 单元测试
```bash
# 测试 JWT 验证
go test ./internal/auth/...

# 测试路径重写
go test ./internal/proxy/...

# 测试负载均衡
go test ./internal/loadbalancer/...
```

### 集成测试
```bash
# 启动测试环境
docker-compose -f docker-compose.test.yml up -d

# 运行集成测试
go test ./tests/integration/...
```

### 压力测试
```bash
# 使用 wrk 进行压力测试
wrk -t4 -c100 -d30s \
  -H "Authorization: Bearer <token>" \
  http://localhost:3000/comfy/queue

# 目标：
# - 延迟 < 10ms (p99)
# - 吞吐量 > 5000 req/s
# - 错误率 < 0.1%
```

---

## 性能目标

| 指标 | 目标值 |
|------|--------|
| 请求延迟 (p50) | < 5ms |
| 请求延迟 (p99) | < 20ms |
| 吞吐量 | > 5000 req/s |
| 内存占用 | < 50MB |
| CPU 占用 | < 10% (空闲时) |
| 并发连接 | > 10000 |

---

## 下一步

1. **立即开始 Phase 1** - 搭建基础框架
2. **每天完成一个 Phase** - 保持节奏
3. **每个 Phase 都要测试** - 确保质量
4. **遇到问题及时调整** - 灵活应对

准备好开始了吗？我可以帮你：
- 生成初始项目结构
- 编写 Phase 1 的代码
- 配置开发环境
