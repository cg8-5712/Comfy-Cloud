# Comfy-Cloud 代理层

基于 Go + GORM + PostgreSQL 的 ComfyUI 云平台代理层，采用标准的 Go 项目架构。

## ✅ 已完成功能（Phase 0）

- ✅ 标准 Go 项目结构（参考 Easy-Stream）
- ✅ 分层架构（Handler → Service → Repository）
- ✅ 用户认证（注册/登录/JWT）
- ✅ 数据库管理（GORM + PostgreSQL）
- ✅ 用户等级系统（basic/pro/enterprise）
- ✅ 订阅管理模型
- ✅ 使用记录和计费模型
- ✅ 模型权限模型
- ✅ 配置管理（Viper）
- ✅ 日志系统（Zap）
- ✅ 依赖注入

## 项目结构

```
Comfy-Cloud/
├── cmd/server/              # 主程序入口
│   └── main.go
├── internal/                # 私有应用代码
│   ├── auth/               # 认证工具
│   ├── config/             # 配置管理
│   ├── database/           # 数据库连接
│   ├── handler/            # HTTP 处理器
│   ├── middleware/         # 中间件
│   ├── models/             # 数据模型
│   ├── repository/         # 数据访问层
│   └── service/            # 业务逻辑层
├── pkg/                    # 公共包
│   └── logger/             # 日志工具
├── configs/                # 配置文件
│   └── config.yaml
└── go.mod
```

## 架构分层

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
```

## 快速开始

### 1. 启动 PostgreSQL

```bash
docker run -d \
  --name comfy-postgres \
  -e POSTGRES_DB=comfy_cloud \
  -e POSTGRES_USER=comfy \
  -e POSTGRES_PASSWORD=comfy123 \
  -p 5432:5432 \
  postgres:15
```

### 2. 配置文件

编辑 `configs/config.yaml`：

```yaml
database:
  host: localhost
  port: 5432
  user: comfy
  password: comfy123
  dbname: comfy_cloud
  sslmode: disable

jwt:
  secret: "your-jwt-secret-key-change-in-production"
  expiration: 24h
```

### 3. 运行项目

```bash
# 方式 1：直接运行
go run cmd/server/main.go

# 方式 2：编译后运行
go build -o comfy-cloud cmd/server/main.go
./comfy-cloud
```

服务将在 `http://localhost:3000` 启动。

## API 接口

### 健康检查

```bash
GET /health
```

### 用户注册

```bash
POST /api/auth/register
Content-Type: application/json

{
  "username": "testuser",
  "email": "test@example.com",
  "password": "password123"
}
```

### 用户登录

```bash
POST /api/auth/login
Content-Type: application/json

{
  "username": "testuser",
  "password": "password123"
}
```

### 验证 Token

```bash
GET /api/auth/verify
Authorization: Bearer <your-token>
```

## 测试

```bash
# 注册
curl -X POST http://localhost:3000/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","email":"test@example.com","password":"123456"}'

# 登录
curl -X POST http://localhost:3000/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"123456"}'

# 验证（替换 token）
curl -X GET http://localhost:3000/api/auth/verify \
  -H "Authorization: Bearer <token>"
```

## 数据库模型

- **users** - 用户表
- **subscriptions** - 订阅表
- **usage_records** - 使用记录表
- **model_permissions** - 模型权限表

详见 [DATABASE_DESIGN.md](DATABASE_DESIGN.md)

## 技术栈

- **Go 1.21+** - 主语言
- **Gin** - Web 框架
- **GORM** - ORM 框架
- **PostgreSQL** - 数据库
- **JWT** - 认证
- **Viper** - 配置管理
- **Zap** - 日志系统
- **bcrypt** - 密码加密

## 开发进度

- [x] **Phase 0: 数据库设计和 GORM 集成**
  - [x] 标准项目结构
  - [x] 分层架构
  - [x] 数据库模型
  - [x] 用户认证
  - [x] JWT Token
  - [x] 配置管理
  - [x] 日志系统
- [ ] Phase 1: 基础框架搭建
- [ ] Phase 2: JWT 认证中间件
- [ ] Phase 3: 路径重写
- [ ] Phase 4: 模型权限控制
- [ ] Phase 5: 负载均衡
- [ ] Phase 6: WebSocket 支持
- [ ] Phase 7: 监控和日志
- [ ] Phase 8: 配置和部署

## 相关文档

- [项目结构说明](PROJECT_STRUCTURE.md) - 详细的架构说明
- [开发路线](CONDUCT_OF_CODE.md) - 开发计划
- [数据库设计](DATABASE_DESIGN.md) - 数据库设计文档
- [项目背景](CLAUDE.md) - 项目背景和架构

## 许可证

MIT
