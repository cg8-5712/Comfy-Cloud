package service

import (
	"comfy-cloud/internal/auth"
	"comfy-cloud/internal/config"
	"comfy-cloud/internal/models"
	"comfy-cloud/internal/repository"
)

type AuthService struct {
	userRepo *repository.UserRepository
	cfg      *config.Config
}

func NewAuthService(userRepo *repository.UserRepository, cfg *config.Config) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

// Register 用户注册
func (s *AuthService) Register(req auth.RegisterRequest) (*models.User, string, error) {
	// 调用 auth 包的注册逻辑
	user, err := auth.Register(req)
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
