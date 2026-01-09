package service

import (
	"context"
	"log"

	"github.com/ashwinyue/next-ai/internal/config"
	"github.com/ashwinyue/next-ai/internal/repository"
	"github.com/ashwinyue/next-ai/internal/service/agent"
	"github.com/ashwinyue/next-ai/internal/service/auth"
	"github.com/ashwinyue/next-ai/internal/service/callback"
	"github.com/ashwinyue/next-ai/internal/service/chat"
	"github.com/ashwinyue/next-ai/internal/service/file"
	"github.com/ashwinyue/next-ai/internal/service/initialization"
	svcmcp "github.com/ashwinyue/next-ai/internal/service/mcp"
	svcModel "github.com/ashwinyue/next-ai/internal/service/model"
	"github.com/ashwinyue/next-ai/internal/service/session"
	svctenant "github.com/ashwinyue/next-ai/internal/service/tenant"
	"github.com/ashwinyue/next-ai/internal/service/tool"
	ecomodel "github.com/cloudwego/eino/components/model"
	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/redis/go-redis/v9"
)

// Services 服务集合
type Services struct {
	// 业务服务
	Auth           *auth.Service
	Chat           *chat.ServiceWithAgent
	Agent          *agent.Service
	Tool           *tool.Service
	Initialization *initialization.Service // 初始化服务
	Model          *svcModel.Service       // 模型管理服务
	MCP            *svcmcp.Service         // MCP 服务管理
	Tenant         *svctenant.Service      // 租户管理
	File           *file.Service           // 文件存储服务

	// 配置
	Config     *config.Config
	SessionMgr *session.Manager

	// Eino 组件（直接使用 eino 类型，无封装）
	AllTools  []einotool.BaseTool
	ChatModel ecomodel.ChatModel // 用于查询处理的 ChatModel
}

// NewServices 创建所有服务
// 参考 eino-examples，使用简单的 newXxx() 函数直接初始化 eino 组件
func NewServices(repo *repository.Repositories, cfg *config.Config, redisClient *redis.Client) (*Services, error) {
	ctx := context.Background()

	// 设置 Eino 全局回调（用于日志追踪）
	callback.SetupGlobalCallbacks(cfg.App.Debug)

	// 创建 Session 管理器
	sessionMgr := session.NewManager(redisClient)

	// 创建 ChatModel
	chatModel, err := newChatModel(ctx, cfg)
	if err != nil {
		log.Printf("Warning: failed to create chat model: %v", err)
	}

	// 初始化工具（不依赖知识库）
	allTools := newTools(ctx, cfg, repo)
	log.Printf("Initialized %d tools", len(allTools))

	// 创建文件存储服务
	fileSvc := newFileService(repo, cfg)

	// 创建 Agent 服务（不再需要 EventBus）
	agentSvc := agent.NewService(repo, cfg, allTools)

	// 创建 Chat 服务
	chatSvc := chat.NewService(repo, chatModel)

	// 创建 Agent 服务适配器
	agentSvcAdapter := newAgentServiceAdapter(agentSvc)

	// 创建带 Agent 集成的 Chat 服务
	chatSvcWithAgent := chat.NewServiceWithAgent(chatSvc, agentSvcAdapter)

	return &Services{
		Auth:           auth.NewService(repo),
		Chat:           chatSvcWithAgent,
		Agent:          agentSvc,
		Tool:           tool.NewService(repo),
		Initialization: initialization.NewService(repo, chatModel),
		Model:          svcModel.NewService(repo.Model),
		MCP:            svcmcp.NewService(repo),
		Tenant:         svctenant.NewService(repo),
		File:           fileSvc,

		Config:     cfg,
		SessionMgr: sessionMgr,

		AllTools:  allTools,
		ChatModel: chatModel,
	}, nil
}
