package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/ashwinyue/next-ai/internal/model"
)

// TagRepository 标签仓库
type TagRepository struct {
	db *gorm.DB
}

// NewTagRepository 创建标签仓库
func NewTagRepository(db *gorm.DB) *TagRepository {
	return &TagRepository{db: db}
}

// Create 创建标签
func (r *TagRepository) Create(ctx context.Context, tag *model.KnowledgeTag) error {
	return r.db.WithContext(ctx).Create(tag).Error
}

// Update 更新标签
func (r *TagRepository) Update(ctx context.Context, tag *model.KnowledgeTag) error {
	return r.db.WithContext(ctx).Save(tag).Error
}

// GetByID 根据 ID 获取标签
func (r *TagRepository) GetByID(ctx context.Context, id string) (*model.KnowledgeTag, error) {
	var tag model.KnowledgeTag
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&tag).Error
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

// GetByKnowledgeBaseID 获取知识库的所有标签
func (r *TagRepository) GetByKnowledgeBaseID(ctx context.Context, kbID string) ([]*model.KnowledgeTag, error) {
	var tags []*model.KnowledgeTag
	err := r.db.WithContext(ctx).
		Where("knowledge_base_id = ?", kbID).
		Order("sort_order ASC, created_at ASC").
		Find(&tags).Error
	return tags, err
}

// GetByName 根据知识库 ID 和名称获取标签
func (r *TagRepository) GetByName(ctx context.Context, kbID, name string) (*model.KnowledgeTag, error) {
	var tag model.KnowledgeTag
	err := r.db.WithContext(ctx).
		Where("knowledge_base_id = ? AND name = ?", kbID, name).
		First(&tag).Error
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

// Delete 删除标签
func (r *TagRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.KnowledgeTag{}, "id = ?", id).Error
}

// List 分页查询标签
func (r *TagRepository) List(ctx context.Context, kbID string, page, pageSize int, keyword string) ([]*model.KnowledgeTag, int64, error) {
	var tags []*model.KnowledgeTag
	var total int64

	query := r.db.WithContext(ctx).Model(&model.KnowledgeTag{}).Where("knowledge_base_id = ?", kbID)

	if keyword != "" {
		query = query.Where("name LIKE ?", "%"+keyword+"%")
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count tags: %w", err)
	}

	if page > 0 && pageSize > 0 {
		offset := (page - 1) * pageSize
		query = query.Offset(offset).Limit(pageSize)
	}

	err = query.Order("sort_order ASC, created_at ASC").Find(&tags).Error
	return tags, total, err
}
