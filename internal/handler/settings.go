package handler

import (
	"comfy-cloud/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SettingsHandler struct {
	settingsService *service.SettingsService
}

func NewSettingsHandler(settingsService *service.SettingsService) *SettingsHandler {
	return &SettingsHandler{
		settingsService: settingsService,
	}
}

// GetSettings 获取用户设置
// GET /api/settings
func (h *SettingsHandler) GetSettings(c *gin.Context) {
	userID := c.GetUint("user_id")

	settings, err := h.settingsService.GetSettings(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, settings)
}

// UpdateSettings 更新用户设置
// PATCH /api/settings
func (h *SettingsHandler) UpdateSettings(c *gin.Context) {
	userID := c.GetUint("user_id")

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.settingsService.UpdateSettings(userID, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Settings updated successfully"})
}

// ChangePassword 修改密码
// POST /api/settings/password
func (h *SettingsHandler) ChangePassword(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.settingsService.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}

// SetupRoutes 设置路由
func (h *SettingsHandler) SetupRoutes(r *gin.Engine, authMiddleware gin.HandlerFunc) {
	settings := r.Group("/api/settings")
	settings.Use(authMiddleware)
	{
		settings.GET("", h.GetSettings)
		settings.PATCH("", h.UpdateSettings)
		settings.POST("/password", h.ChangePassword)
	}
}
