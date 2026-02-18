package handler

import (
	"comfy-cloud/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SubscriptionHandler struct {
	subscriptionService *service.SubscriptionService
}

func NewSubscriptionHandler(subscriptionService *service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{
		subscriptionService: subscriptionService,
	}
}

// GetTiers 获取所有订阅等级配置（公开接口）
// GET /api/tiers
func (h *SubscriptionHandler) GetTiers(c *gin.Context) {
	tiers := h.subscriptionService.GetTiers()
	c.JSON(http.StatusOK, tiers)
}

// GetSubscription 获取用户订阅信息
// GET /api/subscription
func (h *SubscriptionHandler) GetSubscription(c *gin.Context) {
	userID := c.GetUint("user_id")

	subscription, err := h.subscriptionService.GetSubscription(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, subscription)
}

// UpgradeSubscription 升级订阅
// POST /api/subscription/upgrade
func (h *SubscriptionHandler) UpgradeSubscription(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req struct {
		TargetTier string `json:"target_tier" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subscription, err := h.subscriptionService.UpgradeSubscription(userID, req.TargetTier)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, subscription)
}

// SetupRoutes 设置路由
func (h *SubscriptionHandler) SetupRoutes(r *gin.Engine, authMiddleware gin.HandlerFunc) {
	// 公开接口
	r.GET("/api/tiers", h.GetTiers)

	// 需要认证的接口
	subscription := r.Group("/api/subscription")
	subscription.Use(authMiddleware)
	{
		subscription.GET("", h.GetSubscription)
		subscription.POST("/upgrade", h.UpgradeSubscription)
	}
}
