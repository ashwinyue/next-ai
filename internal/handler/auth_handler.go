package handler

import (
	"strings"

	"github.com/ashwinyue/next-ai/internal/service"
	"github.com/ashwinyue/next-ai/internal/service/auth"
	"github.com/gin-gonic/gin"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	svc *service.Services
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(svc *service.Services) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// Register 用户注册
func (h *AuthHandler) Register(c *gin.Context) {
	var req auth.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "Invalid parameters: "+err.Error())
		return
	}

	resp, err := h.svc.Auth.Register(c.Request.Context(), &req)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	Created(c, resp)
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req auth.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "Invalid parameters: "+err.Error())
		return
	}

	resp, err := h.svc.Auth.Login(c.Request.Context(), &req)
	if err != nil {
		Error(c, err)
		return
	}

	if !resp.Success {
		Success(c, resp)
		return
	}

	Success(c, resp)
}

// ValidateToken 验证令牌
func (h *AuthHandler) ValidateToken(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		BadRequest(c, "Missing Authorization header")
		return
	}

	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		BadRequest(c, "Invalid Authorization header format")
		return
	}

	token := tokenParts[1]
	user, err := h.svc.Auth.ValidateToken(c.Request.Context(), token)
	if err != nil {
		BadRequest(c, "Invalid or expired token")
		return
	}

	Success(c, user.ToUserInfo())
}

// RefreshToken 刷新令牌
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refreshToken" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "Invalid parameters")
		return
	}

	accessToken, newRefreshToken, err := h.svc.Auth.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		BadRequest(c, "Invalid refresh token")
		return
	}

	Success(c, gin.H{
		"access_token":  accessToken,
		"refresh_token": newRefreshToken,
	})
}

// Logout 用户登出
func (h *AuthHandler) Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		BadRequest(c, "Missing Authorization header")
		return
	}

	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		BadRequest(c, "Invalid Authorization header format")
		return
	}

	token := tokenParts[1]
	if err := h.svc.Auth.RevokeToken(c.Request.Context(), token); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// GetCurrentUser 获取当前用户
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	user, err := h.svc.Auth.GetCurrentUser(c.Request.Context())
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, user.ToUserInfo())
}

// ChangePassword 修改密码
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req auth.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "Invalid parameters")
		return
	}

	user, err := h.svc.Auth.GetCurrentUser(c.Request.Context())
	if err != nil {
		Error(c, err)
		return
	}

	if err := h.svc.Auth.ChangePassword(c.Request.Context(), user.ID, req.OldPassword, req.NewPassword); err != nil {
		BadRequest(c, err.Error())
		return
	}

	Success(c, nil)
}
