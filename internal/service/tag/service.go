package tag

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/ashwinyue/next-ai/internal/model"
	"github.com/ashwinyue/next-ai/internal/repository"
)

// Service 标签服务
type Service struct {
	repo *repository.Repositories
}

// NewService 创建标签服务
func NewService(repo *repository.Repositories) *Service {
	return &Service{repo: repo}
}

// CreateTagRequest 创建标签请求
type CreateTagRequest struct {
	Name      string `json:"name" binding:"required"`
	Color     string `json:"color"`
	SortOrder int    `json:"sort_order"`
}

// UpdateTagRequest 更新标签请求
type UpdateTagRequest struct {
	Name      *string `json:"name"`
	Color     *string `json:"color"`
	SortOrder *int    `json:"sort_order"`
}

// ListTagsResponse 标签列表响应
type ListTagsResponse struct {
	Tags  []*model.KnowledgeTag `json:"tags"`
	Total int64                 `json:"total"`
}

// CreateTag 创建标签
func (s *Service) CreateTag(ctx context.Context, kbID string, req *CreateTagRequest) (*model.KnowledgeTag, error) {
	// 检查名称是否已存在
	existing, err := s.repo.Tag.GetByName(ctx, kbID, req.Name)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("tag with name '%s' already exists", req.Name)
	}

	tag := &model.KnowledgeTag{
		ID:              uuid.New().String(),
		KnowledgeBaseID: kbID,
		Name:            req.Name,
		Color:           req.Color,
		SortOrder:       req.SortOrder,
	}

	if err := s.repo.Tag.Create(ctx, tag); err != nil {
		return nil, fmt.Errorf("failed to create tag: %w", err)
	}

	return tag, nil
}

// GetTag 获取标签
func (s *Service) GetTag(ctx context.Context, id string) (*model.KnowledgeTag, error) {
	tag, err := s.repo.Tag.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}
	return tag, nil
}

// ListTags 列出标签
func (s *Service) ListTags(ctx context.Context, kbID string, page, pageSize int, keyword string) (*ListTagsResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	tags, total, err := s.repo.Tag.List(ctx, kbID, page, pageSize, keyword)
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}

	return &ListTagsResponse{
		Tags:  tags,
		Total: total,
	}, nil
}

// GetTagsByKnowledgeBaseID 获取知识库的所有标签（不分页）
func (s *Service) GetTagsByKnowledgeBaseID(ctx context.Context, kbID string) ([]*model.KnowledgeTag, error) {
	tags, err := s.repo.Tag.GetByKnowledgeBaseID(ctx, kbID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}
	return tags, nil
}

// UpdateTag 更新标签
func (s *Service) UpdateTag(ctx context.Context, id string, req *UpdateTagRequest) (*model.KnowledgeTag, error) {
	tag, err := s.repo.Tag.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("tag not found: %w", err)
	}

	if req.Name != nil {
		// 检查新名称是否与其他标签冲突
		existing, err := s.repo.Tag.GetByName(ctx, tag.KnowledgeBaseID, *req.Name)
		if err == nil && existing != nil && existing.ID != id {
			return nil, fmt.Errorf("tag with name '%s' already exists", *req.Name)
		}
		tag.Name = *req.Name
	}
	if req.Color != nil {
		tag.Color = *req.Color
	}
	if req.SortOrder != nil {
		tag.SortOrder = *req.SortOrder
	}

	if err := s.repo.Tag.Update(ctx, tag); err != nil {
		return nil, fmt.Errorf("failed to update tag: %w", err)
	}

	return tag, nil
}

// DeleteTag 删除标签
func (s *Service) DeleteTag(ctx context.Context, id string) error {
	if err := s.repo.Tag.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}
	return nil
}

// FindOrCreateTagByName 根据名称查找或创建标签
func (s *Service) FindOrCreateTagByName(ctx context.Context, kbID, name string) (*model.KnowledgeTag, error) {
	tag, err := s.repo.Tag.GetByName(ctx, kbID, name)
	if err == nil && tag != nil {
		return tag, nil
	}

	// 创建新标签
	newTag := &model.KnowledgeTag{
		ID:              uuid.New().String(),
		KnowledgeBaseID: kbID,
		Name:            name,
		Color:           "",
		SortOrder:       0,
	}

	if err := s.repo.Tag.Create(ctx, newTag); err != nil {
		return nil, fmt.Errorf("failed to create tag: %w", err)
	}

	return newTag, nil
}

// ========== 文档-标签关联管理 ==========

// GetDocumentTags 获取文档的所有标签
func (s *Service) GetDocumentTags(ctx context.Context, documentID string) ([]*model.KnowledgeTag, error) {
	tags, err := s.repo.Tag.GetTagsByDocumentID(ctx, documentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get document tags: %w", err)
	}
	return tags, nil
}

// SetDocumentTags 设置文档的标签
func (s *Service) SetDocumentTags(ctx context.Context, documentID string, tagIDs []string) error {
	if err := s.repo.Tag.SetDocumentTags(ctx, documentID, tagIDs); err != nil {
		return fmt.Errorf("failed to set document tags: %w", err)
	}
	return nil
}

// TagUpdate 标签更新项
type TagUpdate struct {
	KnowledgeID string   `json:"knowledge_id"`
	TagIDs      []string `json:"tag_ids"`
}

// BatchUpdateDocumentTags 批量更新文档标签（WeKnora API 兼容）
func (s *Service) BatchUpdateDocumentTags(ctx context.Context, updates []TagUpdate) error {
	// 转换为 repository 格式
	repoUpdates := make([]map[string]interface{}, 0, len(updates))
	for _, u := range updates {
		repoUpdates = append(repoUpdates, map[string]interface{}{
			"document_id": u.KnowledgeID,
			"tag_ids":     u.TagIDs,
		})
	}

	if err := s.repo.Tag.BatchUpdateDocumentTags(ctx, repoUpdates); err != nil {
		return fmt.Errorf("failed to batch update document tags: %w", err)
	}
	return nil
}

// AddTagToDocument 为文档添加单个标签
func (s *Service) AddTagToDocument(ctx context.Context, documentID, tagID string) error {
	if err := s.repo.Tag.AddTagToDocument(ctx, documentID, tagID); err != nil {
		return fmt.Errorf("failed to add tag to document: %w", err)
	}
	return nil
}

// RemoveTagFromDocument 移除文档的单个标签
func (s *Service) RemoveTagFromDocument(ctx context.Context, documentID, tagID string) error {
	if err := s.repo.Tag.RemoveTagFromDocument(ctx, documentID, tagID); err != nil {
		return fmt.Errorf("failed to remove tag from document: %w", err)
	}
	return nil
}
