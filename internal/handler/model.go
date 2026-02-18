package handler

import (
	"comfy-cloud/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ModelHandler struct {
	modelService *service.ModelService
}

func NewModelHandler(modelService *service.ModelService) *ModelHandler {
	return &ModelHandler{
		modelService: modelService,
	}
}

// GetPrivateModels 获取用户私有模型列表
// GET /api/models/private
func (h *ModelHandler) GetPrivateModels(c *gin.Context) {
	userID := c.GetUint("user_id")

	models, err := h.modelService.GetPrivateModels(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"models": models})
}

// UploadModel 上传模型
// POST /api/models/upload
func (h *ModelHandler) UploadModel(c *gin.Context) {
	userID := c.GetUint("user_id")

	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	// 获取模型类型
	modelType := c.PostForm("type")
	if modelType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "type is required"})
		return
	}

	// 打开文件
	fileData, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer fileData.Close()

	// 上传模型
	model, err := h.modelService.UploadModel(userID, file.Filename, modelType, fileData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, model)
}

// DeleteModel 删除模型
// DELETE /api/models/private/:id
func (h *ModelHandler) DeleteModel(c *gin.Context) {
	userID := c.GetUint("user_id")
	modelID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model id"})
		return
	}

	if err := h.modelService.DeleteModel(userID, uint(modelID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Model deleted successfully"})
}

// SetupRoutes 设置路由
func (h *ModelHandler) SetupRoutes(r *gin.Engine, authMiddleware gin.HandlerFunc) {
	models := r.Group("/api/models")
	models.Use(authMiddleware)
	{
		models.GET("/private", h.GetPrivateModels)
		models.POST("/upload", h.UploadModel)
		models.DELETE("/private/:id", h.DeleteModel)
	}
}
