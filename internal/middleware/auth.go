package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"comfy-cloud/internal/auth"
)

// AuthMiddleware JWT 认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 提取 Token
		tokenString := extractToken(c)
		if tokenString == "" {
			c.JSON(401, gin.H{"error": "No token provided"})
			c.Abort()
			return
		}

		// 验证 Token
		claims, err := auth.ParseToken(tokenString)
		if err != nil {
			c.JSON(401, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// 注入用户信息到 Context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("user_tier", claims.Tier)
		c.Next()
	}
}

// ComfyAuthMiddleware ComfyUI 代理认证中间件
// 所有请求可选认证：有 token 就解析注入 user_id，没有也放行
// ComfyUI 前端自身的 auth store 负责处理未登录状态
func ComfyAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := extractToken(c)
		if tokenString != "" {
			if claims, err := auth.ParseToken(tokenString); err == nil {
				c.Set("user_id", claims.UserID)
				c.Set("username", claims.Username)
				c.Set("user_tier", claims.Tier)
			}
		}
		c.Next()
	}
}
func ExtractToken(c *gin.Context) string {
	return extractToken(c)
}

func extractToken(c *gin.Context) string {
	// 从 Authorization Header 提取
	bearerToken := c.GetHeader("Authorization")
	if bearerToken != "" {
		// Bearer token 格式: "Bearer <token>"
		parts := strings.Split(bearerToken, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}

	// 从 Query 参数提取（用于 WebSocket）
	token := c.Query("token")
	if token != "" {
		return token
	}

	return ""
}
