// Package repository 数据访问层
package repository

import (
	"fmt"

	"github.com/ashwinyue/next-ai/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TenantRepository 租户仓库
type TenantRepository struct {
	db *gorm.DB
}

// NewTenantRepository 创建租户仓库
func NewTenantRepository(db *gorm.DB) *TenantRepository {
	return &TenantRepository{db: db}
}

// Create 创建租户
func (r *TenantRepository) Create(tenant *model.Tenant) error {
	return r.db.Create(tenant).Error
}

// GetByID 根据 ID 获取租户
func (r *TenantRepository) GetByID(id string) (*model.Tenant, error) {
	var tenant model.Tenant
	err := r.db.Where("id = ?", id).First(&tenant).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

// GetByAPIKey 根据 API Key 获取租户
func (r *TenantRepository) GetByAPIKey(apiKey string) (*model.Tenant, error) {
	var tenant model.Tenant
	err := r.db.Where("api_key = ?", apiKey).First(&tenant).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

// List 列出所有租户
func (r *TenantRepository) List() ([]*model.Tenant, error) {
	var tenants []*model.Tenant
	err := r.db.Find(&tenants).Error
	return tenants, err
}

// Update 更新租户
func (r *TenantRepository) Update(tenant *model.Tenant) error {
	return r.db.Save(tenant).Error
}

// Delete 删除租户（软删除）
func (r *TenantRepository) Delete(id string) error {
	return r.db.Delete(&model.Tenant{}, "id = ?", id).Error
}

// UpdateConfig 更新租户配置
func (r *TenantRepository) UpdateConfig(id string, configType string, config interface{}) error {
	updates := make(map[string]interface{})
	switch configType {
	case "agent":
		updates["agent_config"] = config
	case "context":
		updates["context_config"] = config
	case "web_search":
		updates["web_search_config"] = config
	case "conversation":
		updates["conversation_config"] = config
	default:
		return fmt.Errorf("unknown config type: %s", configType)
	}
	return r.db.Model(&model.Tenant{}).Where("id = ?", id).Updates(updates).Error
}

// GenerateAPIKey 生成新的 API Key
func (r *TenantRepository) GenerateAPIKey(id string) (string, error) {
	apiKey := "tenant_" + uuid.New().String()
	err := r.db.Model(&model.Tenant{}).Where("id = ?", id).Update("api_key", apiKey).Error
	if err != nil {
		return "", err
	}
	return apiKey, nil
}

// UpdateStorageUsed 更新存储使用量
func (r *TenantRepository) UpdateStorageUsed(id string, delta int64) error {
	return r.db.Model(&model.Tenant{}).Where("id = ?", id).
		Update("storage_used", gorm.Expr("storage_used + ?", delta)).Error
}
