package services

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"time"
	
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	
	"bkp-drive/internal/middleware"
	"bkp-drive/internal/models"
	"bkp-drive/pkg/database"
)

type UserService struct {
	jwtSecret string
}

func NewUserService(jwtSecret string) *UserService {
	return &UserService{
		jwtSecret: jwtSecret,
	}
}

// generateUserID 生成用户ID (bkp-开头的8位字母)
func (s *UserService) generateUserID() (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyz"
	const length = 8
	
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	
	for i := range b {
		b[i] = charset[b[i]%byte(len(charset))]
	}
	
	return "bkp-" + string(b), nil
}

// checkUserIDExists 检查用户ID是否已存在
func (s *UserService) checkUserIDExists(userID string) (bool, error) {
	var count int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM users WHERE user_id = $1", userID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// generateUniqueUserID 生成唯一的用户ID
func (s *UserService) generateUniqueUserID() (string, error) {
	for i := 0; i < 10; i++ { // 最多尝试10次
		userID, err := s.generateUserID()
		if err != nil {
			return "", err
		}
		
		exists, err := s.checkUserIDExists(userID)
		if err != nil {
			return "", err
		}
		
		if !exists {
			return userID, nil
		}
	}
	return "", fmt.Errorf("生成唯一用户ID失败")
}

// RegisterUser 注册用户
func (s *UserService) RegisterUser(req *models.UserRegisterRequest) (*models.User, error) {
	// 检查用户名是否已存在
	var count int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM users WHERE username = $1", req.Username).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("检查用户名失败: %w", err)
	}
	if count > 0 {
		return nil, fmt.Errorf("用户名已存在")
	}

	// 生成用户ID
	userID, err := s.generateUniqueUserID()
	if err != nil {
		return nil, fmt.Errorf("生成用户ID失败: %w", err)
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	// 插入数据库并返回自增ID (PostgreSQL使用RETURNING)
	var insertedID int
	err = database.DB.QueryRow(
		"INSERT INTO users (user_id, username, password, created_at, updated_at) VALUES ($1, $2, $3, NOW(), NOW()) RETURNING id",
		userID, req.Username, string(hashedPassword),
	).Scan(&insertedID)
	if err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	// 返回用户信息
	user := &models.User{
		ID:        insertedID,
		UserID:    userID,
		Username:  req.Username,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return user, nil
}

// LoginUser 用户登录
func (s *UserService) LoginUser(req *models.UserLoginRequest) (*models.User, string, error) {
	var user models.User
	var hashedPassword string

	err := database.DB.QueryRow(
		"SELECT id, user_id, username, password, created_at, updated_at FROM users WHERE username = $1",
		req.Username,
	).Scan(&user.ID, &user.UserID, &user.Username, &hashedPassword, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, "", fmt.Errorf("用户名或密码错误")
		}
		return nil, "", fmt.Errorf("查询用户失败: %w", err)
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password)); err != nil {
		return nil, "", fmt.Errorf("用户名或密码错误")
	}

	// 生成JWT token
	token, err := s.generateToken(&user)
	if err != nil {
		return nil, "", fmt.Errorf("生成令牌失败: %w", err)
	}

	return &user, token, nil
}

// generateToken 生成JWT token
func (s *UserService) generateToken(user *models.User) (string, error) {
	claims := &middleware.Claims{
		UserID:   user.UserID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 24小时过期
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}