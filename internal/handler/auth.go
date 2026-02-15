package handler

import (
	"github.com/gin-gonic/gin"

	"comfy-cloud/internal/auth"
	"comfy-cloud/internal/middleware"
	"comfy-cloud/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register 用户注册
func (h *AuthHandler) Register(c *gin.Context) {
	var req auth.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	user, token, err := h.authService.Register(req)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"token": token,
		"user":  user,
	})
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req auth.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	user, token, err := h.authService.Login(req)
	if err != nil {
		c.JSON(401, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"token": token,
		"user":  user,
	})
}

// Verify 验证 Token
func (h *AuthHandler) Verify(c *gin.Context) {
	userID := c.GetUint("userId")
	username := c.GetString("username")
	userTier := c.GetString("userTier")

	c.JSON(200, gin.H{
		"user_id":  userID,
		"username": username,
		"tier":     userTier,
	})
}

// SetupRoutes 设置认证路由
func (h *AuthHandler) SetupRoutes(r *gin.Engine) {
	authGroup := r.Group("/api/auth")
	{
		authGroup.POST("/register", h.Register)
		authGroup.POST("/login", h.Login)
		authGroup.GET("/verify", middleware.AuthMiddleware(), h.Verify)
	}
}
