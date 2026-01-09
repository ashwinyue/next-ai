package repository

import "gorm.io/gorm"

// Repositories 仓库集合，用于统一管理所有仓库
// 使用接口类型便于依赖注入和单元测试
type Repositories struct {
	DB     *gorm.DB // 直接访问数据库
	Chat   *ChatRepository
	Agent  *AgentRepository
	Tool   *ToolRepository
	Auth   *AuthRepository
	Model  *ModelRepository
	File   *FileRepository
	Tenant *TenantRepository
	MCP    *MCPServiceRepository
}

// NewRepositories 创建所有仓库
func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		DB:     db,
		Chat:   NewChatRepository(db),
		Agent:  NewAgentRepository(db),
		Tool:   NewToolRepository(db),
		Auth:   NewAuthRepository(db),
		Model:  NewModelRepository(db),
		File:   NewFileRepository(db),
		Tenant: NewTenantRepository(db),
		MCP:    NewMCPServiceRepository(db),
	}
}
