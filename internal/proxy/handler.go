package proxy

import (
	"comfy-cloud/internal/models"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ProxyHandler 代理处理器
type ProxyHandler struct {
	sharedPool         *InstancePool
	dedicatedInstances map[string]*Instance // subdomain -> instance
	db                 *gorm.DB
}

// NewProxyHandler 创建代理处理器
func NewProxyHandler(sharedPool *InstancePool, db *gorm.DB) *ProxyHandler {
	handler := &ProxyHandler{
		sharedPool:         sharedPool,
		dedicatedInstances: make(map[string]*Instance),
		db:                 db,
	}

	// 加载独占实例映射
	handler.loadDedicatedInstances()

	return handler
}

// loadDedicatedInstances 加载独占实例映射
func (h *ProxyHandler) loadDedicatedInstances() {
	var instances []models.DedicatedInstance
	h.db.Where("status = ?", "active").Find(&instances)

	for _, inst := range instances {
		if inst.InstanceIDs != "" {
			// 假设 InstanceIDs 格式为 "comfyui-7:8188"
			h.dedicatedInstances[inst.Subdomain] = &Instance{
				ID:     inst.InstanceIDs,
				URL:    fmt.Sprintf("http://%s", inst.InstanceIDs),
				Status: "healthy",
			}
		}
	}
}

// Route 路由请求
func (h *ProxyHandler) Route(c *gin.Context) {
	host := c.Request.Host

	// 移除端口号（如果有）
	host = strings.Split(host, ":")[0]

	// 判断是共享模式还是独占模式
	if host == "comfy-cloud.com" || host == "localhost" || host == "127.0.0.1" {
		// 共享模式
		h.handleSharedMode(c)
	} else if strings.HasSuffix(host, ".comfy-cloud.com") {
		// 独占模式
		subdomain := strings.TrimSuffix(host, ".comfy-cloud.com")
		h.handleDedicatedMode(c, subdomain)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid domain"})
	}
}

// handleSharedMode 处理共享模式请求
func (h *ProxyHandler) handleSharedMode(c *gin.Context) {
	// 选择实例（负载均衡）
	instance := h.sharedPool.SelectInstance()
	if instance == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "No healthy instances available"})
		return
	}

	// 转发请求
	h.proxyTo(c, instance)
}

// handleDedicatedMode 处理独占模式请求
func (h *ProxyHandler) handleDedicatedMode(c *gin.Context, subdomain string) {
	// 查询专属实例
	instance, exists := h.dedicatedInstances[subdomain]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Dedicated instance not found"})
		return
	}

	// 可选：验证用户是否有权访问这个子域名
	// userID := c.GetUint("user_id")
	// if !h.verifyAccess(subdomain, userID) {
	//     c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
	//     return
	// }

	// 直连到专属实例
	h.proxyTo(c, instance)
}

// proxyTo 转发请求到指定实例
func (h *ProxyHandler) proxyTo(c *gin.Context, instance *Instance) {
	// 解析目标 URL
	target, err := url.Parse(instance.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid target URL"})
		return
	}

	// 创建反向代理
	proxy := httputil.NewSingleHostReverseProxy(target)

	// 自定义 Director（修改请求）
	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Host = target.Host

		// 去掉 /comfy 前缀，同时保留原始编码（%2F、%20 等）
		// RawPath 保留编码字符，Path 是解码后的版本
		// ComfyUI 的 userdata API 依赖 %2F 保持编码
		rawPath := req.URL.RawPath
		if rawPath == "" {
			rawPath = req.URL.Path
		}
		rawPath = strings.TrimPrefix(rawPath, "/comfy")
		if rawPath == "" {
			rawPath = "/"
		}
		req.URL.RawPath = rawPath
		req.URL.Path = strings.TrimPrefix(req.URL.Path, "/comfy")
		if req.URL.Path == "" {
			req.URL.Path = "/"
		}

		req.Header.Set("X-Forwarded-Host", c.Request.Host)
		req.Header.Set("X-Instance-ID", instance.ID)

		// 强制设置 Comfy-User（覆盖前端发来的空值）
		if comfyUser := c.Request.Header.Get("Comfy-User"); comfyUser != "" {
			req.Header.Set("Comfy-User", comfyUser)
		}
	}

	// 错误处理
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		// 标记实例为不健康
		h.sharedPool.UpdateStatus(instance.ID, "unhealthy")

		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(fmt.Sprintf("Proxy error: %v", err)))
	}

	// 执行代理
	proxy.ServeHTTP(c.Writer, c.Request)
}

// verifyAccess 验证用户是否有权访问子域名
func (h *ProxyHandler) verifyAccess(subdomain string, userID uint) bool {
	var instance models.DedicatedInstance
	err := h.db.Where("subdomain = ? AND user_id = ? AND status = ?", subdomain, userID, "active").
		First(&instance).Error
	return err == nil
}

// ReloadDedicatedInstances 重新加载独占实例映射
func (h *ProxyHandler) ReloadDedicatedInstances() {
	h.dedicatedInstances = make(map[string]*Instance)
	h.loadDedicatedInstances()
}
