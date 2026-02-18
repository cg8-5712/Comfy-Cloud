# Comfy-Cloud API 规范

本文档定义了 Comfy-Cloud 前后端交互的 API 接口规范。

## 基础信息

- **Base URL**: `/api`
- **认证方式**: JWT Bearer Token
- **Content-Type**: `application/json`
- **字符编码**: UTF-8

## 认证相关 API

### 1. 用户登录

**请求**
```
POST /api/auth/login
Content-Type: application/json

{
  "username": "string",
  "password": "string"
}
```

**响应 - 成功 (200)**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 123,
    "username": "john_doe",
    "email": "john@example.com",
    "tier": "pro",
    "role": "user",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

**响应 - 失败 (401)**
```json
{
  "error": "Invalid credentials"
}
```

### 2. 用户注册

**请求**
```
POST /api/auth/register
Content-Type: application/json

{
  "username": "string",
  "email": "string",
  "password": "string"
}
```

**响应 - 成功 (201)**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 123,
    "username": "john_doe",
    "email": "john@example.com",
    "tier": "basic",
    "role": "user",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

**响应 - 失败 (400)**
```json
{
  "error": "Username already exists"
}
```

### 3. 用户登出

**请求**
```
POST /api/auth/logout
Authorization: Bearer <token>
```

**响应 - 成功 (200)**
```json
{
  "message": "Logged out successfully"
}
```

### 4. 刷新 Token

**请求**
```
POST /api/auth/refresh
Authorization: Bearer <token>
```

**响应 - 成功 (200)**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

## 用户信息 API

### 5. 获取当前用户信息

**请求**
```
GET /api/user/info
Authorization: Bearer <token>
```

**响应 - 成功 (200)**
```json
{
  "id": 123,
  "username": "john_doe",
  "email": "john@example.com",
  "tier": "pro",
  "balance": 150.50,
  "storage_used": 5368709120,
  "storage_limit": 107374182400,
  "created_at": "2024-01-01T00:00:00Z",
  "subscription": {
    "tier": "pro",
    "status": "active",
    "expires_at": "2025-01-01T00:00:00Z"
  }
}
```

### 6. 获取用户余额

**请求**
```
GET /api/user/balance
Authorization: Bearer <token>
```

**响应 - 成功 (200)**
```json
{
  "balance": 150.50,
  "currency": "USD",
  "last_updated": "2024-02-16T10:30:00Z"
}
```

### 7. 获取用户使用统计

**请求**
```
GET /api/user/usage?period=month
Authorization: Bearer <token>

Query Parameters:
- period: "day" | "week" | "month" | "year"
```

**响应 - 成功 (200)**
```json
{
  "period": "month",
  "start_date": "2024-02-01T00:00:00Z",
  "end_date": "2024-02-29T23:59:59Z",
  "gpu_seconds": 3600,
  "storage_gb_hours": 240,
  "total_cost": 25.50,
  "task_count": 150
}
```

## 订阅等级配置 API

### 8. 获取订阅等级列表

获取平台所有可用的订阅等级配置。此接口为公开接口，无需认证，前端用于动态渲染等级标签、定价卡片等。

**请求**
```
GET /api/tiers
```

**响应 - 成功 (200)**
```json
[
  {
    "key": "basic",
    "label": "基础版",
    "color": "bg-muted text-muted-foreground",
    "price": "免费",
    "features": [
      "每月 100 次任务",
      "5 GB 存储空间",
      "基础模型访问",
      "社区支持"
    ],
    "popular": false
  },
  {
    "key": "pro",
    "label": "专业版",
    "color": "bg-primary/10 text-primary",
    "price": "¥99/月",
    "features": [
      "无限任务",
      "50 GB 存储空间",
      "VIP 模型访问",
      "优先队列",
      "邮件支持"
    ],
    "popular": true
  },
  {
    "key": "enterprise",
    "label": "企业版",
    "color": "bg-amber-500/10 text-amber-600",
    "price": "¥299/月",
    "features": [
      "无限任务",
      "200 GB 存储空间",
      "全部模型访问",
      "最高优先级",
      "专属支持",
      "团队协作"
    ],
    "popular": false
  }
]
```

**说明**
- 此接口不需要 Authorization header
- 前端在用户进入账户页面时调用一次，结果缓存在 store 中
- `key` 字段与 User.tier 对应
- `color` 字段为前端 Tailwind CSS 类名，用于 Badge 样式
- `popular` 字段为 true 时前端会标记为"推荐"
- 后端可通过数据库或配置文件管理等级，新增/修改等级无需前端发版

## 订阅管理 API

### 9. 获取订阅信息


**请求**
```
GET /api/subscription
Authorization: Bearer <token>
```

**响应 - 成功 (200)**
```json
{
  "tier": "pro",
  "status": "active",
  "started_at": "2024-01-01T00:00:00Z",
  "expires_at": "2025-01-01T00:00:00Z",
  "auto_renew": true,
  "features": {
    "gpu_access": true,
    "vip_models": true,
    "storage_limit_gb": 100,
    "concurrent_tasks": 3
  }
}
```

### 10. 更新订阅

**请求**
```
POST /api/subscription/upgrade
Authorization: Bearer <token>
Content-Type: application/json

{
  "target_tier": "enterprise"
}
```

**响应 - 成功 (200)**
```json
{
  "tier": "enterprise",
  "status": "active",
  "started_at": "2024-02-16T10:30:00Z",
  "expires_at": "2025-02-16T10:30:00Z"
}
```

## 充值 API

### 11. 创建充值订单

**请求**
```
POST /api/recharge
Authorization: Bearer <token>
Content-Type: application/json

{
  "amount": 100.00,
  "payment_method": "stripe"
}
```

**响应 - 成功 (201)**
```json
{
  "order_id": "ord_abc123",
  "amount": 100.00,
  "currency": "USD",
  "payment_url": "https://checkout.stripe.com/...",
  "status": "pending",
  "created_at": "2024-02-16T10:30:00Z"
}
```

### 12. 获取充值记录

**请求**
```
GET /api/recharge/history?limit=20&offset=0
Authorization: Bearer <token>

Query Parameters:
- limit: number (default: 20, max: 100)
- offset: number (default: 0)
```

**响应 - 成功 (200)**
```json
{
  "total": 50,
  "records": [
    {
      "id": 1,
      "amount": 100.00,
      "currency": "USD",
      "payment_method": "stripe",
      "status": "completed",
      "created_at": "2024-02-16T10:30:00Z",
      "completed_at": "2024-02-16T10:31:00Z"
    }
  ]
}
```

## 使用记录 API

### 13. 获取使用记录列表

**请求**
```
GET /api/usage/records?start_date=2024-02-01&end_date=2024-02-29&limit=50&offset=0
Authorization: Bearer <token>

Query Parameters:
- start_date: ISO 8601 date (optional)
- end_date: ISO 8601 date (optional)
- limit: number (default: 50, max: 100)
- offset: number (default: 0)
```

**响应 - 成功 (200)**
```json
{
  "total": 150,
  "records": [
    {
      "id": 1001,
      "task_id": "task_abc123",
      "type": "gpu_usage",
      "started_at": "2024-02-16T10:00:00Z",
      "ended_at": "2024-02-16T10:05:00Z",
      "duration_seconds": 300,
      "cost": 0.50,
      "details": {
        "gpu_type": "RTX 4090",
        "vram_used_gb": 8.5,
        "model": "sd_v1.5"
      }
    }
  ]
}
```

## 模型管理 API

### 14. 获取用户私有模型列表

**请求**
```
GET /api/models/private
Authorization: Bearer <token>
```

**响应 - 成功 (200)**
```json
{
  "models": [
    {
      "id": 1,
      "name": "my_custom_lora.safetensors",
      "type": "lora",
      "size_bytes": 143654912,
      "uploaded_at": "2024-02-10T15:30:00Z",
      "storage_cost_per_day": 0.01
    }
  ]
}
```

### 15. 上传私有模型

**请求**
```
POST /api/models/upload
Authorization: Bearer <token>
Content-Type: multipart/form-data

file: <binary>
type: "checkpoint" | "lora" | "vae" | "embedding"
```

**响应 - 成功 (201)**
```json
{
  "id": 2,
  "name": "my_model.safetensors",
  "type": "lora",
  "size_bytes": 143654912,
  "uploaded_at": "2024-02-16T10:30:00Z",
  "path": "/users/123/models/my_model.safetensors"
}
```

### 16. 删除私有模型

**请求**
```
DELETE /api/models/private/:id
Authorization: Bearer <token>
```

**响应 - 成功 (200)**
```json
{
  "message": "Model deleted successfully"
}
```

## 设置 API

### 17. 获取用户设置

**请求**
```
GET /api/settings
Authorization: Bearer <token>
```

**响应 - 成功 (200)**
```json
{
  "notifications": {
    "email_on_task_complete": true,
    "email_on_low_balance": true,
    "low_balance_threshold": 10.00
  },
  "preferences": {
    "language": "en",
    "timezone": "UTC"
  }
}
```

### 18. 更新用户设置

**请求**
```
PATCH /api/settings
Authorization: Bearer <token>
Content-Type: application/json

{
  "notifications": {
    "email_on_task_complete": false
  }
}
```

**响应 - 成功 (200)**
```json
{
  "message": "Settings updated successfully"
}
```

### 19. 修改密码

**请求**
```
POST /api/settings/password
Authorization: Bearer <token>
Content-Type: application/json

{
  "old_password": "string",
  "new_password": "string"
}
```

**响应 - 成功 (200)**
```json
{
  "message": "Password updated successfully"
}
```

## 错误响应格式

所有错误响应遵循统一格式：

```json
{
  "error": "Error message",
  "code": "ERROR_CODE",
  "details": {}
}
```

## Admin API

> 以下接口需要管理员权限（`role: "admin"`）

### 20. 获取管理统计

**请求**
```
GET /api/admin/stats
Authorization: Bearer <token>
```

**响应 - 成功 (200)**
```json
{
  "total_users": 1250,
  "active_today": 89,
  "tasks_today": 456,
  "total_revenue": 125000.00,
  "online_instances": 3,
  "avg_queue_length": 2.5,
  "gpu_utilization": 72.5,
  "chart_data": [
    { "date": "2024-02-01", "tasks": 120, "revenue": 3500 }
  ]
}
```

### 21. 获取用户列表（Admin）

**请求**
```
GET /api/admin/users?search=keyword&limit=20&offset=0
Authorization: Bearer <token>

Query Parameters:
- search: string (可选，按用户名/邮箱搜索)
- limit: number (default: 20, max: 100)
- offset: number (default: 0)
```

**响应 - 成功 (200)**
```json
{
  "total": 1250,
  "users": [
    {
      "id": 1,
      "username": "john_doe",
      "email": "john@example.com",
      "tier": "pro",
      "balance": 150.50,
      "status": "active",
      "role": "user",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

### 22. 编辑用户（Admin）

**请求**
```
PATCH /api/admin/users/:id
Authorization: Bearer <token>
Content-Type: application/json

{
  "tier": "enterprise",
  "status": "active",
  "role": "admin",
  "balance": 200.00
}
```

**响应 - 成功 (200)**
```json
{
  "id": 1,
  "username": "john_doe",
  "tier": "enterprise",
  "status": "active",
  "role": "admin",
  "balance": 200.00
}
```

### 23. 获取实例列表（Admin）

**请求**
```
GET /api/admin/instances
Authorization: Bearer <token>
```

**响应 - 成功 (200)**
```json
{
  "instances": [
    {
      "id": "comfyui-1",
      "url": "http://10.0.0.1:8188",
      "status": "online",
      "gpu_model": "RTX 4090",
      "vram_total_gb": 24,
      "vram_used_gb": 16.5,
      "queue_length": 3,
      "uptime_hours": 168,
      "gpu_utilization": 75
    }
  ]
}
```

### 24. 获取模型列表（Admin）

**请求**
```
GET /api/admin/models?visibility=base
Authorization: Bearer <token>

Query Parameters:
- visibility: "base" | "vip" | "private" (可选，筛选可见性)
```

**响应 - 成功 (200)**
```json
{
  "total": 9,
  "models": [
    {
      "id": 1,
      "name": "sd_v1.5.safetensors",
      "type": "checkpoint",
      "size_bytes": 4265000000,
      "visibility": "base",
      "status": "active",
      "user_id": 0,
      "username": "system",
      "uploaded_at": "2024-01-01T00:00:00Z",
      "storage_cost_per_day": 0
    }
  ]
}
```

### 25. 编辑模型（Admin）

**请求**
```
PATCH /api/admin/models/:id
Authorization: Bearer <token>
Content-Type: application/json

{
  "visibility": "vip",
  "status": "active"
}
```

**响应 - 成功 (200)**
```json
{
  "id": 1,
  "name": "sd_v1.5.safetensors",
  "visibility": "vip",
  "status": "active"
}
```

### 26. 删除模型（Admin）

**请求**
```
DELETE /api/admin/models/:id
Authorization: Bearer <token>
```

**响应 - 成功 (200)**
```json
{
  "message": "Model deleted successfully"
}
```

### 27. 获取财务统计（Admin）

**请求**
```
GET /api/admin/finance/stats
Authorization: Bearer <token>
```

**响应 - 成功 (200)**
```json
{
  "total_revenue": 125000.00,
  "revenue_today": 3500.00,
  "revenue_this_week": 18500.00,
  "revenue_this_month": 65000.00,
  "total_recharges": 3200,
  "avg_recharge_amount": 85.50
}
```

### 28. 获取充值记录（Admin）

**请求**
```
GET /api/admin/finance/recharges?limit=20&offset=0
Authorization: Bearer <token>

Query Parameters:
- limit: number (default: 20, max: 100)
- offset: number (default: 0)
```

**响应 - 成功 (200)**
```json
{
  "total": 3200,
  "records": [
    {
      "id": 1,
      "user_id": 123,
      "username": "john_doe",
      "amount": 100.00,
      "currency": "CNY",
      "payment_method": "alipay",
      "status": "completed",
      "created_at": "2024-02-16T10:30:00Z",
      "completed_at": "2024-02-16T10:31:00Z"
    }
  ]
}
```

### 29. 获取系统配置（Admin）

**请求**
```
GET /api/admin/config
Authorization: Bearer <token>
```

**响应 - 成功 (200)**
```json
{
  "billing": {
    "gpu_price_per_second": 0.005,
    "storage_price_per_gb_day": 0.02,
    "bandwidth_price_per_gb": 0.10
  },
  "instance_pool": {
    "max_queue_per_instance": 10,
    "health_check_interval_seconds": 30,
    "auto_scale_enabled": false
  },
  "system": {
    "max_upload_size_mb": 2048,
    "allowed_model_types": ["checkpoint", "lora", "vae", "embedding"],
    "maintenance_mode": false
  }
}
```

### 30. 更新系统配置（Admin）

**请求**
```
PATCH /api/admin/config
Authorization: Bearer <token>
Content-Type: application/json

{
  "billing": {
    "gpu_price_per_second": 0.006
  },
  "instance_pool": {
    "auto_scale_enabled": true
  }
}
```

**响应 - 成功 (200)**
返回完整的更新后配置（格式同 GET /api/admin/config）

### 31. 获取系统日志（Admin）

**请求**
```
GET /api/admin/logs?level=error&source=auth&limit=50&offset=0
Authorization: Bearer <token>

Query Parameters:
- level: "info" | "warn" | "error" (可选)
- source: "auth" | "proxy" | "billing" | "system" | "admin" (可选)
- limit: number (default: 50, max: 200)
- offset: number (default: 0)
```

**响应 - 成功 (200)**
```json
{
  "total": 500,
  "logs": [
    {
      "id": 1,
      "level": "error",
      "source": "auth",
      "message": "Login failed: invalid password",
      "user_id": 123,
      "username": "john_doe",
      "created_at": "2024-02-16T10:30:00Z",
      "details": {
        "ip": "192.168.1.1",
        "user_agent": "Mozilla/5.0..."
      }
    }
  ]
}
```

### 常见错误码

| HTTP Status | Code | Description |
|-------------|------|-------------|
| 400 | BAD_REQUEST | 请求参数错误 |
| 401 | UNAUTHORIZED | 未认证或 token 无效 |
| 403 | FORBIDDEN | 无权限访问 |
| 404 | NOT_FOUND | 资源不存在 |
| 409 | CONFLICT | 资源冲突（如用户名已存在） |
| 429 | RATE_LIMIT_EXCEEDED | 请求频率超限 |
| 500 | INTERNAL_ERROR | 服务器内部错误 |

## JWT Token 格式

### Token Payload

```json
{
  "user_id": 123,
  "username": "john_doe",
  "tier": "pro",
  "iat": 1708077000,
  "exp": 1708163400
}
```

### Token 有效期

- 默认有效期：24 小时
- 刷新机制：在 token 过期前 1 小时可刷新
- 过期后需要重新登录

## 数据类型定义

### TierConfig

```typescript
interface TierConfig {
  key: string           // 等级标识，如 "basic"、"pro"、"enterprise"
  label: string         // 显示名称，如 "基础版"
  color: string         // 前端 Badge 样式类
  price: string         // 显示价格，如 "¥99/月" 或 "免费"
  features: string[]    // 功能列表
  popular?: boolean     // 是否标记为推荐
}
```

### User

```typescript
interface User {
  id: number
  username: string
  email: string
  tier: string          // 对应 TierConfig.key
  role: 'user' | 'admin'
  balance: number
  storage_used: number
  storage_limit: number
  created_at: string // ISO 8601
  subscription?: Subscription
}
```

### Subscription

```typescript
interface Subscription {
  tier: string          // 对应 TierConfig.key
  status: 'active' | 'expired' | 'cancelled'
  started_at: string // ISO 8601
  expires_at: string // ISO 8601
  auto_renew: boolean
  features: {
    gpu_access: boolean
    vip_models: boolean
    storage_limit_gb: number
    concurrent_tasks: number
  }
}
```

### UsageRecord

```typescript
interface UsageRecord {
  id: number
  task_id: string
  type: 'gpu_usage' | 'storage' | 'bandwidth'
  started_at: string // ISO 8601
  ended_at: string // ISO 8601
  duration_seconds: number
  cost: number
  details: Record<string, any>
}
```

### PrivateModel

```typescript
interface PrivateModel {
  id: number
  name: string
  type: 'checkpoint' | 'lora' | 'vae' | 'embedding'
  size_bytes: number
  uploaded_at: string // ISO 8601
  storage_cost_per_day: number
}
```

### AdminModel

```typescript
interface AdminModel extends PrivateModel {
  user_id: number
  username: string
  visibility: 'base' | 'vip' | 'private'
  status: 'active' | 'pending' | 'disabled'
}
```

### RechargeRecord

```typescript
interface RechargeRecord {
  id: number
  user_id: number
  username?: string
  amount: number
  currency: string
  payment_method: string
  status: 'pending' | 'completed' | 'failed' | 'refunded'
  created_at: string // ISO 8601
  completed_at?: string // ISO 8601
}
```

### FinanceStats

```typescript
interface FinanceStats {
  total_revenue: number
  revenue_today: number
  revenue_this_week: number
  revenue_this_month: number
  total_recharges: number
  avg_recharge_amount: number
}
```

### SystemConfig

```typescript
interface SystemConfig {
  billing: {
    gpu_price_per_second: number
    storage_price_per_gb_day: number
    bandwidth_price_per_gb: number
  }
  instance_pool: {
    max_queue_per_instance: number
    health_check_interval_seconds: number
    auto_scale_enabled: boolean
  }
  system: {
    max_upload_size_mb: number
    allowed_model_types: string[]
    maintenance_mode: boolean
  }
}
```

### SystemLog

```typescript
interface SystemLog {
  id: number
  level: 'info' | 'warn' | 'error'
  source: string
  message: string
  user_id?: number
  username?: string
  created_at: string // ISO 8601
  details?: Record<string, unknown>
}
```

### AdminUser

```typescript
interface AdminUser {
  id: number
  username: string
  email: string
  tier: string
  balance: number
  status: 'active' | 'suspended' | 'banned'
  role: 'user' | 'admin'
  created_at: string // ISO 8601
}
```

### Instance

```typescript
interface Instance {
  id: string
  url: string
  status: 'online' | 'offline' | 'maintenance'
  gpu_model: string
  vram_total_gb: number
  vram_used_gb: number
  queue_length: number
  uptime_hours: number
  gpu_utilization: number
}
```

## 版本控制

- 当前版本：v1
- API 路径：`/api/v1/*`（可选，当前使用 `/api/*`）
- 版本更新策略：向后兼容，废弃的 API 会提前通知

## 速率限制

- 认证 API：10 次/分钟/IP
- 其他 API：100 次/分钟/用户
- 超限返回 429 状态码

## CORS 配置

- 允许的源：配置文件指定
- 允许的方法：GET, POST, PUT, PATCH, DELETE, OPTIONS
- 允许的 Headers：Authorization, Content-Type
- 凭证支持：是
