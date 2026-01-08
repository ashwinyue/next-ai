// Package repository 定义数据访问接口
// 接口抽象使依赖注入和单元测试成为可能
package repository

import "github.com/ashwinyue/next-ai/internal/model"

// ========== KnowledgeRepository 接口 ==========

// KnowledgeRepository 知识库数据访问接口
// 接口定义使 Service 层可以轻松 mock 进行单元测试
type KnowledgeRepository interface {
	// 知识库操作
	CreateKnowledgeBase(kb *model.KnowledgeBase) error
	GetKnowledgeBaseByID(id string) (*model.KnowledgeBase, error)
	ListKnowledgeBases(offset, limit int) ([]*model.KnowledgeBase, error)
	UpdateKnowledgeBase(kb *model.KnowledgeBase) error
	DeleteKnowledgeBase(id string) error

	// 文档操作
	CreateDocument(doc *model.Document) error
	GetDocumentByID(id string) (*model.Document, error)
	ListDocuments(kbID string, offset, limit int) ([]*model.Document, error)
	UpdateDocument(doc *model.Document) error
	DeleteDocument(id string) error

	// 分块操作
	CreateChunks(chunks []*model.DocumentChunk) error
	GetChunksByDocumentID(docID string) ([]*model.DocumentChunk, error)
	GetChunkByID(chunkID string) (*model.DocumentChunk, error)
	ListChunksByKnowledgeBaseID(kbID string, offset, limit int) ([]*model.DocumentChunk, int64, error)
	UpdateChunk(chunk *model.DocumentChunk) error
	DeleteChunk(chunkID string) error
	DeleteChunksByDocumentID(docID string) error
	DeleteChunksByKnowledgeBaseID(kbID string) error

	// 父子分块操作（WeKnora 对齐）
	ListChunksByParentID(parentID string) ([]*model.DocumentChunk, error)
	GetParentChunk(chunkID string) (*model.DocumentChunk, error)
	UpdateChunkMetadata(chunkID string, metadata model.JSON) error
	DeleteQuestionsFromChunkMetadata(chunkID string) error
}

// 确保 knowledgeRepositoryImpl 实现了接口
var _ KnowledgeRepository = (*knowledgeRepositoryImpl)(nil)
