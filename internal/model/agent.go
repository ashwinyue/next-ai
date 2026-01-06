package model

import "time"

// Agent AI代理配置
type Agent struct {
	ID          string         `gorm:"primaryKey;size:36"`
	Name        string         `gorm:"size:100;uniqueIndex"`
	DisplayName string         `gorm:"size:255"`
	Description string         `gorm:"type:text"`
	Config      string         `gorm:"type:jsonb"` // 存储为 JSON 配置
	IsActive    bool           `gorm:"index;default:true"`
	CreatedAt   time.Time      `gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"`
}

// AgentConfig Agent配置结构
type AgentConfig struct {
	SystemPrompt string   `json:"system_prompt"`
	Temperature  float64  `json:"temperature"`
	MaxTokens    int      `json:"max_tokens"`
	Model        string   `json:"model"`
	Tools        []string `json:"tools"`
}

func (Agent) TableName() string {
	return "agents"
}
