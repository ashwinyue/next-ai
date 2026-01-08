// Package tenant 提供租户管理服务
package tenant

import (
	"context"
	"fmt"

	"github.com/ashwinyue/next-ai/internal/model"
	"github.com/ashwinyue/next-ai/internal/repository"
)

// Service 租户服务
type Service struct {
	repo *repository.Repositories
}

// NewService 创建租户服务
func NewService(repo *repository.Repositories) *Service {
	return &Service{
		repo: repo,
	}
}

// CreateTenantRequest 创建租户请求
type CreateTenantRequest struct {
	Name         string `json:"name" binding:"required"`
	Description  string `json:"description"`
	Business     string `json:"business"`
	StorageQuota int64  `json:"storage_quota"`
}

// UpdateTenantRequest 更新租户请求
type UpdateTenantRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Business    string `json:"business"`
	Status      string `json:"status"`
}

// CreateTenant 创建租户
func (s *Service) CreateTenant(ctx context.Context, req *CreateTenantRequest) (*model.Tenant, error) {
	// 检查名称是否已存在
	existing, _ := s.repo.Tenant.GetByAPIKey("tenant_" + req.Name)
	if existing != nil {
		return nil, fmt.Errorf("tenant already exists")
	}

	tenant := &model.Tenant{
		Name:        req.Name,
		Description: req.Description,
		Business:    req.Business,
		Status:      "active",
	}

	if req.StorageQuota > 0 {
		tenant.StorageQuota = req.StorageQuota
	} else {
		tenant.StorageQuota = 10737418240 // 默认 10GB
	}

	if err := s.repo.Tenant.Create(tenant); err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	return tenant, nil
}

// GetTenant 获取租户详情
func (s *Service) GetTenant(ctx context.Context, id string) (*model.Tenant, error) {
	tenant, err := s.repo.Tenant.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}
	return tenant, nil
}

// ListTenants 列出租户
func (s *Service) ListTenants(ctx context.Context) ([]*model.Tenant, error) {
	tenants, err := s.repo.Tenant.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}
	return tenants, nil
}

// UpdateTenant 更新租户
func (s *Service) UpdateTenant(ctx context.Context, id string, req *UpdateTenantRequest) (*model.Tenant, error) {
	tenant, err := s.repo.Tenant.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}

	if req.Name != "" {
		tenant.Name = req.Name
	}
	if req.Description != "" {
		tenant.Description = req.Description
	}
	if req.Business != "" {
		tenant.Business = req.Business
	}
	if req.Status != "" {
		tenant.Status = req.Status
	}

	if err := s.repo.Tenant.Update(tenant); err != nil {
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}

	return tenant, nil
}

// DeleteTenant 删除租户
func (s *Service) DeleteTenant(ctx context.Context, id string) error {
	if err := s.repo.Tenant.Delete(id); err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}
	return nil
}

// GetTenantByAPIKey 根据 API Key 获取租户
func (s *Service) GetTenantByAPIKey(ctx context.Context, apiKey string) (*model.Tenant, error) {
	tenant, err := s.repo.Tenant.GetByAPIKey(apiKey)
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}
	return tenant, nil
}

// UpdateTenantConfigRequest 更新租户配置请求
type UpdateTenantConfigRequest struct {
	ConfigType string      `json:"config_type" binding:"required"` // agent, context, web_search, conversation
	Config     interface{} `json:"config" binding:"required"`
}

// UpdateTenantConfig 更新租户配置
func (s *Service) UpdateTenantConfig(ctx context.Context, id string, req *UpdateTenantConfigRequest) error {
	// 验证租户存在
	_, err := s.repo.Tenant.GetByID(id)
	if err != nil {
		return fmt.Errorf("tenant not found: %w", err)
	}

	if err := s.repo.Tenant.UpdateConfig(id, req.ConfigType, req.Config); err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	return nil
}

// GetTenantConfig 获取租户配置
func (s *Service) GetTenantConfig(ctx context.Context, id string, configType string) (interface{}, error) {
	tenant, err := s.repo.Tenant.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}

	switch configType {
	case "agent":
		return tenant.AgentConfig, nil
	case "context":
		return tenant.ContextConfig, nil
	case "web_search":
		return tenant.WebSearchConfig, nil
	case "conversation":
		return tenant.ConversationConfig, nil
	default:
		return nil, fmt.Errorf("unknown config type: %s", configType)
	}
}

// RegenerateAPIKey 重新生成 API Key
func (s *Service) RegenerateAPIKey(ctx context.Context, id string) (string, error) {
	apiKey, err := s.repo.Tenant.GenerateAPIKey(id)
	if err != nil {
		return "", fmt.Errorf("failed to regenerate api key: %w", err)
	}
	return apiKey, nil
}

// GetStorageStats 获取存储统计
func (s *Service) GetStorageStats(ctx context.Context, id string) (used, quota int64, err error) {
	tenant, err := s.repo.Tenant.GetByID(id)
	if err != nil {
		return 0, 0, fmt.Errorf("tenant not found: %w", err)
	}
	return tenant.StorageUsed, tenant.StorageQuota, nil
}
