package service

import (
	"fmt"
	"os"
	"path/filepath"
)

// UserDirectoryService 用户目录管理服务
type UserDirectoryService struct {
	baseDir string // 用户数据根目录
}

// NewUserDirectoryService 创建用户目录服务
func NewUserDirectoryService(baseDir string) *UserDirectoryService {
	return &UserDirectoryService{
		baseDir: baseDir,
	}
}

// InitializeUserDirectory 初始化用户目录结构
func (s *UserDirectoryService) InitializeUserDirectory(userID uint) error {
	userDir := s.GetUserDirectory(userID)

	// 创建用户根目录
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return fmt.Errorf("failed to create user directory: %w", err)
	}

	// 创建子目录
	subdirs := []string{
		"output",    // 生成的图片
		"workflows", // 保存的工作流
		"models",    // 私有模型
		"uploads",   // 上传的图片
		"temp",      // 临时文件
	}

	for _, subdir := range subdirs {
		path := filepath.Join(userDir, subdir)
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to create subdirectory %s: %w", subdir, err)
		}
	}

	return nil
}

// GetUserDirectory 获取用户目录路径
func (s *UserDirectoryService) GetUserDirectory(userID uint) string {
	return filepath.Join(s.baseDir, fmt.Sprintf("%d", userID))
}

// GetUserOutputDirectory 获取用户输出目录
func (s *UserDirectoryService) GetUserOutputDirectory(userID uint) string {
	return filepath.Join(s.GetUserDirectory(userID), "output")
}

// GetUserWorkflowDirectory 获取用户工作流目录
func (s *UserDirectoryService) GetUserWorkflowDirectory(userID uint) string {
	return filepath.Join(s.GetUserDirectory(userID), "workflows")
}

// GetUserModelDirectory 获取用户模型目录
func (s *UserDirectoryService) GetUserModelDirectory(userID uint) string {
	return filepath.Join(s.GetUserDirectory(userID), "models")
}

// GetUserUploadDirectory 获取用户上传目录
func (s *UserDirectoryService) GetUserUploadDirectory(userID uint) string {
	return filepath.Join(s.GetUserDirectory(userID), "uploads")
}

// DeleteUserDirectory 删除用户目录（谨慎使用）
func (s *UserDirectoryService) DeleteUserDirectory(userID uint) error {
	userDir := s.GetUserDirectory(userID)
	return os.RemoveAll(userDir)
}

// GetDirectorySize 获取目录大小（字节）
func (s *UserDirectoryService) GetDirectorySize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

// GetUserStorageUsage 获取用户存储使用情况（GB）
func (s *UserDirectoryService) GetUserStorageUsage(userID uint) (float64, error) {
	userDir := s.GetUserDirectory(userID)
	size, err := s.GetDirectorySize(userDir)
	if err != nil {
		return 0, err
	}
	// 转换为 GB
	return float64(size) / (1024 * 1024 * 1024), nil
}

// CleanupTempFiles 清理临时文件（超过指定天数）
func (s *UserDirectoryService) CleanupTempFiles(userID uint, days int) error {
	tempDir := filepath.Join(s.GetUserDirectory(userID), "temp")
	// TODO: 实现清理逻辑
	return filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// 删除超过指定天数的文件
		// if time.Since(info.ModTime()).Hours() > float64(days*24) {
		//     return os.Remove(path)
		// }
		return nil
	})
}
