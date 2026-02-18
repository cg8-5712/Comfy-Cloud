package handler

import (
	"comfy-cloud/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetUserInfo 获取用户完整信息
// GET /api/user/info
func (h *UserHandler) GetUserInfo(c *gin.Context) {
	userID := c.GetUint("user_id")

	info, err := h.userService.GetUserInfo(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, info)
}

// GetUserBalance 获取用户余额
// GET /api/user/balance
func (h *UserHandler) GetUserBalance(c *gin.Context) {
	userID := c.GetUint("user_id")

	balance, err := h.userService.GetUserBalance(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, balance)
}

// GetUsageStats 获取用户使用统计
// GET /api/user/usage
func (h *UserHandler) GetUsageStats(c *gin.Context) {
	userID := c.GetUint("user_id")
	period := c.DefaultQuery("period", "month")

	stats, err := h.userService.GetUserUsage(userID, period)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// SetupRoutes 设置路由
func (h *UserHandler) SetupRoutes(r *gin.Engine, authMiddleware gin.HandlerFunc) {
	user := r.Group("/api/user")
	user.Use(authMiddleware)
	{
		user.GET("/info", h.GetUserInfo)
		user.GET("/balance", h.GetUserBalance)
		user.GET("/usage", h.GetUsageStats)
	}
}
