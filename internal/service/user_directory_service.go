package service

import (
	"fmt"
	"os"
	"path/filepath"
)

// UserDirectoryService 用户目录管理服务
//
// 只管理 Go 后端自身需要的用户数据（私有模型暂存等）。
// ComfyUI 侧的用户隔离（output、input、workflows）由 PathRewriteMiddleware 处理，
// ComfyUI 会在写文件时自动创建子目录，不需要预建。
type UserDirectoryService struct {
	baseDir string // 用户数据根目录，如 ./data/users
}

func NewUserDirectoryService(baseDir string) *UserDirectoryService {
	return &UserDirectoryService{baseDir: baseDir}
}

// userDir returns the per-user directory path: {baseDir}/user_{id}
func (s *UserDirectoryService) userDir(userID uint) string {
	return filepath.Join(s.baseDir, fmt.Sprintf("user_%d", userID))
}

// InitializeUserDirectory 注册时调用，创建后端管理的用户目录
func (s *UserDirectoryService) InitializeUserDirectory(userID uint) error {
	dir := s.userDir(userID)

	subdirs := []string{
		"models", // 私有模型（用户通过管理面板上传，再同步到 ComfyUI）
		"temp",   // 临时文件（模型上传中转、导出打包等）
	}

	for _, sub := range subdirs {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0755); err != nil {
			return fmt.Errorf("failed to create %s: %w", sub, err)
		}
	}
	return nil
}

// GetUserDirectory 获取用户根目录路径
func (s *UserDirectoryService) GetUserDirectory(userID uint) string {
	return s.userDir(userID)
}

// GetUserModelDirectory 获取用户私有模型目录
func (s *UserDirectoryService) GetUserModelDirectory(userID uint) string {
	return filepath.Join(s.userDir(userID), "models")
}

// DeleteUserDirectory 删除用户目录（谨慎使用）
func (s *UserDirectoryService) DeleteUserDirectory(userID uint) error {
	return os.RemoveAll(s.userDir(userID))
}

// GetUserStorageUsage 获取用户存储使用量（GB）
func (s *UserDirectoryService) GetUserStorageUsage(userID uint) (float64, error) {
	var size int64
	err := filepath.Walk(s.userDir(userID), func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return float64(size) / (1024 * 1024 * 1024), nil
}
