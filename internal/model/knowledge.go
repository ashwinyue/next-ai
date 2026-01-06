package model

import "time"

// KnowledgeBase 知识库
type KnowledgeBase struct {
	ID          string     `gorm:"primaryKey;size:36"`
	Name        string     `gorm:"size:100;uniqueIndex"`
	Description string     `gorm:"type:text"`
	EmbedModel  string     `gorm:"size:50"`
	IndexName   string     `gorm:"size:100;index"`
	CreatedAt   time.Time  `gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime"`
	Documents   []Document `gorm:"foreignKey:KnowledgeBaseID"`
}

// Document 文档
type Document struct {
	ID             string         `gorm:"primaryKey;size:36"`
	KnowledgeBaseID string        `gorm:"index;size:36"`
	Title          string         `gorm:"size:255"`
	FileName       string         `gorm:"size:255"`
	FilePath       string         `gorm:"size:500"`
	FileSize       int64          `gorm:"default:0"`
	Status         string         `gorm:"size:20;index:default:pending"`
	ChunkCount     int            `gorm:"default:0"`
	ErrorMsg       string         `gorm:"type:text"`
	CreatedAt      time.Time      `gorm:"autoCreateTime"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime"`
	Chunks         []DocumentChunk `gorm:"foreignKey:DocumentID"`
}

// DocumentChunk 文档分块
type DocumentChunk struct {
	ID              string    `gorm:"primaryKey;size:36"`
	DocumentID      string    `gorm:"index;size:36"`
	ChunkIndex      int       `gorm:"index"`
	Content         string    `gorm:"type:text"`
	Embedding       string    `gorm:"type:text"` // 存储向量字符串
	TokenCount      int       `gorm:"default:0"`
	Metadata        string    `gorm:"type:jsonb"`
	CreatedAt       time.Time `gorm:"autoCreateTime"`
}

// IndexMapping 索引映射
type IndexMapping struct {
	ID          string    `gorm:"primaryKey;size:36"`
	IndexName   string    `gorm:"uniqueIndex;size:100"`
	EmbedModel  string    `gorm:"size:50"`
	Dimension   int       `gorm:"default:0"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
}

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
