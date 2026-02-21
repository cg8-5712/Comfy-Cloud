# Comfy-Cloud 前端修改总结

本文档记录了为集成 Comfy-Cloud 认证系统对官方 ComfyUI 前端所做的所有修改。

## 修改概览

- **修改策略**: 添加新的 distribution 类型 `comfy-cloud`，最小化侵入性修改
- **总代码量**: 约 450 行核心代码 + UI 组件
- **修改文件数**: 2 个现有文件修改 + 3 个新文件
- **兼容性**: 完全向后兼容，不影响官方的 cloud/desktop/localhost 版本

## 修改清单

### 1. 修改现有文件

#### 1.1 `src/platform/distribution/types.ts`

**修改位置**: 第 6 行和第 18 行

**修改前**:
```typescript
type Distribution = 'desktop' | 'localhost' | 'cloud'

export const isDesktop = DISTRIBUTION === 'desktop'
export const isCloud = DISTRIBUTION === 'cloud'
```

**修改后**:
```typescript
type Distribution = 'desktop' | 'localhost' | 'cloud' | 'comfy-cloud'

export const isDesktop = DISTRIBUTION === 'desktop'
export const isCloud = DISTRIBUTION === 'cloud'
export const isComfyCloud = DISTRIBUTION === 'comfy-cloud'
```

**变更说明**:
- 添加 `'comfy-cloud'` 到 Distribution 联合类型
- 导出 `isComfyCloud` 常量用于条件判断
- 保持向后兼容，不影响现有 distribution

---

#### 1.2 `src/scripts/api.ts`

**修改 1: 导入语句** (第 11 行)

**修改前**:
```typescript
import { isCloud } from '@/platform/distribution/types'
```

**修改后**:
```typescript
import { isCloud, isComfyCloud } from '@/platform/distribution/types'
```

---

**修改 2: 添加 getComfyCloudAuth 方法** (在 getAuthStore 方法后，约第 395 行)

**新增代码**:
```typescript
/**
 * Gets the Comfy-Cloud auth header if available.
 * Returns null for non-comfy-cloud distributions or if user is not logged in.
 * @returns The auth header object, or null
 */
private async getComfyCloudAuth() {
  if (isComfyCloud) {
    try {
      const { useComfyCloudAuthStore } = await import(
        '@/stores/comfyCloudAuthStore'
      )
      const authStore = useComfyCloudAuthStore()
      return authStore.getAuthHeader()
    } catch (error) {
      console.warn('Failed to get Comfy-Cloud auth header:', error)
      return null
    }
  }
  return null
}
```

**插入位置**: 在 `getAuthStore()` 方法之后

---

**修改 3: fetchApi 方法** (约第 418-459 行)

**修改前**:
```typescript
async fetchApi(route: string, options?: RequestInit) {
  const headers: HeadersInit = options?.headers ?? {}

  if (isCloud) {
    await this.waitForAuthInitialization()

    // Get Firebase JWT token if user is logged in
    const getAuthHeaderIfAvailable = async (): Promise<AuthHeader | null> => {
      try {
        const authStore = await this.getAuthStore()
        return authStore ? await authStore.getAuthHeader() : null
      } catch (error) {
        console.warn('Failed to get auth header:', error)
        return null
      }
    }

    const authHeader = await getAuthHeaderIfAvailable()

    if (authHeader) {
      for (const [key, value] of Object.entries(authHeader)) {
        addHeaderEntry(headers, key, value)
      }
    }
  }

  addHeaderEntry(headers, 'Comfy-User', this.user)
  return fetch(this.apiURL(route), {
    cache: 'no-cache',
    ...options,
    headers
  })
}
```

**修改后**:
```typescript
async fetchApi(route: string, options?: RequestInit) {
  const headers: HeadersInit = options?.headers ?? {}

  if (isCloud) {
    await this.waitForAuthInitialization()

    // Get Firebase JWT token if user is logged in
    const getAuthHeaderIfAvailable = async (): Promise<AuthHeader | null> => {
      try {
        const authStore = await this.getAuthStore()
        return authStore ? await authStore.getAuthHeader() : null
      } catch (error) {
        console.warn('Failed to get auth header:', error)
        return null
      }
    }

    const authHeader = await getAuthHeaderIfAvailable()

    if (authHeader) {
      for (const [key, value] of Object.entries(authHeader)) {
        addHeaderEntry(headers, key, value)
      }
    }
  } else if (isComfyCloud) {
    // Get Comfy-Cloud JWT token if user is logged in
    const authHeader = await this.getComfyCloudAuth()

    if (authHeader) {
      for (const [key, value] of Object.entries(authHeader)) {
        addHeaderEntry(headers, key, value)
      }
    }
  }

  addHeaderEntry(headers, 'Comfy-User', this.user)
  return fetch(this.apiURL(route), {
    cache: 'no-cache',
    ...options,
    headers
  })
}
```

**变更说明**:
- 在 `if (isCloud)` 后添加 `else if (isComfyCloud)` 分支
- 调用 `getComfyCloudAuth()` 获取认证 header
- 注入 `Authorization: Bearer <token>` 到请求头

---

**修改 4: createSocket 方法** (约第 561-594 行)

**修改前**:
```typescript
// Get auth token and set cloud params if available
// Uses workspace token (if enabled) or Firebase token
if (isCloud) {
  try {
    const authStore = await this.getAuthStore()
    const authToken = await authStore?.getAuthToken()
    if (authToken) {
      params.set('token', authToken)
    }
  } catch (error) {
    // Continue without auth token if there's an error
    console.warn(
      'Could not get auth token for WebSocket connection:',
      error
    )
  }
}
```

**修改后**:
```typescript
// Get auth token and set cloud params if available
// Uses workspace token (if enabled) or Firebase token
if (isCloud) {
  try {
    const authStore = await this.getAuthStore()
    const authToken = await authStore?.getAuthToken()
    if (authToken) {
      params.set('token', authToken)
    }
  } catch (error) {
    // Continue without auth token if there's an error
    console.warn(
      'Could not get auth token for WebSocket connection:',
      error
    )
  }
} else if (isComfyCloud) {
  // Get Comfy-Cloud JWT token for WebSocket
  try {
    const { useComfyCloudAuthStore } = await import(
      '@/stores/comfyCloudAuthStore'
    )
    const authStore = useComfyCloudAuthStore()
    const authToken = authStore.getAuthToken()
    if (authToken) {
      params.set('token', authToken)
    }
  } catch (error) {
    console.warn(
      'Could not get Comfy-Cloud auth token for WebSocket connection:',
      error
    )
  }
}
```

**变更说明**:
- 在 WebSocket 连接时添加 `isComfyCloud` 分支
- 获取 Comfy-Cloud token 并添加到 URL 查询参数

---

### 2. 新增文件

#### 2.1 `src/types/comfyCloudTypes.ts` (新增 110 行)

**文件用途**: 定义 Comfy-Cloud API 的 TypeScript 类型

**主要类型**:
```typescript
// 用户相关
export interface ComfyCloudUser { ... }
export interface ComfyCloudSubscription { ... }
export interface ComfyCloudBalance { ... }

// 使用记录
export interface ComfyCloudUsageStats { ... }
export interface ComfyCloudUsageRecord { ... }

// 模型管理
export interface ComfyCloudModel { ... }

// 设置
export interface ComfyCloudSettings { ... }

// API 请求/响应
export interface LoginRequest { ... }
export interface LoginResponse { ... }
export interface RegisterRequest { ... }
export interface RegisterResponse { ... }
export interface ApiErrorResponse { ... }
export interface ComfyCloudAuthHeader { ... }
```

**完整代码**: 见文件 `src/types/comfyCloudTypes.ts`

---

#### 2.2 `src/stores/comfyCloudAuthStore.ts` (新增 300 行)

**文件用途**: Comfy-Cloud 认证状态管理 (Pinia Store)

**核心功能**:
- JWT Token 管理（localStorage 持久化）
- 用户登录/注册/登出
- 用户信息和余额获取
- Token 刷新
- 认证 header 提供（HTTP 和 WebSocket）

**主要方法**:
```typescript
// 状态
const token = useLocalStorage<string | null>(STORAGE_KEY, null)
const currentUser = ref<ComfyCloudUser | null>(null)
const balance = ref<ComfyCloudBalance | null>(null)
const isAuthenticated = computed(() => !!currentUser.value && !!token.value)

// 方法
login(username, password): Promise<LoginResponse>
register(username, email, password): Promise<RegisterResponse>
logout(): Promise<void>
fetchUserInfo(): Promise<ComfyCloudUser>
fetchBalance(): Promise<ComfyCloudBalance | null>
refreshToken(): Promise<string | null>
getAuthHeader(): ComfyCloudAuthHeader | null
getAuthToken(): string | null
```

**特性**:
- 使用 VueUse 的 `useLocalStorage` 持久化 token
- 自动监听 token 变化，更新用户状态
- 401 响应自动清除 token 并跳转登录
- 错误处理和 Toast 提示
- 完全类型安全

**完整代码**: 见文件 `src/stores/comfyCloudAuthStore.ts`

---

#### 2.3 `src/locales/en/comfy-cloud-i18n-patch.json` (新增)

**文件用途**: Comfy-Cloud 相关的英文翻译文本

**需要合并到**: `src/locales/en/main.json`

**主要翻译键**:
```json
{
  "auth": {
    "login": { ... },
    "register": { ... },
    "logout": { ... },
    "errors": { ... }
  },
  "user": {
    "balance": "Balance",
    "tier": "Subscription",
    "myAccount": "My Account",
    ...
  },
  "toastMessages": {
    "userNotAuthenticated": "Please log in to continue",
    ...
  }
}
```

**完整内容**: 见文件 `src/locales/en/comfy-cloud-i18n-patch.json`

---

## 构建配置修改（待实施）

### `vite.config.mts`

需要添加 `comfy-cloud` 构建配置：

```typescript
// 在 define 配置中添加
define: {
  __DISTRIBUTION__: JSON.stringify('comfy-cloud'),
  __IS_NIGHTLY__: false
}
```

或者创建单独的构建脚本：

```json
// package.json
{
  "scripts": {
    "build:comfy-cloud": "vite build --mode comfy-cloud"
  }
}
```

---

## 版本更新迁移指南

当 ComfyUI 官方前端更新时，按以下步骤迁移修改：

### 步骤 1: 检查冲突文件

```bash
# 检查我们修改的文件是否有更新
git diff origin/main -- src/platform/distribution/types.ts
git diff origin/main -- src/scripts/api.ts
```

### 步骤 2: 合并 types.ts

1. 打开 `src/platform/distribution/types.ts`
2. 确保 Distribution 类型包含 `'comfy-cloud'`
3. 确保导出 `isComfyCloud` 常量

### 步骤 3: 合并 api.ts

**关键点**:
- 导入 `isComfyCloud`
- 添加 `getComfyCloudAuth()` 方法
- 在 `fetchApi()` 中添加 `else if (isComfyCloud)` 分支
- 在 `createSocket()` 中添加 `else if (isComfyCloud)` 分支

**搜索关键字**:
- 搜索 `if (isCloud)` 找到需要添加 `else if` 的位置
- 确保在 Firebase Auth 逻辑之后添加 Comfy-Cloud 逻辑

### 步骤 4: 复制新增文件

直接复制以下文件到新版本：
- `src/types/comfyCloudTypes.ts`
- `src/stores/comfyCloudAuthStore.ts`
- `src/locales/en/comfy-cloud-i18n-patch.json`（需要合并到 main.json）

### 步骤 5: 测试

```bash
# 安装依赖
pnpm install

# 类型检查
pnpm typecheck

# Lint 检查
pnpm lint

# 构建测试
pnpm build
```

---

## 代码审查检查清单

在版本更新后，确保以下内容正确：

- [ ] `types.ts` 包含 `'comfy-cloud'` 类型
- [ ] `types.ts` 导出 `isComfyCloud` 常量
- [ ] `api.ts` 导入 `isComfyCloud`
- [ ] `api.ts` 包含 `getComfyCloudAuth()` 方法
- [ ] `fetchApi()` 有 `else if (isComfyCloud)` 分支
- [ ] `createSocket()` 有 `else if (isComfyCloud)` 分支
- [ ] `comfyCloudAuthStore.ts` 文件存在且完整
- [ ] `comfyCloudTypes.ts` 文件存在且完整
- [ ] i18n 翻译已合并到 `main.json`
- [ ] 类型检查通过
- [ ] Lint 检查通过
- [ ] 构建成功

---

## 关键设计决策

### 为什么使用新的 Distribution 类型？

1. **语义清晰**: `comfy-cloud` 明确表示这是我们的私有云服务
2. **隔离变更**: 不影响官方的 `cloud`/`desktop`/`localhost` 逻辑
3. **易于维护**: 所有 Comfy-Cloud 相关代码都在 `if (isComfyCloud)` 分支中
4. **向后兼容**: 完全不影响现有功能

### 为什么使用动态 import？

```typescript
const { useComfyCloudAuthStore } = await import('@/stores/comfyCloudAuthStore')
```

**原因**:
1. **按需加载**: 只有 `comfy-cloud` 版本才会加载这个 store
2. **减小打包体积**: 其他 distribution 不会包含这些代码
3. **避免循环依赖**: 动态导入可以避免模块循环引用问题

### 为什么参考 Firebase Auth 的结构？

1. **一致性**: 保持与官方代码风格一致
2. **成熟方案**: Firebase Auth 的架构经过验证
3. **易于理解**: 开发者熟悉的模式
4. **功能完整**: 包含所有必要的认证功能

---

## 常见问题

### Q: 如果官方也添加了新的 distribution 类型怎么办？

A: 只需确保我们的 `'comfy-cloud'` 仍在 Distribution 联合类型中即可。

### Q: 如果 api.ts 的 fetchApi 方法签名改变了？

A: 需要相应调整我们的 `else if (isComfyCloud)` 分支，但核心逻辑（获取 auth header 并注入）不变。

### Q: 如果官方重构了认证系统？

A: 我们的认证系统是独立的，不依赖官方的 Firebase Auth，所以不受影响。只需确保 `fetchApi` 和 `createSocket` 中的注入逻辑仍然有效。

### Q: 如何验证修改是否正确？

A:
1. 类型检查通过：`pnpm typecheck`
2. Lint 通过：`pnpm lint`
3. 构建成功：`pnpm build`
4. 运行时测试：登录功能正常，API 请求带有 Authorization header

---

## 相关文档

- [API 规范](./API_SPECIFICATION.md) - 后端 API 接口定义
- [开发路线](./CONDUCT_OF_CODE.md) - 完整的开发计划
- [ComfyUI 官方文档](https://github.com/comfyanonymous/ComfyUI)

---

## 修改历史

| 日期 | 版本 | 修改内容 | 修改人 |
|------|------|----------|--------|
| 2024-02-16 | 1.0.0 | 初始版本，添加 comfy-cloud distribution 支持 | Claude |
| 2026-02-21 | 1.1.0 | 多用户隔离：后端自动注册、前端路由守卫、role 权限 | Claude |

---

## 联系方式

如有问题，请参考：
- 项目 README
- API 规范文档
- 开发路线文档

---

## v1.1.0 新增修改（多用户隔离）

### 3. ComfyUI 后端修改 (`comfy/`)

#### 3.1 `app/user_manager.py`

**修改位置**: `get_request_user_id` 方法内，`SYSTEM_USER_PREFIX` 检查之后、`if user not in self.users: raise KeyError` 之前。

**目的**: `--multi-user` 模式下，Go 代理传来的 `Comfy-User` header 中的用户自动注册到 `users.json`，无需手动调 `POST /users`。

**新增代码**:
```python
            # Auto-register proxy users (Comfy-Cloud integration)
            if user and user not in self.users:
                self.users[user] = user
                try:
                    with open(self.get_users_file(), "w") as f:
                        json.dump(self.users, f)
                except Exception as e:
                    logging.warning(f"Failed to save users.json: {e}")
```

**前提条件**: ComfyUI 必须以 `--multi-user` 参数启动。

---

### 4. ComfyUI 前端新增修改 (`comfy-frontend/`)

#### 4.1 `src/types/comfyCloudTypes.ts`

**改动**: `ComfyCloudUser` 接口新增 `role` 字段。

```typescript
export interface ComfyCloudUser {
  id: number
  username: string
  email: string
  role: string          // ← v1.1.0 新增
  tier: SubscriptionTier
  // ...其余不变
}
```

---

#### 4.2 `src/stores/comfyCloudAuthStore.ts`

**改动 1**: 顶部新增 import

```typescript
import { api } from '@/scripts/api'
```

**改动 2**: computed 区域新增 `isAdmin`

```typescript
const isAdmin = computed(() => currentUser.value?.role === 'admin')
```

**改动 3**: watch token 回调中，`fetchUserInfo()` 成功后设置 ComfyUI 用户身份

```typescript
await fetchUserInfo()
// Set ComfyUI user identity (proxy also overrides this header)
if (currentUser.value?.id) {
  api.user = `user_${currentUser.value.id}`
}
isInitialized.value = true
```

**改动 4**: return 对象中导出 `isAdmin`

---

#### 4.3 `src/router.ts`

**改动 1**: GraphView 路由的 `beforeEnter`，`isComfyCloud` 时初始化 userStore 并自动登录，跳过用户选择界面。

> 注意：不能跳过 `userStore.initialize()`，GraphView 内部组件依赖 `userStore.initialized` 为 `true` 才能渲染。

```typescript
beforeEnter: async (_to, _from, next) => {
  const userStore = useUserStore()
  await userStore.initialize()

  // Comfy-Cloud: auto-login with proxy user identity, skip user selection
  if (isComfyCloud) {
    if (userStore.needsLogin) {
      const { useComfyCloudAuthStore } = await import(
        '@/stores/comfyCloudAuthStore'
      )
      const authStore = useComfyCloudAuthStore()
      if (authStore.userId) {
        const uid = `user_${authStore.userId}`
        await userStore.login({ userId: uid, username: uid })
      }
    }
    return next()
  }

  if (userStore.needsLogin) {
    next('/user-select')
  } else {
    next()
  }
}
```

**改动 2**: user-select 路由新增 `beforeEnter`，等待 auth 初始化完成后检查权限，只有 admin 可访问，普通用户重定向回 `/`。

```typescript
{
  path: 'user-select',
  name: 'UserSelectView',
  component: () => import('@/views/UserSelectView.vue'),
  beforeEnter: async (_to, _from, next) => {
    if (isComfyCloud) {
      const { useComfyCloudAuthStore } =
        await import('@/stores/comfyCloudAuthStore')
      const authStore = useComfyCloudAuthStore()

      // Wait for auth initialization
      if (!authStore.isInitialized) {
        const { storeToRefs } = await import('pinia')
        const { until } = await import('@vueuse/core')
        const { isInitialized } = storeToRefs(authStore)
        await until(isInitialized).toBe(true, { timeout: 10_000 })
      }

      return authStore.isAdmin ? next() : next('/')
    }
    next()
  }
}
```

---

#### 4.4 管理平台 URL 默认值修复

三处引用 `VITE_ADMIN_URL` 的地方，默认值从硬编码域名统一改为 `window.location.origin`，确保跳转到当前访问的域名。

**`src/components/user/ComfyCloudUserButton.vue`** (第 83-84 行)

```typescript
// 修改前
const ADMIN_BASE_URL =
  import.meta.env.VITE_ADMIN_URL || 'https://admin.your-domain.com'

// 修改后
const ADMIN_BASE_URL =
  import.meta.env.VITE_ADMIN_URL || window.location.origin
```

**`src/router.ts`** (isComfyCloud 未认证重定向)

```typescript
// 修改前
const adminUrl =
  import.meta.env.VITE_ADMIN_URL || 'https://admin.your-domain.com'

// 修改后
const adminUrl =
  import.meta.env.VITE_ADMIN_URL || window.location.origin
```

**`src/stores/comfyCloudAuthStore.ts`** (token 清除后跳转)

```typescript
// 修改前
const adminUrl = import.meta.env.VITE_ADMIN_URL || ''

// 修改后
const adminUrl =
  import.meta.env.VITE_ADMIN_URL || window.location.origin
```

---

### v1.1.0 代码审查检查清单

- [ ] ComfyUI 以 `--multi-user` 启动
- [ ] `user_manager.py` 包含自动注册逻辑
- [ ] `comfyCloudTypes.ts` 的 `ComfyCloudUser` 包含 `role` 字段
- [ ] `comfyCloudAuthStore.ts` 导出 `isAdmin`
- [ ] `comfyCloudAuthStore.ts` 在 auth 初始化后设置 `api.user`
- [ ] `router.ts` GraphView 初始化 userStore 并自动登录（不能跳过 initialize）
- [ ] `router.ts` user-select 等待 auth 初始化后检查 admin 权限
- [ ] Go 后端 `/api/user/info` 返回 `role` 字段
- [ ] 三处 `VITE_ADMIN_URL` 默认值均为 `window.location.origin`
