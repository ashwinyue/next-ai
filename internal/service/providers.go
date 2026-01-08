package service

import (
	"context"
	"time"

	"github.com/ashwinyue/next-ai/internal/service/agent"
	"github.com/ashwinyue/next-ai/internal/service/chat"
	"github.com/ashwinyue/next-ai/internal/service/event"
	"github.com/cloudwego/eino-ext/components/retriever/es8"
	ecomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	estypes "github.com/elastic/go-elasticsearch/v8/typedapi/types"
)

// ========== Provider 适配器（用于 Agent 服务依赖注入）==========

// eventBusProvider 事件总线提供者适配器
type eventBusProvider struct {
	eventBus *event.EventBus
}

func (p *eventBusProvider) Publish(ctx context.Context, evt *agent.AgentEvent) error {
	if p.eventBus == nil {
		return nil
	}
	// 将 AgentEvent 转换为 event.Event
	e := &event.Event{
		ID:        evt.ID,
		SessionID: evt.SessionID,
		AgentID:   evt.AgentID,
		EventType: event.EventType(evt.Type),
		Component: evt.ToolName,
		Data:      evt.Data,
		Metadata:  evt.Metadata,
	}
	return p.eventBus.Publish(ctx, e)
}

// agentServiceAdapter Agent 服务适配器（将 agent.Service 适配为 chat.AgentService）
type agentServiceAdapter struct {
	agentSvc *agent.Service
}

func (a *agentServiceAdapter) StreamWithContext(ctx context.Context, agentID string, req interface{}) (<-chan interface{}, error) {
	// 直接调用 agent 服务的适配方法
	return a.agentSvc.StreamWithContextForChat(ctx, agentID, req)
}

// ========== Provider 实现 ==========

// eventBusProviderImpl EventBus 提供者实现
type eventBusProviderImpl struct {
	bus *event.EventBus
}

// newEventBusProvider 创建 EventBus 提供者
func newEventBusProvider(bus *event.EventBus) agent.EventBusProvider {
	return &eventBusProviderImpl{bus: bus}
}

// Publish 发布事件
func (p *eventBusProviderImpl) Publish(ctx context.Context, evt *agent.AgentEvent) error {
	if p.bus == nil {
		return nil
	}
	// 转换为 event.Event
	einoEvt := &event.Event{
		ID:        evt.ID,
		SessionID: evt.SessionID,
		EventType: event.EventType(evt.Type),
		Data:      evt.Data,
		Timestamp: time.Now(),
	}
	// 添加额外字段
	if evt.ToolName != "" {
		if einoEvt.Metadata == nil {
			einoEvt.Metadata = make(map[string]interface{})
		}
		einoEvt.Metadata["tool_name"] = evt.ToolName
	}
	return p.bus.Publish(ctx, einoEvt)
}

// retrieverProviderImpl Retriever 提供者实现
type retrieverProviderImpl struct {
	retriever *es8.Retriever
}

// newRetrieverProvider 创建 Retriever 提供者
func newRetrieverProvider(r *es8.Retriever) agent.RetrieverProvider {
	return &retrieverProviderImpl{
		retriever: r,
	}
}

// filteredRetrieverWrapper 带过滤条件的 Retriever 包装器
// 使用 Eino WithFilters 选项实现知识库过滤
type filteredRetrieverWrapper struct {
	retriever retriever.Retriever
	filters   []estypes.Query // ES 查询过滤条件
}

// Retrieve 实现 retriever.Retriever 接口
func (w *filteredRetrieverWrapper) Retrieve(ctx context.Context, query string, opts ...interface{}) ([]*schema.Document, error) {
	// 将自定义 opts 转换为 retriever.Option
	einoOpts := make([]retriever.Option, 0, len(opts)+1)
	for _, opt := range opts {
		if ro, ok := opt.(retriever.Option); ok {
			einoOpts = append(einoOpts, ro)
		}
	}

	// 添加过滤选项
	if len(w.filters) > 0 {
		einoOpts = append(einoOpts, es8.WithFilters(w.filters))
	}

	return w.retriever.Retrieve(ctx, query, einoOpts...)
}

// GetRetriever 获取 Retriever
// 根据 knowledgeBaseIDs 创建带过滤的 Retriever（使用 Eino WithFilters）
func (p *retrieverProviderImpl) GetRetriever(ctx context.Context, knowledgeBaseIDs []string, tenantID string) (interface{}, error) {
	// 如果没有指定知识库 ID，返回全局检索器
	if len(knowledgeBaseIDs) == 0 {
		return p.retriever, nil
	}

	// 创建带过滤的 Retriever 包装器（使用 Eino 方式）
	return &filteredRetrieverWrapper{
		retriever: p.retriever,
		filters:   buildKnowledgeBaseFilters(knowledgeBaseIDs),
	}, nil
}

// buildKnowledgeBaseFilters 构建知识库 ID 过滤条件
func buildKnowledgeBaseFilters(kbIDs []string) []estypes.Query {
	if len(kbIDs) == 0 {
		return nil
	}
	// 将 []string 转换为 []estypes.FieldValue
	ids := make([]estypes.FieldValue, len(kbIDs))
	for i, id := range kbIDs {
		ids[i] = id
	}
	return []estypes.Query{
		{
			Terms: &estypes.TermsQuery{
				TermsQuery: map[string]estypes.TermsQueryField{
					"knowledge_base_id": ids,
				},
			},
		},
	}
}

// chatModelProviderImpl ChatModel 提供者实现
type chatModelProviderImpl struct {
	chatModel ecomodel.ChatModel
}

// newChatModelProvider 创建 ChatModel 提供者
func newChatModelProvider(cm ecomodel.ChatModel) agent.ChatModelProvider {
	return &chatModelProviderImpl{chatModel: cm}
}

// GetChatModel 获取 ChatModel
func (p *chatModelProviderImpl) GetChatModel(ctx context.Context, config interface{}) (interface{}, error) {
	return p.chatModel, nil
}

// ========== 适配器创建 ==========

// newAgentServiceAdapter 创建 Agent 服务适配器
func newAgentServiceAdapter(agentSvc *agent.Service) chat.AgentService {
	return &agentServiceAdapter{agentSvc: agentSvc}
}
