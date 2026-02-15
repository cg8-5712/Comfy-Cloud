package auth

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"comfy-cloud/internal/database"
	"comfy-cloud/internal/models"
)

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string        `json:"token"`
	User  *models.User  `json:"user"`
}

// Register 注册用户
func Register(req RegisterRequest) (*models.User, error) {
	// 检查用户名是否存在
	var existingUser models.User
	if err := database.DB.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		return nil, errors.New("username already exists")
	}

	// 检查邮箱是否存在
	if err := database.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return nil, errors.New("email already exists")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 创建用户
	user := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Tier:         "basic",
		Balance:      10.00, // 新用户赠送 10 元
		Status:       "active",
	}

	if err := database.DB.Create(user).Error; err != nil {
		return nil, err
	}

	// 创建默认订阅
	subscription := &models.Subscription{
		UserID:    user.ID,
		Plan:      "basic",
		Status:    "active",
		StartedAt: time.Now(),
	}
	database.DB.Create(subscription)

	return user, nil
}

// Login 登录验证
func Login(req LoginRequest) (*models.User, error) {
	var user models.User

	// 查找用户
	if err := database.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid username or password")
		}
		return nil, err
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid username or password")
	}

	// 检查账户状态
	if user.Status != "active" {
		return nil, errors.New("account is suspended or deleted")
	}

	return &user, nil
}

// GetUserByID 根据 ID 获取用户
func GetUserByID(userID uint) (*models.User, error) {
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
