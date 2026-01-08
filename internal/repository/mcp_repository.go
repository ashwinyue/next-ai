// Package repository 数据访问层
package repository

import (
	"github.com/ashwinyue/next-ai/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MCPServiceRepository MCP 服务仓库
type MCPServiceRepository struct {
	db *gorm.DB
}

// NewMCPServiceRepository 创建 MCP 服务仓库
func NewMCPServiceRepository(db *gorm.DB) *MCPServiceRepository {
	return &MCPServiceRepository{db: db}
}

// Create 创建 MCP 服务
func (r *MCPServiceRepository) Create(svc *model.MCPService) error {
	if svc.ID == "" {
		svc.ID = uuid.New().String()
	}
	return r.db.Create(svc).Error
}

// GetByID 根据 ID 获取 MCP 服务
func (r *MCPServiceRepository) GetByID(id string) (*model.MCPService, error) {
	var svc model.MCPService
	err := r.db.Where("id = ?", id).First(&svc).Error
	if err != nil {
		return nil, err
	}
	return &svc, nil
}

// List 列出所有 MCP 服务
func (r *MCPServiceRepository) List() ([]*model.MCPService, error) {
	var svcs []*model.MCPService
	err := r.db.Find(&svcs).Error
	return svcs, err
}

// ListEnabled 列出启用的 MCP 服务
func (r *MCPServiceRepository) ListEnabled() ([]*model.MCPService, error) {
	var svcs []*model.MCPService
	err := r.db.Where("enabled = ?", true).Find(&svcs).Error
	return svcs, err
}

// Update 更新 MCP 服务
func (r *MCPServiceRepository) Update(svc *model.MCPService) error {
	return r.db.Save(svc).Error
}

// Delete 删除 MCP 服务
func (r *MCPServiceRepository) Delete(id string) error {
	return r.db.Delete(&model.MCPService{}, "id = ?", id).Error
}

// UpdateEnabled 更新启用状态
func (r *MCPServiceRepository) UpdateEnabled(id string, enabled bool) error {
	return r.db.Model(&model.MCPService{}).Where("id = ?", id).Update("enabled", enabled).Error
}

// GetByName 根据名称获取 MCP 服务
func (r *MCPServiceRepository) GetByName(name string) (*model.MCPService, error) {
	var svc model.MCPService
	err := r.db.Where("name = ?", name).First(&svc).Error
	if err != nil {
		return nil, err
	}
	return &svc, nil
}
