// Package model 提供租户相关的数据模型
package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Tenant 租户
type Tenant struct {
	ID             string             `json:"id" gorm:"type:varchar(36);primaryKey"`
	Name           string             `json:"name" gorm:"type:varchar(255);not null"`
	Description    string             `json:"description" gorm:"type:text"`
	APIKey         string             `json:"api_key" gorm:"type:varchar(255);uniqueIndex"`
	Status         string             `json:"status" gorm:"type:varchar(50);default:'active'"`
	Business       string             `json:"business" gorm:"type:varchar(255)"`
	StorageQuota   int64              `json:"storage_quota" gorm:"default:10737418240"` // 10GB
	StorageUsed    int64              `json:"storage_used" gorm:"default:0"`

	// 配置字段（JSON）
	AgentConfig        *AgentConfig        `json:"agent_config,omitempty" gorm:"type:jsonb"`
	ContextConfig      *ContextConfig      `json:"context_config,omitempty" gorm:"type:jsonb"`
	WebSearchConfig    *WebSearchConfig    `json:"web_search_config,omitempty" gorm:"type:jsonb"`
	ConversationConfig *ConversationConfig `json:"conversation_config,omitempty" gorm:"type:jsonb"`

	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// BeforeCreate GORM 钩子
func (t *Tenant) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	if t.APIKey == "" {
		t.APIKey = "tenant_" + t.ID
	}
	return nil
}

// TableName 指定表名
func (Tenant) TableName() string {
	return "tenants"
}

// AgentConfig Agent 配置
type AgentConfig struct {
	MaxIterations     int      `json:"max_iterations"`
	ReflectionEnabled bool     `json:"reflection_enabled"`
	AllowedTools      []string `json:"allowed_tools"`
	Temperature       float64  `json:"temperature"`
	SystemPrompt      string   `json:"system_prompt,omitempty"`
}

// ContextConfig 上下文配置
type ContextConfig struct {
	MaxRounds        int     `json:"max_rounds"`
	EmbeddingTopK    int     `json:"embedding_top_k"`
	KeywordThreshold float64 `json:"keyword_threshold"`
	VectorThreshold  float64 `json:"vector_threshold"`
}

// WebSearchConfig 网络搜索配置
type WebSearchConfig struct {
	Enabled   bool   `json:"enabled"`
	MaxResults int    `json:"max_results"`
	Provider   string `json:"provider"`
}

// ConversationConfig 对话配置
type ConversationConfig struct {
	Prompt               string  `json:"prompt"`
	ContextTemplate      string  `json:"context_template"`
	Temperature          float64 `json:"temperature"`
	MaxCompletionTokens  int     `json:"max_completion_tokens"`
	MaxRounds            int     `json:"max_rounds"`
	EmbeddingTopK        int     `json:"embedding_top_k"`
	KeywordThreshold     float64 `json:"keyword_threshold"`
	VectorThreshold      float64 `json:"vector_threshold"`
	RerankTopK           int     `json:"rerank_top_k"`
	RerankThreshold      float64 `json:"rerank_threshold"`
	EnableRewrite        bool    `json:"enable_rewrite"`
	EnableQueryExpansion bool    `json:"enable_query_expansion"`
	FallbackStrategy     string  `json:"fallback_strategy"`
	FallbackResponse     string  `json:"fallback_response"`
}

// Value 实现 driver.Valuer for AgentConfig
func (c *AgentConfig) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c)
}

// Scan 实现 sql.Scanner for AgentConfig
func (c *AgentConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, c)
}

// Value 实现 driver.Valuer for ContextConfig
func (c *ContextConfig) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c)
}

// Scan 实现 sql.Scanner for ContextConfig
func (c *ContextConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, c)
}

// Value 实现 driver.Valuer for WebSearchConfig
func (c *WebSearchConfig) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c)
}

// Scan 实现 sql.Scanner for WebSearchConfig
func (c *WebSearchConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, c)
}

// Value 实现 driver.Valuer for ConversationConfig
func (c *ConversationConfig) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c)
}

// Scan 实现 sql.Scanner for ConversationConfig
func (c *ConversationConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, c)
}
