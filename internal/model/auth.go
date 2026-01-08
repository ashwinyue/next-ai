package model

import "time"

// User 用户
type User struct {
	ID           string    `gorm:"primaryKey;size:36" json:"id"`
	Username     string    `gorm:"uniqueIndex;size:100;not null" json:"username"`
	Email        string    `gorm:"uniqueIndex;size:255;not null" json:"email"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	TenantID     string    `gorm:"index;size:36" json:"tenant_id"` // 租户 ID
	Avatar       string    `gorm:"size:500" json:"avatar"`
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// AuthToken 认证令牌
type AuthToken struct {
	ID        string    `gorm:"primaryKey;size:36" json:"id"`
	UserID    string    `gorm:"index;size:36;not null" json:"user_id"`
	Token     string    `gorm:"type:text;not null" json:"-"`
	TokenType string    `gorm:"size:50;not null" json:"token_type"` // access_token, refresh_token
	ExpiresAt time.Time `json:"expires_at"`
	IsRevoked bool      `gorm:"default:false" json:"is_revoked"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName 指定表名
func (AuthToken) TableName() string {
	return "auth_tokens"
}

// UserInfo 用户信息（不含敏感数据）
type UserInfo struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Avatar    string    `json:"avatar"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToUserInfo 转换为 UserInfo
func (u *User) ToUserInfo() *UserInfo {
	return &UserInfo{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Avatar:    u.Avatar,
		IsActive:  u.IsActive,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
