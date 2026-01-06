package chunk

import (
	"context"
	"errors"

	"github.com/ashwinyue/next-ai/internal/model"
	"github.com/ashwinyue/next-ai/internal/repository"
)

var (
	ErrChunkNotFound = errors.New("chunk not found")
)

// Service 分块服务
type Service struct {
	repo *repository.Repositories
}

// NewService 创建分块服务
func NewService(repo *repository.Repositories) *Service {
	return &Service{repo: repo}
}

// GetChunkByID 获取单个分块
func (s *Service) GetChunkByID(ctx context.Context, chunkID string) (*model.DocumentChunk, error) {
	chunk, err := s.repo.Knowledge.GetChunkByID(chunkID)
	if err != nil {
		return nil, ErrChunkNotFound
	}
	return chunk, nil
}

// ListChunksByKnowledgeBaseID 获取知识库的所有分块（支持分页）
func (s *Service) ListChunksByKnowledgeBaseID(ctx context.Context, kbID string, page, pageSize int) ([]*model.DocumentChunk, int64, error) {
	// 验证知识库是否存在
	_, err := s.repo.Knowledge.GetKnowledgeBaseByID(kbID)
	if err != nil {
		return nil, 0, errors.New("knowledge base not found")
	}

	offset := (page - 1) * pageSize
	return s.repo.Knowledge.ListChunksByKnowledgeBaseID(kbID, offset, pageSize)
}

// UpdateChunkRequest 更新分块请求
type UpdateChunkRequest struct {
	Content    string `json:"content"`
	ChunkIndex *int   `json:"chunk_index,omitempty"`
}

// UpdateChunk 更新分块
func (s *Service) UpdateChunk(ctx context.Context, chunkID string, req *UpdateChunkRequest) (*model.DocumentChunk, error) {
	chunk, err := s.repo.Knowledge.GetChunkByID(chunkID)
	if err != nil {
		return nil, ErrChunkNotFound
	}

	// 更新内容
	if req.Content != "" {
		chunk.Content = req.Content
	}

	// 更新分块索引
	if req.ChunkIndex != nil {
		chunk.ChunkIndex = *req.ChunkIndex
	}

	if err := s.repo.Knowledge.UpdateChunk(chunk); err != nil {
		return nil, err
	}

	return chunk, nil
}

// DeleteChunk 删除单个分块
func (s *Service) DeleteChunk(ctx context.Context, chunkID string) error {
	// 先检查分块是否存在
	_, err := s.repo.Knowledge.GetChunkByID(chunkID)
	if err != nil {
		return ErrChunkNotFound
	}

	return s.repo.Knowledge.DeleteChunk(chunkID)
}

// DeleteChunksByDocumentID 删除文档的所有分块
func (s *Service) DeleteChunksByDocumentID(ctx context.Context, docID string) error {
	return s.repo.Knowledge.DeleteChunksByDocumentID(docID)
}

// DeleteChunksByKnowledgeBaseID 删除知识库的所有分块
func (s *Service) DeleteChunksByKnowledgeBaseID(ctx context.Context, kbID string) error {
	// 验证知识库是否存在
	_, err := s.repo.Knowledge.GetKnowledgeBaseByID(kbID)
	if err != nil {
		return errors.New("knowledge base not found")
	}

	return s.repo.Knowledge.DeleteChunksByKnowledgeBaseID(kbID)
}
