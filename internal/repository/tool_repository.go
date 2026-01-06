package repository

import (
	"github.com/ashwinyue/next-rag/next-ai/internal/model"
	"gorm.io/gorm"
)

// ToolRepository 工具数据访问
type ToolRepository struct {
	db *gorm.DB
}

// NewToolRepository 创建工具仓库
func NewToolRepository(db *gorm.DB) *ToolRepository {
	return &ToolRepository{db: db}
}

// Create 创建工具
func (r *ToolRepository) Create(tool *model.Tool) error {
	return r.db.Create(tool).Error
}

// GetByID 获取工具
func (r *ToolRepository) GetByID(id string) (*model.Tool, error) {
	var tool model.Tool
	err := r.db.Where("id = ?", id).First(&tool).Error
	if err != nil {
		return nil, err
	}
	return &tool, nil
}

// GetByName 获取工具
func (r *ToolRepository) GetByName(name string) (*model.Tool, error) {
	var tool model.Tool
	err := r.db.Where("name = ?", name).First(&tool).Error
	if err != nil {
		return nil, err
	}
	return &tool, nil
}

// List 列出工具
func (r *ToolRepository) List(offset, limit int) ([]*model.Tool, error) {
	var tools []*model.Tool
	err := r.db.Order("created_at DESC").Offset(offset).Limit(limit).Find(&tools).Error
	return tools, err
}

// ListActive 列出活跃工具
func (r *ToolRepository) ListActive() ([]*model.Tool, error) {
	var tools []*model.Tool
	err := r.db.Where("is_active = ?", true).Order("created_at DESC").Find(&tools).Error
	return tools, err
}

// Update 更新工具
func (r *ToolRepository) Update(tool *model.Tool) error {
	return r.db.Save(tool).Error
}

// Delete 删除工具
func (r *ToolRepository) Delete(id string) error {
	return r.db.Delete(&model.Tool{}, "id = ?", id).Error
}
