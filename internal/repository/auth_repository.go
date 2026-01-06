package repository

import (
	"time"

	"github.com/ashwinyue/next-ai/internal/model"
	"gorm.io/gorm"
)

// AuthRepository 认证数据访问
type AuthRepository struct {
	db *gorm.DB
}

// NewAuthRepository 创建认证仓库
func NewAuthRepository(db *gorm.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

// CreateUser 创建用户
func (r *AuthRepository) CreateUser(user *model.User) error {
	return r.db.Create(user).Error
}

// GetUserByID 获取用户
func (r *AuthRepository) GetUserByID(id string) (*model.User, error) {
	var user model.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail 获取用户
func (r *AuthRepository) GetUserByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByUsername 获取用户
func (r *AuthRepository) GetUserByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser 更新用户
func (r *AuthRepository) UpdateUser(user *model.User) error {
	return r.db.Save(user).Error
}

// CreateToken 创建令牌
func (r *AuthRepository) CreateToken(token *model.AuthToken) error {
	return r.db.Create(token).Error
}

// GetTokenByValue 获取令牌
func (r *AuthRepository) GetTokenByValue(tokenValue string) (*model.AuthToken, error) {
	var token model.AuthToken
	err := r.db.Where("token = ? AND is_revoked = ?", tokenValue, false).
		Where("expires_at > ?", time.Now()).
		First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// GetTokensByUserID 获取用户的所有令牌
func (r *AuthRepository) GetTokensByUserID(userID string) ([]*model.AuthToken, error) {
	var tokens []*model.AuthToken
	err := r.db.Where("user_id = ?", userID).Find(&tokens).Error
	return tokens, err
}

// RevokeToken 撤销令牌
func (r *AuthRepository) RevokeToken(tokenID string) error {
	return r.db.Model(&model.AuthToken{}).Where("id = ?", tokenID).Update("is_revoked", true).Error
}

// RevokeTokensByUserID 撤销用户的所有令牌
func (r *AuthRepository) RevokeTokensByUserID(userID string) error {
	return r.db.Model(&model.AuthToken{}).Where("user_id = ?", userID).Update("is_revoked", true).Error
}

// DeleteExpiredTokens 删除过期令牌
func (r *AuthRepository) DeleteExpiredTokens() error {
	return r.db.Where("expires_at < ? OR is_revoked = ?", time.Now(), true).Delete(&model.AuthToken{}).Error
}
