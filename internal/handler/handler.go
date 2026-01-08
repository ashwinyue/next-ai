package handler

import (
	"github.com/ashwinyue/next-ai/internal/service"
)

// Handlers 处理器集合
type Handlers struct {
	Auth           *AuthHandler
	Chat           *ChatHandler
	Agent          *AgentHandler
	Knowledge      *KnowledgeHandler
	Chunk          *ChunkHandler
	Tool           *ToolHandler
	FAQ            *FAQHandler
	RAG            *RAGHandler
	Initialization *InitializationHandler
	Model          *ModelHandler
	Evaluation     *EvaluationHandler
	MCPService     *MCPServiceHandler
	Tenant         *TenantHandler
	Tag            *TagHandler
	File           *FileHandler
	Dataset        *DatasetHandler
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
		Knowledge:      NewKnowledgeHandler(svc),
		Chunk:          NewChunkHandler(svc.Chunk),
		Tool:           NewToolHandler(svc),
		FAQ:            NewFAQHandler(svc),
		RAG:            NewRAGHandler(svc),
		Initialization: NewInitializationHandler(svc),
		Model:          NewModelHandler(svc.Model),
		Evaluation:     NewEvaluationHandler(svc),
		MCPService:     NewMCPServiceHandler(svc),
		Tenant:         NewTenantHandler(svc),
		Tag:            NewTagHandler(svc.Tag),
		File:           NewFileHandler(svc.File),
		Dataset:        NewDatasetHandler(svc.Dataset),
		System:         NewSystemHandler(svc),
		Message:        NewMessageHandler(svc.Chat),
		WebSearch:      NewWebSearchHandler(),
	}
}
