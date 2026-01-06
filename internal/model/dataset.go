package model

import (
	"time"
)

// Dataset 数据集
type Dataset struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"index"`
	Description string    `json:"description"`
	Type        string    `json:"type" gorm:"index"` // qa, evaluation
	Source      string    `json:"source"`             // file, manual, api
	SourcePath  string    `json:"source_path"`        // 文件路径
	RecordCount int       `json:"record_count"`
	TenantID    string    `json:"tenant_id" gorm:"index"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 指定表名
func (Dataset) TableName() string {
	return "datasets"
}

// QAPair QA 对
type QAPair struct {
	ID          string   `json:"id" gorm:"primaryKey"`
	DatasetID   string   `json:"dataset_id" gorm:"index"`
	Question    string   `json:"question"`
	Answer      string   `json:"answer"`
	PassageIDs  []string `json:"passage_ids" gorm:"type:text[]"`
	Passages    []string `json:"passages" gorm:"type:text[]"`
	Metadata    JSON     `json:"metadata"` // 额外元数据
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 指定表名
func (QAPair) TableName() string {
	return "qa_pairs"
}
