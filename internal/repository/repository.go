package repository

import "gorm.io/gorm"

// Repositories 仓库集合，用于统一管理所有仓库
type Repositories struct {
	DB        *gorm.DB // 直接访问数据库
	Chat      *ChatRepository
	Agent     *AgentRepository
	Knowledge *KnowledgeRepository
	Tool      *ToolRepository
	FAQ       *FAQRepository
	Auth      *AuthRepository
	Model     *ModelRepository
	Tag       *TagRepository
	File      *FileRepository
	Dataset   *DatasetRepository
}

// NewRepositories 创建所有仓库
func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		DB:        db,
		Chat:      NewChatRepository(db),
		Agent:     NewAgentRepository(db),
		Knowledge: NewKnowledgeRepository(db),
		Tool:      NewToolRepository(db),
		FAQ:       NewFAQRepository(db),
		Auth:      NewAuthRepository(db),
		Model:     NewModelRepository(db),
		Tag:       NewTagRepository(db),
		File:      NewFileRepository(db),
		Dataset:   NewDatasetRepository(db),
	}
}
