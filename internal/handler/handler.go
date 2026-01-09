package handler

import (
	"github.com/ashwinyue/next-ai/internal/service"
)

// Handlers 处理器集合
type Handlers struct {
	Auth           *AuthHandler
	Chat           *ChatHandler
	Agent          *AgentHandler
	Tool           *ToolHandler
	Initialization *InitializationHandler
	Model          *ModelHandler
	MCPService     *MCPServiceHandler
	Tenant         *TenantHandler
	File           *FileHandler
	System         *SystemHandler
	Message        *MessageHandler
	WebSearch      *WebSearchHandler
}

// NewHandlers 创建所有处理器
func NewHandlers(svc *service.Services) *Handlers {
	return &Handlers{
		Auth:           NewAuthHandler(svc),
		Chat:           NewChatHandler(svc),
		Agent:          NewAgentHandler(svc),
		Tool:           NewToolHandler(svc),
		Initialization: NewInitializationHandler(svc),
		Model:          NewModelHandler(svc.Model),
		MCPService:     NewMCPServiceHandler(svc),
		Tenant:         NewTenantHandler(svc),
		File:           NewFileHandler(svc.File),
		System:         NewSystemHandler(svc),
		Message:        NewMessageHandler(svc.Chat),
		WebSearch:      NewWebSearchHandler(),
	}
}
