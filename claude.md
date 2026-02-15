# Comfy-Cloud 项目上下文

本文档为 Claude AI 助手提供项目背景和开发指南。

## 项目背景

这是一个商业化的 ComfyUI 云托管平台项目。创建者拥有算力和存储资源，希望构建一个 SaaS 服务来出售 ComfyUI 的使用权。

### 核心需求

1. **完整的原生 ComfyUI 体验**
   - 用户必须能使用 100% 原生的 ComfyUI WebUI
   - 不能阉割功能，不能重新开发前端
   - 保持与官方版本同步更新

2. **成本控制**
   - 多用户共享 ComfyUI 实例池
   - 不能每个用户一个独占容器（太贵）
   - 通过代理层实现多租户隔离

3. **数据隔离**
   - 基础模型：所有用户共享（节省存储）
   - 用户数据：图片、工作流、私有模型完全隔离
   - VIP 模型：付费用户可访问

4. **管理功能**
   - 用户管理和认证
   - 灵活的计费系统
   - 团队协作（可选）

## 核心架构方案（重点）

### 设计理念

**通过代理层实现多租户共享，而不是每用户独占实例**

```
传统方案（成本高）：
用户A → 独占容器A (GPU 利用率 20%)
用户B → 独占容器B (GPU 利用率 20%)
用户C → 独占容器C (GPU 利用率 20%)
→ 需要 3 个 GPU，成本高

我们的方案（成本优化）：
用户A ──┐
用户B ──┤→ 代理层（数据隔离）→ 共享实例池（3-5个）
用户C ──┘                        (GPU 利用率 70-80%)
→ 3 个 GPU 可服务 100+ 用户
```

### 完整流程（核心）

```
步骤 1: 用户访问
  用户打开浏览器 → 访问 comfy-cloud.com

步骤 2: 登录认证
  显示登录页 → 用户输入账号密码 → 调用认证 API
  ↓
  认证成功 → 签发 JWT Token → 存储到 localStorage

步骤 3: 进入 ComfyUI
  重定向到 /comfy/ → 加载原生 ComfyUI WebUI
  (用户看到的是 100% 官方界面，无任何修改)

步骤 4: 前端请求注入 Token
  ComfyUI 前端的 api.js 被轻微修改（10-20 行代码）
  所有 HTTP 请求自动添加：Authorization: Bearer <token>
  所有 WebSocket 连接自动添加：?token=<token>

步骤 5: 代理层拦截
  用户的每个请求都先到达代理层
  ↓
  代理层执行：
  a) 验证 Token → 提取 user_id 和 user_tier
  b) 路径重写（根据 user_id）
     /output/image.png → /users/123/output/image.png
     /workflows/my.json → /users/123/workflows/my.json
  c) 模型权限检查
     - 基础模型：所有人可访问
     - VIP 模型：检查 user_tier
     - 私有模型：检查所有权
  d) 负载均衡
     - 查询各实例队列长度
     - 选择队列最短的实例
     - 转发请求

步骤 6: ComfyUI 处理
  请求到达 ComfyUI 实例（共享的）
  ↓
  ComfyUI 读取文件时，使用代理层重写后的路径
  - 读取 /data/users/123/workflows/my.json
  - 加载 /data/models/base/sd_v1.5.safetensors (共享)
  - 加载 /data/users/123/models/custom.safetensors (私有)
  ↓
  任务进入队列，等待 GPU 执行（串行）

步骤 7: 返回结果
  生成的图片保存到 /data/users/123/output/
  ↓
  通过 WebSocket 推送进度给用户
  ↓
  用户在 ComfyUI WebUI 中看到结果

步骤 8: 计费
  后台记录用户的使用时长/任务数
  ↓
  根据计费模式扣费
```

## 技术实现细节

### 1. ComfyUI 前端修改（最小化）

**文件：comfyui/web/scripts/api.js**

```javascript
// 原始代码
class ComfyApi {
    async fetchApi(route, options = {}) {
        return fetch(route, options);
    }
}

// 修改后（添加约 15 行）
class ComfyApi {
    constructor() {
        // 从 localStorage 读取 token
        this.token = localStorage.getItem('comfy_token');
    }

    async fetchApi(route, options = {}) {
        // 注入 Authorization header
        if (!options.headers) {
            options.headers = {};
        }
        if (this.token) {
            options.headers['Authorization'] = `Bearer ${this.token}`;
        }
        return fetch(route, options);
    }

    // WebSocket 连接也要带 token
    socket = new WebSocket(`ws://...?token=${this.token}`);
}
```

**就这么简单！前端只需要改这一个文件。**

### 2. 登录页（新增）

**文件：web/login.html**

```html
<!DOCTYPE html>
<html>
<head>
    <title>Comfy Cloud - 登录</title>
</head>
<body>
    <h1>Comfy Cloud</h1>
    <form id="loginForm">
        <input type="text" name="username" placeholder="用户名" required />
        <input type="password" name="password" placeholder="密码" required />
        <button type="submit">登录</button>
    </form>

    <script>
    document.getElementById('loginForm').onsubmit = async (e) => {
        e.preventDefault();
        const formData = new FormData(e.target);

        // 调用认证 API
        const res = await fetch('/api/auth/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                username: formData.get('username'),
                password: formData.get('password')
            })
        });

        if (res.ok) {
            const { token } = await res.json();
            // 存储 token
            localStorage.setItem('comfy_token', token);
            // 跳转到 ComfyUI
            window.location.href = '/comfy/';
        } else {
            alert('登录失败');
        }
    };
    </script>
</body>
</html>
```

### 3. 认证服务（Node.js 示例）

**文件：auth-service/src/routes/auth.js**

```javascript
const express = require('express');
const jwt = require('jsonwebtoken');
const bcrypt = require('bcrypt');
const router = express.Router();

// 登录
router.post('/login', async (req, res) => {
    const { username, password } = req.body;

    // 从数据库查询用户
    const user = await db.query('SELECT * FROM users WHERE username = $1', [username]);

    if (!user || !bcrypt.compareSync(password, user.password_hash)) {
        return res.status(401).json({ error: 'Invalid credentials' });
    }

    // 签发 JWT Token
    const token = jwt.sign(
        {
            userId: user.id,
            username: user.username,
            tier: user.tier  // basic/pro/enterprise
        },
        process.env.JWT_SECRET,
        { expiresIn: '24h' }
    );

    res.json({ token });
});

module.exports = router;
```

### 4. 代理层（核心，Node.js 示例）

**文件：proxy-layer/src/proxy.js**

```javascript
const express = require('express');
const { createProxyMiddleware } = require('http-proxy-middleware');
const jwt = require('jsonwebtoken');

const app = express();

// ComfyUI 实例池
const comfyInstances = [
    { url: 'http://comfyui-1:8188', queueSize: 0 },
    { url: 'http://comfyui-2:8188', queueSize: 0 },
    { url: 'http://comfyui-3:8188', queueSize: 0 }
];

// 定期更新队列长度
setInterval(async () => {
    for (let instance of comfyInstances) {
        try {
            const res = await fetch(`${instance.url}/queue`);
            const data = await res.json();
            instance.queueSize = data.queue_pending.length;
        } catch (err) {
            console.error(`Failed to fetch queue from ${instance.url}`);
        }
    }
}, 2000);

// 1. Token 验证中间件
function authMiddleware(req, res, next) {
    // 从 header 或 query 获取 token
    const token = req.headers.authorization?.replace('Bearer ', '')
                  || req.query.token;

    if (!token) {
        return res.status(401).json({ error: 'No token provided' });
    }

    try {
        const decoded = jwt.verify(token, process.env.JWT_SECRET);
        req.userId = decoded.userId;
        req.userTier = decoded.tier;
        next();
    } catch (err) {
        return res.status(401).json({ error: 'Invalid token' });
    }
}

// 2. 路径重写中间件
function pathRewriteMiddleware(req, res, next) {
    const userId = req.userId;

    // 重写输出路径
    if (req.url.includes('/output/')) {
        req.url = req.url.replace('/output/', `/users/${userId}/output/`);
    }

    // 重写工作流路径
    if (req.url.includes('/workflows/')) {
        req.url = req.url.replace('/workflows/', `/users/${userId}/workflows/`);
    }

    // 重写上传路径
    if (req.url.includes('/upload/')) {
        req.url = req.url.replace('/upload/', `/users/${userId}/upload/`);
    }

    next();
}

// 3. 模型权限检查中间件
function modelAccessMiddleware(req, res, next) {
    if (req.url.includes('/models/')) {
        const modelPath = req.url.split('/models/')[1];

        // 基础模型：所有人可访问
        if (modelPath.startsWith('base/')) {
            return next();
        }

        // 私有模型：检查所有权
        if (modelPath.startsWith(`user_${req.userId}/`)) {
            return next();
        }

        // VIP 模型：检查等级
        if (modelPath.startsWith('vip/')) {
            if (req.userTier === 'pro' || req.userTier === 'enterprise') {
                return next();
            }
            return res.status(403).json({ error: 'VIP models require Pro subscription' });
        }

        return res.status(403).json({ error: 'No access to this model' });
    }

    next();
}

// 4. 负载均衡：选择队列最短的实例
function selectInstance() {
    return comfyInstances.reduce((min, instance) =>
        instance.queueSize < min.queueSize ? instance : min
    );
}

// 代理到 ComfyUI
app.use('/comfy',
    authMiddleware,
    pathRewriteMiddleware,
    modelAccessMiddleware,
    (req, res, next) => {
        // 选择实例
        const instance = selectInstance();
        req.comfyTarget = instance.url;
        next();
    },
    createProxyMiddleware({
        target: (req) => req.comfyTarget,
        changeOrigin: true,
        ws: true,  // 支持 WebSocket
        pathRewrite: {
            '^/comfy': ''  // 移除 /comfy 前缀
        },
        onProxyReq: (proxyReq, req) => {
            // 注入用户信息到 header（可选，ComfyUI 后端可以读取）
            proxyReq.setHeader('X-User-Id', req.userId);
            proxyReq.setHeader('X-User-Tier', req.userTier);
        }
    })
);

app.listen(3000, () => {
    console.log('Proxy layer running on port 3000');
});
```

### 5. 文件系统布局

```bash
/data/
├── models/
│   ├── base/                    # 所有用户共享（只读）
│   │   ├── checkpoints/
│   │   │   ├── sd_v1.5.safetensors
│   │   │   ├── sdxl_base.safetensors
│   │   │   └── ...
│   │   ├── loras/
│   │   ├── vae/
│   │   └── embeddings/
│   │
│   └── vip/                     # VIP 用户可访问（只读）
│       ├── premium_model_1.safetensors
│       └── premium_model_2.safetensors
│
└── users/
    ├── 123/                     # 用户 123 的数据
    │   ├── outputs/            # 生成的图片
    │   │   ├── image_001.png
    │   │   └── image_002.png
    │   ├── workflows/          # 保存的工作流
    │   │   ├── my_workflow.json
    │   │   └── portrait.json
    │   ├── models/             # 私有模型
    │   │   └── custom_lora.safetensors
    │   └── uploads/            # 上传的图片
    │       └── input.png
    │
    ├── 456/                     # 用户 456 的数据
    │   └── ...
    │
    └── 789/                     # 用户 789 的数据
        └── ...
```

### 6. Docker Compose 配置

**文件：docker/docker-compose.yml**

```yaml
version: '3.8'

services:
  # 认证服务
  auth-service:
    build: ../auth-service
    ports:
      - "3001:3000"
    environment:
      - JWT_SECRET=your-secret-key
      - DATABASE_URL=postgresql://user:pass@postgres:5432/comfy_cloud
    depends_on:
      - postgres

  # 代理层
  proxy-layer:
    build: ../proxy-layer
    ports:
      - "80:3000"
    environment:
      - JWT_SECRET=your-secret-key
    depends_on:
      - comfyui-1
      - comfyui-2
      - comfyui-3

  # ComfyUI 实例 1
  comfyui-1:
    build: ./comfyui
    runtime: nvidia
    environment:
      - NVIDIA_VISIBLE_DEVICES=0
    volumes:
      - /data/models/base:/app/models:ro
      - /data/models/vip:/app/models/vip:ro
      - /data/users:/app/users

  # ComfyUI 实例 2
  comfyui-2:
    build: ./comfyui
    runtime: nvidia
    environment:
      - NVIDIA_VISIBLE_DEVICES=1
    volumes:
      - /data/models/base:/app/models:ro
      - /data/models/vip:/app/models/vip:ro
      - /data/users:/app/users

  # ComfyUI 实例 3
  comfyui-3:
    build: ./comfyui
    runtime: nvidia
    environment:
      - NVIDIA_VISIBLE_DEVICES=2
    volumes:
      - /data/models/base:/app/models:ro
      - /data/models/vip:/app/models/vip:ro
      - /data/users:/app/users

  # PostgreSQL
  postgres:
    image: postgres:15
    environment:
      - POSTGRES_DB=comfy_cloud
      - POSTGRES_USER=user
      - POSTGRES_PASSWORD=pass
    volumes:
      - postgres_data:/var/lib/postgresql/data

  # Redis（可选，用于缓存）
  redis:
    image: redis:7
    ports:
      - "6379:6379"

volumes:
  postgres_data:
```

## 关键技术点

### ComfyUI 任务处理机制

**重要：一个 ComfyUI 实例是串行处理，不是并发！**

```python
# ComfyUI 内部逻辑（简化）
class PromptQueue:
    def __init__(self):
        self.queue = []  # 任务队列

    def put(self, prompt):
        self.queue.append(prompt)  # 多个用户可以提交

    def execute(self):
        while True:
            if len(self.queue) > 0:
                prompt = self.queue.pop(0)
                # 执行任务（占用 GPU，串行）
                execute_prompt(prompt)  # 这里会阻塞，直到完成
```

**为什么不能并发？**
- 一个 SD 1.5 任务：约 4-6GB 显存
- 一个 SDXL 任务：约 8-12GB 显存
- RTX 4090 (24GB) 理论上可以跑 2 个 SDXL
- 但实际上：模型加载开销 + 中间激活值 + 显存碎片 → 容易 OOM

**解决方案：**
- 部署 3-5 个 ComfyUI 实例（每个独占一个 GPU）
- 代理层做负载均衡，选择队列最短的实例
- 这样可以"伪并发"，提高整体吞吐量

### 容量规划

**单实例容量：**
- 平均任务耗时：30 秒
- 每小时处理：120 个任务
- 队列长度限制：10 个任务（超过则提示用户稍后再试）

**3 实例容量：**
- 每小时处理：360 个任务
- 同时服务：30 个排队用户
- 峰值承载：可以应对短时间的流量高峰

**成本对比：**
```
独占模式（每用户一个容器）：
- 100 个用户 = 100 个容器 = 100 个 GPU
- 成本：¥1,500,000 (100 x ¥15,000)
- GPU 利用率：20%

共享模式（我们的方案）：
- 100 个用户 = 5 个容器 = 5 个 GPU
- 成本：¥75,000 (5 x ¥15,000)
- GPU 利用率：70-80%

节省成本：95% ！
```

## 开发优先级

### Phase 1: MVP（2-3 周）
1. 认证服务（登录/注册/JWT）
2. 代理层（Token 验证 + 路径重写）
3. 修改 ComfyUI 前端（注入 Token）
4. 部署 1 个 ComfyUI 实例
5. 基础计费（余额扣费）

**目标：** 能让用户登录后使用 ComfyUI，数据隔离

### Phase 2: 多实例（1-2 周）
1. 部署 3 个 ComfyUI 实例
2. 负载均衡（选择队列最短的）
3. 队列状态监控
4. 显示等待时间

**目标：** 提高并发能力，优化用户体验

### Phase 3: 增强功能（2-3 周）
1. 模型权限控制（基础/VIP/私有）
2. 工作流市场
3. 团队协作
4. 详细的使用统计

**目标：** 商业化功能完善

### Phase 4: 优化和扩展（持续）
1. ComfyUI 自动更新系统
2. 性能优化（模型预加载、缓存）
3. 监控和告警
4. 自动扩缩容

## 常见问题

### Q: 用户能看到其他用户的数据吗？
A: 不能。代理层会根据 user_id 重写路径，每个用户只能访问自己的目录。

### Q: 如果用户恶意修改 Token 怎么办？
A: JWT Token 是签名的，无法伪造。如果修改，验证会失败。

### Q: 多个用户同时提交任务会怎样？
A: 任务会进入队列，串行执行。代理层会选择队列最短的实例，实现负载均衡。

### Q: 如何防止用户滥用？
A:
- Rate limiting（限制请求频率）
- 队列长度限制（超过则拒绝）
- 余额不足则拒绝
- 异常检测和封禁

### Q: ComfyUI 更新会破坏现有功能吗？
A: 可能。建议：
- 先在测试环境验证
- 灰度发布（先更新 1 个实例）
- 保留旧版本镜像，支持回滚

### Q: 如何处理 GPU 资源不足？
A:
- 任务排队
- 显示预计等待时间
- 提示用户稍后再试
- 自动扩容（云 GPU）

## 下一步行动

1. **搭建开发环境** - Docker Compose 本地环境
2. **实现认证服务** - 用户登录和 JWT
3. **实现代理层** - Token 验证和路径重写
4. **修改 ComfyUI 前端** - 注入 Token
5. **测试端到端流程** - 登录 → 使用 ComfyUI → 数据隔离
6. **部署多实例** - 负载均衡
7. **添加计费功能** - 按时长/按次数
8. **上线运营** - 监控、客服、迭代

## 参考资源

- ComfyUI 官方: https://github.com/comfyanonymous/ComfyUI
- JWT 文档: https://jwt.io/
- http-proxy-middleware: https://github.com/chimurai/http-proxy-middleware
- Docker 文档: https://docs.docker.com/
- NVIDIA Docker: https://github.com/NVIDIA/nvidia-docker
