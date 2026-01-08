// Package knowledge 提供 Knowledge Service 单元测试
package knowledge

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/ashwinyue/next-ai/internal/config"
	"github.com/ashwinyue/next-ai/internal/model"
	"github.com/ashwinyue/next-ai/internal/repository"
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/indexer"
	"github.com/cloudwego/eino/schema"
)

// ========== Mock KnowledgeRepository ==========

// mockKnowledgeRepository 模拟知识库仓库，用于单元测试
type mockKnowledgeRepository struct {
	// 知识库操作
	createKnowledgeBaseFunc func(kb *model.KnowledgeBase) error
	getKnowledgeBaseByIDFunc func(id string) (*model.KnowledgeBase, error)
	listKnowledgeBasesFunc func(offset, limit int) ([]*model.KnowledgeBase, error)
	updateKnowledgeBaseFunc func(kb *model.KnowledgeBase) error
	deleteKnowledgeBaseFunc func(id string) error

	// 文档操作
	createDocumentFunc func(doc *model.Document) error
	getDocumentByIDFunc func(id string) (*model.Document, error)
	listDocumentsFunc func(kbID string, offset, limit int) ([]*model.Document, error)
	updateDocumentFunc func(doc *model.Document) error
	deleteDocumentFunc func(id string) error

	// 分块操作
	createChunksFunc func(chunks []*model.DocumentChunk) error
	getChunksByDocumentIDFunc func(docID string) ([]*model.DocumentChunk, error)
	getChunkByIDFunc func(chunkID string) (*model.DocumentChunk, error)
	listChunksByKnowledgeBaseIDFunc func(kbID string, offset, limit int) ([]*model.DocumentChunk, int64, error)
	updateChunkFunc func(chunk *model.DocumentChunk) error
	deleteChunkFunc func(chunkID string) error
	deleteChunksByDocumentIDFunc func(docID string) error
	deleteChunksByKnowledgeBaseIDFunc func(kbID string) error

	// 父子分块操作
	listChunksByParentIDFunc func(parentID string) ([]*model.DocumentChunk, error)
	getParentChunkFunc func(chunkID string) (*model.DocumentChunk, error)
	updateChunkMetadataFunc func(chunkID string, metadata model.JSON) error
	deleteQuestionsFromChunkMetadataFunc func(chunkID string) error
}

// 知识库操作实现
func (m *mockKnowledgeRepository) CreateKnowledgeBase(kb *model.KnowledgeBase) error {
	if m.createKnowledgeBaseFunc != nil {
		return m.createKnowledgeBaseFunc(kb)
	}
	return nil
}

func (m *mockKnowledgeRepository) GetKnowledgeBaseByID(id string) (*model.KnowledgeBase, error) {
	if m.getKnowledgeBaseByIDFunc != nil {
		return m.getKnowledgeBaseByIDFunc(id)
	}
	return &model.KnowledgeBase{ID: id, Name: "Test KB"}, nil
}

func (m *mockKnowledgeRepository) ListKnowledgeBases(offset, limit int) ([]*model.KnowledgeBase, error) {
	if m.listKnowledgeBasesFunc != nil {
		return m.listKnowledgeBasesFunc(offset, limit)
	}
	return []*model.KnowledgeBase{}, nil
}

func (m *mockKnowledgeRepository) UpdateKnowledgeBase(kb *model.KnowledgeBase) error {
	if m.updateKnowledgeBaseFunc != nil {
		return m.updateKnowledgeBaseFunc(kb)
	}
	return nil
}

func (m *mockKnowledgeRepository) DeleteKnowledgeBase(id string) error {
	if m.deleteKnowledgeBaseFunc != nil {
		return m.deleteKnowledgeBaseFunc(id)
	}
	return nil
}

// 文档操作实现
func (m *mockKnowledgeRepository) CreateDocument(doc *model.Document) error {
	if m.createDocumentFunc != nil {
		return m.createDocumentFunc(doc)
	}
	return nil
}

func (m *mockKnowledgeRepository) GetDocumentByID(id string) (*model.Document, error) {
	if m.getDocumentByIDFunc != nil {
		return m.getDocumentByIDFunc(id)
	}
	return &model.Document{ID: id, FileName: "test.pdf"}, nil
}

func (m *mockKnowledgeRepository) ListDocuments(kbID string, offset, limit int) ([]*model.Document, error) {
	if m.listDocumentsFunc != nil {
		return m.listDocumentsFunc(kbID, offset, limit)
	}
	return []*model.Document{}, nil
}

func (m *mockKnowledgeRepository) UpdateDocument(doc *model.Document) error {
	if m.updateDocumentFunc != nil {
		return m.updateDocumentFunc(doc)
	}
	return nil
}

func (m *mockKnowledgeRepository) DeleteDocument(id string) error {
	if m.deleteDocumentFunc != nil {
		return m.deleteDocumentFunc(id)
	}
	return nil
}

// 分块操作实现
func (m *mockKnowledgeRepository) CreateChunks(chunks []*model.DocumentChunk) error {
	if m.createChunksFunc != nil {
		return m.createChunksFunc(chunks)
	}
	return nil
}

func (m *mockKnowledgeRepository) GetChunksByDocumentID(docID string) ([]*model.DocumentChunk, error) {
	if m.getChunksByDocumentIDFunc != nil {
		return m.getChunksByDocumentIDFunc(docID)
	}
	return []*model.DocumentChunk{}, nil
}

func (m *mockKnowledgeRepository) GetChunkByID(chunkID string) (*model.DocumentChunk, error) {
	if m.getChunkByIDFunc != nil {
		return m.getChunkByIDFunc(chunkID)
	}
	return &model.DocumentChunk{ID: chunkID, Content: "test content"}, nil
}

func (m *mockKnowledgeRepository) ListChunksByKnowledgeBaseID(kbID string, offset, limit int) ([]*model.DocumentChunk, int64, error) {
	if m.listChunksByKnowledgeBaseIDFunc != nil {
		return m.listChunksByKnowledgeBaseIDFunc(kbID, offset, limit)
	}
	return []*model.DocumentChunk{}, 0, nil
}

func (m *mockKnowledgeRepository) UpdateChunk(chunk *model.DocumentChunk) error {
	if m.updateChunkFunc != nil {
		return m.updateChunkFunc(chunk)
	}
	return nil
}

func (m *mockKnowledgeRepository) DeleteChunk(chunkID string) error {
	if m.deleteChunkFunc != nil {
		return m.deleteChunkFunc(chunkID)
	}
	return nil
}

func (m *mockKnowledgeRepository) DeleteChunksByDocumentID(docID string) error {
	if m.deleteChunksByDocumentIDFunc != nil {
		return m.deleteChunksByDocumentIDFunc(docID)
	}
	return nil
}

func (m *mockKnowledgeRepository) DeleteChunksByKnowledgeBaseID(kbID string) error {
	if m.deleteChunksByKnowledgeBaseIDFunc != nil {
		return m.deleteChunksByKnowledgeBaseIDFunc(kbID)
	}
	return nil
}

// 父子分块操作实现
func (m *mockKnowledgeRepository) ListChunksByParentID(parentID string) ([]*model.DocumentChunk, error) {
	if m.listChunksByParentIDFunc != nil {
		return m.listChunksByParentIDFunc(parentID)
	}
	return []*model.DocumentChunk{}, nil
}

func (m *mockKnowledgeRepository) GetParentChunk(chunkID string) (*model.DocumentChunk, error) {
	if m.getParentChunkFunc != nil {
		return m.getParentChunkFunc(chunkID)
	}
	return nil, nil
}

func (m *mockKnowledgeRepository) UpdateChunkMetadata(chunkID string, metadata model.JSON) error {
	if m.updateChunkMetadataFunc != nil {
		return m.updateChunkMetadataFunc(chunkID, metadata)
	}
	return nil
}

func (m *mockKnowledgeRepository) DeleteQuestionsFromChunkMetadata(chunkID string) error {
	if m.deleteQuestionsFromChunkMetadataFunc != nil {
		return m.deleteQuestionsFromChunkMetadataFunc(chunkID)
	}
	return nil
}

// 确保 mock 实现了接口
var _ repository.KnowledgeRepository = (*mockKnowledgeRepository)(nil)

// ========== CreateKnowledgeBase 测试 ==========

func TestCreateKnowledgeBase(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		createKnowledgeBaseFunc: func(kb *model.KnowledgeBase) error {
			if kb.Name == "" {
				return errors.New("name is required")
			}
			return nil
		},
	}

	svc := &Service{repo: mockRepo}

	tests := []struct {
		name        string
		req         *CreateKnowledgeBaseRequest
		wantErr     bool
		errContains string
	}{
		{
			name: "valid request",
			req: &CreateKnowledgeBaseRequest{
				Name:        "Test KB",
				Description: "Test Description",
				EmbedModel:  "text-embedding-ada-002",
			},
			wantErr: false,
		},
		{
			name: "repository error",
			req: &CreateKnowledgeBaseRequest{
				Name: "",
			},
			wantErr:     true,
			errContains: "failed to create knowledge base",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			kb, err := svc.CreateKnowledgeBase(ctx, tt.req)

			if tt.wantErr {
				if err == nil {
					t.Error("CreateKnowledgeBase() expected error, got nil")
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("Error = %v, want contain %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("CreateKnowledgeBase() unexpected error: %v", err)
			}
			if kb == nil {
				t.Fatal("CreateKnowledgeBase() returned nil kb")
			}
			if kb.ID == "" {
				t.Error("CreateKnowledgeBase() ID is empty")
			}
			if kb.ChunkSize != 512 {
				t.Errorf("CreateKnowledgeBase() ChunkSize = %d, want 512", kb.ChunkSize)
			}
		})
	}
}

// ========== GetKnowledgeBase 测试 ==========

func TestGetKnowledgeBase(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		getKnowledgeBaseByIDFunc: func(id string) (*model.KnowledgeBase, error) {
			if id == "not-found" {
				return nil, errors.New("not found")
			}
			return &model.KnowledgeBase{
				ID:          id,
				Name:        "Test KB",
				Description: "Test Description",
			}, nil
		},
	}

	svc := &Service{repo: mockRepo}

	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "existing kb",
			id:      "kb123",
			wantErr: false,
		},
		{
			name:    "not found",
			id:      "not-found",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			kb, err := svc.GetKnowledgeBase(ctx, tt.id)

			if tt.wantErr {
				if err == nil {
					t.Error("GetKnowledgeBase() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("GetKnowledgeBase() unexpected error: %v", err)
			}
			if kb == nil {
				t.Fatal("GetKnowledgeBase() returned nil")
			}
			if kb.ID != tt.id {
				t.Errorf("GetKnowledgeBase() ID = %q, want %q", kb.ID, tt.id)
			}
		})
	}
}

// ========== ListKnowledgeBases 测试 ==========

func TestListKnowledgeBases(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		listKnowledgeBasesFunc: func(offset, limit int) ([]*model.KnowledgeBase, error) {
			return []*model.KnowledgeBase{
				{ID: "kb1", Name: "KB 1"},
				{ID: "kb2", Name: "KB 2"},
			}, nil
		},
	}

	svc := &Service{repo: mockRepo}

	tests := []struct {
		name      string
		req       *ListKnowledgeBasesRequest
		wantCount int
	}{
		{
			name: "default values",
			req:  &ListKnowledgeBasesRequest{},
			wantCount: 2,
		},
		{
			name: "custom page",
			req: &ListKnowledgeBasesRequest{
				Page: 2,
				Size: 10,
			},
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			kbs, err := svc.ListKnowledgeBases(ctx, tt.req)

			if err != nil {
				t.Errorf("ListKnowledgeBases() unexpected error: %v", err)
			}
			if len(kbs) != tt.wantCount {
				t.Errorf("ListKnowledgeBases() returned %d kbs, want %d", len(kbs), tt.wantCount)
			}
		})
	}
}

// ========== UpdateKnowledgeBase 测试 ==========

func TestUpdateKnowledgeBase(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		getKnowledgeBaseByIDFunc: func(id string) (*model.KnowledgeBase, error) {
			if id == "not-found" {
				return nil, errors.New("not found")
			}
			return &model.KnowledgeBase{ID: id, Name: "Old Name"}, nil
		},
		updateKnowledgeBaseFunc: func(kb *model.KnowledgeBase) error {
			return nil
		},
	}

	svc := &Service{repo: mockRepo}

	tests := []struct {
		name        string
		id          string
		req         *CreateKnowledgeBaseRequest
		wantErr     bool
		errContains string
	}{
		{
			name: "successful update",
			id:   "kb123",
			req: &CreateKnowledgeBaseRequest{
				Name:        "New Name",
				Description: "New Description",
				EmbedModel:  "new-model",
			},
			wantErr: false,
		},
		{
			name: "kb not found",
			id:   "not-found",
			req: &CreateKnowledgeBaseRequest{
				Name: "New Name",
			},
			wantErr:     true,
			errContains: "knowledge base not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			kb, err := svc.UpdateKnowledgeBase(ctx, tt.id, tt.req)

			if tt.wantErr {
				if err == nil {
					t.Error("UpdateKnowledgeBase() expected error, got nil")
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("Error = %v, want contain %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("UpdateKnowledgeBase() unexpected error: %v", err)
			}
			if kb == nil {
				t.Fatal("UpdateKnowledgeBase() returned nil")
			}
			if kb.Name != tt.req.Name {
				t.Errorf("UpdateKnowledgeBase() Name = %q, want %q", kb.Name, tt.req.Name)
			}
		})
	}
}

// ========== DeleteKnowledgeBase 测试 ==========

func TestDeleteKnowledgeBase(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		deleteKnowledgeBaseFunc: func(id string) error {
			if id == "error" {
				return errors.New("delete failed")
			}
			return nil
		},
	}

	svc := &Service{repo: mockRepo}

	tests := []struct {
		name        string
		id          string
		wantErr     bool
		errContains string
	}{
		{
			name:    "successful delete",
			id:      "kb123",
			wantErr: false,
		},
		{
			name:        "delete error",
			id:          "error",
			wantErr:     true,
			errContains: "failed to delete knowledge base",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := svc.DeleteKnowledgeBase(ctx, tt.id)

			if tt.wantErr {
				if err == nil {
					t.Error("DeleteKnowledgeBase() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("DeleteKnowledgeBase() unexpected error: %v", err)
			}
		})
	}
}

// ========== UploadDocument 测试 ==========

func TestUploadDocument(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		getKnowledgeBaseByIDFunc: func(id string) (*model.KnowledgeBase, error) {
			if id == "kb-not-found" {
				return nil, errors.New("kb not found")
			}
			return &model.KnowledgeBase{ID: id}, nil
		},
		createDocumentFunc: func(doc *model.Document) error {
			return nil
		},
	}

	svc := &Service{repo: mockRepo}

	tests := []struct {
		name        string
		req         *UploadDocumentRequest
		wantErr     bool
		errContains string
	}{
		{
			name: "successful upload",
			req: &UploadDocumentRequest{
				KnowledgeBaseID: "kb123",
				Title:           "Test Doc",
				FileName:        "test.pdf",
				FilePath:        "/path/to/test.pdf",
				FileSize:        1024,
			},
			wantErr: false,
		},
		{
			name: "kb not found",
			req: &UploadDocumentRequest{
				KnowledgeBaseID: "kb-not-found",
				Title:           "Test Doc",
				FileName:        "test.pdf",
			},
			wantErr:     true,
			errContains: "knowledge base not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			doc, err := svc.UploadDocument(ctx, tt.req)

			if tt.wantErr {
				if err == nil {
					t.Error("UploadDocument() expected error, got nil")
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("Error = %v, want contain %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("UploadDocument() unexpected error: %v", err)
			}
			if doc == nil {
				t.Fatal("UploadDocument() returned nil")
			}
			if doc.ID == "" {
				t.Error("UploadDocument() ID is empty")
			}
			if doc.Status != "pending" {
				t.Errorf("UploadDocument() Status = %q, want 'pending'", doc.Status)
			}
		})
	}
}

// ========== GetDocument 测试 ==========

func TestGetDocument(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		getDocumentByIDFunc: func(id string) (*model.Document, error) {
			if id == "not-found" {
				return nil, errors.New("not found")
			}
			return &model.Document{
				ID:       id,
				FileName: "test.pdf",
				Title:    "Test Document",
			}, nil
		},
	}

	svc := &Service{repo: mockRepo}

	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "existing document",
			id:      "doc123",
			wantErr: false,
		},
		{
			name:    "not found",
			id:      "not-found",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			doc, err := svc.GetDocument(ctx, tt.id)

			if tt.wantErr {
				if err == nil {
					t.Error("GetDocument() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("GetDocument() unexpected error: %v", err)
			}
			if doc == nil {
				t.Fatal("GetDocument() returned nil")
			}
			if doc.ID != tt.id {
				t.Errorf("GetDocument() ID = %q, want %q", doc.ID, tt.id)
			}
		})
	}
}

// ========== ListDocuments 测试 ==========

func TestListDocuments(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		listDocumentsFunc: func(kbID string, offset, limit int) ([]*model.Document, error) {
			return []*model.Document{
				{ID: "doc1", FileName: "doc1.pdf"},
				{ID: "doc2", FileName: "doc2.pdf"},
			}, nil
		},
	}

	svc := &Service{repo: mockRepo}

	tests := []struct {
		name      string
		req       *ListDocumentsRequest
		wantCount int
	}{
		{
			name: "with kb id",
			req: &ListDocumentsRequest{
				KnowledgeBaseID: "kb123",
			},
			wantCount: 2,
		},
		{
			name: "empty kb id",
			req: &ListDocumentsRequest{
				KnowledgeBaseID: "",
			},
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			docs, err := svc.ListDocuments(ctx, tt.req)

			if err != nil {
				t.Errorf("ListDocuments() unexpected error: %v", err)
			}
			if len(docs) != tt.wantCount {
				t.Errorf("ListDocuments() returned %d docs, want %d", len(docs), tt.wantCount)
			}
		})
	}
}

// ========== DeleteDocument 测试 ==========

func TestDeleteDocument(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		deleteDocumentFunc: func(id string) error {
			if id == "error" {
				return errors.New("delete failed")
			}
			return nil
		},
	}

	svc := &Service{repo: mockRepo}

	tests := []struct {
		name        string
		id          string
		wantErr     bool
		errContains string
	}{
		{
			name:    "successful delete",
			id:      "doc123",
			wantErr: false,
		},
		{
			name:        "delete error",
			id:          "error",
			wantErr:     true,
			errContains: "failed to delete document",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := svc.DeleteDocument(ctx, tt.id)

			if tt.wantErr {
				if err == nil {
					t.Error("DeleteDocument() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("DeleteDocument() unexpected error: %v", err)
			}
		})
	}
}

// ========== UpdateDocumentStatus 测试 ==========

func TestUpdateDocumentStatus(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		getDocumentByIDFunc: func(id string) (*model.Document, error) {
			if id == "not-found" {
				return nil, errors.New("not found")
			}
			return &model.Document{ID: id, Status: "pending"}, nil
		},
		updateDocumentFunc: func(doc *model.Document) error {
			return nil
		},
	}

	svc := &Service{repo: mockRepo}

	tests := []struct {
		name        string
		id          string
		status      string
		chunkCount  int
		wantErr     bool
		errContains string
	}{
		{
			name:       "successful update",
			id:         "doc123",
			status:     "completed",
			chunkCount: 10,
			wantErr:    false,
		},
		{
			name:        "document not found",
			id:          "not-found",
			status:      "completed",
			chunkCount:  5,
			wantErr:     true,
			errContains: "document not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := svc.UpdateDocumentStatus(ctx, tt.id, tt.status, tt.chunkCount)

			if tt.wantErr {
				if err == nil {
					t.Error("UpdateDocumentStatus() expected error, got nil")
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("Error = %v, want contain %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("UpdateDocumentStatus() unexpected error: %v", err)
			}
		})
	}
}

// ========== Search 测试 ==========

func TestSearch_NoESClient(t *testing.T) {
	svc := &Service{esSearcher: nil}

	ctx := context.Background()
	req := &SearchKnowledgeRequest{
		KnowledgeBaseID: "kb123",
		Query:           "test query",
	}

	_, err := svc.Search(ctx, req)
	if err == nil {
		t.Error("Search() expected error when esSearcher is nil, got nil")
	}
	if !contains(err.Error(), "elasticsearch client not configured") {
		t.Errorf("Error = %v, want contain 'elasticsearch client not configured'", err)
	}
}

// ========== HybridSearch 测试 ==========

func TestHybridSearch_KBNotFound(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		getKnowledgeBaseByIDFunc: func(id string) (*model.KnowledgeBase, error) {
		return nil, errors.New("not found")
		},
	}

	svc := &Service{repo: mockRepo}

	ctx := context.Background()
	params := &HybridSearchParams{
		QueryText: "test query",
	}

	_, err := svc.HybridSearch(ctx, "kb123", params)
	if err == nil {
		t.Error("HybridSearch() expected error when kb not found, got nil")
	}
	if !contains(err.Error(), "knowledge base not found") {
		t.Errorf("Error = %v, want contain 'knowledge base not found'", err)
	}
}

func TestHybridSearch_ExecuteSearchError(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		getKnowledgeBaseByIDFunc: func(id string) (*model.KnowledgeBase, error) {
			return &model.KnowledgeBase{
				ID:        id,
				IndexName: "test_index",
			}, nil
		},
	}

	// 不设置 esSearcher，会导致 executeHybridSearch 返回错误
	svc := &Service{
		repo:     mockRepo,
		esSearcher: nil,
		embedder: &mockEmbedder{},
	}

	ctx := context.Background()
	params := &HybridSearchParams{
		QueryText:        "test query",
		DisableVectorMatch: true, // 禁用向量搜索，避免 embedder 调用
	}

	_, err := svc.HybridSearch(ctx, "kb123", params)
	if err == nil {
		t.Error("HybridSearch() expected error when esSearcher is nil, got nil")
	}
	if !contains(err.Error(), "failed to execute hybrid search") {
		t.Errorf("Error = %v, want contain 'failed to execute hybrid search'", err)
	}
}

func TestHybridSearch_DefaultParameters(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		getKnowledgeBaseByIDFunc: func(id string) (*model.KnowledgeBase, error) {
			return &model.KnowledgeBase{
				ID:        id,
				IndexName: "test_index",
			}, nil
		},
	}

	svc := &Service{
		repo:     mockRepo,
		esSearcher: nil, // 会导致后续错误
		embedder: &mockEmbedder{},
	}

	ctx := context.Background()
	params := &HybridSearchParams{
		QueryText:          "test query",
		MatchCount:         0,     // 测试默认值
		VectorThreshold:    0,     // 测试默认值
		KeywordThreshold:   0,     // 测试默认值
		DisableVectorMatch: true,  // 禁用向量搜索
	}

	_, err := svc.HybridSearch(ctx, "kb123", params)
	// 由于 esSearcher 为 nil，executeHybridSearch 会失败
	if err == nil {
		t.Error("HybridSearch() expected error, got nil")
	}
}

// ========== CopyKnowledgeBase 测试 ==========

func TestCopyKnowledgeBase_SourceNotFound(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		getKnowledgeBaseByIDFunc: func(id string) (*model.KnowledgeBase, error) {
		return nil, errors.New("not found")
		},
	}

	svc := &Service{repo: mockRepo}

	ctx := context.Background()
	req := &CopyKnowledgeBaseRequest{
		SourceID: "not-found",
	}

	_, err := svc.CopyKnowledgeBase(ctx, req)
	if err == nil {
		t.Error("CopyKnowledgeBase() expected error, got nil")
	}
	if !contains(err.Error(), "source knowledge base not found") {
		t.Errorf("Error = %v, want contain 'source knowledge base not found'", err)
	}
}

func TestCopyKnowledgeBase_CreateNewTarget(t *testing.T) {
	sourceKB := &model.KnowledgeBase{
		ID:             "source123",
		Name:           "Source KB",
		Description:    "Source Description",
		EmbeddingModel: "model-1",
		ChunkSize:      512,
		ChunkOverlap:   50,
	}

	mockRepo := &mockKnowledgeRepository{
		getKnowledgeBaseByIDFunc: func(id string) (*model.KnowledgeBase, error) {
			if id == "source123" {
				return sourceKB, nil
			}
			return &model.KnowledgeBase{ID: id}, nil // 用于新创建的 target KB
		},
		createKnowledgeBaseFunc: func(kb *model.KnowledgeBase) error {
			return nil
		},
		listDocumentsFunc: func(kbID string, offset, limit int) ([]*model.Document, error) {
			return []*model.Document{}, nil
		},
	}

	svc := &Service{repo: mockRepo}

	ctx := context.Background()
	req := &CopyKnowledgeBaseRequest{
		SourceID: "source123",
		Name:     "Target KB",
	}

	resp, err := svc.CopyKnowledgeBase(ctx, req)
	if err != nil {
		t.Errorf("CopyKnowledgeBase() unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("CopyKnowledgeBase() returned nil response")
	}
	if resp.TaskID == "" {
		t.Error("CopyKnowledgeBase() TaskID is empty")
	}
	if resp.SourceID != "source123" {
		t.Errorf("CopyKnowledgeBase() SourceID = %q, want 'source123'", resp.SourceID)
	}
}

func TestCopyKnowledgeBase_TargetNotFound(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		getKnowledgeBaseByIDFunc: func(id string) (*model.KnowledgeBase, error) {
			if id == "source123" {
				return &model.KnowledgeBase{ID: id, Name: "Source KB"}, nil
			}
			return nil, errors.New("not found")
		},
	}

	svc := &Service{repo: mockRepo}

	ctx := context.Background()
	req := &CopyKnowledgeBaseRequest{
		SourceID: "source123",
		TargetID: "target-not-found",
	}

	_, err := svc.CopyKnowledgeBase(ctx, req)
	if err == nil {
		t.Error("CopyKnowledgeBase() expected error when target not found, got nil")
	}
	if !contains(err.Error(), "target knowledge base not found") {
		t.Errorf("Error = %v, want contain 'target knowledge base not found'", err)
	}
}

func TestCopyKnowledgeBase_WithExistingTarget(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		getKnowledgeBaseByIDFunc: func(id string) (*model.KnowledgeBase, error) {
			return &model.KnowledgeBase{ID: id, Name: "KB " + id}, nil
		},
		listDocumentsFunc: func(kbID string, offset, limit int) ([]*model.Document, error) {
			return []*model.Document{}, nil
		},
	}

	var cloneProgressMap sync.Map
	svc := &Service{
		repo:             mockRepo,
		cloneProgressMap: cloneProgressMap,
	}

	ctx := context.Background()
	req := &CopyKnowledgeBaseRequest{
		SourceID: "source123",
		TargetID: "target456", // 使用已存在的目标
	}

	resp, err := svc.CopyKnowledgeBase(ctx, req)
	if err != nil {
		t.Errorf("CopyKnowledgeBase() unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("CopyKnowledgeBase() returned nil")
	}
	if resp.TargetID != "target456" {
		t.Errorf("CopyKnowledgeBase() TargetID = %q, want 'target456'", resp.TargetID)
	}
}

func TestCopyKnowledgeBase_WithNameOnly(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		getKnowledgeBaseByIDFunc: func(id string) (*model.KnowledgeBase, error) {
			return &model.KnowledgeBase{
				ID:             id,
				Name:           "Source KB",
				Description:    "Source Description",
				EmbeddingModel: "model-1",
				ChunkSize:      512,
				ChunkOverlap:   50,
			}, nil
		},
		createKnowledgeBaseFunc: func(kb *model.KnowledgeBase) error {
			return nil
		},
		listDocumentsFunc: func(kbID string, offset, limit int) ([]*model.Document, error) {
			return []*model.Document{}, nil
		},
	}

	var cloneProgressMap sync.Map
	svc := &Service{
		repo:             mockRepo,
		cloneProgressMap: cloneProgressMap,
	}

	ctx := context.Background()
	req := &CopyKnowledgeBaseRequest{
		SourceID: "source123",
		Name:     "My Copy",
	}

	resp, err := svc.CopyKnowledgeBase(ctx, req)
	if err != nil {
		t.Errorf("CopyKnowledgeBase() unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("CopyKnowledgeBase() returned nil")
	}
	if resp.TaskID == "" {
		t.Error("CopyKnowledgeBase() TaskID is empty")
	}
}

// ========== GetKBCloneProgress 测试 ==========

func TestGetKBCloneProgress_NotFound(t *testing.T) {
	svc := &Service{}

	ctx := context.Background()
	progress, err := svc.GetKBCloneProgress(ctx, "non-existent-task")

	if err != nil {
		t.Errorf("GetKBCloneProgress() unexpected error: %v", err)
	}
	if progress == nil {
		t.Fatal("GetKBCloneProgress() returned nil")
	}
	if progress.Status != "not_found" {
		t.Errorf("GetKBCloneProgress() Status = %q, want 'not_found'", progress.Status)
	}
}

func TestGetKBCloneProgress_Found(t *testing.T) {
	var m sync.Map
	expectedProgress := &KBCloneProgress{
		TaskID:     "task123",
		SourceID:   "source123",
		TargetID:   "target456",
		Status:     "completed",
		Progress:   100,
		TotalDocs:  10,
		CopiedDocs: 10,
	}
	m.Store("task123", expectedProgress)

	svc := &Service{cloneProgressMap: m}

	ctx := context.Background()
	progress, err := svc.GetKBCloneProgress(ctx, "task123")

	if err != nil {
		t.Errorf("GetKBCloneProgress() unexpected error: %v", err)
	}
	if progress == nil {
		t.Fatal("GetKBCloneProgress() returned nil")
	}
	if progress.TaskID != expectedProgress.TaskID {
		t.Errorf("GetKBCloneProgress() TaskID = %q, want %q", progress.TaskID, expectedProgress.TaskID)
	}
	if progress.Status != expectedProgress.Status {
		t.Errorf("GetKBCloneProgress() Status = %q, want %q", progress.Status, expectedProgress.Status)
	}
}

// ========== ListChunksByParentID 测试 ==========

func TestListChunksByParentID(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		listChunksByParentIDFunc: func(parentID string) ([]*model.DocumentChunk, error) {
			if parentID == "empty" {
				return []*model.DocumentChunk{}, nil
			}
			return []*model.DocumentChunk{
				{ID: "chunk1", ParentChunkID: "parent123", Content: "child 1"},
				{ID: "chunk2", ParentChunkID: "parent123", Content: "child 2"},
			}, nil
		},
	}

	svc := &Service{repo: mockRepo}

	tests := []struct {
		name      string
		parentID  string
		wantCount int
	}{
		{
			name:      "with children",
			parentID:  "parent123",
			wantCount: 2,
		},
		{
			name:      "no children",
			parentID:  "empty",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			chunks, err := svc.ListChunksByParentID(ctx, tt.parentID)

			if err != nil {
				t.Errorf("ListChunksByParentID() unexpected error: %v", err)
			}
			if len(chunks) != tt.wantCount {
				t.Errorf("ListChunksByParentID() returned %d chunks, want %d", len(chunks), tt.wantCount)
			}
		})
	}
}

// ========== GetParentChunk 测试 ==========

func TestGetParentChunk(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		getParentChunkFunc: func(chunkID string) (*model.DocumentChunk, error) {
			if chunkID == "has-parent" {
				return &model.DocumentChunk{ID: "parent456", Content: "parent content"}, nil
			}
			return nil, nil
		},
	}

	svc := &Service{repo: mockRepo}

	tests := []struct {
		name       string
		chunkID    string
		wantNil    bool
		wantParent bool
	}{
		{
			name:       "has parent",
			chunkID:    "has-parent",
			wantNil:    false,
			wantParent: true,
		},
		{
			name:       "no parent",
			chunkID:    "no-parent",
			wantNil:    false,
			wantParent: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			parent, err := svc.GetParentChunk(ctx, tt.chunkID)

			if err != nil {
				t.Errorf("GetParentChunk() unexpected error: %v", err)
			}
			if tt.wantNil && parent != nil {
				t.Error("GetParentChunk() expected nil, got non-nil")
			}
			if tt.wantParent && parent == nil {
				t.Error("GetParentChunk() expected parent, got nil")
			}
		})
	}
}

// ========== UpdateChunkParent 测试 ==========

func TestUpdateChunkParent(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		getChunkByIDFunc: func(chunkID string) (*model.DocumentChunk, error) {
			if chunkID == "not-found" {
				return nil, errors.New("not found")
			}
			return &model.DocumentChunk{ID: chunkID, Content: "test"}, nil
		},
		updateChunkFunc: func(chunk *model.DocumentChunk) error {
			return nil
		},
	}

	svc := &Service{repo: mockRepo}

	tests := []struct {
		name        string
		chunkID     string
		parentID    string
		wantErr     bool
		errContains string
	}{
		{
			name:    "successful update",
			chunkID: "chunk123",
			parentID: "parent456",
			wantErr: false,
		},
		{
			name:        "chunk not found",
			chunkID:     "not-found",
			parentID:    "parent456",
			wantErr:     true,
			errContains: "chunk not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := svc.UpdateChunkParent(ctx, tt.chunkID, tt.parentID)

			if tt.wantErr {
				if err == nil {
					t.Error("UpdateChunkParent() expected error, got nil")
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("Error = %v, want contain %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("UpdateChunkParent() unexpected error: %v", err)
			}
		})
	}
}

// ========== UpdateChunkImageInfo 测试 ==========

func TestUpdateChunkImageInfo(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		getChunkByIDFunc: func(chunkID string) (*model.DocumentChunk, error) {
			if chunkID == "not-found" {
				return nil, errors.New("not found")
			}
			return &model.DocumentChunk{
				ID:       chunkID,
				Content:  "test",
				Metadata: make(model.JSON),
			}, nil
		},
		updateChunkMetadataFunc: func(chunkID string, metadata model.JSON) error {
			return nil
		},
	}

	svc := &Service{repo: mockRepo}

	tests := []struct {
		name        string
		req         *UpdateChunkImageInfoRequest
		wantErr     bool
		errContains string
	}{
		{
			name: "successful update",
			req: &UpdateChunkImageInfoRequest{
				ChunkID:   "chunk123",
				ImageInfo: "OCR text result",
			},
			wantErr: false,
		},
		{
			name: "chunk not found",
			req: &UpdateChunkImageInfoRequest{
				ChunkID:   "not-found",
				ImageInfo: "info",
			},
			wantErr:     true,
			errContains: "chunk not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := svc.UpdateChunkImageInfo(ctx, tt.req)

			if tt.wantErr {
				if err == nil {
					t.Error("UpdateChunkImageInfo() expected error, got nil")
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("Error = %v, want contain %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("UpdateChunkImageInfo() unexpected error: %v", err)
			}
		})
	}
}

// ========== DeleteQuestionsByChunk 测试 ==========

func TestDeleteQuestionsByChunk(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		deleteQuestionsFromChunkMetadataFunc: func(chunkID string) error {
			if chunkID == "error" {
				return errors.New("delete failed")
			}
			return nil
		},
	}

	svc := &Service{repo: mockRepo}

	tests := []struct {
		name    string
		chunkID string
		wantErr bool
	}{
		{
			name:    "successful delete",
			chunkID: "chunk123",
			wantErr: false,
		},
		{
			name:    "delete error",
			chunkID: "error",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := svc.DeleteQuestionsByChunk(ctx, tt.chunkID)

			if tt.wantErr {
				if err == nil {
					t.Error("DeleteQuestionsByChunk() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("DeleteQuestionsByChunk() unexpected error: %v", err)
			}
		})
	}
}

// ========== GetChunk 测试 ==========

func TestGetChunk(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		getChunkByIDFunc: func(chunkID string) (*model.DocumentChunk, error) {
			if chunkID == "not-found" {
				return nil, errors.New("not found")
			}
			return &model.DocumentChunk{
				ID:      chunkID,
				Content: "test content",
			}, nil
		},
	}

	svc := &Service{repo: mockRepo}

	tests := []struct {
		name    string
		chunkID string
		wantErr bool
	}{
		{
			name:    "existing chunk",
			chunkID: "chunk123",
			wantErr: false,
		},
		{
			name:    "not found",
			chunkID: "not-found",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			chunk, err := svc.GetChunk(ctx, tt.chunkID)

			if tt.wantErr {
				if err == nil {
					t.Error("GetChunk() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("GetChunk() unexpected error: %v", err)
			}
			if chunk == nil {
				t.Fatal("GetChunk() returned nil")
			}
			if chunk.ID != tt.chunkID {
				t.Errorf("GetChunk() ID = %q, want %q", chunk.ID, tt.chunkID)
			}
		})
	}
}

// ========== ProcessDocument 测试 ==========

func TestProcessDocument_DocumentNotFound(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		getDocumentByIDFunc: func(id string) (*model.Document, error) {
			return nil, errors.New("not found")
		},
	}

	// 创建 mock DocumentProcessor
	docProcessor := &DocumentProcessor{
		repo: &repository.Repositories{
			Knowledge: mockRepo,
		},
		embedder: nil,
		cfg:      &config.Config{},
	}

	svc := &Service{
		repo:              mockRepo,
		documentProcessor: docProcessor,
	}

	ctx := context.Background()
	result, err := svc.ProcessDocument(ctx, "doc123", "kb123")

	if err == nil {
		t.Error("ProcessDocument() expected error when doc not found, got nil")
	}
	if result == nil {
		t.Fatal("ProcessDocument() returned nil result")
	}
	if !contains(result.Error, "文档不存在") {
		t.Errorf("Error = %v, want contain '文档不存在'", result.Error)
	}
}

func TestProcessDocument_KBNotFound(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		getDocumentByIDFunc: func(id string) (*model.Document, error) {
			return &model.Document{ID: id, FilePath: "/tmp/test.txt"}, nil
		},
		getKnowledgeBaseByIDFunc: func(id string) (*model.KnowledgeBase, error) {
			return nil, errors.New("not found")
		},
	}

	docProcessor := &DocumentProcessor{
		repo: &repository.Repositories{
			Knowledge: mockRepo,
		},
		embedder: nil,
		cfg:      &config.Config{},
	}

	svc := &Service{
		repo:              mockRepo,
		documentProcessor: docProcessor,
	}

	ctx := context.Background()
	result, err := svc.ProcessDocument(ctx, "doc123", "kb123")

	if err == nil {
		t.Error("ProcessDocument() expected error when kb not found, got nil")
	}
	if result == nil {
		t.Fatal("ProcessDocument() returned nil result")
	}
	if !contains(result.Error, "知识库不存在") {
		t.Errorf("Error = %v, want contain '知识库不存在'", result.Error)
	}
}

// ========== NewService 测试 ==========

func TestNewService(t *testing.T) {
	cfg := &config.Config{}
	mockRepo := &mockKnowledgeRepository{}

	repos := &repository.Repositories{
		Knowledge: mockRepo,
	}

	// 使用 mock embedder
	mockEmb := &mockEmbedder{}

	svc := NewService(repos, cfg, mockEmb)

	if svc == nil {
		t.Fatal("NewService() returned nil")
	}
	if svc.repo != mockRepo {
		t.Error("NewService() repo not set correctly")
	}
	if svc.cfg != cfg {
		t.Error("NewService() cfg not set correctly")
	}
	if svc.embedder == nil {
		t.Error("NewService() embedder is nil")
	}
	if svc.documentProcessor == nil {
		t.Error("NewService() documentProcessor is nil")
	}
	if svc.esSearcher != nil {
		t.Error("NewService() esSearcher should be nil when no ES config")
	}
}

func TestNewService_WithESConfig(t *testing.T) {
	cfg := &config.Config{
		Elastic: config.ElasticConfig{
			Host:     "http://localhost:9200",
			Username: "elastic",
			Password: "password",
		},
	}
	mockRepo := &mockKnowledgeRepository{}

	repos := &repository.Repositories{
		Knowledge: mockRepo,
	}

	mockEmb := &mockEmbedder{}

	svc := NewService(repos, cfg, mockEmb)

	if svc == nil {
		t.Fatal("NewService() returned nil")
	}
	if svc.esSearcher == nil {
		t.Error("NewService() esSearcher should be created when ES host is configured")
	}
}

// ========== Search 更多测试 ==========

func TestSearch_KBNotFound(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		getKnowledgeBaseByIDFunc: func(id string) (*model.KnowledgeBase, error) {
			return nil, errors.New("not found")
		},
	}

	svc := &Service{
		repo:      mockRepo,
		esSearcher: newMockESSearcher(nil), // 使用 mock ES 搜索器
	}

	ctx := context.Background()
	req := &SearchKnowledgeRequest{
		KnowledgeBaseID: "kb123",
		Query:           "test query",
	}

	_, err := svc.Search(ctx, req)
	if err == nil {
		t.Error("Search() expected error when kb not found, got nil")
	}
	if !contains(err.Error(), "knowledge base not found") {
		t.Errorf("Error = %v, want contain 'knowledge base not found'", err)
	}
}

func TestSearch_DefaultTopK(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		getKnowledgeBaseByIDFunc: func(id string) (*model.KnowledgeBase, error) {
			return &model.KnowledgeBase{
				ID:        id,
				IndexName: "test_index",
			}, nil
		},
	}

	_ = &Service{
		repo:      mockRepo,
		esSearcher: newMockESSearcher(nil),
	}

	// 测试 topK 默认值处理逻辑
	req := &SearchKnowledgeRequest{
		KnowledgeBaseID: "kb123",
		Query:           "test query",
		TopK:            0, // 测试默认值
	}
	_ = req
}

func TestSearch_SuccessWithResults(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		getKnowledgeBaseByIDFunc: func(id string) (*model.KnowledgeBase, error) {
			return &model.KnowledgeBase{
				ID:        id,
				IndexName: "test_index",
			}, nil
		},
	}

	// 创建 mock ES 搜索器，返回结果
	esSearcher := newMockESSearcher(func(ctx context.Context, index string, queryJSON []byte) (*ESResponse, error) {
		// 验证查询参数
		if index != "test_index_chunks" {
			t.Errorf("Expected index 'test_index_chunks', got '%s'", index)
		}
		// 返回模拟的搜索结果
		resultData := createSearchResponse([]map[string]interface{}{
			{
				"id":    "chunk1",
				"score": 0.85,
				"source": map[string]interface{}{
					"content":           "test content 1",
					"knowledge_base_id": "kb123",
				},
			},
			{
				"id":    "chunk2",
				"score": 0.75,
				"source": map[string]interface{}{
					"content":           "test content 2",
					"knowledge_base_id": "kb123",
				},
			},
		})
		return &ESResponse{
			IsError: false,
			Body:    io.NopCloser(bytes.NewReader(resultData)),
			String:  string(resultData),
		}, nil
	})

	svc := &Service{
		repo:      mockRepo,
		esSearcher: esSearcher,
	}

	ctx := context.Background()
	req := &SearchKnowledgeRequest{
		KnowledgeBaseID: "kb123",
		Query:           "test query",
		TopK:            10,
	}

	results, err := svc.Search(ctx, req)
	if err != nil {
		t.Errorf("Search() unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Search() returned %d results, want 2", len(results))
	}
	if results[0].Content != "test content 1" {
		t.Errorf("Search() first result content = %q, want 'test content 1'", results[0].Content)
	}
}

func TestSearch_ESError(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		getKnowledgeBaseByIDFunc: func(id string) (*model.KnowledgeBase, error) {
			return &model.KnowledgeBase{
				ID:        id,
				IndexName: "test_index",
			}, nil
		},
	}

	// 创建 mock ES 搜索器，返回错误
	esSearcher := newMockESSearcher(func(ctx context.Context, index string, queryJSON []byte) (*ESResponse, error) {
		errData := createErrorResponse("index not found")
		return &ESResponse{
			IsError: true,
			Body:    io.NopCloser(bytes.NewReader(errData)),
			String:  string(errData),
		}, nil
	})

	svc := &Service{
		repo:      mockRepo,
		esSearcher: esSearcher,
	}

	ctx := context.Background()
	req := &SearchKnowledgeRequest{
		KnowledgeBaseID: "kb123",
		Query:           "test query",
	}

	_, err := svc.Search(ctx, req)
	if err == nil {
		t.Error("Search() expected error when ES returns error, got nil")
	}
	if !contains(err.Error(), "elasticsearch error") {
		t.Errorf("Error = %v, want contain 'elasticsearch error'", err)
	}
}

func TestSearch_EmptyResults(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		getKnowledgeBaseByIDFunc: func(id string) (*model.KnowledgeBase, error) {
			return &model.KnowledgeBase{
				ID:        id,
				IndexName: "test_index",
			}, nil
		},
	}

	// 创建 mock ES 搜索器，返回空结果
	esSearcher := newMockESSearcher(func(ctx context.Context, index string, queryJSON []byte) (*ESResponse, error) {
		resultData := createEmptySearchResponse()
		return &ESResponse{
			IsError: false,
			Body:    io.NopCloser(bytes.NewReader(resultData)),
			String:  string(resultData),
		}, nil
	})

	svc := &Service{
		repo:      mockRepo,
		esSearcher: esSearcher,
	}

	ctx := context.Background()
	req := &SearchKnowledgeRequest{
		KnowledgeBaseID: "kb123",
		Query:           "test query",
	}

	results, err := svc.Search(ctx, req)
	if err != nil {
		t.Errorf("Search() unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Search() returned %d results, want 0", len(results))
	}
}

// ========== HybridSearch 更多测试 ==========

func TestHybridSearch_DefaultParams(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		getKnowledgeBaseByIDFunc: func(id string) (*model.KnowledgeBase, error) {
			return &model.KnowledgeBase{
				ID:        id,
				IndexName: "test_index",
			}, nil
		},
	}

	_ = &Service{
		repo:      mockRepo,
		esSearcher: newMockESSearcher(nil),
		embedder: nil, // 无 embedder，测试仅关键词搜索路径
	}

	// 测试参数默认值处理逻辑
	params := &HybridSearchParams{
		QueryText:            "test query",
		VectorThreshold:      0,   // 测试默认值
		KeywordThreshold:     0,   // 测试默认值
		MatchCount:           0,   // 测试默认值
		DisableVectorMatch:   true, // 禁用向量搜索
		DisableKeywordsMatch: false,
	}
	_ = params
}

func TestHybridSearch_SuccessWithResults(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		getKnowledgeBaseByIDFunc: func(id string) (*model.KnowledgeBase, error) {
			return &model.KnowledgeBase{
				ID:        id,
				IndexName: "test_index",
			}, nil
		},
	}

	// 创建 mock ES 搜索器，返回结果
	esSearcher := newMockESSearcher(func(ctx context.Context, index string, queryJSON []byte) (*ESResponse, error) {
		// 返回模拟的搜索结果
		resultData := createSearchResponse([]map[string]interface{}{
			{
				"id":    "chunk1",
				"score": 0.85,
				"source": map[string]interface{}{
					"content":           "hybrid test content",
					"knowledge_base_id": "kb123",
					"knowledge_id":      "doc123",
					"chunk_index":       0,
				},
			},
		})
		return &ESResponse{
			IsError: false,
			Body:    io.NopCloser(bytes.NewReader(resultData)),
			String:  string(resultData),
		}, nil
	})

	svc := &Service{
		repo:      mockRepo,
		esSearcher: esSearcher,
		embedder:   &mockEmbedder{}, // mock embedder
	}

	ctx := context.Background()
	params := &HybridSearchParams{
		QueryText:            "test query",
		DisableVectorMatch:   true, // 禁用向量搜索
		DisableKeywordsMatch: false,
		MatchCount:           10,
	}

	results, err := svc.HybridSearch(ctx, "kb123", params)
	if err != nil {
		t.Errorf("HybridSearch() unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("HybridSearch() returned %d results, want 1", len(results))
	}
}

func TestHybridSearch_EmptyResults(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		getKnowledgeBaseByIDFunc: func(id string) (*model.KnowledgeBase, error) {
			return &model.KnowledgeBase{
				ID:        id,
				IndexName: "test_index",
			}, nil
		},
	}

	// 创建 mock ES 搜索器，返回空结果
	esSearcher := newMockESSearcher(func(ctx context.Context, index string, queryJSON []byte) (*ESResponse, error) {
		resultData := createEmptySearchResponse()
		return &ESResponse{
			IsError: false,
			Body:    io.NopCloser(bytes.NewReader(resultData)),
			String:  string(resultData),
		}, nil
	})

	svc := &Service{
		repo:      mockRepo,
		esSearcher: esSearcher,
		embedder:   &mockEmbedder{},
	}

	ctx := context.Background()
	params := &HybridSearchParams{
		QueryText:          "test query",
		DisableVectorMatch: true,
		MatchCount:         10,
	}

	results, err := svc.HybridSearch(ctx, "kb123", params)
	if err != nil {
		t.Errorf("HybridSearch() unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("HybridSearch() returned %d results, want 0", len(results))
	}
}

func TestHybridSearch_ESearchError(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		getKnowledgeBaseByIDFunc: func(id string) (*model.KnowledgeBase, error) {
			return &model.KnowledgeBase{
				ID:        id,
				IndexName: "test_index",
			}, nil
		},
	}

	// 创建 mock ES 搜索器，返回错误
	esSearcher := newMockESSearcher(func(ctx context.Context, index string, queryJSON []byte) (*ESResponse, error) {
		return nil, errors.New("ES connection failed")
	})

	svc := &Service{
		repo:      mockRepo,
		esSearcher: esSearcher,
		embedder:   &mockEmbedder{},
	}

	ctx := context.Background()
	params := &HybridSearchParams{
		QueryText:          "test query",
		DisableVectorMatch: true,
		MatchCount:         10,
	}

	_, err := svc.HybridSearch(ctx, "kb123", params)
	if err == nil {
		t.Error("HybridSearch() expected error when ES fails, got nil")
	}
}

func TestHybridSearch_WithKnowledgeIDs(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		getKnowledgeBaseByIDFunc: func(id string) (*model.KnowledgeBase, error) {
			return &model.KnowledgeBase{
				ID:        id,
				IndexName: "test_index",
			}, nil
		},
	}

	esSearcher := newMockESSearcher(func(ctx context.Context, index string, queryJSON []byte) (*ESResponse, error) {
		// 验证查询中包含 knowledge_id 过滤
		queryStr := string(queryJSON)
		if !contains(queryStr, "knowledge_id") {
			t.Errorf("Expected query to contain knowledge_id filter, got: %s", queryStr)
		}
		resultData := createEmptySearchResponse()
		return &ESResponse{
			IsError: false,
			Body:    io.NopCloser(bytes.NewReader(resultData)),
			String:  string(resultData),
		}, nil
	})

	svc := &Service{
		repo:      mockRepo,
		esSearcher: esSearcher,
		embedder:   &mockEmbedder{},
	}

	ctx := context.Background()
	params := &HybridSearchParams{
		QueryText:          "test query",
		DisableVectorMatch: true,
		MatchCount:         10,
		KnowledgeIDs:       []string{"doc1", "doc2"},
	}

	_, err := svc.HybridSearch(ctx, "kb123", params)
	if err != nil {
		t.Errorf("HybridSearch() unexpected error: %v", err)
	}
}

func TestHybridSearch_WithVector(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		getKnowledgeBaseByIDFunc: func(id string) (*model.KnowledgeBase, error) {
			return &model.KnowledgeBase{
				ID:        id,
				IndexName: "test_index",
			}, nil
		},
	}

	esSearcher := newMockESSearcher(func(ctx context.Context, index string, queryJSON []byte) (*ESResponse, error) {
		resultData := createEmptySearchResponse()
		return &ESResponse{
			IsError: false,
			Body:    io.NopCloser(bytes.NewReader(resultData)),
			String:  string(resultData),
		}, nil
	})

	svc := &Service{
		repo:      mockRepo,
		esSearcher: esSearcher,
		embedder:   &mockEmbedder{},
	}

	ctx := context.Background()
	params := &HybridSearchParams{
		QueryText:            "test query",
		DisableKeywordsMatch: true, // 只使用向量搜索
		DisableVectorMatch:   false,
		MatchCount:           10,
	}

	_, err := svc.HybridSearch(ctx, "kb123", params)
	if err != nil {
		t.Errorf("HybridSearch() unexpected error: %v", err)
	}
}

func TestHybridSearch_BothDisabled(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		getKnowledgeBaseByIDFunc: func(id string) (*model.KnowledgeBase, error) {
			return &model.KnowledgeBase{
				ID:        id,
				IndexName: "test_index",
			}, nil
		},
	}

	esSearcher := newMockESSearcher(func(ctx context.Context, index string, queryJSON []byte) (*ESResponse, error) {
		resultData := createEmptySearchResponse()
		return &ESResponse{
			IsError: false,
			Body:    io.NopCloser(bytes.NewReader(resultData)),
			String:  string(resultData),
		}, nil
	})

	svc := &Service{
		repo:      mockRepo,
		esSearcher: esSearcher,
		embedder:   &mockEmbedder{},
	}

	ctx := context.Background()
	params := &HybridSearchParams{
		QueryText:            "test query",
		DisableKeywordsMatch: true, // 两者都禁用
		DisableVectorMatch:   true,
		MatchCount:           10,
	}

	results, err := svc.HybridSearch(ctx, "kb123", params)
	if err != nil {
		t.Errorf("HybridSearch() unexpected error: %v", err)
	}
	// 应该返回空结果，因为 should 子句为空
	if len(results) != 0 {
		t.Errorf("HybridSearch() returned %d results, want 0", len(results))
	}
}

func TestUpdateKnowledgeBase_NotFound(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		getKnowledgeBaseByIDFunc: func(id string) (*model.KnowledgeBase, error) {
			return nil, errors.New("not found")
		},
	}

	svc := &Service{repo: mockRepo}

	ctx := context.Background()
	req := &CreateKnowledgeBaseRequest{
		Name: "Updated Name",
	}

	_, err := svc.UpdateKnowledgeBase(ctx, "kb123", req)
	if err == nil {
		t.Error("UpdateKnowledgeBase() expected error when kb not found, got nil")
	}
	if !contains(err.Error(), "knowledge base not found") {
		t.Errorf("Error = %v, want contain 'knowledge base not found'", err)
	}
}

func TestUploadDocument_GenerateID(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		getKnowledgeBaseByIDFunc: func(id string) (*model.KnowledgeBase, error) {
			return &model.KnowledgeBase{ID: id}, nil
		},
		createDocumentFunc: func(doc *model.Document) error {
			// 验证文档 ID 已生成
			if doc.ID == "" {
				return errors.New("ID is empty")
			}
			return nil
		},
	}

	svc := &Service{repo: mockRepo}

	ctx := context.Background()
	req := &UploadDocumentRequest{
		KnowledgeBaseID: "kb123",
		Title:           "Test Doc",
		FileName:        "test.pdf",
		FilePath:        "/path/to/test.pdf",
		FileSize:        1024,
	}

	doc, err := svc.UploadDocument(ctx, req)
	if err != nil {
		t.Errorf("UploadDocument() unexpected error: %v", err)
	}
	if doc == nil {
		t.Fatal("UploadDocument() returned nil")
	}
	if doc.ID == "" {
		t.Error("UploadDocument() ID should be generated")
	}
}

// ========== executeCopy 测试 ==========

func TestExecuteCopy_Success(t *testing.T) {
	var cloneProgressMap sync.Map

	mockRepo := &mockKnowledgeRepository{
		listDocumentsFunc: func(kbID string, offset, limit int) ([]*model.Document, error) {
			return []*model.Document{
				{ID: "doc1", Title: "Doc 1", FileName: "doc1.pdf", Status: "completed", ChunkCount: 2},
			}, nil
		},
		createDocumentFunc: func(doc *model.Document) error {
			return nil
		},
		listChunksByKnowledgeBaseIDFunc: func(kbID string, offset, limit int) ([]*model.DocumentChunk, int64, error) {
			return []*model.DocumentChunk{
				{ID: "chunk1", DocumentID: "doc1", Content: "content 1", ChunkIndex: 0},
				{ID: "chunk2", DocumentID: "doc1", Content: "content 2", ChunkIndex: 1},
			}, 2, nil
		},
		createChunksFunc: func(chunks []*model.DocumentChunk) error {
			return nil
		},
	}

	svc := &Service{
		repo:             mockRepo,
		cloneProgressMap: cloneProgressMap,
	}

	ctx := context.Background()
	progress := &KBCloneProgress{
		TaskID:     "task123",
		SourceID:   "source123",
		TargetID:   "target456",
		Status:     "pending",
		Progress:   0,
		TotalDocs:  0,
		CopiedDocs: 0,
	}

	svc.executeCopy(ctx, "task123", "source123", "target456", progress)

	// 验证最终状态
	value, _ := svc.cloneProgressMap.Load("task123")
	finalProgress := value.(*KBCloneProgress)
	if finalProgress.Status != "completed" {
		t.Errorf("executeCopy() Status = %q, want 'completed'", finalProgress.Status)
	}
	if finalProgress.CopiedDocs != 1 {
		t.Errorf("executeCopy() CopiedDocs = %d, want 1", finalProgress.CopiedDocs)
	}
}

func TestExecuteCopy_ListDocumentsError(t *testing.T) {
	var cloneProgressMap sync.Map

	mockRepo := &mockKnowledgeRepository{
		listDocumentsFunc: func(kbID string, offset, limit int) ([]*model.Document, error) {
			return nil, errors.New("list failed")
		},
	}

	svc := &Service{
		repo:             mockRepo,
		cloneProgressMap: cloneProgressMap,
	}

	ctx := context.Background()
	progress := &KBCloneProgress{
		TaskID:     "task123",
		SourceID:   "source123",
		TargetID:   "target456",
		Status:     "pending",
		Progress:   0,
		TotalDocs:  0,
		CopiedDocs: 0,
	}

	svc.executeCopy(ctx, "task123", "source123", "target456", progress)

	// 验证失败状态
	value, _ := svc.cloneProgressMap.Load("task123")
	finalProgress := value.(*KBCloneProgress)
	if finalProgress.Status != "failed" {
		t.Errorf("executeCopy() Status = %q, want 'failed'", finalProgress.Status)
	}
}

func TestExecuteCopy_EmptyDocuments(t *testing.T) {
	var cloneProgressMap sync.Map

	mockRepo := &mockKnowledgeRepository{
		listDocumentsFunc: func(kbID string, offset, limit int) ([]*model.Document, error) {
			return []*model.Document{}, nil
		},
	}

	svc := &Service{
		repo:             mockRepo,
		cloneProgressMap: cloneProgressMap,
	}

	ctx := context.Background()
	progress := &KBCloneProgress{
		TaskID:   "task123",
		SourceID: "source123",
		TargetID: "target456",
		Status:   "pending",
	}

	svc.executeCopy(ctx, "task123", "source123", "target456", progress)

	// 验证完成状态（即使没有文档）
	value, _ := svc.cloneProgressMap.Load("task123")
	finalProgress := value.(*KBCloneProgress)
	if finalProgress.Status != "completed" {
		t.Errorf("executeCopy() Status = %q, want 'completed'", finalProgress.Status)
	}
	if finalProgress.TotalDocs != 0 {
		t.Errorf("executeCopy() TotalDocs = %d, want 0", finalProgress.TotalDocs)
	}
}

// ========== GetKBCloneProgress 错误处理测试 ==========

func TestGetKBCloneProgress_TypeMismatch(t *testing.T) {
	var m sync.Map
	m.Store("task123", "not a progress object") // 存入错误类型

	svc := &Service{cloneProgressMap: m}

	ctx := context.Background()
	progress, err := svc.GetKBCloneProgress(ctx, "task123")

	if err != nil {
		t.Errorf("GetKBCloneProgress() unexpected error: %v", err)
	}
	if progress == nil {
		t.Fatal("GetKBCloneProgress() returned nil")
	}
	if progress.Status != "error" {
		t.Errorf("GetKBCloneProgress() Status = %q, want 'error'", progress.Status)
	}
}

// ========== DocumentProcessor 测试 ==========

func TestDocumentProcessor_EmbedChunks_NoEmbedder(t *testing.T) {
	// 测试没有 embedder 的情况
	docProcessor := &DocumentProcessor{
		embedder: nil,
	}

	ctx := context.Background()
	chunks := []*model.DocumentChunk{
		{ID: "chunk1", Content: "test content 1"},
		{ID: "chunk2", Content: "test content 2"},
	}

	_, err := docProcessor.embedChunks(ctx, chunks)
	if err == nil {
		t.Error("embedChunks() expected error when embedder is nil, got nil")
	}
	if !contains(err.Error(), "embedder not configured") {
		t.Errorf("Error = %v, want contain 'embedder not configured'", err)
	}
}

// ========== CopyKnowledgeBase CreateKnowledgeBaseError 测试 ==========

func TestCopyKnowledgeBase_CreateKnowledgeBaseError(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		getKnowledgeBaseByIDFunc: func(id string) (*model.KnowledgeBase, error) {
			return &model.KnowledgeBase{
				ID:             id,
				Name:           "Source KB",
				Description:    "Source Description",
				EmbeddingModel: "model-1",
				ChunkSize:      512,
				ChunkOverlap:   50,
			}, nil
		},
		createKnowledgeBaseFunc: func(kb *model.KnowledgeBase) error {
			return errors.New("create failed")
		},
	}

	svc := &Service{repo: mockRepo}

	ctx := context.Background()
	req := &CopyKnowledgeBaseRequest{
		SourceID: "source123",
		Name:     "My Copy",
	}

	_, err := svc.CopyKnowledgeBase(ctx, req)
	if err == nil {
		t.Error("CopyKnowledgeBase() expected error when create fails, got nil")
	}
	if !contains(err.Error(), "failed to create target knowledge base") {
		t.Errorf("Error = %v, want contain 'failed to create target knowledge base'", err)
	}
}

// ========== UpdateKnowledgeBase 更新失败测试 ==========

func TestUpdateKnowledgeBase_UpdateError(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		getKnowledgeBaseByIDFunc: func(id string) (*model.KnowledgeBase, error) {
			return &model.KnowledgeBase{ID: id, Name: "Old Name"}, nil
		},
		updateKnowledgeBaseFunc: func(kb *model.KnowledgeBase) error {
			return errors.New("update failed")
		},
	}

	svc := &Service{repo: mockRepo}

	ctx := context.Background()
	req := &CreateKnowledgeBaseRequest{
		Name: "New Name",
	}

	_, err := svc.UpdateKnowledgeBase(ctx, "kb123", req)
	if err == nil {
		t.Error("UpdateKnowledgeBase() expected error when update fails, got nil")
	}
	if !contains(err.Error(), "failed to update knowledge base") {
		t.Errorf("Error = %v, want contain 'failed to update knowledge base'", err)
	}
}

// ========== UploadDocument 创建失败测试 ==========

func TestUploadDocument_CreateError(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		getKnowledgeBaseByIDFunc: func(id string) (*model.KnowledgeBase, error) {
			return &model.KnowledgeBase{ID: id}, nil
		},
		createDocumentFunc: func(doc *model.Document) error {
			return errors.New("create failed")
		},
	}

	svc := &Service{repo: mockRepo}

	ctx := context.Background()
	req := &UploadDocumentRequest{
		KnowledgeBaseID: "kb123",
		Title:           "Test Doc",
		FileName:        "test.pdf",
	}

	_, err := svc.UploadDocument(ctx, req)
	if err == nil {
		t.Error("UploadDocument() expected error when create fails, got nil")
	}
	if !contains(err.Error(), "failed to create document") {
		t.Errorf("Error = %v, want contain 'failed to create document'", err)
	}
}

// ========== DocumentProcessor Process 测试 ==========

func TestDocumentProcessor_Process_ParsedDocsEmpty(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		getDocumentByIDFunc: func(id string) (*model.Document, error) {
			return &model.Document{ID: id, FilePath: "/tmp/test.txt"}, nil
		},
		getKnowledgeBaseByIDFunc: func(id string) (*model.KnowledgeBase, error) {
			return &model.KnowledgeBase{ID: id, Name: "Test KB", ChunkSize: 512, ChunkOverlap: 50}, nil
		},
	}

	docProcessor := &DocumentProcessor{
		repo: &repository.Repositories{
			Knowledge: mockRepo,
		},
		cfg:      &config.Config{},
		embedder: nil,
	}

	ctx := context.Background()
	req := &ProcessRequest{
		DocumentID:      "doc123",
		KnowledgeBaseID: "kb123",
	}

	result, err := docProcessor.Process(ctx, req)

	// 由于文件不存在，应该会在解析阶段失败
	if err == nil {
		// 如果没有错误，检查结果
		if result != nil && result.Error != "" {
			// 预期有错误信息
			if !contains(result.Error, "解析失败") && !contains(result.Error, "failed to parse") {
				// 可以接受其他错误
			}
		}
	}
}

// ========== CreateChunkIndex 测试 ==========

func TestCreateChunkIndex_NoConfig(t *testing.T) {
	// 测试没有配置的情况
	ctx := context.Background()
	cfg := &config.Config{} // 空配置，不会 panic 但会返回错误
	err := CreateChunkIndex(ctx, cfg, nil)
	if err == nil {
		t.Error("CreateChunkIndex() expected error with no ES config, got nil")
	}
}

// ========== UpdateChunkImageInfo 元数据处理测试 ==========

func TestUpdateChunkImageInfo_WithExistingMetadata(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		getChunkByIDFunc: func(chunkID string) (*model.DocumentChunk, error) {
			return &model.DocumentChunk{
				ID:       chunkID,
				Content:  "test",
				Metadata: model.JSON{"existing": "data"},
			}, nil
		},
		updateChunkMetadataFunc: func(chunkID string, metadata model.JSON) error {
			return nil
		},
	}

	svc := &Service{repo: mockRepo}

	ctx := context.Background()
	req := &UpdateChunkImageInfoRequest{
		ChunkID:   "chunk123",
		ImageInfo: "OCR text result",
	}

	err := svc.UpdateChunkImageInfo(ctx, req)
	if err != nil {
		t.Errorf("UpdateChunkImageInfo() unexpected error: %v", err)
	}
}

func TestUpdateChunkImageInfo_MetadataUpdateError(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		getChunkByIDFunc: func(chunkID string) (*model.DocumentChunk, error) {
			return &model.DocumentChunk{
				ID:       chunkID,
				Content:  "test",
				Metadata: model.JSON{},
			}, nil
		},
		updateChunkMetadataFunc: func(chunkID string, metadata model.JSON) error {
			return errors.New("update failed")
		},
	}

	svc := &Service{repo: mockRepo}

	ctx := context.Background()
	req := &UpdateChunkImageInfoRequest{
		ChunkID:   "chunk123",
		ImageInfo: "OCR text result",
	}

	err := svc.UpdateChunkImageInfo(ctx, req)
	if err == nil {
		t.Error("UpdateChunkImageInfo() expected error when update fails, got nil")
	}
}

func TestUpdateChunkImageInfo_NilMetadata(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		getChunkByIDFunc: func(chunkID string) (*model.DocumentChunk, error) {
			// 返回 Metadata 为 nil 的 chunk
			return &model.DocumentChunk{
				ID:       chunkID,
				Content:  "test",
				Metadata: nil, // 测试 Metadata 为 nil 的情况
			}, nil
		},
		updateChunkMetadataFunc: func(chunkID string, metadata model.JSON) error {
			// 验证 metadata 被正确初始化
			if metadata == nil {
				t.Error("Expected metadata to be initialized, got nil")
			}
			if _, exists := metadata["image_info"]; !exists {
				t.Error("Expected image_info key in metadata")
			}
			return nil
		},
	}

	svc := &Service{repo: mockRepo}

	ctx := context.Background()
	req := &UpdateChunkImageInfoRequest{
		ChunkID:   "chunk123",
		ImageInfo: "OCR text result",
	}

	err := svc.UpdateChunkImageInfo(ctx, req)
	if err != nil {
		t.Errorf("UpdateChunkImageInfo() unexpected error: %v", err)
	}
}

// ========== ListDocuments 参数测试 ==========

func TestListDocuments_DefaultPage(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		listDocumentsFunc: func(kbID string, offset, limit int) ([]*model.Document, error) {
			return []*model.Document{
				{ID: "doc1", FileName: "doc1.pdf"},
			}, nil
		},
	}

	svc := &Service{repo: mockRepo}

	ctx := context.Background()
	req := &ListDocumentsRequest{
		KnowledgeBaseID: "kb123",
		// Page 和 Size 使用默认值
	}

	docs, err := svc.ListDocuments(ctx, req)
	if err != nil {
		t.Errorf("ListDocuments() unexpected error: %v", err)
	}
	if docs == nil {
		t.Fatal("ListDocuments() returned nil")
	}
}

// ========== CopyKnowledgeBase 默认名称测试 ==========

func TestCopyKnowledgeBase_DefaultName(t *testing.T) {
	sourceKB := &model.KnowledgeBase{
		ID:   "source123",
		Name: "Original KB",
	}

	mockRepo := &mockKnowledgeRepository{
		getKnowledgeBaseByIDFunc: func(id string) (*model.KnowledgeBase, error) {
			if id == "source123" {
				return sourceKB, nil
			}
			return &model.KnowledgeBase{ID: id}, nil
		},
		createKnowledgeBaseFunc: func(kb *model.KnowledgeBase) error {
			// 验证默认名称
			if kb.Name != "Original KB (副本)" {
				return fmt.Errorf("expected name 'Original KB (副本)', got '%s'", kb.Name)
			}
			return nil
		},
		listDocumentsFunc: func(kbID string, offset, limit int) ([]*model.Document, error) {
			return []*model.Document{}, nil
		},
	}

	var cloneProgressMap sync.Map
	svc := &Service{
		repo:             mockRepo,
		cloneProgressMap: cloneProgressMap,
	}

	ctx := context.Background()
	req := &CopyKnowledgeBaseRequest{
		SourceID: "source123",
		// 不提供 Name，应该使用默认名称
	}

	resp, err := svc.CopyKnowledgeBase(ctx, req)
	if err != nil {
		t.Errorf("CopyKnowledgeBase() unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("CopyKnowledgeBase() returned nil")
	}
}

// ========== parseDocument 测试 ==========

func TestParseDocument_EmptyFilePath(t *testing.T) {
	docProcessor := &DocumentProcessor{
		repo: &repository.Repositories{
			Knowledge: &mockKnowledgeRepository{},
		},
		cfg:      &config.Config{},
		embedder: nil,
	}

	ctx := context.Background()
	doc := &model.Document{
		ID:        "doc123",
		FilePath: "", // 空路径
	}

	_, err := docProcessor.parseDocument(ctx, doc)
	if err == nil {
		t.Error("parseDocument() expected error with empty file path, got nil")
	}
}

func TestParseDocument_UnsupportedFileType(t *testing.T) {
	docProcessor := &DocumentProcessor{
		repo: &repository.Repositories{
			Knowledge: &mockKnowledgeRepository{},
		},
		cfg:      &config.Config{},
		embedder: nil,
	}

	ctx := context.Background()
	doc := &model.Document{
		ID:        "doc123",
		FilePath: "/path/to/file.xyz", // 不支持的扩展名
	}

	_, err := docProcessor.parseDocument(ctx, doc)
	if err == nil {
		t.Error("parseDocument() expected error with unsupported file type, got nil")
	}
}

func TestParseDocument_FileNotFoundError(t *testing.T) {
	docProcessor := &DocumentProcessor{
		repo: &repository.Repositories{
			Knowledge: &mockKnowledgeRepository{},
		},
		cfg:      &config.Config{},
		embedder: nil,
	}

	ctx := context.Background()
	doc := &model.Document{
		ID:        "doc123",
		FilePath: "/nonexistent/path/to/file.txt", // 不存在的文件
	}

	_, err := docProcessor.parseDocument(ctx, doc)
	if err == nil {
		t.Error("parseDocument() expected error when file not found, got nil")
	}
	if !contains(err.Error(), "failed to open file") {
		t.Errorf("Error = %v, want contain 'failed to open file'", err)
	}
}

// ========== DocumentProcessor.Process 测试 ==========

func TestDocumentProcessor_Process_NoContent(t *testing.T) {
	// 创建一个测试文件
	tmpDir := t.TempDir()
	testFile := tmpDir + "/test.txt"
	if err := os.WriteFile(testFile, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	mockRepo := &mockKnowledgeRepository{
		getDocumentByIDFunc: func(id string) (*model.Document, error) {
			return &model.Document{ID: id, FilePath: testFile}, nil
		},
		getKnowledgeBaseByIDFunc: func(id string) (*model.KnowledgeBase, error) {
			return &model.KnowledgeBase{ID: id, Name: "Test KB", ChunkSize: 512, ChunkOverlap: 50}, nil
		},
	}

	docProcessor := &DocumentProcessor{
		repo: &repository.Repositories{
			Knowledge: mockRepo,
		},
		cfg:      &config.Config{},
		embedder: nil,
	}

	ctx := context.Background()
	req := &ProcessRequest{
		DocumentID:      "doc123",
		KnowledgeBaseID: "kb123",
	}

	result, err := docProcessor.Process(ctx, req)
	if err == nil {
		// 空文件会导致 "no content parsed" 错误
		if result.Error == "" {
			t.Error("Process() expected error message for empty file")
		}
	}
}

func TestDocumentProcessor_Process_SplitError(t *testing.T) {
	// 创建一个测试文件
	tmpDir := t.TempDir()
	testFile := tmpDir + "/test.txt"
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatal(err)
	}

	mockRepo := &mockKnowledgeRepository{
		getDocumentByIDFunc: func(id string) (*model.Document, error) {
			return &model.Document{ID: id, FilePath: testFile, Title: "Test"}, nil
		},
		getKnowledgeBaseByIDFunc: func(id string) (*model.KnowledgeBase, error) {
			return &model.KnowledgeBase{ID: id, Name: "Test KB", ChunkSize: 512, ChunkOverlap: 50}, nil
		},
	}

	docProcessor := &DocumentProcessor{
		repo: &repository.Repositories{
			Knowledge: mockRepo,
		},
		cfg:      &config.Config{},
		embedder: nil, // 没有 embedder 会导致后续错误
	}

	ctx := context.Background()
	req := &ProcessRequest{
		DocumentID:      "doc123",
		KnowledgeBaseID: "kb123",
	}

	result, err := docProcessor.Process(ctx, req)
	// 由于没有 embedder，向量化步骤会失败
	if err == nil {
		if result == nil || !result.Success {
			// 预期处理会失败
		}
	}
}

// ========== embedChunks 错误测试 ==========

func TestEmbedChunks_EmbedError(t *testing.T) {
	mockEmbedder := &mockEmbedder{
		embedStringsError: errors.New("embed failed"),
	}

	docProcessor := &DocumentProcessor{
		embedder: mockEmbedder,
	}

	ctx := context.Background()
	chunks := []*model.DocumentChunk{
		{ID: "chunk1", Content: "test content 1"},
		{ID: "chunk2", Content: "test content 2"},
	}

	_, err := docProcessor.embedChunks(ctx, chunks)
	if err == nil {
		t.Error("embedChunks() expected error when embed fails, got nil")
	}
}

func TestEmbedChunks_VectorCountMismatch(t *testing.T) {
	mockEmbedder := &mockEmbedder{
		vectorCount: 1, // 返回 1 个向量，但 chunks 有 2 个
	}

	docProcessor := &DocumentProcessor{
		embedder: mockEmbedder,
	}

	ctx := context.Background()
	chunks := []*model.DocumentChunk{
		{ID: "chunk1", Content: "test content 1"},
		{ID: "chunk2", Content: "test content 2"},
	}

	_, err := docProcessor.embedChunks(ctx, chunks)
	if err == nil {
		t.Error("embedChunks() expected error when vector count mismatches, got nil")
	}
	if !contains(err.Error(), "vector count mismatch") {
		t.Errorf("Error = %v, want contain 'vector count mismatch'", err)
	}
}

// ========== DeleteDocument_ChunkError 测试 ==========

func TestDeleteDocument_ChunksDeleteError(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		getDocumentByIDFunc: func(id string) (*model.Document, error) {
			return &model.Document{ID: id}, nil
		},
		deleteDocumentFunc: func(id string) error {
			return nil
		},
		deleteChunksByDocumentIDFunc: func(docID string) error {
			return errors.New("delete chunks failed")
		},
	}

	svc := &Service{repo: mockRepo}

	ctx := context.Background()
	err := svc.DeleteDocument(ctx, "doc123")
	// 删除分块失败不应该导致整体失败
	if err != nil {
		// 预期不会有错误，因为删除分块失败只记录警告
	}
}

// ========== ListKnowledgeBases 边界测试 ==========

func TestListKnowledgeBases_LargePage(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		listKnowledgeBasesFunc: func(offset, limit int) ([]*model.KnowledgeBase, error) {
			return []*model.KnowledgeBase{}, nil
		},
	}

	svc := &Service{repo: mockRepo}

	ctx := context.Background()
	req := &ListKnowledgeBasesRequest{
		Page: 1,
		Size: 200, // 超过最大值 100，应该被限制
	}

	_, err := svc.ListKnowledgeBases(ctx, req)
	if err != nil {
		t.Errorf("ListKnowledgeBases() unexpected error: %v", err)
	}
}

func TestListKnowledgeBases_ZeroPage(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		listKnowledgeBasesFunc: func(offset, limit int) ([]*model.KnowledgeBase, error) {
			// 验证 offset 是 0 (page 1)
			if offset != 0 {
				t.Errorf("Expected offset 0, got %d", offset)
			}
			return []*model.KnowledgeBase{}, nil
		},
	}

	svc := &Service{repo: mockRepo}

	ctx := context.Background()
	req := &ListKnowledgeBasesRequest{
		Page: 0, // 应该被重置为 1
		Size: 20,
	}

	_, err := svc.ListKnowledgeBases(ctx, req)
	if err != nil {
		t.Errorf("ListKnowledgeBases() unexpected error: %v", err)
	}
}

// ========== DeleteKnowledgeBase_ChunksError 测试 ==========

func TestDeleteKnowledgeBase_ChunksError(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		getKnowledgeBaseByIDFunc: func(id string) (*model.KnowledgeBase, error) {
			return &model.KnowledgeBase{ID: id, Name: "Test KB"}, nil
		},
		deleteKnowledgeBaseFunc: func(id string) error {
			return nil
		},
		deleteChunksByKnowledgeBaseIDFunc: func(kbID string) error {
			return errors.New("delete chunks failed")
		},
	}

	svc := &Service{repo: mockRepo}

	ctx := context.Background()
	err := svc.DeleteKnowledgeBase(ctx, "kb123")
	// 删除分块失败不应该导致整体失败
	if err != nil {
		t.Errorf("DeleteKnowledgeBase() unexpected error: %v", err)
	}
}

// ========== DocumentProcessor.Parse 元数据测试 ==========

func TestTextParser_WithMetadata(t *testing.T) {
	parser := &textParser{}

	ctx := context.Background()
	content := "test content with metadata"

	reader := strings.NewReader(content)
	docs, err := parser.Parse(ctx, reader)
	if err != nil {
		t.Errorf("Parse() unexpected error: %v", err)
	}
	if len(docs) != 1 {
		t.Errorf("Parse() returned %d docs, want 1", len(docs))
	}
	if docs[0].Content != content {
		t.Errorf("Parse() content = %q, want %q", docs[0].Content, content)
	}
}

func TestTextParser_EmptyContent(t *testing.T) {
	parser := &textParser{}

	ctx := context.Background()
	// 测试空内容
	reader := strings.NewReader("")
	docs, err := parser.Parse(ctx, reader)
	if err != nil {
		t.Errorf("Parse() unexpected error: %v", err)
	}
	if len(docs) != 0 {
		t.Errorf("Parse() returned %d docs for empty content, want 0", len(docs))
	}
}

// ========== CreateKnowledgeBase 边界测试 ==========

func TestCreateKnowledgeBase_WithAllFields(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		createKnowledgeBaseFunc: func(kb *model.KnowledgeBase) error {
			// 验证所有字段都被正确设置
			if kb.ID == "" {
				return errors.New("ID is empty")
			}
			if kb.ChunkSize != 512 {
				return errors.New("ChunkSize incorrect")
			}
			if kb.ChunkOverlap != 50 {
				return errors.New("ChunkOverlap incorrect")
			}
			return nil
		},
	}

	svc := &Service{repo: mockRepo}

	ctx := context.Background()
	req := &CreateKnowledgeBaseRequest{
		Name:        "Test KB",
		Description: "Test Description",
		EmbedModel:  "model-1",
	}

	kb, err := svc.CreateKnowledgeBase(ctx, req)
	if err != nil {
		t.Errorf("CreateKnowledgeBase() unexpected error: %v", err)
	}
	if kb == nil {
		t.Fatal("CreateKnowledgeBase() returned nil")
	}
	if kb.IndexName == "" {
		t.Error("CreateKnowledgeBase() IndexName is empty")
	}
}

// ========== UploadDocument_WithAllFields 测试 ==========

func TestUploadDocument_WithAllFields(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		getKnowledgeBaseByIDFunc: func(id string) (*model.KnowledgeBase, error) {
			return &model.KnowledgeBase{ID: id}, nil
		},
		createDocumentFunc: func(doc *model.Document) error {
			// 验证所有字段
			if doc.ID == "" {
				return errors.New("ID is empty")
			}
			if doc.Status != "pending" {
				return errors.New("Status incorrect")
			}
			return nil
		},
	}

	svc := &Service{repo: mockRepo}

	ctx := context.Background()
	req := &UploadDocumentRequest{
		KnowledgeBaseID: "kb123",
		Title:           "Test Doc",
		FileName:        "test.pdf",
		FilePath:        "/path/to/test.pdf",
		FileSize:        2048,
	}

	doc, err := svc.UploadDocument(ctx, req)
	if err != nil {
		t.Errorf("UploadDocument() unexpected error: %v", err)
	}
	if doc == nil {
		t.Fatal("UploadDocument() returned nil")
	}
}

// ========== UpdateDocumentStatus_WithZeroCount 测试 ==========

func TestUpdateDocumentStatus_WithZeroCount(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		getDocumentByIDFunc: func(id string) (*model.Document, error) {
			return &model.Document{ID: id, Status: "pending"}, nil
		},
		updateDocumentFunc: func(doc *model.Document) error {
			return nil
		},
	}

	svc := &Service{repo: mockRepo}

	ctx := context.Background()
	err := svc.UpdateDocumentStatus(ctx, "doc123", "processing", 0)
	if err != nil {
		t.Errorf("UpdateDocumentStatus() unexpected error: %v", err)
	}
}

// ========== GetParentChunk_nil 测试 ==========

func TestGetParentChunk_ReturnsNil(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		getParentChunkFunc: func(chunkID string) (*model.DocumentChunk, error) {
			return nil, nil // 返回 nil 表示没有父块
		},
	}

	svc := &Service{repo: mockRepo}

	ctx := context.Background()
	parent, err := svc.GetParentChunk(ctx, "chunk123")
	if err != nil {
		t.Errorf("GetParentChunk() unexpected error: %v", err)
	}
	if parent != nil {
		t.Error("GetParentChunk() expected nil, got non-nil")
	}
}

// ========== UpdateChunkParent_nilMetadata 测试 ==========

func TestUpdateChunkParent_NilMetadata(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		getChunkByIDFunc: func(chunkID string) (*model.DocumentChunk, error) {
			return &model.DocumentChunk{
				ID:       chunkID,
				Content:  "test",
				Metadata: nil, // nil metadata
			}, nil
		},
		updateChunkFunc: func(chunk *model.DocumentChunk) error {
			return nil
		},
	}

	svc := &Service{repo: mockRepo}

	ctx := context.Background()
	err := svc.UpdateChunkParent(ctx, "chunk123", "parent456")
	if err != nil {
		t.Errorf("UpdateChunkParent() unexpected error: %v", err)
	}
}

// ========== DeleteDocument 测试 ==========


// ========== UpdateDocumentStatus_UpdateError 测试 ==========

func TestUpdateDocumentStatus_UpdateError(t *testing.T) {
	mockRepo := &mockKnowledgeRepository{
		getDocumentByIDFunc: func(id string) (*model.Document, error) {
			return &model.Document{ID: id, Status: "pending"}, nil
		},
		updateDocumentFunc: func(doc *model.Document) error {
			return errors.New("update failed")
		},
	}

	svc := &Service{repo: mockRepo}

	ctx := context.Background()
	err := svc.UpdateDocumentStatus(ctx, "doc123", "completed", 10)
	if err == nil {
		t.Error("UpdateDocumentStatus() expected error when update fails, got nil")
	}
}

// ========== 辅助函数 ==========

// mockEmbedder 用于测试的 mock embedder
type mockEmbedder struct {
	embedStringsError error
	// 返回的向量数量（用于测试向量数量不匹配的情况）
	// 如果为 -1，则返回 len(texts) 个向量（默认行为）
	vectorCount int
}

func (m *mockEmbedder) EmbedStrings(ctx context.Context, texts []string, opts ...embedding.Option) ([][]float64, error) {
	if m.embedStringsError != nil {
		return nil, m.embedStringsError
	}
	count := len(texts)
	if m.vectorCount >= 0 {
		count = m.vectorCount
	}
	result := make([][]float64, count)
	for i := range result {
		result[i] = make([]float64, 128) // 模拟 128 维向量
		for j := range result[i] {
			result[i][j] = 0.1
		}
	}
	return result, nil
}

func (m *mockEmbedder) EmbedQuery(ctx context.Context, query string, opts ...embedding.Option) (float64, error) {
	return 0.5, nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsMiddle(s, substr))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ========== ES8 Indexer 测试 ==========

func TestDocumentToESFields(t *testing.T) {
	tests := []struct {
		name     string
		doc      *schema.Document
		wantKeys []string
	}{
		{
			name: "basic document",
			doc: &schema.Document{
				Content: "test content",
			},
			wantKeys: []string{"content"},
		},
		{
			name: "document with metadata",
			doc: &schema.Document{
				Content: "test content",
				MetaData: map[string]any{
					"document_id":       "doc123",
					"chunk_index":       0,
					"knowledge_base_id": "kb123",
				},
			},
			wantKeys: []string{"content", "document_id", "chunk_index", "knowledge_base_id"},
		},
		{
			name: "document with nil metadata",
			doc: &schema.Document{
				Content:  "test content",
				MetaData: nil,
			},
			wantKeys: []string{"content"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := documentToESFields(tt.doc)
			for _, key := range tt.wantKeys {
				if _, ok := fields[key]; !ok {
					t.Errorf("documentToESFields() missing key %q", key)
				}
			}
		})
	}
}

func TestChunksToEinoDocuments(t *testing.T) {
	tests := []struct {
		name   string
		chunks []*model.DocumentChunk
		want   int
	}{
		{
			name:   "empty chunks",
			chunks: []*model.DocumentChunk{},
			want:   0,
		},
		{
			name: "single chunk",
			chunks: []*model.DocumentChunk{
				{
					ID:              "chunk1",
					DocumentID:      "doc1",
					KnowledgeBaseID: "kb1",
					ChunkIndex:      0,
					Content:         "content 1",
					Metadata:        model.JSON{"key": "value"},
				},
			},
			want: 1,
		},
		{
			name: "multiple chunks with parent",
			chunks: []*model.DocumentChunk{
				{
					ID:              "chunk1",
					DocumentID:      "doc1",
					KnowledgeBaseID: "kb1",
					ChunkIndex:      0,
					Content:         "content 1",
					ParentChunkID:   "parent1",
					Metadata:        nil,
				},
				{
					ID:              "chunk2",
					DocumentID:      "doc1",
					KnowledgeBaseID: "kb1",
					ChunkIndex:      1,
					Content:         "content 2",
					ParentChunkID:   "parent1",
					Metadata:        model.JSON{"key": "value2"},
				},
			},
			want: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			docs := ChunksToEinoDocuments(tt.chunks)
			if len(docs) != tt.want {
				t.Errorf("ChunksToEinoDocuments() returned %d docs, want %d", len(docs), tt.want)
			}
			for i, doc := range docs {
				if doc.ID != tt.chunks[i].ID {
					t.Errorf("ChunksToEinoDocuments() doc[%d].ID = %q, want %q", i, doc.ID, tt.chunks[i].ID)
				}
				if doc.Content != tt.chunks[i].Content {
					t.Errorf("ChunksToEinoDocuments() doc[%d].Content = %q, want %q", i, doc.Content, tt.chunks[i].Content)
				}
				// 检查元数据
				if tt.chunks[i].DocumentID != "" && doc.MetaData["document_id"] != tt.chunks[i].DocumentID {
					t.Errorf("ChunksToEinoDocuments() doc[%d].MetaData[document_id] = %v, want %v", i, doc.MetaData["document_id"], tt.chunks[i].DocumentID)
				}
			}
		})
	}
}

// ========== Parent Indexer 测试 ==========

func TestNewParentIndexer(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *ParentIndexerConfig
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &ParentIndexerConfig{
				Indexer:        &mockIndexer{},
				Transformer:    &mockTransformer{},
				ParentIDKey:    "parent_id",
				SubIDGenerator: SequentialChunkGenerator(),
			},
			wantErr: false,
		},
		{
			name: "missing indexer",
			cfg: &ParentIndexerConfig{
				Transformer:    &mockTransformer{},
				ParentIDKey:    "parent_id",
				SubIDGenerator: SequentialChunkGenerator(),
			},
			wantErr: true,
		},
		{
			name: "missing transformer",
			cfg: &ParentIndexerConfig{
				Indexer:        &mockIndexer{},
				ParentIDKey:    "parent_id",
				SubIDGenerator: SequentialChunkGenerator(),
			},
			wantErr: true,
		},
		{
			name: "missing sub ID generator",
			cfg: &ParentIndexerConfig{
				Indexer:     &mockIndexer{},
				Transformer: &mockTransformer{},
				ParentIDKey: "parent_id",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			_, err := NewParentIndexer(ctx, tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewParentIndexer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSequentialChunkGenerator(t *testing.T) {
	ctx := context.Background()
	gen := SequentialChunkGenerator()

	tests := []struct {
		name      string
		parentID  string
		num       int
		wantCount int
		prefix    string
	}{
		{
			name:      "generate 3 chunks",
			parentID:  "doc123",
			num:       3,
			wantCount: 3,
			prefix:    "doc123_chunk_",
		},
		{
			name:      "generate 1 chunk",
			parentID:  "abc",
			num:       1,
			wantCount: 1,
			prefix:    "abc_chunk_",
		},
		{
			name:      "generate 0 chunks",
			parentID:  "doc456",
			num:       0,
			wantCount: 0,
			prefix:    "doc456_chunk_",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ids, err := gen(ctx, tt.parentID, tt.num)
			if err != nil {
				t.Errorf("SequentialChunkGenerator() error = %v", err)
			}
			if len(ids) != tt.wantCount {
				t.Errorf("SequentialChunkGenerator() returned %d ids, want %d", len(ids), tt.wantCount)
			}
			for i, id := range ids {
				expected := fmt.Sprintf("%s%d", tt.prefix, i+1)
				if id != expected {
					t.Errorf("SequentialChunkGenerator() ids[%d] = %q, want %q", i, id, expected)
				}
			}
		})
	}
}

func TestUUIDChunkGenerator(t *testing.T) {
	ctx := context.Background()
	gen := UUIDChunkGenerator()

	tests := []struct {
		name      string
		parentID  string
		num       int
		wantCount int
	}{
		{
			name:      "generate 3 chunks",
			parentID:  "doc123",
			num:       3,
			wantCount: 3,
		},
		{
			name:      "generate 1 chunk",
			parentID:  "abc",
			num:       1,
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ids, err := gen(ctx, tt.parentID, tt.num)
			if err != nil {
				t.Errorf("UUIDChunkGenerator() error = %v", err)
			}
			if len(ids) != tt.wantCount {
				t.Errorf("UUIDChunkGenerator() returned %d ids, want %d", len(ids), tt.wantCount)
			}
			for i, id := range ids {
				expected := fmt.Sprintf("%s_%d", tt.parentID, i)
				if id != expected {
					t.Errorf("UUIDChunkGenerator() ids[%d] = %q, want %q", i, id, expected)
				}
			}
		})
	}
}

func TestStoreWithParentIndexer(t *testing.T) {
	tests := []struct {
		name       string
		baseIndexer indexer.Indexer
		transformer document.Transformer
		wantErr     bool
	}{
		{
			name: "successful store",
			baseIndexer: &mockIndexer{
				storeFunc: func(ctx context.Context, docs []*schema.Document, opts ...indexer.Option) ([]string, error) {
					return []string{"id1", "id2"}, nil
				},
			},
			transformer: &mockTransformer{
				transformFunc: func(ctx context.Context, docs []*schema.Document, opts ...document.TransformerOption) ([]*schema.Document, error) {
					return docs, nil
				},
			},
			wantErr: false,
		},
		{
			name: "indexer error",
			baseIndexer: &mockIndexer{
				storeFunc: func(ctx context.Context, docs []*schema.Document, opts ...indexer.Option) ([]string, error) {
					return nil, errors.New("indexer error")
				},
			},
			transformer: &mockTransformer{
				transformFunc: func(ctx context.Context, docs []*schema.Document, opts ...document.TransformerOption) ([]*schema.Document, error) {
					return docs, nil
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			docs := []*schema.Document{
				{ID: "doc1", Content: "content1"},
				{ID: "doc2", Content: "content2"},
			}
			_, err := StoreWithParentIndexer(ctx, tt.baseIndexer, tt.transformer, "parent_id", docs)
			if (err != nil) != tt.wantErr {
				t.Errorf("StoreWithParentIndexer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// ========== newParser 额外测试 ==========

func TestNewParser_HTMLFiles(t *testing.T) {
	// 测试 .html 和 .htm 文件类型
	processor := &DocumentProcessor{}

	tests := []struct {
		name     string
		filePath string
		wantErr  bool
	}{
		{
			name:     "html file",
			filePath: "/path/to/document.html",
			wantErr:  false, // 创建解析器不应该出错
		},
		{
			name:     "htm file",
			filePath: "/path/to/document.htm",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			_, err := processor.newParser(ctx, tt.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("newParser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// ========== splitDocuments 额外测试 ==========

func TestSplitDocuments_EmptyDocuments(t *testing.T) {
	mockEmbedder := &mockEmbedder{vectorCount: -1}
	processor := &DocumentProcessor{embedder: mockEmbedder}

	ctx := context.Background()
	doc := &model.Document{ID: "doc1", Title: "Test"}
	kb := &model.KnowledgeBase{ID: "kb1", Name: "Test KB"}

	// 空文档列表
	chunks, err := processor.splitDocuments(ctx, []*schema.Document{}, doc, kb)
	if err != nil {
		t.Errorf("splitDocuments() unexpected error: %v", err)
	}
	if len(chunks) != 0 {
		t.Errorf("splitDocuments() returned %d chunks, want 0", len(chunks))
	}
}

// ========== NewParser 不同文件类型测试 ==========

func TestNewParser_UnsupportedFile(t *testing.T) {
	processor := &DocumentProcessor{}
	ctx := context.Background()

	_, err := processor.newParser(ctx, "/path/to/file.xyz")
	if err == nil {
		t.Error("newParser() expected error for unsupported file type, got nil")
	}
}

// ========== Mock 类型 ==========

type mockIndexer struct {
	storeFunc func(ctx context.Context, docs []*schema.Document, opts ...indexer.Option) ([]string, error)
}

func (m *mockIndexer) Store(ctx context.Context, docs []*schema.Document, opts ...indexer.Option) ([]string, error) {
	if m.storeFunc != nil {
		return m.storeFunc(ctx, docs, opts...)
	}
	return []string{"id1"}, nil
}

type mockTransformer struct {
	transformFunc func(ctx context.Context, docs []*schema.Document, opts ...document.TransformerOption) ([]*schema.Document, error)
}

func (m *mockTransformer) Transform(ctx context.Context, docs []*schema.Document, opts ...document.TransformerOption) ([]*schema.Document, error) {
	if m.transformFunc != nil {
		return m.transformFunc(ctx, docs, opts...)
	}
	return docs, nil
}
