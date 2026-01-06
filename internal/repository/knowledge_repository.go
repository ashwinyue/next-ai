package repository

import (
	"github.com/ashwinyue/next-ai/internal/model"
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

// GetChunkByID 获取单个分块
func (r *KnowledgeRepository) GetChunkByID(chunkID string) (*model.DocumentChunk, error) {
	var chunk model.DocumentChunk
	err := r.db.Where("id = ?", chunkID).First(&chunk).Error
	if err != nil {
		return nil, err
	}
	return &chunk, nil
}

// ListChunksByKnowledgeBaseID 获取知识库的所有分块（支持分页）
func (r *KnowledgeRepository) ListChunksByKnowledgeBaseID(kbID string, offset, limit int) ([]*model.DocumentChunk, int64, error) {
	var chunks []*model.DocumentChunk
	var total int64

	query := r.db.Model(&model.DocumentChunk{}).Where("knowledge_base_id = ?", kbID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("chunk_index ASC").Offset(offset).Limit(limit).Find(&chunks).Error
	return chunks, total, err
}

// UpdateChunk 更新分块
func (r *KnowledgeRepository) UpdateChunk(chunk *model.DocumentChunk) error {
	return r.db.Save(chunk).Error
}

// DeleteChunk 删除单个分块
func (r *KnowledgeRepository) DeleteChunk(chunkID string) error {
	return r.db.Delete(&model.DocumentChunk{}, "id = ?", chunkID).Error
}

// DeleteChunksByDocumentID 删除文档的所有分块
func (r *KnowledgeRepository) DeleteChunksByDocumentID(docID string) error {
	return r.db.Delete(&model.DocumentChunk{}, "document_id = ?", docID).Error
}

// DeleteChunksByKnowledgeBaseID 删除知识库的所有分块
func (r *KnowledgeRepository) DeleteChunksByKnowledgeBaseID(kbID string) error {
	return r.db.Delete(&model.DocumentChunk{}, "knowledge_base_id = ?", kbID).Error
}
