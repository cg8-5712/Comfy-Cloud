package service

import (
	"fmt"

	"comfy-cloud/internal/auth"
	"comfy-cloud/internal/config"
	"comfy-cloud/internal/models"
	"comfy-cloud/internal/repository"
)

type AuthService struct {
	userRepo   *repository.UserRepository
	cfg        *config.Config
	userDirSvc *UserDirectoryService
}

func NewAuthService(userRepo *repository.UserRepository, cfg *config.Config, userDirSvc *UserDirectoryService) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		cfg:        cfg,
		userDirSvc: userDirSvc,
	}
}

// Register 用户注册
func (s *AuthService) Register(req auth.RegisterRequest) (*models.User, string, error) {
	user, err := auth.Register(req)
	if err != nil {
		return nil, "", err
	}

	// 自动创建用户目录（models + temp）
	if s.userDirSvc != nil {
		if err := s.userDirSvc.InitializeUserDirectory(user.ID); err != nil {
			fmt.Printf("Warning: failed to create user directory for user %d: %v\n", user.ID, err)
		}
	}

	token, err := auth.GenerateToken(
		user.ID,
		user.Username,
		user.Tier,
		s.cfg.JWT.Expiration,
	)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

// Login 用户登录
func (s *AuthService) Login(req auth.LoginRequest) (*models.User, string, error) {
	// 调用 auth 包的登录逻辑
	user, err := auth.Login(req)
	if err != nil {
		return nil, "", err
	}

	// 生成 Token
	token, err := auth.GenerateToken(
		user.ID,
		user.Username,
		user.Tier,
		s.cfg.JWT.Expiration,
	)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

// GetUserByID 根据 ID 获取用户
func (s *AuthService) GetUserByID(userID uint) (*models.User, error) {
	return auth.GetUserByID(userID)
}
