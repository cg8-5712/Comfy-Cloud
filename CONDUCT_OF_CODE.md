# Comfy-Cloud 开发路线

## 项目架构说明 🏗️

### 系统组成

本项目包含三个独立的系统：

1. **Go 后端服务** - 提供 API、反向代理、认证、计费等核心功能
2. **管理平台前端** - 独立的 Web 应用，负责用户登录、账户管理、Admin 后台
3. **ComfyUI 前端** - 修改后的官方 ComfyUI 前端，保持 100% 原生功能，仅添加认证和用户信息显示

### 用户流程

```
用户访问 ComfyUI
    ↓
检测到未登录
    ↓
重定向到管理平台登录页
    ↓
用户在管理平台登录
    ↓
登录成功，存储 token
    ↓
跳转回 ComfyUI
    ↓
ComfyUI 自动读取 token，注入到所有请求
    ↓
用户使用 ComfyUI（右上角显示用户信息）
    ↓
点击"我的账户" → 跳转回管理平台
```

---

## 已完成 ✅

### Phase 0: 数据库设计和基础架构
- [x] 标准 Go 项目结构（handler/service/repository）
- [x] 数据库模型（User, Subscription, UsageRecord, ModelPermission, BillingConfig, DedicatedPricing, DedicatedInstance）
- [x] 用户认证（注册/登录/JWT）
- [x] 配置管理（Viper）
- [x] 日志系统（Zap）
- [x] 计费服务（按 billing.md 实现）
- [x] 使用记录仓储

### Phase 1: 反向代理基础
- [x] 实现反向代理中间件（httputil.ReverseProxy）
- [x] 配置 ComfyUI 实例池
- [x] 简单的负载均衡（选择队列最短的实例）
- [x] 健康检查（定期检查实例状态）
- [x] 子域名路由（独占模式支持）

### Phase 2: 路径重写和数据隔离
- [x] 实现路径重写中间件
  - `/output/` → `/users/{user_id}/output/`
  - `/workflows/` → `/users/{user_id}/workflows/`
  - `/upload/` → `/users/{user_id}/upload/`
  - Workflow JSON 路径重写
- [x] 文件系统布局设计
- [x] 用户目录初始化服务

### Phase 2.5: ComfyUI 前端集成 ✅

**重要说明**：ComfyUI 前端只做最小化修改，保持 100% 原生功能

#### 核心认证（已完成）
- [x] 修改 Distribution 类型定义（添加 `comfy-cloud`）
- [x] 创建 TypeScript 类型定义（`comfyCloudTypes.ts` - 110 行）
- [x] 创建认证 Store（`comfyCloudAuthStore.ts` - 300 行）
- [x] 修改 API 核心文件（`api.ts` - HTTP 和 WebSocket 认证注入）
- [x] 创建 i18n 翻译文本（`comfy-cloud-i18n-patch.json` - 60 行）
- [x] 创建详细的修改文档（`COMFY_FRONTEND_MODIFICATIONS.md`）

#### UI 组件（已完成）
- [x] 创建用户信息组件（`ComfyCloudUserButton.vue` - 180 行）
  - 可拖动的浮动组件（位置保存到 localStorage）
  - 显示用户名、余额、订阅等级
  - 下拉菜单：余额详情、"我的账户"链接、"退出登录"按钮
  - 每 30 秒自动刷新余额
  - 余额不足时红色警告动画
  - 拖动手柄，支持鼠标和触摸操作

- [x] 添加路由守卫（`router.ts` - 30 行）
  - 全局认证检查
  - 未登录用户重定向到管理平台
  - 等待认证初始化（10 秒超时）

- [x] 主应用集成（`App.vue` + `main.ts` - 40 行）
  - App.vue 添加 ComfyCloudUserButton 组件
  - main.ts 初始化认证 store（修复：确保使用同一个 pinia 实例）
  - 等待认证完成后再渲染应用

- [x] 构建配置（`vite.config.mts` + `package.json` - 6 行）
  - 添加 `comfy-cloud` distribution 支持
  - 新增 `build:comfy-cloud` 构建脚本

#### 代码统计
- 核心认证：470 行
- UI 组件：256 行
- **ComfyUI 前端总计：约 726 行**

#### 相关文档
- `COMFY_FRONTEND_MODIFICATIONS.md` - 详细修改说明和迁移指南
- `PHASE_2.5_COMPLETION_SUMMARY.md` - 完成总结和测试清单
- `WEBSOCKET_AUTH_OPTIONS.md` - WebSocket 认证方案说明

---

## 待开发

---

### Phase 2.6: 管理平台前端开发 🌐

**重要说明**：这是一个独立的前端项目，与 ComfyUI 前端完全分离

#### 技术栈选择
- 框架：React / Vue / Next.js / Nuxt（自选）
- UI 库：Ant Design / Element Plus / Tailwind CSS（自选）
- 状态管理：Redux / Pinia / Zustand（自选）

#### 2.6.1 认证页面
- [ ] 登录页面 `/login`
  - 用户名/密码登录
  - 记住我选项
  - 忘记密码链接
  - 调用 `POST /api/auth/login`
  - 登录成功后：
    - 存储 token 到 localStorage（key: `comfy_cloud_token`）
    - 跳转到 ComfyUI：`window.location.href = 'https://comfyui.your-domain.com'`

- [ ] 注册页面 `/register`
  - 用户名、邮箱、密码
  - 密码确认
  - 调用 `POST /api/auth/register`

- [ ] 忘记密码页面 `/forgot-password`（可选）

#### 2.6.2 账户管理页面
- [ ] 账户概览 `/account/dashboard`
  - 当前余额
  - 订阅状态和到期时间
  - 本月使用统计（GPU 时长、存储用量、费用）
  - 快捷操作按钮（充值、升级订阅、跳转到 ComfyUI）

- [ ] 充值页面 `/account/recharge`
  - 充值金额选择（预设金额 + 自定义）
  - 支付方式选择（支付宝/微信/信用卡）
  - 调用 `POST /api/recharge`
  - 充值记录列表

- [ ] 使用记录 `/account/usage`
  - 使用记录列表（时间、任务类型、GPU 时长、费用）
  - 日期范围筛选
  - 导出 CSV
  - 调用 `GET /api/usage/records`

- [ ] 订阅管理 `/account/subscription`
  - 当前订阅详情
  - 订阅计划对比（Basic/Pro/Enterprise）
  - 升级/降级订阅
  - 订阅历史
  - 调用 `GET /api/subscription` 和 `POST /api/subscription/upgrade`

- [ ] 个人设置 `/account/settings`
  - 修改密码
  - 修改邮箱
  - 通知设置
  - API Key 管理（可选）
  - 调用 `GET /api/settings` 和 `PATCH /api/settings`

- [ ] 模型管理 `/account/models`
  - 我的私有模型列表
  - 上传模型（支持拖拽）
  - 删除模型
  - 模型大小和存储费用显示
  - 调用 `GET /api/models/private`、`POST /api/models/upload`、`DELETE /api/models/private/:id`

#### 2.6.3 Admin 管理后台
- [ ] Admin 登录 `/admin/login`
  - 独立的管理员登录页面
  - 验证管理员权限

- [ ] 用户管理 `/admin/users`
  - 用户列表（分页、搜索、筛选）
  - 用户详情（余额、订阅、使用统计）
  - 用户操作（封禁、解封、删除、手动充值、调整订阅）
  - 批量操作

- [ ] 订阅管理 `/admin/subscriptions`
  - 订阅计划配置（价格、配额、功能）
  - 订阅记录查询
  - 手动调整用户订阅

- [ ] 模型管理 `/admin/models`
  - 基础模型管理（上传、删除、标记为 VIP）
  - VIP 模型管理
  - 用户私有模型审核
  - 模型存储统计

- [ ] 实例监控 `/admin/instances` 📊
  - 实例列表和状态（在线/离线、IP、端口、运行时长）
  - 实时使用情况（队列长度、GPU 利用率、显存、CPU、内存）
  - 硬件信息（GPU 型号、显存容量、CUDA 版本、驱动版本）
  - 历史负载曲线（GPU 利用率、显存、队列长度 - 1h/24h/7d/30d）
  - 性能指标（总任务数、平均耗时、成功率）
  - 实例操作（重启、暂停、恢复、查看日志）
  - 告警设置
  - 调用 `/api/admin/instances/*` 系列接口

- [ ] 财务报表 `/admin/finance`
  - 收入统计（日/周/月）
  - 充值记录
  - 消费记录
  - 用户消费排行
  - 导出财务报表

- [ ] 系统配置 `/admin/config`
  - 计费规则配置
  - 实例池配置
  - 系统参数调整

- [ ] 日志查看 `/admin/logs`
  - 系统日志查询
  - 错误日志
  - 用户操作日志
  - 按时间/用户/类型筛选

#### 代码量估算
- 认证页面：400 行
- 账户管理页面：1500 行
- Admin 管理后台：2500 行
- **管理平台前端总计：约 4400 行**

---

### Phase 3: 后端 API 实现

#### 3.1 认证 API（已有基础，需完善）
- [ ] `POST /api/auth/login` - 登录
- [ ] `POST /api/auth/register` - 注册
- [ ] `POST /api/auth/logout` - 登出
- [ ] `POST /api/auth/refresh` - 刷新 token

#### 3.2 用户信息 API
- [ ] `GET /api/user/info` - 获取用户信息
- [ ] `GET /api/user/balance` - 获取余额
- [ ] `GET /api/user/usage` - 获取使用统计

#### 3.3 订阅管理 API
- [ ] `GET /api/subscription` - 获取订阅信息
- [ ] `POST /api/subscription/upgrade` - 升级订阅

#### 3.4 充值 API
- [ ] `POST /api/recharge` - 创建充值订单
- [ ] `GET /api/recharge/history` - 充值记录

#### 3.5 使用记录 API
- [ ] `GET /api/usage/records` - 使用记录列表

#### 3.6 模型管理 API
- [ ] `GET /api/models/private` - 私有模型列表
- [ ] `POST /api/models/upload` - 上传模型
- [ ] `DELETE /api/models/private/:id` - 删除模型

#### 3.7 设置 API
- [ ] `GET /api/settings` - 获取设置
- [ ] `PATCH /api/settings` - 更新设置
- [ ] `POST /api/settings/password` - 修改密码

#### 3.8 Admin API
- [ ] 用户管理接口
- [ ] 订阅管理接口
- [ ] 模型管理接口
- [ ] 实例监控接口（`/api/admin/instances/*`）
- [ ] 财务报表接口
- [ ] 系统配置接口

**详细 API 规范见**: `API_SPECIFICATION.md`

---

### Phase 4: 模型权限控制
- [ ] 模型访问权限检查中间件
  - 基础模型：所有用户可访问
  - VIP 模型：检查订阅等级
  - 私有模型：检查所有权
- [ ] 模型上传接口
- [ ] 模型列表接口

### Phase 5: 智能调度（模型亲和性）
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

### Phase 6: 计费集成
- [ ] 任务提交时预扣费
- [ ] 任务完成时实际扣费
- [ ] 余额不足拒绝任务
- [ ] 使用记录自动生成

### Phase 7: WebSocket 代理
- [ ] WebSocket 连接代理
- [ ] Token 验证（query 参数）
- [ ] 进度推送
- [ ] 连接管理

### Phase 8: 独占模式
- [ ] 独占实例分配
- [ ] 资源隔离
- [ ] 专属路由
- [ ] 独占模式计费

### Phase 9: 监控和运维
- [ ] 性能监控
  - GPU 利用率
  - 队列长度
  - 响应时间
- [ ] 日志聚合
- [ ] 告警系统
- [ ] 自动扩缩容（可选）
- [ ] 监控数据采集服务
  - 每个 ComfyUI 实例部署监控 Agent
  - 定期上报硬件信息和性能指标
  - 使用 Prometheus + Grafana（可选）
  - 或自建时序数据库（InfluxDB/TimescaleDB）

### Phase 10: 部署
- [ ] Docker Compose 配置（8 个 ComfyUI 实例）
- [ ] 环境变量配置
- [ ] 数据备份策略
- [ ] 生产环境优化
- [ ] HTTPS 配置
- [ ] 域名配置（管理平台 + ComfyUI）

---

## 技术栈

### 后端
- **Go 1.21+** - 主语言
- **Gin** - Web 框架
- **GORM** - ORM
- **PostgreSQL 15** - 数据库
- **JWT** - 认证
- **Viper** - 配置
- **Zap** - 日志
- **Docker** - 容器化

### ComfyUI 前端（修改官方）
- **Vue 3.5+** - 框架
- **TypeScript** - 类型安全
- **Pinia** - 状态管理
- **Tailwind 4** - 样式
- **VueUse** - 工具函数

### 管理平台前端（独立开发）
- 框架：自选（React/Vue/Next.js/Nuxt）
- UI 库：自选（Ant Design/Element Plus/Tailwind）
- 状态管理：自选（Redux/Pinia/Zustand）

---

## 核心功能

1. **多租户共享** - 多用户共享 ComfyUI 实例池
2. **数据隔离** - 路径重写实现用户数据隔离
3. **智能调度** - 模型亲和性缓存，减少加载时间
4. **灵活计费** - 按量计费（GPU + VRAM + 存储）+ 等待折扣
5. **独占模式** - 支持独占 GPU 资源
6. **双前端架构** - 管理平台 + ComfyUI 分离，各司其职

---

## 开发优先级

### 第一阶段：核心功能（必须）
1. ~~**Phase 2.5** - ComfyUI 前端集成~~ ✅ **已完成**
2. **Phase 3** - 后端 API 实现（18 个端点）🔥 **进行中**
3. **Phase 2.6** - 管理平台前端开发（认证 + 账户管理）

### 第二阶段：增强功能（重要）
4. **Phase 4** - 模型权限控制
5. **Phase 5** - 智能调度
6. **Phase 6** - 计费集成
7. **Phase 2.6** - 管理平台前端（Admin 后台）

### 第三阶段：优化和扩展（可选）
8. **Phase 7** - WebSocket 代理
9. **Phase 8** - 独占模式
10. **Phase 9** - 监控和运维
11. **Phase 10** - 部署

---

## 相关文档

- [API 规范](./API_SPECIFICATION.md) - 后端 API 接口定义
- [ComfyUI 前端修改文档](./COMFY_FRONTEND_MODIFICATIONS.md) - 前端修改详细说明
- [Phase 2.5 完成总结](./PHASE_2.5_COMPLETION_SUMMARY.md) - ComfyUI 前端集成完成报告
- [WebSocket 认证方案](./WEBSOCKET_AUTH_OPTIONS.md) - WebSocket 认证技术方案对比
- [数据库设计](./DATABASE_DESIGN.md) - 数据库表结构
- [计费系统](./billing.md) - 计费规则和算法
- [文件系统布局](./FILESYSTEM_LAYOUT.md) - 数据存储结构
