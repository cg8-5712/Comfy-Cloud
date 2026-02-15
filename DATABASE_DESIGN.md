# 数据库设计文档

## 概述

本文档描述 Comfy-Cloud 项目的数据库设计，使用 PostgreSQL + GORM。

## 技术栈

- **数据库**: PostgreSQL 15+
- **ORM**: GORM v2
- **驱动**: gorm.io/driver/postgres
- **密码加密**: bcrypt

## 数据表设计

### 1. users（用户表）

存储用户基本信息和账户状态。

| 字段 | 类型 | 说明 |
|------|------|------|
| id | SERIAL | 主键 |
| username | VARCHAR(50) | 用户名（唯一） |
| email | VARCHAR(100) | 邮箱（唯一） |
| password_hash | VARCHAR(255) | 密码哈希（bcrypt） |
| tier | VARCHAR(20) | 用户等级：basic/pro/enterprise |
| balance | DECIMAL(10,2) | 账户余额 |
| status | VARCHAR(20) | 账户状态：active/suspended/deleted |
| created_at | TIMESTAMP | 创建时间 |
| updated_at | TIMESTAMP | 更新时间 |
| deleted_at | TIMESTAMP | 软删除时间 |

**索引:**
- `username` (唯一索引)
- `email` (唯一索引)

### 2. subscriptions（订阅表）

存储用户的订阅信息。

| 字段 | 类型 | 说明 |
|------|------|------|
| id | SERIAL | 主键 |
| user_id | INTEGER | 外键 → users.id |
| plan | VARCHAR(20) | 订阅计划：basic/pro/enterprise |
| status | VARCHAR(20) | 订阅状态：active/cancelled/expired |
| started_at | TIMESTAMP | 开始时间 |
| expires_at | TIMESTAMP | 过期时间（NULL 表示永久） |
| created_at | TIMESTAMP | 创建时间 |
| updated_at | TIMESTAMP | 更新时间 |
| deleted_at | TIMESTAMP | 软删除时间 |

**索引:**
- `user_id` (普通索引)

### 3. usage_records（使用记录表）

记录用户的每次使用行为，用于计费和统计。

| 字段 | 类型 | 说明 |
|------|------|------|
| id | SERIAL | 主键 |
| user_id | INTEGER | 外键 → users.id |
| task_type | VARCHAR(50) | 任务类型：prompt/upload/download |
| cost | DECIMAL(10,4) | 本次费用 |
| duration | INTEGER | 耗时（秒） |
| metadata | JSONB | 额外信息（JSON 格式） |
| created_at | TIMESTAMP | 创建时间 |

**索引:**
- `user_id` (普通索引)
- `created_at` (普通索引，用于时间范围查询)

**metadata 示例:**
```json
{
  "model": "sd_v1.5",
  "width": 512,
  "height": 512,
  "steps": 20,
  "instance": "comfyui-1"
}
```

### 4. model_permissions（模型权限表）

记录用户上传的私有模型。

| 字段 | 类型 | 说明 |
|------|------|------|
| id | SERIAL | 主键 |
| user_id | INTEGER | 外键 → users.id |
| model_path | VARCHAR(255) | 模型路径（相对路径） |
| model_name | VARCHAR(100) | 模型名称 |
| model_type | VARCHAR(50) | 模型类型：checkpoint/lora/vae/embedding |
| file_size | BIGINT | 文件大小（字节） |
| created_at | TIMESTAMP | 创建时间 |

**索引:**
- `user_id` (普通索引)
- `(user_id, model_path)` (唯一索引)

## GORM 模型关系

```
User (1) ──< (N) Subscription
User (1) ──< (N) UsageRecord
User (1) ──< (N) ModelPermission
```

## 用户等级说明

| 等级 | 说明 | 权限 |
|------|------|------|
| basic | 基础用户 | 访问基础模型 |
| pro | 专业用户 | 访问基础模型 + VIP 模型 |
| enterprise | 企业用户 | 访问所有模型 + 优先队列 |

## 计费逻辑

### 按次计费
- 每次提交 prompt：0.10 元
- 每次上传文件：0.01 元
- 每次下载文件：0.01 元

### 按时长计费
- 每分钟 GPU 使用：0.05 元

### 扣费流程
1. 用户提交任务
2. 检查余额是否充足
3. 预扣费（冻结金额）
4. 任务完成后，根据实际使用计算费用
5. 扣除实际费用，退还多余金额
6. 记录到 `usage_records` 表

## 数据库迁移

使用 GORM 的 AutoMigrate 功能：

```go
db.AutoMigrate(
    &models.User{},
    &models.Subscription{},
    &models.UsageRecord{},
    &models.ModelPermission{},
)
```

## 性能优化建议

### 1. 索引优化
- 已为常用查询字段添加索引
- 使用 `EXPLAIN ANALYZE` 分析慢查询

### 2. 连接池配置
```go
sqlDB.SetMaxIdleConns(10)      // 最大空闲连接
sqlDB.SetMaxOpenConns(100)     // 最大打开连接
sqlDB.SetConnMaxLifetime(time.Hour)  // 连接最大生命周期
```

### 3. 查询优化
- 使用 `Select()` 只查询需要的字段
- 使用 `Preload()` 预加载关联数据
- 避免 N+1 查询问题

### 4. 分区表（未来优化）
对于 `usage_records` 表，可以按月分区：
```sql
CREATE TABLE usage_records_2024_01 PARTITION OF usage_records
FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');
```

## 备份策略

### 1. 定期备份
```bash
# 每天凌晨 2 点备份
0 2 * * * pg_dump -U comfy comfy_cloud > /backup/comfy_$(date +\%Y\%m\%d).sql
```

### 2. 增量备份
使用 WAL 归档实现增量备份。

### 3. 备份保留
- 每日备份：保留 7 天
- 每周备份：保留 4 周
- 每月备份：保留 12 个月

## 安全建议

1. **密码安全**
   - 使用 bcrypt 加密（cost=12）
   - 永远不要记录明文密码

2. **SQL 注入防护**
   - GORM 自动参数化查询
   - 避免使用 `Raw()` 拼接 SQL

3. **数据加密**
   - 敏感字段考虑加密存储
   - 使用 SSL/TLS 连接数据库

4. **访问控制**
   - 数据库用户最小权限原则
   - 生产环境禁用 `DROP` 权限

## 监控指标

需要监控的关键指标：

1. **连接数**
   - 当前活跃连接
   - 连接池使用率

2. **查询性能**
   - 慢查询（> 1s）
   - 平均查询时间

3. **表大小**
   - `usage_records` 增长速度
   - 定期清理历史数据

4. **锁等待**
   - 死锁检测
   - 长事务监控

## 常用查询示例

### 查询用户余额
```go
var user models.User
db.Select("id", "username", "balance").First(&user, userID)
```

### 查询用户使用记录（最近 30 天）
```go
var records []models.UsageRecord
db.Where("user_id = ? AND created_at > ?", userID, time.Now().AddDate(0, 0, -30)).
   Order("created_at DESC").
   Find(&records)
```

### 统计用户总消费
```go
var totalCost float64
db.Model(&models.UsageRecord{}).
   Where("user_id = ?", userID).
   Select("SUM(cost)").
   Scan(&totalCost)
```

### 查询用户私有模型
```go
var models []models.ModelPermission
db.Where("user_id = ?", userID).Find(&models)
```

## 数据清理策略

### 1. 使用记录归档
- 保留最近 3 个月的详细记录
- 3 个月前的数据归档到冷存储
- 只保留汇总统计

### 2. 软删除清理
- 定期清理 30 天前的软删除记录
```go
db.Unscoped().Where("deleted_at < ?", time.Now().AddDate(0, 0, -30)).Delete(&models.User{})
```

## 扩展性考虑

### 1. 读写分离
- 主库：写操作
- 从库：读操作（统计、报表）

### 2. 分库分表
当单表数据量超过 1000 万时，考虑分表：
- 按用户 ID 哈希分表
- 按时间范围分表

### 3. 缓存层
- Redis 缓存用户信息
- Redis 缓存热点数据
- 减少数据库压力
