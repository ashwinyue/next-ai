package model

import (
	"time"
)

// StoredFile 存储的文件信息
type StoredFile struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	TenantID    string    `json:"tenant_id" gorm:"index"`
	KnowledgeID string    `json:"knowledge_id" gorm:"index"`
	FileName    string    `json:"file_name"`
	FileSize    int64     `json:"file_size"`
	ContentType string    `json:"content_type"`
	StorageType string    `json:"storage_type"` // local, minio, cos
	FilePath    string    `json:"file_path"`    // 存储路径或URL
	FileID      string    `json:"file_id"`      // 存储系统中的文件ID
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 指定表名
func (StoredFile) TableName() string {
	return "stored_files"
}
