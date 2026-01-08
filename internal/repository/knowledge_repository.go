package repository

import (
	"github.com/ashwinyue/next-ai/internal/model"
	"gorm.io/gorm"
)

// knowledgeRepositoryImpl 知识库数据访问实现
type knowledgeRepositoryImpl struct {
	db *gorm.DB
}

// NewKnowledgeRepository 创建知识库仓库
func NewKnowledgeRepository(db *gorm.DB) KnowledgeRepository {
	return &knowledgeRepositoryImpl{db: db}
}

// CreateKnowledgeBase 创建知识库
func (r *knowledgeRepositoryImpl) CreateKnowledgeBase(kb *model.KnowledgeBase) error {
	return r.db.Create(kb).Error
}

// GetKnowledgeBaseByID 获取知识库
func (r *knowledgeRepositoryImpl) GetKnowledgeBaseByID(id string) (*model.KnowledgeBase, error) {
	var kb model.KnowledgeBase
	err := r.db.Preload("Documents").Where("id = ?", id).First(&kb).Error
	if err != nil {
		return nil, err
	}
	return &kb, nil
}

// ListKnowledgeBases 列出知识库
func (r *knowledgeRepositoryImpl) ListKnowledgeBases(offset, limit int) ([]*model.KnowledgeBase, error) {
	var kbs []*model.KnowledgeBase
	err := r.db.Order("created_at DESC").Offset(offset).Limit(limit).Find(&kbs).Error
	return kbs, err
}

// UpdateKnowledgeBase 更新知识库
func (r *knowledgeRepositoryImpl) UpdateKnowledgeBase(kb *model.KnowledgeBase) error {
	return r.db.Save(kb).Error
}

// DeleteKnowledgeBase 删除知识库
func (r *knowledgeRepositoryImpl) DeleteKnowledgeBase(id string) error {
	return r.db.Delete(&model.KnowledgeBase{}, "id = ?", id).Error
}

// CreateDocument 创建文档
func (r *knowledgeRepositoryImpl) CreateDocument(doc *model.Document) error {
	return r.db.Create(doc).Error
}

// GetDocumentByID 获取文档
func (r *knowledgeRepositoryImpl) GetDocumentByID(id string) (*model.Document, error) {
	var doc model.Document
	err := r.db.Preload("Chunks").Where("id = ?", id).First(&doc).Error
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

// ListDocuments 列出文档
func (r *knowledgeRepositoryImpl) ListDocuments(kbID string, offset, limit int) ([]*model.Document, error) {
	var docs []*model.Document
	query := r.db.Order("created_at DESC").Offset(offset).Limit(limit)
	if kbID != "" {
		query = query.Where("knowledge_base_id = ?", kbID)
	}
	err := query.Find(&docs).Error
	return docs, err
}

// UpdateDocument 更新文档
func (r *knowledgeRepositoryImpl) UpdateDocument(doc *model.Document) error {
	return r.db.Save(doc).Error
}

// DeleteDocument 删除文档
func (r *knowledgeRepositoryImpl) DeleteDocument(id string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&model.DocumentChunk{}, "document_id = ?", id).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Document{}, "id = ?", id).Error
	})
}

// CreateChunks 创建文档分块
func (r *knowledgeRepositoryImpl) CreateChunks(chunks []*model.DocumentChunk) error {
	if len(chunks) == 0 {
		return nil
	}
	return r.db.CreateInBatches(chunks, 100).Error
}

// GetChunksByDocumentID 获取文档分块
func (r *knowledgeRepositoryImpl) GetChunksByDocumentID(docID string) ([]*model.DocumentChunk, error) {
	var chunks []*model.DocumentChunk
	err := r.db.Where("document_id = ?", docID).Order("chunk_index ASC").Find(&chunks).Error
	return chunks, err
}

// GetChunkByID 获取单个分块
func (r *knowledgeRepositoryImpl) GetChunkByID(chunkID string) (*model.DocumentChunk, error) {
	var chunk model.DocumentChunk
	err := r.db.Where("id = ?", chunkID).First(&chunk).Error
	if err != nil {
		return nil, err
	}
	return &chunk, nil
}

// ListChunksByKnowledgeBaseID 获取知识库的所有分块（支持分页）
func (r *knowledgeRepositoryImpl) ListChunksByKnowledgeBaseID(kbID string, offset, limit int) ([]*model.DocumentChunk, int64, error) {
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
func (r *knowledgeRepositoryImpl) UpdateChunk(chunk *model.DocumentChunk) error {
	return r.db.Save(chunk).Error
}

// DeleteChunk 删除单个分块
func (r *knowledgeRepositoryImpl) DeleteChunk(chunkID string) error {
	return r.db.Delete(&model.DocumentChunk{}, "id = ?", chunkID).Error
}

// DeleteChunksByDocumentID 删除文档的所有分块
func (r *knowledgeRepositoryImpl) DeleteChunksByDocumentID(docID string) error {
	return r.db.Delete(&model.DocumentChunk{}, "document_id = ?", docID).Error
}

// DeleteChunksByKnowledgeBaseID 删除知识库的所有分块
func (r *knowledgeRepositoryImpl) DeleteChunksByKnowledgeBaseID(kbID string) error {
	return r.db.Delete(&model.DocumentChunk{}, "knowledge_base_id = ?", kbID).Error
}

// ListChunksByParentID 获取指定父 Chunk 的所有子 Chunk（WeKnora 对齐）
// 用于获取某个 Chunk 的所有子分块
func (r *knowledgeRepositoryImpl) ListChunksByParentID(parentID string) ([]*model.DocumentChunk, error) {
	var chunks []*model.DocumentChunk
	err := r.db.Where("parent_chunk_id = ?", parentID).Order("chunk_index ASC").Find(&chunks).Error
	return chunks, err
}

// GetParentChunk 获取指定 Chunk 的父 Chunk（WeKnora 对齐）
// 用于获取某个 Chunk 的父分块
func (r *knowledgeRepositoryImpl) GetParentChunk(chunkID string) (*model.DocumentChunk, error) {
	// 先获取当前 Chunk 的 ParentChunkID
	chunk, err := r.GetChunkByID(chunkID)
	if err != nil {
		return nil, err
	}
	if chunk.ParentChunkID == "" {
		return nil, nil // 没有父 Chunk
	}
	return r.GetChunkByID(chunk.ParentChunkID)
}

// UpdateChunkMetadata 更新分块元数据（用于图像信息更新等）
func (r *knowledgeRepositoryImpl) UpdateChunkMetadata(chunkID string, metadata model.JSON) error {
	return r.db.Model(&model.DocumentChunk{}).
		Where("id = ?", chunkID).
		Update("metadata", metadata).Error
}

// DeleteQuestionsFromChunkMetadata 从分块元数据中删除问题
func (r *knowledgeRepositoryImpl) DeleteQuestionsFromChunkMetadata(chunkID string) error {
	chunk, err := r.GetChunkByID(chunkID)
	if err != nil {
		return err
	}

	// 从 Metadata 中移除 questions 字段
	if chunk.Metadata != nil {
		delete(chunk.Metadata, "questions")
		return r.UpdateChunkMetadata(chunkID, chunk.Metadata)
	}

	return nil
}
