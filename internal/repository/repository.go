package repository

import "gorm.io/gorm"

// Repositories 仓库集合，用于统一管理所有仓库
// 使用接口类型便于依赖注入和单元测试
type Repositories struct {
	DB        *gorm.DB // 直接访问数据库
	Chat      *ChatRepository
	Agent     *AgentRepository
	Knowledge KnowledgeRepository // 接口类型，支持 mock
	Tool      *ToolRepository
	FAQ       *FAQRepository
	Auth      *AuthRepository
	Model     *ModelRepository
	Tag       *TagRepository
	File      *FileRepository
	Dataset   *DatasetRepository
	Tenant    *TenantRepository
	Evaluation *EvaluationTaskRepository
	MCP       *MCPServiceRepository
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
		Tenant:    NewTenantRepository(db),
		Evaluation: NewEvaluationTaskRepository(db),
		MCP:       NewMCPServiceRepository(db),
	}
}
