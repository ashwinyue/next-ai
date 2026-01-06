// Package model 提供评估相关的数据模型
package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EvaluationTaskStatus 评估任务状态
type EvaluationTaskStatus string

const (
	EvaluationStatusPending   EvaluationTaskStatus = "pending"    // 待执行
	EvaluationStatusRunning   EvaluationTaskStatus = "running"    // 执行中
	EvaluationStatusCompleted EvaluationTaskStatus = "completed"  // 已完成
	EvaluationStatusFailed    EvaluationTaskStatus = "failed"     // 失败
)

// EvaluationTask 评估任务
type EvaluationTask struct {
	ID              string               `json:"id" gorm:"type:varchar(36);primaryKey"`
	DatasetID       string               `json:"dataset_id" gorm:"type:varchar(36);not null;index"`         // 数据集ID
	KnowledgeBaseID string               `json:"knowledge_base_id" gorm:"type:varchar(36);not null;index"` // 知识库ID
	ChatModelID     string               `json:"chat_model_id" gorm:"type:varchar(36)"`                   // 对话模型ID
	RerankModelID   string               `json:"rerank_model_id" gorm:"type:varchar(36)"`                 // 重排序模型ID
	Status          EvaluationTaskStatus `json:"status" gorm:"type:varchar(20);default:'pending'"`
	Progress        int                  `json:"progress" gorm:"default:0"`              // 进度 0-100
	TotalQuestions  int                  `json:"total_questions" gorm:"default:0"`      // 总问题数
	CompletedCount  int                  `json:"completed_count" gorm:"default:0"`      // 已完成数量

	// 评估结果
	Result *EvaluationResult `json:"result,omitempty" gorm:"embedded;embeddedPrefix:result_"`

	// 错误信息
	ErrorMsg string `json:"error_msg,omitempty" gorm:"type:text"`

	// 时间戳
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	StartedAt *time.Time     `json:"started_at,omitempty"`
	CompletedAt *time.Time   `json:"completed_at,omitempty"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// EvaluationResult 评估结果
type EvaluationResult struct {
	Precision    float64 `json:"precision"`     // 精确率
	Recall       float64 `json:"recall"`        // 召回率
	F1Score      float64 `json:"f1_score"`      // F1 分数
	AvgResponseTime float64 `json:"avg_response_time"` // 平均响应时间(ms)
	TotalCorrect int     `json:"total_correct"` // 正确数量
	TotalWrong   int     `json:"total_wrong"`   // 错误数量
}

// BeforeCreate GORM 钩子，创建前生成 UUID
func (e *EvaluationTask) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuid.New().String()
	}
	return nil
}

// TableName 指定表名
func (EvaluationTask) TableName() string {
	return "evaluation_tasks"
}

// EvaluationDataset 评估数据集
type EvaluationDataset struct {
	ID          string    `json:"id" gorm:"type:varchar(36);primaryKey"`
	Name        string    `json:"name" gorm:"type:varchar(255);not null"`
	Description string    `json:"description" gorm:"type:text"`
	Questions   int       `json:"questions" gorm:"default:0"` // 问题数量
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// BeforeCreate GORM 钩子
func (e *EvaluationDataset) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuid.New().String()
	}
	return nil
}

// TableName 指定表名
func (EvaluationDataset) TableName() string {
	return "evaluation_datasets"
}
