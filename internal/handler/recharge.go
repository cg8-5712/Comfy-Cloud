package handler

import (
	"comfy-cloud/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type RechargeHandler struct {
	rechargeService *service.RechargeService
}

func NewRechargeHandler(rechargeService *service.RechargeService) *RechargeHandler {
	return &RechargeHandler{
		rechargeService: rechargeService,
	}
}

// CreateRecharge 创建充值订单
// POST /api/recharge
func (h *RechargeHandler) CreateRecharge(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req struct {
		Amount        float64 `json:"amount" binding:"required,gt=0"`
		PaymentMethod string  `json:"payment_method" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := h.rechargeService.CreateRecharge(userID, req.Amount, req.PaymentMethod)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, order)
}

// GetRechargeHistory 获取充值记录
// GET /api/recharge/history
func (h *RechargeHandler) GetRechargeHistory(c *gin.Context) {
	userID := c.GetUint("user_id")

	limit := 20
	offset := 0
	if l := c.Query("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 && val <= 100 {
			limit = val
		}
	}
	if o := c.Query("offset"); o != "" {
		if val, err := strconv.Atoi(o); err == nil && val >= 0 {
			offset = val
		}
	}

	records, total, err := h.rechargeService.GetRechargeHistory(userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"records": records,
		"total":   total,
	})
}

// SetupRoutes 设置路由
func (h *RechargeHandler) SetupRoutes(r *gin.Engine, authMiddleware gin.HandlerFunc) {
	recharge := r.Group("/api/recharge")
	recharge.Use(authMiddleware)
	{
		recharge.POST("", h.CreateRecharge)
		recharge.GET("/history", h.GetRechargeHistory)
	}
}
