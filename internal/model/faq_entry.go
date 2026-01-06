package model

import (
	"encoding/json"
	"time"
)

// FAQEntry FAQ条目（完整版）
type FAQEntry struct {
	ID                string    `gorm:"primaryKey;size:36"`
	StandardQuestion string    `gorm:"type:text;not null"`        // 标准问题
	SimilarQuestions string    `gorm:"type:text"`                 // 相似问题（JSON数组）
	NegativeQuestions string    `gorm:"type:text"`                 // 反例问题（JSON数组）
	Answers           string    `gorm:"type:text;not null"`        // 答案（JSON数组）
	AnswerStrategy    string    `gorm:"size:20;default:all"`       // 答案策略: all, random
	Category          string    `gorm:"size:100;index"`             // 分类/标签
	IsEnabled         bool      `gorm:"index;default:true"`         // 是否启用
	IsRecommended     bool      `gorm:"default:false"`             // 是否推荐
	Priority          int       `gorm:"default:0"`                 // 优先级
	HitCount          int       `gorm:"default:0"`                 // 命中次数
	Source            string    `gorm:"size:100"`                  // 来源
	Version           int       `gorm:"default:1"`                 // 版本号
	ContentHash       string    `gorm:"size:64;index"`             // 内容哈希（用于去重）
	CreatedAt         time.Time `gorm:"autoCreateTime"`
	UpdatedAt         time.Time `gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (FAQEntry) TableName() string {
	return "faq_entries"
}

// GetSimilarQuestions 获取相似问题列表
func (f *FAQEntry) GetSimilarQuestions() []string {
	if f.SimilarQuestions == "" {
		return nil
	}
	var questions []string
	_ = json.Unmarshal([]byte(f.SimilarQuestions), &questions)
	return questions
}

// SetSimilarQuestions 设置相似问题列表
func (f *FAQEntry) SetSimilarQuestions(questions []string) error {
	if len(questions) == 0 {
		f.SimilarQuestions = ""
		return nil
	}
	data, err := json.Marshal(questions)
	if err != nil {
		return err
	}
	f.SimilarQuestions = string(data)
	return nil
}

// GetNegativeQuestions 获取反例问题列表
func (f *FAQEntry) GetNegativeQuestions() []string {
	if f.NegativeQuestions == "" {
		return nil
	}
	var questions []string
	_ = json.Unmarshal([]byte(f.NegativeQuestions), &questions)
	return questions
}

// SetNegativeQuestions 设置反例问题列表
func (f *FAQEntry) SetNegativeQuestions(questions []string) error {
	if len(questions) == 0 {
		f.NegativeQuestions = ""
		return nil
	}
	data, err := json.Marshal(questions)
	if err != nil {
		return err
	}
	f.NegativeQuestions = string(data)
	return nil
}

// GetAnswers 获取答案列表
func (f *FAQEntry) GetAnswers() []string {
	if f.Answers == "" {
		return nil
	}
	var answers []string
	_ = json.Unmarshal([]byte(f.Answers), &answers)
	return answers
}

// SetAnswers 设置答案列表
func (f *FAQEntry) SetAnswers(answers []string) error {
	if len(answers) == 0 {
		f.Answers = ""
		return nil
	}
	data, err := json.Marshal(answers)
	if err != nil {
		return err
	}
	f.Answers = string(data)
	return nil
}

// AnswerStrategy 答案策略常量
const (
	AnswerStrategyAll    = "all"    // 返回所有答案
	AnswerStrategyRandom = "random" // 随机返回一个答案
)

// FAQImportProgress FAQ导入进度
type FAQImportProgress struct {
	TaskID    string `gorm:"primaryKey;size:36"`
	Status    string `gorm:"size:20;index"` // pending, processing, completed, failed
	Progress  int    `gorm:"default:0"`      // 0-100
	Total     int    `gorm:"default:0"`
	Processed int    `gorm:"default:0"`
	Message   string `gorm:"type:text"`
	Error     string `gorm:"type:text"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName 指定表名
func (FAQImportProgress) TableName() string {
	return "faq_import_progress"
}

// FAQImportTaskStatus 导入任务状态常量
const (
	FAQImportStatusPending    = "pending"
	FAQImportStatusProcessing = "processing"
	FAQImportStatusCompleted  = "completed"
	FAQImportStatusFailed     = "failed"
)

// FAQEntryFieldsUpdate 单个FAQ条目的字段更新
type FAQEntryFieldsUpdate struct {
	IsEnabled     *bool   `json:"is_enabled,omitempty"`
	IsRecommended *bool   `json:"is_recommended,omitempty"`
	Category      *string `json:"category,omitempty"`
}

// FAQEntryFieldsBatchUpdate 批量更新FAQ条目字段的请求
type FAQEntryFieldsBatchUpdate struct {
	ByID      map[string]FAQEntryFieldsUpdate `json:"by_id,omitempty"`      // 按条目ID更新
	ByCategory map[string]FAQEntryFieldsUpdate `json:"by_category,omitempty"` // 按分类批量更新
	ExcludeIDs []string                         `json:"exclude_ids,omitempty"` // 在ByCategory操作中需要排除的ID列表
}
