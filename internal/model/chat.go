package model

import "time"

// ChatSession 聊天会话
type ChatSession struct {
	ID        string        `gorm:"primaryKey;size:36"`
	UserID    string        `gorm:"index;size:36"`
	AgentID   string        `gorm:"index;size:36"`
	Title     string        `gorm:"size:255"`
	Status    string        `gorm:"index;size:20;default:active"`
	CreatedAt time.Time     `gorm:"autoCreateTime"`
	UpdatedAt time.Time     `gorm:"autoUpdateTime"`
	Messages  []ChatMessage `gorm:"foreignKey:SessionID"`
}

// ChatMessage 聊天消息
type ChatMessage struct {
	ID        string    `gorm:"primaryKey;size:36"`
	SessionID string    `gorm:"index;size:36"`
	Role      string    `gorm:"size:20;index"` // user, assistant, system, tool
	Content   string    `gorm:"type:text"`
	TokenUsed int       `gorm:"default:0"`
	CreatedAt time.Time `gorm:"autoCreateTime;index"`
}

// TableName 指定表名
func (ChatSession) TableName() string {
	return "chat_sessions"
}

func (ChatMessage) TableName() string {
	return "chat_messages"
}
