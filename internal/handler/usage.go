package handler

import (
	"comfy-cloud/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UsageHandler struct {
	usageService *service.UsageService
}

func NewUsageHandler(usageService *service.UsageService) *UsageHandler {
	return &UsageHandler{
		usageService: usageService,
	}
}

// GetUsageRecords 获取使用记录列表
// GET /api/usage/records
func (h *UsageHandler) GetUsageRecords(c *gin.Context) {
	userID := c.GetUint("user_id")

	// 解析查询参数
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	records, total, err := h.usageService.GetUsageRecords(userID, limit, offset, startDate, endDate)
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
func (h *UsageHandler) SetupRoutes(r *gin.Engine, authMiddleware gin.HandlerFunc) {
	usage := r.Group("/api/usage")
	usage.Use(authMiddleware)
	{
		usage.GET("/records", h.GetUsageRecords)
	}
}
