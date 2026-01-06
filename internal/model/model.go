// Package model 提供模型相关的数据模型
package model

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ModelType 模型类型
type ModelType string

const (
	ModelTypeEmbedding   ModelType = "Embedding"   // 向量化模型
	ModelTypeRerank      ModelType = "Rerank"      // 重排序模型
	ModelTypeChatModel   ModelType = "ChatModel"   // 对话模型
)

// ModelStatus 模型状态
type ModelStatus string

const (
	ModelStatusActive         ModelStatus = "active"          // 活跃
	ModelStatusDownloading    ModelStatus = "downloading"     // 下载中
	ModelStatusDownloadFailed ModelStatus = "download_failed" // 下载失败
)

// ModelSource 模型来源
type ModelSource string

const (
	ModelSourceLocal       ModelSource = "local"       // 本地模型 (Ollama)
	ModelSourceRemote      ModelSource = "remote"      // 远程模型 (API)
	ModelSourceAliyun      ModelSource = "aliyun"      // 阿里云 DashScope
	ModelSourceZhipu       ModelSource = "zhipu"       // 智谱
	ModelSourceVolcengine  ModelSource = "volcengine"  // 火山引擎
	ModelSourceDeepseek    ModelSource = "deepseek"    // DeepSeek
	ModelSourceHunyuan     ModelSource = "hunyuan"     // 混元
	ModelSourceMinimax     ModelSource = "minimax"     // MiniMax
	ModelSourceOpenAI      ModelSource = "openai"      // OpenAI
	ModelSourceGemini      ModelSource = "gemini"      // Google Gemini
	ModelSourceMimo        ModelSource = "mimo"        // 面壁智能
	ModelSourceSiliconFlow ModelSource = "siliconflow" // SiliconFlow
	ModelSourceJina        ModelSource = "jina"        // Jina AI
	ModelSourceOpenRouter  ModelSource = "openrouter"  // OpenRouter
)

// EmbeddingParameters 向量化参数
type EmbeddingParameters struct {
	Dimension            int `json:"dimension"`               // 向量维度
	TruncatePromptTokens int `json:"truncate_prompt_tokens"` // 截断提示词 tokens
}

// ModelParameters 模型参数
type ModelParameters struct {
	BaseURL             string                `json:"base_url"`              // API 基础 URL
	APIKey              string                `json:"api_key"`               // API 密钥
	InterfaceType       string                `json:"interface_type"`        // 接口类型
	EmbeddingParameters EmbeddingParameters  `json:"embedding_parameters"` // 向量化参数
	ParameterSize       string                `json:"parameter_size"`        // 参数大小 (如 "7B", "13B")
	Provider            string                `json:"provider"`              // 提供商标识
	ExtraConfig         map[string]string     `json:"extra_config"`          // 额外配置
}

// Value 实现 driver.Valuer 接口
func (m ModelParameters) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// Scan 实现 sql.Scanner 接口
func (m *ModelParameters) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, m)
}

// Model AI 模型
type Model struct {
	ID          string           `json:"id" gorm:"type:varchar(36);primaryKey"`
	Name        string           `json:"name" gorm:"type:varchar(255);not null"`
	Type        ModelType        `json:"type" gorm:"type:varchar(50);not null"`
	Source      ModelSource      `json:"source" gorm:"type:varchar(50);not null"`
	Description string           `json:"description" gorm:"type:text"`
	Parameters  ModelParameters  `json:"parameters" gorm:"type:json;not null"`
	IsDefault   bool             `json:"is_default" gorm:"default:false"`
	IsBuiltin   bool             `json:"is_builtin" gorm:"default:false"`
	Status      ModelStatus      `json:"status" gorm:"type:varchar(50);default:'active'"`
	CreatedAt   int64            `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   int64            `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt    `json:"deleted_at,omitempty" gorm:"index"`
}

// BeforeCreate GORM 钩子，创建前生成 UUID
func (m *Model) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	return nil
}

// TableName 指定表名
func (Model) TableName() string {
	return "models"
}
