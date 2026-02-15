# 文件系统布局设计

## 目录结构

```
/data/
├── models/                          # 共享模型目录（所有用户只读）
│   └── base/
│       ├── checkpoints/
│       │   ├── sd_v1-5.safetensors
│       │   ├── sdxl_base.safetensors
│       │   └── flux_dev.safetensors
│       ├── loras/
│       │   ├── style_lora_1.safetensors
│       │   └── character_lora_2.safetensors
│       ├── vae/
│       │   └── vae-ft-mse-840000.safetensors
│       ├── embeddings/
│       └── controlnet/
│
└── users/                           # 用户数据目录
    ├── 123/                         # 用户 ID 123
    │   ├── output/                  # 生成的图片
    │   │   ├── image_001.png
    │   │   ├── image_002.png
    │   │   └── ...
    │   ├── workflows/               # 保存的工作流
    │   │   ├── my_workflow.json
    │   │   ├── portrait.json
    │   │   └── ...
    │   ├── models/                  # 私有模型
    │   │   ├── my_lora.safetensors
    │   │   └── custom_checkpoint.safetensors
    │   ├── uploads/                 # 上传的图片（用于 img2img）
    │   │   ├── input_001.png
    │   │   └── reference.jpg
    │   └── temp/                    # 临时文件
    │
    ├── 456/                         # 用户 ID 456
    │   └── ...
    │
    └── 789/                         # 用户 ID 789
        └── ...
```

## Docker Volume 映射

### ComfyUI 实例配置

```yaml
# docker-compose.yml
services:
  comfyui-1:
    image: comfyui:latest
    runtime: nvidia
    environment:
      - NVIDIA_VISIBLE_DEVICES=0
    volumes:
      # 共享模型（只读）
      - /data/models/base:/app/models:ro

      # 用户数据（读写）
      - /data/users:/app/users
    ports:
      - "8188:8188"

  comfyui-2:
    image: comfyui:latest
    runtime: nvidia
    environment:
      - NVIDIA_VISIBLE_DEVICES=1
    volumes:
      - /data/models/base:/app/models:ro
      - /data/users:/app/users
    ports:
      - "8189:8188"

  # ... 其他实例
```

### 路径映射说明

| 用户请求路径 | 代理层重写后 | ComfyUI 实际路径 |
|------------|------------|----------------|
| `/output/image.png` | `/users/123/output/image.png` | `/app/users/123/output/image.png` |
| `/workflows/my.json` | `/users/123/workflows/my.json` | `/app/users/123/workflows/my.json` |
| `/models/base/sd_v1-5.safetensors` | 不重写 | `/app/models/base/sd_v1-5.safetensors` |
| `/models/user_123/my_lora.safetensors` | `/users/123/models/my_lora.safetensors` | `/app/users/123/models/my_lora.safetensors` |

## 路径重写规则

### 1. 输出路径（SaveImage 节点）

**用户 workflow：**
```json
{
  "3": {
    "class_type": "SaveImage",
    "inputs": {
      "filename_prefix": "myimage",
      "images": ["2", 0]
    }
  }
}
```

**代理层重写后：**
```json
{
  "3": {
    "class_type": "SaveImage",
    "inputs": {
      "filename_prefix": "users/123/output/myimage",
      "images": ["2", 0]
    }
  }
}
```

**ComfyUI 保存到：** `/app/users/123/output/myimage_00001.png`

### 2. 输入路径（LoadImage 节点）

**用户 workflow：**
```json
{
  "1": {
    "class_type": "LoadImage",
    "inputs": {
      "image": "input.png"
    }
  }
}
```

**代理层重写后：**
```json
{
  "1": {
    "class_type": "LoadImage",
    "inputs": {
      "image": "users/123/uploads/input.png"
    }
  }
}
```

**ComfyUI 读取：** `/app/users/123/uploads/input.png`

### 3. 模型路径

**共享模型（不重写）：**
```json
{
  "inputs": {
    "ckpt_name": "sd_v1-5.safetensors"
  }
}
```
→ ComfyUI 从 `/app/models/base/checkpoints/sd_v1-5.safetensors` 读取

**私有模型（重写）：**
```json
{
  "inputs": {
    "lora_name": "user_123/my_lora.safetensors"
  }
}
```
→ 重写为 `users/123/models/my_lora.safetensors`
→ ComfyUI 从 `/app/users/123/models/my_lora.safetensors` 读取

## 权限控制

### 文件系统权限

```bash
# 共享模型（只读）
chmod -R 755 /data/models/base
chown -R root:root /data/models/base

# 用户数据（读写）
chmod -R 755 /data/users
chown -R comfyui:comfyui /data/users

# 每个用户目录
chmod 755 /data/users/123
chown comfyui:comfyui /data/users/123
```

### 代理层权限检查

```go
// 检查用户是否有权访问路径
func checkPathAccess(userID uint, path string) bool {
    // 1. 共享模型：所有人可访问
    if strings.HasPrefix(path, "/models/base/") {
        return true
    }

    // 2. 用户自己的数据：只能访问自己的
    if strings.HasPrefix(path, fmt.Sprintf("/users/%d/", userID)) {
        return true
    }

    // 3. 其他路径：拒绝
    return false
}
```

## 存储配额管理

### 配置

```yaml
# configs/config.yaml
storage:
  user_data_dir: ./data/users
  shared_models_dir: ./data/models
  max_user_storage_gb: 100  # 每个用户最大 100GB
```

### 检查存储使用

```go
// 获取用户存储使用情况
usage, err := userDirService.GetUserStorageUsage(userID)
if usage > cfg.Storage.MaxUserStorageGB {
    return errors.New("storage quota exceeded")
}
```

### 清理策略

1. **临时文件**：7 天后自动删除
2. **输出图片**：用户可手动删除
3. **工作流**：永久保存
4. **私有模型**：永久保存

## 初始化流程

### 用户注册时

```go
// 1. 创建用户记录
user := CreateUser(...)

// 2. 初始化用户目录
userDirService.InitializeUserDirectory(user.ID)

// 目录结构：
// /data/users/123/
//   ├── output/
//   ├── workflows/
//   ├── models/
//   ├── uploads/
//   └── temp/
```

### 首次使用时

```bash
# 用户首次访问 ComfyUI
POST /api/user/init-directory
Authorization: Bearer <token>

# 响应
{
  "message": "User directory initialized",
  "paths": {
    "output": "/users/123/output",
    "workflows": "/users/123/workflows",
    "models": "/users/123/models",
    "uploads": "/users/123/uploads"
  }
}
```

## 备份策略

### 用户数据备份

```bash
# 每日备份用户数据
rsync -av /data/users/ /backup/users/$(date +%Y%m%d)/

# 保留最近 30 天的备份
find /backup/users/ -type d -mtime +30 -exec rm -rf {} \;
```

### 共享模型备份

```bash
# 共享模型不需要频繁备份（只在更新时备份）
rsync -av /data/models/ /backup/models/
```

## 性能优化

### 1. 使用 SSD

```bash
# 用户数据使用 SSD（高 IOPS）
/data/users → SSD

# 共享模型可以使用 HDD（大容量）
/data/models → HDD
```

### 2. 文件系统选择

- **推荐**：XFS 或 ext4
- **不推荐**：NTFS（性能差）

### 3. 缓存策略

```bash
# 增加文件系统缓存
echo 'vm.vfs_cache_pressure=50' >> /etc/sysctl.conf
sysctl -p
```

## 监控

### 存储使用监控

```bash
# 监控每个用户的存储使用
du -sh /data/users/*

# 监控总存储使用
df -h /data
```

### 告警阈值

- 单用户存储 > 90GB：发送警告
- 单用户存储 > 100GB：禁止上传
- 总存储 > 80%：扩容告警
