package handler

import (
	"comfy-cloud/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	adminService    *service.AdminService
	instanceService *service.InstanceService
	modelService    *service.ModelService
}

func NewAdminHandler(adminService *service.AdminService, instanceService *service.InstanceService, modelService *service.ModelService) *AdminHandler {
	return &AdminHandler{
		adminService:    adminService,
		instanceService: instanceService,
		modelService:    modelService,
	}
}

// AdminMiddleware 管理员权限中间件
func (h *AdminHandler) AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("user_id")

		if err := h.adminService.CheckAdminPermission(userID); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetStats 获取管理统计
// GET /api/admin/stats
func (h *AdminHandler) GetStats(c *gin.Context) {
	stats, err := h.adminService.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetUsers 获取用户列表
// GET /api/admin/users
func (h *AdminHandler) GetUsers(c *gin.Context) {
	search := c.Query("search")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	users, total, err := h.adminService.GetAllUsers(search, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"total": total,
	})
}

// UpdateUser 更新用户
// PATCH /api/admin/users/:id
func (h *AdminHandler) UpdateUser(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.adminService.UpdateUser(uint(userID), updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// GetInstances 获取实例列表
// GET /api/admin/instances
func (h *AdminHandler) GetInstances(c *gin.Context) {
	instances, err := h.instanceService.GetAllInstances()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, instances)
}

// GetModels 获取模型列表
// GET /api/admin/models
func (h *AdminHandler) GetModels(c *gin.Context) {
	visibility := c.Query("visibility")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	models, total, err := h.modelService.GetAllModels(visibility, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"models": models,
		"total":  total,
	})
}

// UpdateModel 更新模型
// PATCH /api/admin/models/:id
func (h *AdminHandler) UpdateModel(c *gin.Context) {
	modelID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model id"})
		return
	}

	var req struct {
		Visibility string `json:"visibility"`
		Status     string `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	model, err := h.modelService.UpdateModel(uint(modelID), req.Visibility, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, model)
}

// DeleteModel 删除模型
// DELETE /api/admin/models/:id
func (h *AdminHandler) DeleteModel(c *gin.Context) {
	modelID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model id"})
		return
	}

	// Admin 可以删除任何模型，传入 userID = 0
	if err := h.modelService.DeleteModel(0, uint(modelID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Model deleted successfully"})
}

// GetFinanceStats 获取财务统计
// GET /api/admin/finance/stats
func (h *AdminHandler) GetFinanceStats(c *gin.Context) {
	stats, err := h.adminService.GetFinanceStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetRechargeRecords 获取充值记录
// GET /api/admin/finance/recharges
func (h *AdminHandler) GetRechargeRecords(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	records, total, err := h.adminService.GetRechargeRecords(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"records": records,
		"total":   total,
	})
}

// GetConfig 获取系统配置
// GET /api/admin/config
func (h *AdminHandler) GetConfig(c *gin.Context) {
	config, err := h.adminService.GetSystemConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, config)
}

// UpdateConfig 更新系统配置
// PATCH /api/admin/config
func (h *AdminHandler) UpdateConfig(c *gin.Context) {
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config, err := h.adminService.UpdateSystemConfig(updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, config)
}

// GetLogs 获取系统日志
// GET /api/admin/logs
func (h *AdminHandler) GetLogs(c *gin.Context) {
	level := c.Query("level")
	source := c.Query("source")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	logs, total, err := h.adminService.GetSystemLogs(level, source, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":  logs,
		"total": total,
	})
}

// SetupRoutes 设置路由
func (h *AdminHandler) SetupRoutes(r *gin.Engine, authMiddleware gin.HandlerFunc) {
	admin := r.Group("/api/admin")
	admin.Use(authMiddleware)
	admin.Use(h.AdminMiddleware())
	{
		admin.GET("/stats", h.GetStats)
		admin.GET("/users", h.GetUsers)
		admin.PATCH("/users/:id", h.UpdateUser)
		admin.GET("/instances", h.GetInstances)
		admin.GET("/models", h.GetModels)
		admin.PATCH("/models/:id", h.UpdateModel)
		admin.DELETE("/models/:id", h.DeleteModel)
		admin.GET("/finance/stats", h.GetFinanceStats)
		admin.GET("/finance/recharges", h.GetRechargeRecords)
		admin.GET("/config", h.GetConfig)
		admin.PATCH("/config", h.UpdateConfig)
		admin.GET("/logs", h.GetLogs)
	}
}
