package handler

import (
	"github.com/ashwinyue/next-rag/next-ai/internal/service"
)

// Handlers 处理器集合
type Handlers struct {
	Chat      *ChatHandler
	Agent     *AgentHandler
	Knowledge *KnowledgeHandler
	Tool      *ToolHandler
	FAQ       *FAQHandler
	RAG       *RAGHandler
}

// NewHandlers 创建所有处理器
func NewHandlers(svc *service.Services) *Handlers {
	return &Handlers{
		Chat:      NewChatHandler(svc),
		Agent:     NewAgentHandler(svc),
		Knowledge: NewKnowledgeHandler(svc),
		Tool:      NewToolHandler(svc),
		FAQ:       NewFAQHandler(svc),
		RAG:       NewRAGHandler(svc),
	}
}
