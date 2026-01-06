package knowledge

import (
	"context"
	"fmt"

	"github.com/ashwinyue/next-ai/internal/config"
	"github.com/ashwinyue/next-ai/internal/model"
	"github.com/ashwinyue/next-ai/internal/repository"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/google/uuid"
)

// Service 知识库服务
type Service struct {
	repo              *repository.Repositories
	cfg               *config.Config
	embedder          embedding.Embedder
	documentProcessor *DocumentProcessor
}

// NewService 创建知识库服务
func NewService(repo *repository.Repositories, cfg *config.Config, embedder embedding.Embedder) *Service {
	docProcessor := NewDocumentProcessor(repo, cfg, embedder)
	return &Service{
		repo:              repo,
		cfg:               cfg,
		embedder:          embedder,
		documentProcessor: docProcessor,
	}
}

// CreateKnowledgeBaseRequest 创建知识库请求
type CreateKnowledgeBaseRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	EmbedModel  string `json:"embed_model"`
}

// CreateKnowledgeBase 创建知识库
func (s *Service) CreateKnowledgeBase(ctx context.Context, req *CreateKnowledgeBaseRequest) (*model.KnowledgeBase, error) {
	kb := &model.KnowledgeBase{
		ID:             uuid.New().String(),
		Name:           req.Name,
		Description:    req.Description,
		EmbeddingModel: req.EmbedModel,
		IndexName:      "kb_" + uuid.New().String(),
		ChunkSize:      512,
		ChunkOverlap:   50,
	}

	if err := s.repo.Knowledge.CreateKnowledgeBase(kb); err != nil {
		return nil, fmt.Errorf("failed to create knowledge base: %w", err)
	}

	return kb, nil
}

// GetKnowledgeBase 获取知识库
func (s *Service) GetKnowledgeBase(ctx context.Context, id string) (*model.KnowledgeBase, error) {
	return s.repo.Knowledge.GetKnowledgeBaseByID(id)
}

// ListKnowledgeBasesRequest 列出知识库请求
type ListKnowledgeBasesRequest struct {
	Page int `json:"page"`
	Size int `json:"size"`
}

// ListKnowledgeBases 列出知识库
func (s *Service) ListKnowledgeBases(ctx context.Context, req *ListKnowledgeBasesRequest) ([]*model.KnowledgeBase, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 || req.Size > 100 {
		req.Size = 20
	}

	offset := (req.Page - 1) * req.Size
	return s.repo.Knowledge.ListKnowledgeBases(offset, req.Size)
}

// UpdateKnowledgeBase 更新知识库
func (s *Service) UpdateKnowledgeBase(ctx context.Context, id string, req *CreateKnowledgeBaseRequest) (*model.KnowledgeBase, error) {
	kb, err := s.repo.Knowledge.GetKnowledgeBaseByID(id)
	if err != nil {
		return nil, fmt.Errorf("knowledge base not found: %w", err)
	}

	kb.Name = req.Name
	kb.Description = req.Description
	kb.EmbeddingModel = req.EmbedModel

	if err := s.repo.Knowledge.UpdateKnowledgeBase(kb); err != nil {
		return nil, fmt.Errorf("failed to update knowledge base: %w", err)
	}

	return kb, nil
}

// DeleteKnowledgeBase 删除知识库
func (s *Service) DeleteKnowledgeBase(ctx context.Context, id string) error {
	if err := s.repo.Knowledge.DeleteKnowledgeBase(id); err != nil {
		return fmt.Errorf("failed to delete knowledge base: %w", err)
	}
	return nil
}

// UploadDocumentRequest 上传文档请求
type UploadDocumentRequest struct {
	KnowledgeBaseID string `json:"knowledge_base_id" binding:"required"`
	Title           string `json:"title" binding:"required"`
	FileName        string `json:"file_name" binding:"required"`
	FilePath        string `json:"file_path"`
	FileSize        int64  `json:"file_size"`
}

// UploadDocument 上传文档
func (s *Service) UploadDocument(ctx context.Context, req *UploadDocumentRequest) (*model.Document, error) {
	// 检查知识库是否存在
	_, err := s.repo.Knowledge.GetKnowledgeBaseByID(req.KnowledgeBaseID)
	if err != nil {
		return nil, fmt.Errorf("knowledge base not found: %w", err)
	}

	doc := &model.Document{
		ID:              uuid.New().String(),
		KnowledgeBaseID: req.KnowledgeBaseID,
		Title:           req.Title,
		FileName:        req.FileName,
		FilePath:        req.FilePath,
		FileSize:        req.FileSize,
		Status:          "pending",
	}

	if err := s.repo.Knowledge.CreateDocument(doc); err != nil {
		return nil, fmt.Errorf("failed to create document: %w", err)
	}

	return doc, nil
}

// GetDocument 获取文档
func (s *Service) GetDocument(ctx context.Context, id string) (*model.Document, error) {
	return s.repo.Knowledge.GetDocumentByID(id)
}

// ListDocumentsRequest 列出文档请求
type ListDocumentsRequest struct {
	KnowledgeBaseID string `json:"knowledge_base_id"`
	Page            int    `json:"page"`
	Size            int    `json:"size"`
}

// ListDocuments 列出文档
func (s *Service) ListDocuments(ctx context.Context, req *ListDocumentsRequest) ([]*model.Document, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 || req.Size > 100 {
		req.Size = 20
	}

	offset := (req.Page - 1) * req.Size
	return s.repo.Knowledge.ListDocuments(req.KnowledgeBaseID, offset, req.Size)
}

// DeleteDocument 删除文档
func (s *Service) DeleteDocument(ctx context.Context, id string) error {
	if err := s.repo.Knowledge.DeleteDocument(id); err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}
	return nil
}

// UpdateDocumentStatus 更新文档状态
func (s *Service) UpdateDocumentStatus(ctx context.Context, id, status string, chunkCount int) error {
	doc, err := s.repo.Knowledge.GetDocumentByID(id)
	if err != nil {
		return fmt.Errorf("document not found: %w", err)
	}

	doc.Status = status
	if chunkCount > 0 {
		doc.ChunkCount = chunkCount
	}

	return s.repo.Knowledge.UpdateDocument(doc)
}

// ProcessDocument 处理文档（解析、分块、向量化、索引）
// 直接调用 DocumentProcessor
func (s *Service) ProcessDocument(ctx context.Context, documentID, knowledgeBaseID string) (*ProcessResult, error) {
	return s.documentProcessor.Process(ctx, &ProcessRequest{
		DocumentID:      documentID,
		KnowledgeBaseID: knowledgeBaseID,
	})
}

// SearchKnowledgeRequest 搜索知识库请求
type SearchKnowledgeRequest struct {
	KnowledgeBaseID string `json:"knowledge_base_id"`
	Query           string `json:"query" binding:"required"`
	TopK            int    `json:"top_k"`
}

// SearchResult 搜索结果
type SearchResult struct {
	Content  string                 `json:"content"`
	Score    float64                `json:"score"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Search 搜索知识库（使用 Retriever）
// 这个方法在 service.go 中通过 newES8Retriever 提供的检索器来实现
func (s *Service) Search(ctx context.Context, req *SearchKnowledgeRequest) ([]*SearchResult, error) {
	// TODO: 实现搜索逻辑
	// 可以使用 service.go 中创建的 Retriever
	return nil, fmt.Errorf("search not implemented yet, use retriever from service.go")
}
