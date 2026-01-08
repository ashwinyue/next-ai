package model

import (
	"time"

	"gorm.io/gorm"
)

// KnowledgeTag 知识库标签
type KnowledgeTag struct {
	ID              string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	KnowledgeBaseID string         `gorm:"type:varchar(36);index:idx_knowledge_base_tag;not null" json:"knowledge_base_id"`
	Name            string         `gorm:"type:varchar(128);not null" json:"name"`
	Color           string         `gorm:"type:varchar(32)" json:"color,omitempty"`
	SortOrder       int            `gorm:"default:0" json:"sort_order"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (KnowledgeTag) TableName() string {
	return "knowledge_tags"
}

// DocumentTag 文档-标签关联表
type DocumentTag struct {
	ID         string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	DocumentID string    `gorm:"type:varchar(36);not null;index:idx_doc_tag;index:idx_doc_tag_unique,unique" json:"document_id"`
	TagID      string    `gorm:"type:varchar(36);not null;index:idx_doc_tag_unique,unique" json:"tag_id"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName 指定表名
func (DocumentTag) TableName() string {
	return "document_tags"
}
