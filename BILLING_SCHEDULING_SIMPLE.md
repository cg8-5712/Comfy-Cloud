# Comfy-Cloud 计费和调度方案（简化版）

## 一、计费模式

### 按量计费（Pay-as-you-go）

**计费公式：**
```
总费用 = (GPU费用 + 显存费用 + 存储费用) × 等待时间折扣系数

其中：
- GPU费用 = (基础价格 + GPU使用率 × 单位价格) × 使用时长
- 显存费用 = 显存占用(GB) × 使用时长 × 显存单价
- 存储费用 = 存储占用(GB) × 天数 × 存储单价
- 等待时间折扣 = 根据等待时长计算补偿
```

**定价配置：**
```javascript
const PRICING = {
    // GPU 使用（按分钟）
    gpu: {
        base: 0.05,        // ¥0.05/分钟（基础价格）
        perPercent: 0.001  // 每 1% 使用率额外 ¥0.001/分钟
    },

    // 显存使用（按分钟）
    vram: {
        perGB: 0.01        // ¥0.01/GB/分钟
    },

    // 存储使用（按天）
    storage: {
        perGB: 0.01        // ¥0.01/GB/天
    },

    // 等待时间折扣
    waitDiscount: {
        threshold: 60,     // 等待超过 60 秒开始折扣
        rate: 0.01         // 每多等待 10 秒，折扣 1%
    }
};
```

**计费示例：**

**场景 1: 简单任务（SD 1.5）**
```
- GPU 使用率: 60%
- 显存占用: 4GB
- 执行时长: 20 秒
- 等待时长: 30 秒

计算：
GPU 费用 = (0.05 + 60 × 0.001) × (20/60) = ¥0.0367
显存费用 = 0.01 × 4 × (20/60) = ¥0.0133
小计 = ¥0.05
等待折扣 = 1.0 (无折扣，等待 < 60 秒)
总费用 = ¥0.05
```

**场景 2: 复杂任务（SDXL）**
```
- GPU 使用率: 95%
- 显存占用: 12GB
- 执行时长: 60 秒
- 等待时长: 180 秒

计算：
GPU 费用 = (0.05 + 95 × 0.001) × (60/60) = ¥0.145
显存费用 = 0.01 × 12 × (60/60) = ¥0.12
小计 = ¥0.265
等待折扣 = 0.88 (等待 180 秒，折扣 12%)
总费用 = ¥0.233
节省 = ¥0.032
```

**场景 3: 大模型任务（Flux）**
```
- GPU 使用率: 98%
- 显存占用: 20GB
- 执行时长: 120 秒
- 等待时长: 300 秒

计算：
GPU 费用 = (0.05 + 98 × 0.001) × (120/60) = ¥0.296
显存费用 = 0.01 × 20 × (120/60) = ¥0.40
小计 = ¥0.696
等待折扣 = 0.76 (等待 300 秒，折扣 24%)
总费用 = ¥0.529
节省 = ¥0.167
```

**优点：**
- ✅ **公平计费** - 按实际资源使用付费
- ✅ **用户友好** - 等待时间长有补偿
- ✅ **精细化** - GPU、显存、存储分别计费
- ✅ **透明** - 用户可以看到详细的费用明细
- ✅ **灵活** - 可以根据实际情况调整定价

### 独占模式计费

**单卡独占：**
- 按小时：10 元/小时
- 按天：200 元/天
- 按月：5000 元/月

**双卡独占：**
- 按小时：18 元/小时
- 按天：360 元/天
- 按月：9000 元/月

**四卡独占：**
- 按小时：32 元/小时
- 按天：640 元/天
- 按月：16000 元/月

---

## 二、资源模式

### 1. 排队模式（共享池）

```
特点:
- 共享 GPU 资源池
- 按队列排队
- 按时长计费
- 成本低

适用场景:
- 偶尔使用
- 对响应时间要求不高
- 成本敏感用户
```

### 2. 独占模式

```
特点:
- 独占 N 张 GPU
- 无排队，即时响应
- 专属实例
- 成本高

适用场景:
- 频繁使用
- 对响应时间要求高
- 企业用户
```

---

## 三、智能调度策略（核心创新）

### 3.1 模型亲和性调度

**原理：**
```
实例 1 刚处理完 Flux 任务
  ↓
Flux 模型还在显存中（约 10GB）
  ↓
新的 Flux 任务到来
  ↓
优先分配给实例 1（无需重新加载）
  ↓
节省 10-15 秒加载时间 ✅
```

**数据结构：**
```go
type InstanceState struct {
    ID            string
    GPUID         int
    QueueSize     int
    CurrentModel  string      // 当前加载的模型
    LoadedAt      time.Time   // 加载时间
    LastUsedAt    time.Time   // 最后使用时间
    MemoryUsage   int64       // 显存占用（MB）
}
```

**调度算法：**
```go
func SelectInstance(task *Task) *Instance {
    // 1. 优先选择已加载相同模型的实例（权重 50%）
    for _, inst := range instances {
        if inst.CurrentModel == task.Model && inst.QueueSize < MAX_QUEUE {
            return inst  // 命中缓存，优先返回
        }
    }

    // 2. 选择队列最短的实例（权重 30%）
    minQueue := instances[0]
    for _, inst := range instances {
        if inst.QueueSize < minQueue.QueueSize {
            minQueue = inst
        }
    }

    return minQueue
}
```

**性能提升：**
```
无缓存:
- SD 1.5 加载: 3-5 秒
- SDXL 加载: 8-10 秒
- Flux 加载: 10-15 秒

有缓存:
- 加载时间: 0 秒 ✅
- 用户体验提升 30-50%
- GPU 利用率提升 20-30%
```

---

### 3.2 动态实例分配

**场景：Flux 模型请求量突然增大**

```yaml
初始状态:
  - 8 个实例，随机分配任务
  - Flux 请求占比 10%

检测到 Flux 请求激增:
  - Flux 请求占比 > 40%
  - 触发动态调整

调整策略:
  - 将 4 个实例固定为 Flux 专用
  - 4 个实例处理其他模型
  - 持续监控，动态调整
```

**实现逻辑：**
```go
type DynamicAllocator struct {
    instances      []*Instance
    modelStats     map[string]*ModelStats
    checkInterval  time.Duration  // 检查间隔（如 5 分钟）
}

type ModelStats struct {
    ModelName      string
    RequestCount   int      // 总请求数
    LastHour       int      // 最近 1 小时请求数
    Percentage     float64  // 占比
    DedicatedCount int      // 专用实例数
}

func (a *DynamicAllocator) Monitor() {
    ticker := time.NewTicker(5 * time.Minute)
    for range ticker.C {
        a.adjustAllocation()
    }
}

func (a *DynamicAllocator) adjustAllocation() {
    // 1. 统计最近 1 小时的模型请求分布
    stats := a.calculateModelStats()

    // 2. 如果某个模型请求占比 > 40%，增加专用实例
    for model, stat := range stats {
        if stat.Percentage > 0.4 {
            // 计算目标实例数
            targetCount := int(float64(len(a.instances)) * stat.Percentage)

            // 如果需要更多实例
            if targetCount > stat.DedicatedCount {
                a.allocateDedicatedInstances(model, targetCount)
            }
        }
    }
}

func (a *DynamicAllocator) allocateDedicatedInstances(model string, count int) {
    // 选择最适合的实例转为专用
    for i := 0; i < count; i++ {
        // 优先选择已经加载该模型的实例
        inst := a.selectBestInstanceForModel(model)
        inst.DedicatedModel = model
        inst.Priority = HIGH_PRIORITY
    }
}
```

**效果：**
```
场景 1: Flux 请求激增
- 检测到 Flux 占比 60%
- 自动分配 5 个实例专用于 Flux
- Flux 任务平均等待时间减少 50%

场景 2: 请求均衡
- SD 1.5: 30%, SDXL: 40%, Flux: 30%
- 不触发专用分配
- 保持灵活调度
```

---

## 四、部署配置

### 8 卡 V100 资源分配

**方案 1：全部共享池（推荐）**
```yaml
共享池: 8 个实例
  - GPU 0-7
  - 端口 8188-8195
  - 智能调度
  - 模型亲和性
  - 动态分配

优点:
  - 资源利用率最高
  - 灵活性最好
  - 适合多用户场景
```

**方案 2：混合模式**
```yaml
共享池: 6 个实例
  - GPU 0-5
  - 端口 8188-8193

独占池: 2 张卡
  - GPU 6-7
  - 用于独占模式用户

优点:
  - 兼顾灵活性和专属性
  - 可以提供独占服务
```

### Docker Compose 配置

```yaml
version: '3.8'

services:
  # 代理层
  proxy:
    build: .
    ports:
      - "3000:3000"
    environment:
      - JWT_SECRET=your-secret
      - SCHEDULER_MODE=smart
    depends_on:
      - postgres
      - comfyui-1
      - comfyui-2
      - comfyui-3
      - comfyui-4
      - comfyui-5
      - comfyui-6
      - comfyui-7
      - comfyui-8

  # ComfyUI 实例 1-8
  comfyui-1:
    image: comfyui:latest
    runtime: nvidia
    environment:
      - NVIDIA_VISIBLE_DEVICES=0
    ports:
      - "8188:8188"
    volumes:
      - ./data/models:/app/models:ro
      - ./data/users:/app/users

  # ... 重复 comfyui-2 到 comfyui-8
```

### 配置文件

```yaml
# configs/config.yaml
scheduler:
  mode: 'smart'  # 智能调度

  # 模型亲和性
  model_affinity:
    enabled: true
    cache_ttl: 3600  # 缓存有效期（秒）

  # 动态分配
  dynamic_allocation:
    enabled: true
    check_interval: 300  # 检查间隔（秒）
    threshold: 0.4       # 触发阈值（40%）

  # 实例配置
  instances:
    - url: http://comfyui-1:8188
      gpu_id: 0
    - url: http://comfyui-2:8188
      gpu_id: 1
    # ... 8 个实例

billing:
  pay_as_you_go:
    rate_per_minute: 0.05
    pre_charge:
      sd15: 2
      sdxl: 5
      flux: 10

  dedicated:
    single_gpu:
      hourly: 10.00
      daily: 200.00
      monthly: 5000.00
```

---

## 五、API 接口

### 提交任务

```bash
POST /api/task/submit
Authorization: Bearer <token>

{
  "workflow": {...},        # ComfyUI workflow JSON
  "model": "flux",          # sd15/sdxl/flux
  "priority": 0             # 优先级（独占用户可设置）
}

Response:
{
  "task_id": "abc123",
  "estimated_cost": 0.50,   # 预估费用
  "estimated_time": 10,     # 预估时间（分钟）
  "queue_position": 3       # 队列位置
}
```

### 查询任务状态

```bash
GET /api/task/status/:task_id

Response:
{
  "task_id": "abc123",
  "status": "running",      # queued/running/completed/failed
  "progress": 50,           # 进度百分比
  "queue_time": 120,        # 排队时长（秒）
  "execution_time": 60,     # 执行时长（秒）
  "current_cost": 0.15,     # 当前费用
  "instance_id": "comfyui-2"
}
```

### 申请独占模式

```bash
POST /api/dedicated/request

{
  "gpu_count": 1,           # 1/2/4
  "duration": "monthly"     # hourly/daily/monthly
}

Response:
{
  "dedicated_id": "ded123",
  "gpu_ids": [6],
  "instance_urls": ["http://comfyui-7:8188"],
  "cost": 5000.00,
  "expires_at": "2024-02-15T00:00:00Z"
}
```

---

## 六、性能预估

### 8 卡 V100 容量

**排队模式：**
```
并发任务数: 8 个
平均任务时间: 30 秒
每小时处理: 960 个任务
支持并发用户: 100-200 人
GPU 利用率: 70-80%
```

**智能调度优化：**
```
模型缓存命中率: 60-70%
平均加载时间节省: 8 秒
用户体验提升: 30-40%
GPU 利用率提升: 20-30%
```

### 成本分析

**硬件成本：**
```
服务器成本: ¥50,000/月
运维成本: ¥10,000/月
总成本: ¥60,000/月
```

**收入预估（排队模式）：**
```
场景: 200 个活跃用户
平均每用户每月: ¥500
总收入: ¥100,000/月
利润: ¥40,000/月
利润率: 40% ✅
```

**收入预估（混合模式）：**
```
排队用户: 150 人 × ¥400 = ¥60,000
独占用户: 2 人 × ¥5,000 = ¥10,000
总收入: ¥70,000/月
利润: ¥10,000/月
利润率: 14%
```

---

## 七、实现优先级

### Phase 1: 基础计费（1-2 天）
- [ ] 按时长计费数据模型
- [ ] 余额扣费逻辑
- [ ] 预扣费机制
- [ ] 使用记录

### Phase 2: 智能调度（3-4 天）
- [ ] 模型缓存记录
- [ ] 亲和性调度算法
- [ ] 动态实例分配
- [ ] 性能监控

### Phase 3: 独占模式（2-3 天）
- [ ] 独占实例分配
- [ ] 资源隔离
- [ ] 专属路由
- [ ] 独占计费

---

## 八、总结

### 核心优势

1. **简单透明的计费** - 按时长计费，用户容易理解
2. **智能调度** - 模型亲和性，减少加载时间
3. **动态分配** - 根据负载自动调整
4. **灵活选择** - 排队模式 or 独占模式

### 技术亮点

1. **模型缓存** - 节省 10-15 秒加载时间
2. **动态分配** - 自动适应负载变化
3. **资源隔离** - 独占模式保证性能

### 商业价值

1. **成本优化** - 智能调度提升 GPU 利用率 20-30%
2. **用户体验** - 缓存命中减少等待时间 30-40%
3. **灵活定价** - 满足不同用户需求
