package service

import (
	"context"

	"github.com/ashwinyue/next-ai/internal/service/agent"
	"github.com/ashwinyue/next-ai/internal/service/chat"
)

// ========== Provider 适配器（用于 Agent 服务依赖注入）==========

// agentServiceAdapter Agent 服务适配器（将 agent.Service 适配为 chat.AgentService）
type agentServiceAdapter struct {
	agentSvc *agent.Service
}

func (a *agentServiceAdapter) StreamWithContext(ctx context.Context, agentID string, req interface{}) (<-chan interface{}, error) {
	// 直接调用 agent 服务的适配方法
	return a.agentSvc.StreamWithContextForChat(ctx, agentID, req)
}

// ========== 适配器创建 ==========

// newAgentServiceAdapter 创建 Agent 服务适配器
func newAgentServiceAdapter(agentSvc *agent.Service) chat.AgentService {
	return &agentServiceAdapter{agentSvc: agentSvc}
}
