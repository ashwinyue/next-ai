package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/ashwinyue/next-ai/internal/model"
	"github.com/ashwinyue/next-ai/internal/repository"
)

var (
	jwtSecretOnce sync.Once
	jwtSecret     string
)

// getJwtSecret 获取 JWT 密钥
func getJwtSecret() string {
	jwtSecretOnce.Do(func() {
		if envSecret := strings.TrimSpace(os.Getenv("JWT_SECRET")); envSecret != "" {
			jwtSecret = envSecret
			return
		}

		randomBytes := make([]byte, 32)
		if _, err := rand.Read(randomBytes); err != nil {
			panic(fmt.Sprintf("failed to generate JWT secret: %v", err))
		}
		jwtSecret = base64.StdEncoding.EncodeToString(randomBytes)
	})

	return jwtSecret
}

// Service 认证服务
type Service struct {
	repo *repository.Repositories
}

// NewService 创建认证服务
func NewService(repo *repository.Repositories) *Service {
	return &Service{repo: repo}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Success      bool         `json:"success"`
	Message      string       `json:"message,omitempty"`
	User         *model.User   `json:"user,omitempty"`
	Token        string       `json:"token,omitempty"`
	RefreshToken string       `json:"refresh_token,omitempty"`
}

// RegisterResponse 注册响应
type RegisterResponse struct {
	Success bool       `json:"success"`
	Message string     `json:"message,omitempty"`
	User    *model.User `json:"user,omitempty"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// Register 注册用户
func (s *Service) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	// 检查邮箱是否已存在
	existingUser, _ := s.repo.Auth.GetUserByEmail(req.Email)
	if existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	// 检查用户名是否已存在
	existingUser, _ = s.repo.Auth.GetUserByUsername(req.Username)
	if existingUser != nil {
		return nil, errors.New("user with this username already exists")
	}

	// 哈希密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 创建用户
	user := &model.User{
		ID:           uuid.New().String(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		IsActive:     true,
	}

	if err := s.repo.Auth.CreateUser(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &RegisterResponse{
		Success: true,
		Message: "Registration successful",
		User:    user,
	}, nil
}

// Login 用户登录
func (s *Service) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	// 获取用户
	user, err := s.repo.Auth.GetUserByEmail(req.Email)
	if err != nil {
		return &LoginResponse{
			Success: false,
			Message: "Invalid email or password",
		}, nil
	}

	// 检查用户是否激活
	if !user.IsActive {
		return &LoginResponse{
			Success: false,
			Message: "Account is disabled",
		}, nil
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return &LoginResponse{
			Success: false,
			Message: "Invalid email or password",
		}, nil
	}

	// 生成令牌
	accessToken, refreshToken, err := s.generateTokens(ctx, user)
	if err != nil {
		return &LoginResponse{
			Success: false,
			Message: "Login failed",
		}, err
	}

	return &LoginResponse{
		Success:      true,
		Message:      "Login successful",
		User:         user,
		Token:        accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// ValidateToken 验证令牌
func (s *Service) ValidateToken(ctx context.Context, tokenString string) (*model.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(getJwtSecret()), nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, errors.New("invalid user ID in token")
	}

	// 检查令牌是否被撤销
	tokenRecord, err := s.repo.Auth.GetTokenByValue(tokenString)
	if err != nil || tokenRecord == nil || tokenRecord.IsRevoked {
		return nil, errors.New("token is revoked")
	}

	return s.repo.Auth.GetUserByID(userID)
}

// RefreshToken 刷新令牌
func (s *Service) RefreshToken(ctx context.Context, refreshTokenString string) (string, string, error) {
	token, err := jwt.Parse(refreshTokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(getJwtSecret()), nil
	})

	if err != nil || !token.Valid {
		return "", "", errors.New("invalid refresh token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", errors.New("invalid token claims")
	}

	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != "refresh" {
		return "", "", errors.New("not a refresh token")
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", "", errors.New("invalid user ID in token")
	}

	// 检查令牌是否被撤销
	tokenRecord, err := s.repo.Auth.GetTokenByValue(refreshTokenString)
	if err != nil || tokenRecord == nil || tokenRecord.IsRevoked {
		return "", "", errors.New("refresh token is revoked")
	}

	// 获取用户
	user, err := s.repo.Auth.GetUserByID(userID)
	if err != nil {
		return "", "", err
	}

	// 撤销旧的刷新令牌
	_ = s.repo.Auth.RevokeToken(tokenRecord.ID)

	// 生成新令牌
	return s.generateTokens(ctx, user)
}

// RevokeToken 撤销令牌
func (s *Service) RevokeToken(ctx context.Context, tokenString string) error {
	tokenRecord, err := s.repo.Auth.GetTokenByValue(tokenString)
	if err != nil {
		return err
	}

	return s.repo.Auth.RevokeToken(tokenRecord.ID)
}

// GetCurrentUser 从 context 获取当前用户
func (s *Service) GetCurrentUser(ctx context.Context) (*model.User, error) {
	user, ok := ctx.Value("user").(*model.User)
	if !ok {
		return nil, errors.New("user not found in context")
	}
	return user, nil
}

// ChangePassword 修改密码
func (s *Service) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	user, err := s.repo.Auth.GetUserByID(userID)
	if err != nil {
		return err
	}

	// 验证旧密码
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword))
	if err != nil {
		return errors.New("invalid old password")
	}

	// 哈希新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hashedPassword)
	return s.repo.Auth.UpdateUser(user)
}

// generateTokens 生成访问令牌和刷新令牌
func (s *Service) generateTokens(ctx context.Context, user *model.User) (string, string, error) {
	// 生成访问令牌（24小时有效）
	accessClaims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
		"type":    "access",
	}

	accessTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err := accessTokenObj.SignedString([]byte(getJwtSecret()))
	if err != nil {
		return "", "", err
	}

	// 生成刷新令牌（7天有效）
	refreshClaims := jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
		"type":    "refresh",
	}

	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err := refreshTokenObj.SignedString([]byte(getJwtSecret()))
	if err != nil {
		return "", "", err
	}

	// 存储令牌到数据库
	accessTokenRecord := &model.AuthToken{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		Token:     accessToken,
		TokenType: "access_token",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	refreshTokenRecord := &model.AuthToken{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		Token:     refreshToken,
		TokenType: "refresh_token",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	_ = s.repo.Auth.CreateToken(accessTokenRecord)
	_ = s.repo.Auth.CreateToken(refreshTokenRecord)

	return accessToken, refreshToken, nil
}
