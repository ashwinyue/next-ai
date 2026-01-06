package repository

import (
	"github.com/ashwinyue/next-rag/next-ai/internal/model"
	"gorm.io/gorm"
)

// KnowledgeRepository 知识库数据访问
type KnowledgeRepository struct {
	db *gorm.DB
}

// NewKnowledgeRepository 创建知识库仓库
func NewKnowledgeRepository(db *gorm.DB) *KnowledgeRepository {
	return &KnowledgeRepository{db: db}
}

// CreateKnowledgeBase 创建知识库
func (r *KnowledgeRepository) CreateKnowledgeBase(kb *model.KnowledgeBase) error {
	return r.db.Create(kb).Error
}

// GetKnowledgeBaseByID 获取知识库
func (r *KnowledgeRepository) GetKnowledgeBaseByID(id string) (*model.KnowledgeBase, error) {
	var kb model.KnowledgeBase
	err := r.db.Preload("Documents").Where("id = ?", id).First(&kb).Error
	if err != nil {
		return nil, err
	}
	return &kb, nil
}

// ListKnowledgeBases 列出知识库
func (r *KnowledgeRepository) ListKnowledgeBases(offset, limit int) ([]*model.KnowledgeBase, error) {
	var kbs []*model.KnowledgeBase
	err := r.db.Order("created_at DESC").Offset(offset).Limit(limit).Find(&kbs).Error
	return kbs, err
}

// UpdateKnowledgeBase 更新知识库
func (r *KnowledgeRepository) UpdateKnowledgeBase(kb *model.KnowledgeBase) error {
	return r.db.Save(kb).Error
}

// DeleteKnowledgeBase 删除知识库
func (r *KnowledgeRepository) DeleteKnowledgeBase(id string) error {
	return r.db.Delete(&model.KnowledgeBase{}, "id = ?", id).Error
}

// CreateDocument 创建文档
func (r *KnowledgeRepository) CreateDocument(doc *model.Document) error {
	return r.db.Create(doc).Error
}

// GetDocumentByID 获取文档
func (r *KnowledgeRepository) GetDocumentByID(id string) (*model.Document, error) {
	var doc model.Document
	err := r.db.Preload("Chunks").Where("id = ?", id).First(&doc).Error
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

// ListDocuments 列出文档
func (r *KnowledgeRepository) ListDocuments(kbID string, offset, limit int) ([]*model.Document, error) {
	var docs []*model.Document
	query := r.db.Order("created_at DESC").Offset(offset).Limit(limit)
	if kbID != "" {
		query = query.Where("knowledge_base_id = ?", kbID)
	}
	err := query.Find(&docs).Error
	return docs, err
}

// UpdateDocument 更新文档
func (r *KnowledgeRepository) UpdateDocument(doc *model.Document) error {
	return r.db.Save(doc).Error
}

// DeleteDocument 删除文档
func (r *KnowledgeRepository) DeleteDocument(id string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&model.DocumentChunk{}, "document_id = ?", id).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Document{}, "id = ?", id).Error
	})
}

// CreateChunks 创建文档分块
func (r *KnowledgeRepository) CreateChunks(chunks []*model.DocumentChunk) error {
	if len(chunks) == 0 {
		return nil
	}
	return r.db.CreateInBatches(chunks, 100).Error
}

// GetChunksByDocumentID 获取文档分块
func (r *KnowledgeRepository) GetChunksByDocumentID(docID string) ([]*model.DocumentChunk, error) {
	var chunks []*model.DocumentChunk
	err := r.db.Where("document_id = ?", docID).Order("chunk_index ASC").Find(&chunks).Error
	return chunks, err
}
