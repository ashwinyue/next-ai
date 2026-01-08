package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

// Agent 模式常量
const (
	AgentModeQuickAnswer    = "quick-answer"    // RAG 快速问答模式
	AgentModeSmartReasoning = "smart-reasoning" // ReAct 多步推理模式
)

// 内置 Agent ID 常量
const (
	BuiltinQuickAnswerID          = "builtin-quick-answer"
	BuiltinSmartReasoningID       = "builtin-smart-reasoning"
	BuiltinDeepResearcherID       = "builtin-deep-researcher"
	BuiltinDataAnalystID          = "builtin-data-analyst"
	BuiltinKnowledgeGraphExpertID = "builtin-knowledge-graph-expert"
	BuiltinDocumentAssistantID    = "builtin-document-assistant"
)

// 内置 Agent ID 列表
var BuiltinAgentIDs = []string{
	BuiltinQuickAnswerID,
	BuiltinSmartReasoningID,
	BuiltinDeepResearcherID,
	BuiltinDataAnalystID,
	BuiltinKnowledgeGraphExpertID,
	BuiltinDocumentAssistantID,
}

// IsBuiltinAgentID 检查是否是内置 Agent ID
func IsBuiltinAgentID(id string) bool {
	for _, builtinID := range BuiltinAgentIDs {
		if id == builtinID {
			return true
		}
	}
	return false
}

// ModelConfig represents AI model configuration
type ModelConfig struct {
	Provider   string                 `json:"provider"`
	Model      string                 `json:"model"`
	APIKey     string                 `json:"api_key,omitempty"`
	BaseURL    string                 `json:"base_url,omitempty"`
	Parameters map[string]interface{} `json:"parameters"`
}

// Agent AI代理配置
type Agent struct {
	ID           string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	Name         string         `gorm:"size:255;not null;uniqueIndex" json:"name"`
	Description  string         `gorm:"type:text" json:"description"`
	Avatar       string         `gorm:"size:64" json:"avatar,omitempty"`                // 头像/图标
	IsBuiltin    bool           `gorm:"default:false" json:"is_builtin"`                // 是否内置 Agent
	AgentMode    string         `gorm:"size:32;default:quick-answer" json:"agent_mode"` // Agent 模式
	SystemPrompt string         `gorm:"type:text" json:"system_prompt"`
	ModelConfig  ModelConfig    `gorm:"type:jsonb;serializer:json" json:"model_config"`
	Tools        JSON           `gorm:"type:jsonb" json:"tools"`
	MaxIter      int            `gorm:"default:10" json:"max_iterations"`
	Temperature  float64        `gorm:"default:0.7" json:"temperature"`          // 温度参数
	KnowledgeIDs pq.StringArray `gorm:"type:varchar(36)[]" json:"knowledge_ids"` // 关联的知识库 ID
	IsActive     bool           `gorm:"index;default:true" json:"is_active"`
	Metadata     JSON           `gorm:"type:jsonb" json:"metadata"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Agent) TableName() string {
	return "agents"
}

// ModelConfig 实现 driver.Valuer 和 sql.Scanner
func (m ModelConfig) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *ModelConfig) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, m)
}

func (ModelConfig) GormDataType() string {
	return "jsonb"
}
