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
		c.Set("userId", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("userTier", claims.Tier)
		c.Next()
	}
}

// extractToken 从 Header 或 Query 提取 Token
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
