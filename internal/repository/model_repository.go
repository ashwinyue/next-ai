// Package repository 提供模型数据访问层
package repository

import (
	"context"
	"fmt"

	"github.com/ashwinyue/next-ai/internal/model"
	"gorm.io/gorm"
)

// ModelRepository 模型数据访问
type ModelRepository struct {
	db *gorm.DB
}

// NewModelRepository 创建模型仓库
func NewModelRepository(db *gorm.DB) *ModelRepository {
	return &ModelRepository{db: db}
}

// Create 创建模型
func (r *ModelRepository) Create(ctx context.Context, m *model.Model) error {
	return r.db.WithContext(ctx).Create(m).Error
}

// GetByID 根据 ID 获取模型
func (r *ModelRepository) GetByID(ctx context.Context, id string) (*model.Model, error) {
	var m model.Model
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// List 列出模型
func (r *ModelRepository) List(ctx context.Context, modelType *model.ModelType, source *model.ModelSource) ([]*model.Model, error) {
	var models []*model.Model
	query := r.db.WithContext(ctx).Model(&model.Model{})

	if modelType != nil {
		query = query.Where("type = ?", *modelType)
	}
	if source != nil {
		query = query.Where("source = ?", *source)
	}

	err := query.Order("created_at DESC").Find(&models).Error
	return models, err
}

// Update 更新模型
func (r *ModelRepository) Update(ctx context.Context, m *model.Model) error {
	return r.db.WithContext(ctx).Save(m).Error
}

// Delete 删除模型
func (r *ModelRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.Model{}, "id = ?", id).Error
}

// ClearDefaultByType 清除指定类型的默认标记
func (r *ModelRepository) ClearDefaultByType(ctx context.Context, modelType model.ModelType, excludeID string) error {
	query := r.db.WithContext(ctx).Model(&model.Model{}).
		Where("type = ?", modelType).
		Where("is_default = ?", true)

	if excludeID != "" {
		query = query.Where("id != ?", excludeID)
	}

	return query.Update("is_default", false).Error
}

// GetDefaultByType 获取指定类型的默认模型
func (r *ModelRepository) GetDefaultByType(ctx context.Context, modelType model.ModelType) (*model.Model, error) {
	var m model.Model
	err := r.db.WithContext(ctx).
		Where("type = ?", modelType).
		Where("is_default = ?", true).
		First(&m).Error
	if err != nil {
		return nil, fmt.Errorf("no default model found for type %s: %w", modelType, err)
	}
	return &m, nil
}
