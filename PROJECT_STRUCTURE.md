# Comfy-Cloud 项目结构

参考 Easy-Stream 的标准 Go 项目架构重构后的目录结构。

## 目录结构

```
Comfy-Cloud/
├── cmd/
│   └── server/              # 主程序入口
│       └── main.go          # 启动文件
│
├── internal/                # 私有应用代码
│   ├── auth/               # 认证相关
│   │   ├── jwt.go          # JWT 工具函数
│   │   └── service.go      # 认证业务逻辑
│   │
│   ├── config/             # 配置管理
│   │   └── config.go       # Viper 配置加载
│   │
│   ├── database/           # 数据库连接
│   │   └── db.go           # GORM 连接和迁移
│   │
│   ├── handler/            # HTTP 处理器（Controller 层）
│   │   └── auth.go         # 认证相关接口
│   │
│   ├── middleware/         # 中间件
│   │   └── auth.go         # JWT 认证中间件
│   │
│   ├── models/             # 数据模型
│   │   ├── user.go         # 用户模型
│   │   ├── subscription.go # 订阅模型
│   │   ├── usage.go        # 使用记录
│   │   └── model_permission.go # 模型权限
│   │
│   ├── repository/         # 数据访问层（DAO）
│   │   └── user_repository.go # 用户数据访问
│   │
│   └── service/            # 业务逻辑层
│       └── auth_service.go # 认证服务
│
├── pkg/                    # 可复用的公共包
│   ├── logger/             # 日志工具
│   │   └── logger.go       # Zap 日志
│   └── utils/              # 工具函数
│
├── configs/                # 配置文件
│   └── config.yaml         # 主配置文件
│
├── migrations/             # 数据库迁移文件
│
├── scripts/                # 脚本文件
│
├── go.mod
├── go.sum
└── README.md
```

## 架构分层

### 1. Handler 层（HTTP 处理器）
- 位置：`internal/handler/`
- 职责：
  - 接收 HTTP 请求
  - 参数验证和绑定
  - 调用 Service 层
  - 返回 HTTP 响应
- 示例：`auth.go` 处理注册、登录、验证接口

### 2. Service 层（业务逻辑）
- 位置：`internal/service/`
- 职责：
  - 实现核心业务逻辑
  - 调用 Repository 层
  - 事务管理
  - 业务规则验证
- 示例：`auth_service.go` 处理用户注册、登录逻辑

### 3. Repository 层（数据访问）
- 位置：`internal/repository/`
- 职责：
  - 数据库 CRUD 操作
  - 数据查询
  - 与 GORM 交互
- 示例：`user_repository.go` 提供用户数据访问方法

### 4. Model 层（数据模型）
- 位置：`internal/models/`
- 职责：
  - 定义数据结构
  - GORM 标签
  - 数据验证规则

### 5. Middleware 层（中间件）
- 位置：`internal/middleware/`
- 职责：
  - 请求拦截
  - 认证授权
  - 日志记录
  - 错误处理

## 数据流向

```
HTTP 请求
    ↓
Handler (接收请求、参数验证)
    ↓
Service (业务逻辑处理)
    ↓
Repository (数据访问)
    ↓
Database (数据库)
    ↓
Repository (返回数据)
    ↓
Service (业务处理)
    ↓
Handler (返回响应)
    ↓
HTTP 响应
```

## 依赖注入

在 `main.go` 中按顺序初始化：

```go
// 1. 初始化 Repository
userRepo := repository.NewUserRepository(database.DB)

// 2. 初始化 Service（注入 Repository）
authService := service.NewAuthService(userRepo, cfg)

// 3. 初始化 Handler（注入 Service）
authHandler := handler.NewAuthHandler(authService)

// 4. 设置路由
authHandler.SetupRoutes(r)
```

## 与 Easy-Stream 的对比

| 目录 | Comfy-Cloud | Easy-Stream | 说明 |
|------|-------------|-------------|------|
| 入口 | cmd/server/ | cmd/server/ | ✅ 一致 |
| 处理器 | internal/handler/ | internal/handler/ | ✅ 一致 |
| 服务层 | internal/service/ | internal/service/ | ✅ 一致 |
| 数据层 | internal/repository/ | internal/repository/ | ✅ 一致 |
| 模型 | internal/models/ | internal/model/ | ⚠️ 复数形式 |
| 中间件 | internal/middleware/ | internal/middleware/ | ✅ 一致 |
| 配置 | internal/config/ | internal/config/ | ✅ 一致 |
| 公共包 | pkg/ | pkg/ | ✅ 一致 |

## 优势

1. **清晰的分层架构** - 职责明确，易于维护
2. **依赖注入** - 便于测试和扩展
3. **标准 Go 项目结构** - 符合社区最佳实践
4. **易于扩展** - 新增功能只需添加对应层的代码

## 下一步

- [ ] 添加更多 Handler（Stream、Model、Billing 等）
- [ ] 完善 Service 层业务逻辑
- [ ] 添加单元测试
- [ ] 添加 API 文档（Swagger）
