package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// JSON 类型用于存储 JSONB 数据
type JSON map[string]interface{}

func (j JSON) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSON) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

func (JSON) GormDataType() string {
	return "jsonb"
}

// KnowledgeBase 知识库
type KnowledgeBase struct {
	ID             string     `gorm:"type:varchar(36);primaryKey" json:"id"`
	Name           string     `gorm:"type:varchar(255);not null;uniqueIndex" json:"name"`
	Description    string     `gorm:"type:text" json:"description"`
	IndexName      string     `gorm:"type:varchar(255);not null;index" json:"index_name"`
	EmbeddingModel string     `gorm:"type:varchar(100)" json:"embedding_model"`
	ChunkSize      int        `gorm:"type:int;default:512" json:"chunk_size"`
	ChunkOverlap   int        `gorm:"type:int;default:50" json:"chunk_overlap"`
	CreatedAt      time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	Documents      []Document `gorm:"foreignKey:KnowledgeBaseID" json:"documents,omitempty"`
}

// Document 文档
type Document struct {
	ID              string          `gorm:"type:varchar(36);primaryKey" json:"id"`
	KnowledgeBaseID string          `gorm:"type:varchar(36);not null;index" json:"knowledge_base_id"`
	Title           string          `gorm:"type:varchar(500)" json:"title"`
	FileName        string          `gorm:"type:varchar(500);not null" json:"file_name"`
	FilePath        string          `gorm:"type:varchar(1000);not null" json:"file_path"`
	FileSize        int64           `gorm:"type:bigint" json:"file_size"`
	ContentType     string          `gorm:"type:varchar(100)" json:"content_type"`
	Status          string          `gorm:"type:varchar(50);default:'pending';index" json:"status"`
	Error           string          `gorm:"type:text" json:"error,omitempty"`
	Metadata        JSON            `gorm:"type:jsonb" json:"metadata,omitempty"`
	ChunkCount      int             `gorm:"type:int;default:0" json:"chunk_count"`
	ProcessedAt     *time.Time      `gorm:"type:timestamp" json:"processed_at,omitempty"`
	CreatedAt       time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
	Chunks          []DocumentChunk `gorm:"foreignKey:DocumentID" json:"chunks,omitempty"`
}

// DocumentChunk 文档分块
type DocumentChunk struct {
	ID              string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	DocumentID      string    `gorm:"type:varchar(36);not null;index" json:"document_id"`
	KnowledgeBaseID string    `gorm:"type:varchar(36);not null;index" json:"knowledge_base_id"`
	ChunkIndex      int       `gorm:"type:int;not null" json:"chunk_index"`
	Content         string    `gorm:"type:text;not null" json:"content"`
	Metadata        JSON      `gorm:"type:jsonb" json:"metadata,omitempty"`
	VectorID        string    `gorm:"type:varchar(255);index" json:"vector_id,omitempty"`
	ParentChunkID   string    `gorm:"type:varchar(36);index" json:"parent_chunk_id,omitempty"` // 父 Chunk ID（WeKnora 对齐）
	StartAt         int       `gorm:"type:int;default:0" json:"start_at,omitempty"`            // 在原文档中的起始位置
	EndAt           int       `gorm:"type:int;default:0" json:"end_at,omitempty"`              // 在原文档中的结束位置
	PreChunkID      string    `gorm:"type:varchar(36)" json:"pre_chunk_id,omitempty"`          // 前一个 Chunk ID
	NextChunkID     string    `gorm:"type:varchar(36)" json:"next_chunk_id,omitempty"`         // 下一个 Chunk ID
	IsEnabled       bool      `gorm:"default:true" json:"is_enabled"`                          // 是否启用
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// DocumentStatus constants
const (
	DocumentStatusPending    = "pending"
	DocumentStatusProcessing = "processing"
	DocumentStatusCompleted  = "completed"
	DocumentStatusFailed     = "failed"
)

// IndexMapping 索引映射
type IndexMapping struct {
	ID         string    `gorm:"primaryKey;size:36"`
	IndexName  string    `gorm:"uniqueIndex;size:100"`
	EmbedModel string    `gorm:"size:50"`
	Dimension  int       `gorm:"default:0"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
}

// TableName 指定表名
func (KnowledgeBase) TableName() string {
	return "knowledge_bases"
}

func (Document) TableName() string {
	return "documents"
}

func (DocumentChunk) TableName() string {
	return "document_chunks"
}

func (IndexMapping) TableName() string {
	return "index_mappings"
}
