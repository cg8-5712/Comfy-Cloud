package service

import (
	"comfy-cloud/internal/models"
	"comfy-cloud/internal/repository"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"gorm.io/gorm"
)

type ModelService struct {
	modelRepo *repository.ModelRepository
	db        *gorm.DB
	baseDir   string
}

func NewModelService(modelRepo *repository.ModelRepository, db *gorm.DB, baseDir string) *ModelService {
	return &ModelService{
		modelRepo: modelRepo,
		db:        db,
		baseDir:   baseDir,
	}
}

// GetPrivateModels 获取用户私有模型列表
func (s *ModelService) GetPrivateModels(userID uint) ([]map[string]interface{}, error) {
	models, err := s.modelRepo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, len(models))
	for i, model := range models {
		result[i] = map[string]interface{}{
			"id":                   model.ID,
			"name":                 model.Name,
			"type":                 model.Type,
			"size_bytes":           model.SizeBytes,
			"uploaded_at":          model.UploadedAt,
			"storage_cost_per_day": model.StorageCostPerDay,
		}
	}

	return result, nil
}

// UploadModel 上传模型
func (s *ModelService) UploadModel(userID uint, filename string, modelType string, fileData io.Reader) (map[string]interface{}, error) {
	// 验证模型类型
	validTypes := map[string]bool{
		"checkpoint": true,
		"lora":       true,
		"vae":        true,
		"embedding":  true,
	}

	if !validTypes[modelType] {
		return nil, fmt.Errorf("invalid model type")
	}

	// 创建用户模型目录
	userModelDir := filepath.Join(s.baseDir, "users", fmt.Sprintf("%d", userID), "models")
	if err := os.MkdirAll(userModelDir, 0755); err != nil {
		return nil, err
	}

	// 保存文件
	filePath := filepath.Join(userModelDir, filename)
	file, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 复制文件内容
	size, err := io.Copy(file, fileData)
	if err != nil {
		return nil, err
	}

	// 计算存储成本（每 GB 每天 0.02 元）
	storageCostPerDay := float64(size) / (1024 * 1024 * 1024) * 0.02

	// 创建数据库记录
	model := &models.PrivateModel{
		UserID:            userID,
		Name:              filename,
		Type:              modelType,
		SizeBytes:         size,
		FilePath:          filePath,
		Visibility:        "private",
		Status:            "active",
		StorageCostPerDay: storageCostPerDay,
		UploadedAt:        time.Now(),
	}

	if err := s.modelRepo.Create(model); err != nil {
		// 删除文件
		os.Remove(filePath)
		return nil, err
	}

	return map[string]interface{}{
		"id":                   model.ID,
		"name":                 model.Name,
		"type":                 model.Type,
		"size_bytes":           model.SizeBytes,
		"uploaded_at":          model.UploadedAt,
		"storage_cost_per_day": model.StorageCostPerDay,
		"path":                 fmt.Sprintf("/users/%d/models/%s", userID, filename),
	}, nil
}

// DeleteModel 删除模型
func (s *ModelService) DeleteModel(userID, modelID uint) error {
	model, err := s.modelRepo.FindByID(modelID)
	if err != nil {
		return err
	}

	// 验证所有权
	if model.UserID != userID {
		return fmt.Errorf("permission denied")
	}

	// 删除文件
	if err := os.Remove(model.FilePath); err != nil {
		// 文件可能已经不存在，继续删除数据库记录
	}

	// 删除数据库记录
	return s.modelRepo.Delete(modelID)
}

// GetAllModels 获取所有模型（Admin）
func (s *ModelService) GetAllModels(visibility string, limit, offset int) ([]map[string]interface{}, int64, error) {
	models, total, err := s.modelRepo.FindAll(visibility, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	result := make([]map[string]interface{}, len(models))
	for i, model := range models {
		username := "system"
		if model.User.Username != "" {
			username = model.User.Username
		}

		result[i] = map[string]interface{}{
			"id":                   model.ID,
			"name":                 model.Name,
			"type":                 model.Type,
			"size_bytes":           model.SizeBytes,
			"uploaded_at":          model.UploadedAt,
			"storage_cost_per_day": model.StorageCostPerDay,
			"user_id":              model.UserID,
			"username":             username,
			"visibility":           model.Visibility,
			"status":               model.Status,
		}
	}

	return result, total, nil
}

// UpdateModel 更新模型（Admin）
func (s *ModelService) UpdateModel(modelID uint, visibility, status string) (map[string]interface{}, error) {
	model, err := s.modelRepo.FindByID(modelID)
	if err != nil {
		return nil, err
	}

	if visibility != "" {
		model.Visibility = visibility
	}
	if status != "" {
		model.Status = status
	}

	if err := s.modelRepo.Update(model); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"id":         model.ID,
		"name":       model.Name,
		"visibility": model.Visibility,
		"status":     model.Status,
	}, nil
}
