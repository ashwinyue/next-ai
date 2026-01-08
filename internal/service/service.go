package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/ashwinyue/next-ai/internal/config"
	"github.com/ashwinyue/next-ai/internal/repository"
	"github.com/ashwinyue/next-ai/internal/service/agent"
	"github.com/ashwinyue/next-ai/internal/service/auth"
	"github.com/ashwinyue/next-ai/internal/service/callback"
	"github.com/ashwinyue/next-ai/internal/service/chat"
	"github.com/ashwinyue/next-ai/internal/service/chunk"
	"github.com/ashwinyue/next-ai/internal/service/dataset"
	"github.com/ashwinyue/next-ai/internal/service/evaluation"
	"github.com/ashwinyue/next-ai/internal/service/event"
	"github.com/ashwinyue/next-ai/internal/service/faq"
	"github.com/ashwinyue/next-ai/internal/service/file"
	"github.com/ashwinyue/next-ai/internal/service/initialization"
	"github.com/ashwinyue/next-ai/internal/service/knowledge"
	svcmcp "github.com/ashwinyue/next-ai/internal/service/mcp"
	svcModel "github.com/ashwinyue/next-ai/internal/service/model"
	"github.com/ashwinyue/next-ai/internal/service/rag"
	"github.com/ashwinyue/next-ai/internal/service/rewrite"
	"github.com/ashwinyue/next-ai/internal/service/session"
	svctag "github.com/ashwinyue/next-ai/internal/service/tag"
	svctenant "github.com/ashwinyue/next-ai/internal/service/tenant"
	"github.com/ashwinyue/next-ai/internal/service/tool"
	"github.com/ashwinyue/next-ai/internal/service/types"
	"github.com/cloudwego/eino-ext/components/retriever/es8"
	einoembed "github.com/cloudwego/eino/components/embedding"
	ecomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/retriever"
	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/redis/go-redis/v9"
)

// Services 服务集合
type Services struct {
	// 业务服务
	Auth           *auth.Service
	Chat           *chat.ServiceWithAgent
	Agent          *agent.Service
	Knowledge      *knowledge.Service
	Chunk          *chunk.Service
	Tool           *tool.Service
	FAQ            *faq.Service
	FAQEntry       *faq.EntryService       // 增强版FAQ服务
	Initialization *initialization.Service // 初始化服务
	Model          *svcModel.Service       // 模型管理服务
	Evaluation     *evaluation.Service     // 评估服务
	MCP            *svcmcp.Service         // MCP 服务管理
	Tenant         *svctenant.Service      // 租户管理
	Tag            *svctag.Service         // 标签管理服务
	File           *file.Service           // 文件存储服务
	Dataset        *dataset.Service        // 数据集服务

	// 新增服务
	RewriteSvc *rewrite.Service
	EventBus   *event.EventBus

	// 配置
	Config     *config.Config
	SessionMgr *session.Manager

	// Eino 组件（直接使用 eino 类型，无封装）
	AllTools []einotool.BaseTool
	Embedder einoembed.Embedder

	// RAG 组件
	ChatModel ecomodel.ChatModel  // 用于查询处理和重排的 ChatModel
	Retriever retriever.Retriever // ES8 检索器
	Rerankers []types.Reranker    // 重排器列表（使用 types 包避免循环导入）
}

// NewServices 创建所有服务
// 参考 eino-examples，使用简单的 newXxx() 函数直接初始化 eino 组件
func NewServices(repo *repository.Repositories, cfg *config.Config, redisClient *redis.Client) (*Services, error) {
	ctx := context.Background()

	// 设置 Eino 全局回调（用于日志追踪）
	callback.SetupGlobalCallbacks(cfg.App.Debug)

	// 创建 Session 管理器
	sessionMgr := session.NewManager(redisClient)

	// 创建 ChatModel (用于查询处理和重排)
	chatModel, err := newChatModel(ctx, cfg)
	if err != nil {
		log.Printf("Warning: failed to create chat model: %v", err)
	}

	// 创建 Embedding 器
	embedder := newEmbedder(ctx, cfg)

	// 创建 ES8 Retriever 及相关组件
	var retriever *es8.Retriever
	if embedder != nil {
		retriever, _, _ = newES8Retriever(ctx, cfg, embedder)
	}

	// 创建重排器（使用 rag 包的函数）
	rerankers := rag.NewRerankers(ctx, cfg, chatModel)

	// 初始化工具
	allTools := newTools(ctx, cfg, retriever, repo)
	log.Printf("Initialized %d tools", len(allTools))

	// 创建查询重写服务
	rewriteSvc := rewrite.NewService(chatModel, rewrite.DefaultConfig())

	// 创建事件总线
	eventBus := event.NewEventBus(newEventStore(redisClient))

	// 创建文件存储服务
	fileSvc := newFileService(repo, cfg)

	// 创建 Provider 适配器（使用 providers.go 中的工厂函数）
	eventBusProvider := newEventBusProvider(eventBus)
	retrieverProvider := newRetrieverProvider(retriever)
	chatModelProvider := newChatModelProvider(chatModel)

	// 创建 Agent 服务（带依赖注入）
	agentSvc := agent.NewService(repo, cfg, allTools, retrieverProvider, chatModelProvider, eventBusProvider)

	// 创建 Chat 服务
	chatSvc := chat.NewService(repo, chatModel)

	// 创建 Agent 服务适配器（使用 providers.go 中的工厂函数）
	agentSvcAdapter := newAgentServiceAdapter(agentSvc)

	// 创建带 Agent 集成的 Chat 服务（传入 retrieverProvider 用于知识库搜索）
	chatSvcWithAgent := chat.NewServiceWithAgent(chatSvc, agentSvcAdapter, retrieverProvider)

	return &Services{
		Auth:           auth.NewService(repo),
		Chat:           chatSvcWithAgent, // 使用带 Agent 集成的服务
		Agent:          agentSvc,
		Knowledge:      knowledge.NewService(repo, cfg, embedder),
		Chunk:          chunk.NewService(repo),
		Tool:           tool.NewService(repo),
		FAQ:            faq.NewService(repo),
		FAQEntry:       faq.NewEntryService(repo),
		Initialization: initialization.NewService(repo, chatModel),
		Model:          svcModel.NewService(repo.Model),
		Evaluation:     evaluation.NewService(repo),
		MCP:            svcmcp.NewService(repo),
		Tenant:         svctenant.NewService(repo),
		Tag:            svctag.NewService(repo),
		File:           fileSvc,
		Dataset:        dataset.NewService(repo),

		// 新增服务
		RewriteSvc: rewriteSvc,
		EventBus:   eventBus,

		Config:     cfg,
		SessionMgr: sessionMgr,

		AllTools:  allTools,
		Embedder:  embedder,
		ChatModel: chatModel,
		Retriever: retriever,
		Rerankers: rerankers,
	}, nil
}

// ========== 存储实现 ==========

// eventStoreImpl 事件存储实现
type eventStoreImpl struct {
	redisClient *redis.Client
}

// newEventStore 创建事件存储
func newEventStore(redisClient *redis.Client) event.Store {
	return &eventStoreImpl{redisClient: redisClient}
}

func (s *eventStoreImpl) SaveEvent(ctx context.Context, evt *event.Event) error {
	if s.redisClient == nil {
		return nil
	}

	key := fmt.Sprintf("events:%s", evt.SessionID)
	data, err := json.Marshal(evt)
	if err != nil {
		return err
	}

	// 使用 LPUSH 存储事件
	return s.redisClient.LPush(ctx, key, data).Err()
}

func (s *eventStoreImpl) GetEvents(ctx context.Context, sessionID string) ([]*event.Event, error) {
	if s.redisClient == nil {
		return []*event.Event{}, nil
	}

	key := fmt.Sprintf("events:%s", sessionID)
	values, err := s.redisClient.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	events := make([]*event.Event, 0, len(values))
	for _, v := range values {
		var evt event.Event
		if err := json.Unmarshal([]byte(v), &evt); err != nil {
			continue
		}
		events = append(events, &evt)
	}

	return events, nil
}

func (s *eventStoreImpl) GetEventsByType(ctx context.Context, sessionID string, eventType event.EventType) ([]*event.Event, error) {
	events, err := s.GetEvents(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	filtered := make([]*event.Event, 0)
	for _, evt := range events {
		if evt.EventType == eventType {
			filtered = append(filtered, evt)
		}
	}

	return filtered, nil
}

func (s *eventStoreImpl) ClearEvents(ctx context.Context, sessionID string) error {
	if s.redisClient == nil {
		return nil
	}

	key := fmt.Sprintf("events:%s", sessionID)
	return s.redisClient.Del(ctx, key).Err()
}

// containsIgnoreCase 大小写不敏感搜索
func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(substr) == 0 ||
		containsIgnoreCaseWorker(s, substr))
}

func containsIgnoreCaseWorker(s, substr string) bool {
	// 简化版大小写不敏感搜索
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			c1 := s[i+j]
			c2 := substr[j]
			if c1 >= 'A' && c1 <= 'Z' {
				c1 += 32
			}
			if c2 >= 'A' && c2 <= 'Z' {
				c2 += 32
			}
			if c1 != c2 {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
