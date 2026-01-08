package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
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

// ========== 文档-标签关联 ==========

// GetTagsByDocumentID 获取文档的所有标签
func (r *TagRepository) GetTagsByDocumentID(ctx context.Context, documentID string) ([]*model.KnowledgeTag, error) {
	var tags []*model.KnowledgeTag
	err := r.db.WithContext(ctx).
		Joins("JOIN document_tags ON document_tags.tag_id = knowledge_tags.id").
		Where("document_tags.document_id = ?", documentID).
		Order("knowledge_tags.sort_order ASC").
		Find(&tags).Error
	return tags, err
}

// AddTagToDocument 为文档添加标签
func (r *TagRepository) AddTagToDocument(ctx context.Context, documentID, tagID string) error {
	docTag := &model.DocumentTag{
		ID:         uuid.New().String(),
		DocumentID: documentID,
		TagID:      tagID,
	}
	return r.db.WithContext(ctx).Create(docTag).Error
}

// RemoveTagFromDocument 移除文档的标签
func (r *TagRepository) RemoveTagFromDocument(ctx context.Context, documentID, tagID string) error {
	return r.db.WithContext(ctx).
		Where("document_id = ? AND tag_id = ?", documentID, tagID).
		Delete(&model.DocumentTag{}).Error
}

// SetDocumentTags 设置文档的标签（先删除后添加）
func (r *TagRepository) SetDocumentTags(ctx context.Context, documentID string, tagIDs []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除现有的标签关联
		if err := tx.Where("document_id = ?", documentID).Delete(&model.DocumentTag{}).Error; err != nil {
			return fmt.Errorf("failed to remove existing tags: %w", err)
		}

		// 添加新的标签关联
		for _, tagID := range tagIDs {
			docTag := &model.DocumentTag{
				ID:         uuid.New().String(),
				DocumentID: documentID,
				TagID:      tagID,
			}
			if err := tx.Create(docTag).Error; err != nil {
				return fmt.Errorf("failed to add tag %s: %w", tagID, err)
			}
		}

		return nil
	})
}

// BatchUpdateDocumentTags 批量更新多个文档的标签
func (r *TagRepository) BatchUpdateDocumentTags(ctx context.Context, updates []map[string]interface{}) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, update := range updates {
			documentID, _ := update["document_id"].(string)
			tagIDs, _ := update["tag_ids"].([]string)

			if documentID == "" {
				continue
			}

			// 删除现有的标签关联
			if err := tx.Where("document_id = ?", documentID).Delete(&model.DocumentTag{}).Error; err != nil {
				return fmt.Errorf("failed to remove tags for document %s: %w", documentID, err)
			}

			// 添加新的标签关联
			for _, tagID := range tagIDs {
				docTag := &model.DocumentTag{
					ID:         uuid.New().String(),
					DocumentID: documentID,
					TagID:      tagID,
				}
				if err := tx.Create(docTag).Error; err != nil {
					return fmt.Errorf("failed to add tag %s to document %s: %w", tagID, documentID, err)
				}
			}
		}
		return nil
	})
}
