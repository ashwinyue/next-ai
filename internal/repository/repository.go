package repository

import "gorm.io/gorm"

// Repositories 仓库集合，用于统一管理所有仓库
type Repositories struct {
	Chat       *ChatRepository
	Agent      *AgentRepository
	Knowledge  *KnowledgeRepository
	Tool       *ToolRepository
	FAQ        *FAQRepository
}

// NewRepositories 创建所有仓库
func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		Chat:      NewChatRepository(db),
		Agent:     NewAgentRepository(db),
		Knowledge: NewKnowledgeRepository(db),
		Tool:      NewToolRepository(db),
		FAQ:       NewFAQRepository(db),
	}
}
