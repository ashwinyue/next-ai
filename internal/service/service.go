package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/ashwinyue/next-ai/internal/service/file"
	"github.com/ashwinyue/next-ai/internal/config"
	"github.com/ashwinyue/next-ai/internal/model"
	"github.com/ashwinyue/next-ai/internal/repository"
	"github.com/ashwinyue/next-ai/internal/service/agent"
	"github.com/ashwinyue/next-ai/internal/service/auth"
	"github.com/ashwinyue/next-ai/internal/service/chat"
	"github.com/ashwinyue/next-ai/internal/service/callback"
	"github.com/ashwinyue/next-ai/internal/service/chunk"
	"github.com/ashwinyue/next-ai/internal/service/dataset"
	"github.com/ashwinyue/next-ai/internal/service/event"
	"github.com/ashwinyue/next-ai/internal/service/evaluation"
	"github.com/ashwinyue/next-ai/internal/service/faq"
	"github.com/ashwinyue/next-ai/internal/service/initialization"
	svcModel "github.com/ashwinyue/next-ai/internal/service/model"
	"github.com/ashwinyue/next-ai/internal/service/knowledge"
	svcmcp "github.com/ashwinyue/next-ai/internal/service/mcp"
	"github.com/ashwinyue/next-ai/internal/service/query"
	"github.com/ashwinyue/next-ai/internal/service/rewrite"
	"github.com/ashwinyue/next-ai/internal/service/session"
	svctag "github.com/ashwinyue/next-ai/internal/service/tag"
	svctenant "github.com/ashwinyue/next-ai/internal/service/tenant"
	"github.com/ashwinyue/next-ai/internal/service/tool"
	"github.com/ashwinyue/next-ai/internal/service/types"
	"github.com/cloudwego/eino-ext/components/embedding/dashscope"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino-ext/components/retriever/es8"
	"github.com/cloudwego/eino-ext/components/retriever/es8/search_mode"
	duckduckgov2 "github.com/cloudwego/eino-ext/components/tool/duckduckgo/v2"
	httptool "github.com/cloudwego/eino-ext/components/tool/httprequest"
	sequencethinking "github.com/cloudwego/eino-ext/components/tool/sequentialthinking"
	wikipediatool "github.com/cloudwego/eino-ext/components/tool/wikipedia"
	"github.com/cloudwego/eino/components/embedding"
	ecomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/retriever"
	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/redis/go-redis/v9"
)

// Services 服务集合
type Services struct {
	// 业务服务
	Auth      *auth.Service
	Chat      *chat.Service
	Agent     *agent.Service
	Knowledge *knowledge.Service
	Chunk     *chunk.Service
	Tool      *tool.Service
	FAQ       *faq.Service
	FAQEntry  *faq.EntryService // 增强版FAQ服务
	Initialization *initialization.Service // 初始化服务
	Model     *svcModel.Service // 模型管理服务
	Evaluation *evaluation.Service // 评估服务
	MCP       *svcmcp.Service // MCP 服务管理
	Tenant    *svctenant.Service // 租户管理
	Tag       *svctag.Service // 标签管理服务
	File      *file.Service // 文件存储服务
	Dataset   *dataset.Service // 数据集服务

	// 新增服务
	RewriteSvc *rewrite.Service
	EventBus   *event.EventBus

	// 配置
	Config     *config.Config
	SessionMgr *session.Manager

	// Eino 组件（直接使用 eino 类型，无封装）
	AllTools []einotool.BaseTool
	Embedder embedding.Embedder

	// RAG 组件
	ChatModel  ecomodel.ChatModel             // 用于查询处理和重排的 ChatModel
	Retriever  retriever.Retriever         // ES8 检索器
	Query      *query.Optimizer            // 查询优化器
	Rerankers  []types.Reranker            // 重排器列表（使用 types 包避免循环导入）
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

	// 创建 ES8 Retriever
	var retriever *es8.Retriever
	if embedder != nil {
		retriever = newES8Retriever(ctx, cfg, embedder)
	}

	// 创建查询优化器
	var queryOptimizer *query.Optimizer
	if chatModel != nil {
		queryOptimizer = query.NewOptimizer(chatModel, 3)
	}

	// 创建重排器
	rerankers := newRerankers(ctx, cfg, chatModel)

	// 初始化工具
	allTools := newTools(ctx, cfg, retriever, repo)
	log.Printf("Initialized %d tools", len(allTools))

	// 创建查询重写服务
	rewriteSvc := rewrite.NewService(chatModel, rewrite.DefaultConfig())

	// 创建事件总线
	eventBus := event.NewEventBus(newEventStore(redisClient))

	// 创建文件存储服务
	fileSvc := newFileService(repo, cfg)

	return &Services{
		Auth:      auth.NewService(repo),
		Chat:      chat.NewService(repo, chatModel),
		Agent:     agent.NewService(repo, cfg, allTools),
		Knowledge: knowledge.NewService(repo, cfg, embedder),
		Chunk:     chunk.NewService(repo),
		Tool:      tool.NewService(repo),
		FAQ:       faq.NewService(repo),
		FAQEntry:  faq.NewEntryService(repo),
		Initialization: initialization.NewService(repo),
		Model:     svcModel.NewService(repo.Model),
		Evaluation: evaluation.NewService(repo),
		MCP:       svcmcp.NewService(repo),
		Tenant:    svctenant.NewService(repo),
		Tag:       svctag.NewService(repo),
		File:      fileSvc,
		Dataset:   dataset.NewService(repo),

		// 新增服务
		RewriteSvc: rewriteSvc,
		EventBus:   eventBus,

		Config:     cfg,
		SessionMgr: sessionMgr,

		AllTools:  allTools,
		Embedder:  embedder,
		ChatModel: chatModel,
		Retriever: retriever,
		Query:     queryOptimizer,
		Rerankers: rerankers,
	}, nil
}

// newChatModel 创建 ChatModel
func newChatModel(ctx context.Context, cfg *config.Config) (ecomodel.ChatModel, error) {
	aiCfg := cfg.AI

	var apiKey, baseURL, modelName string

	switch aiCfg.Provider {
	case "openai":
		apiKey = aiCfg.OpenAI.APIKey
		baseURL = aiCfg.OpenAI.BaseURL
		modelName = aiCfg.OpenAI.Model
	case "alibaba", "qwen", "dashscope":
		apiKey = aiCfg.Alibaba.AccessKeySecret
		baseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
		modelName = aiCfg.Alibaba.Model
	case "deepseek":
		apiKey = aiCfg.DeepSeek.APIKey
		baseURL = aiCfg.DeepSeek.BaseURL
		modelName = aiCfg.DeepSeek.Model
	default:
		return nil, fmt.Errorf("unsupported ai provider: %s", aiCfg.Provider)
	}

	if apiKey == "" {
		return nil, fmt.Errorf("api_key is required for provider: %s", aiCfg.Provider)
	}

	if modelName == "" {
		modelName = "gpt-4o-mini"
	}

	return openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  apiKey,
		BaseURL: baseURL,
		Model:   modelName,
	})
}

// newToolCallingChatModel 创建支持工具调用的 ChatModel
func newToolCallingChatModel(ctx context.Context, cfg *config.Config) (ecomodel.ToolCallingChatModel, error) {
	aiCfg := cfg.AI

	var apiKey, baseURL, modelName string

	switch aiCfg.Provider {
	case "openai":
		apiKey = aiCfg.OpenAI.APIKey
		baseURL = aiCfg.OpenAI.BaseURL
		modelName = aiCfg.OpenAI.Model
	case "alibaba", "qwen", "dashscope":
		apiKey = aiCfg.Alibaba.AccessKeySecret
		baseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
		modelName = aiCfg.Alibaba.Model
	case "deepseek":
		apiKey = aiCfg.DeepSeek.APIKey
		baseURL = aiCfg.DeepSeek.BaseURL
		modelName = aiCfg.DeepSeek.Model
	default:
		return nil, fmt.Errorf("unsupported ai provider: %s", aiCfg.Provider)
	}

	if apiKey == "" {
		return nil, fmt.Errorf("api_key is required for provider: %s", aiCfg.Provider)
	}

	if modelName == "" {
		modelName = "gpt-4o-mini"
	}

	temperature := float32(0.7)

	return openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:      apiKey,
		BaseURL:     baseURL,
		Model:       modelName,
		Temperature: &temperature,
	})
}

// newEmbedder 创建 Embedding 器
func newEmbedder(ctx context.Context, cfg *config.Config) embedding.Embedder {
	embCfg := cfg.AI.Embedding

	var apiKey, model string
	var timeout int

	switch embCfg.Provider {
	case "alibaba", "qwen", "dashscope", "":
		apiKey = embCfg.APIKey
		model = embCfg.Model
		timeout = embCfg.Timeout
	case "openai":
		apiKey = embCfg.APIKey
		model = embCfg.Model
		timeout = embCfg.Timeout
	default:
		log.Printf("Warning: unsupported embedding provider: %s", embCfg.Provider)
		return nil
	}

	if apiKey == "" {
		log.Printf("Warning: embedding api_key is empty")
		return nil
	}

	if model == "" {
		model = "text-embedding-v3"
	}

	embConfig := &dashscope.EmbeddingConfig{
		APIKey: apiKey,
		Model:  model,
	}

	if timeout > 0 {
		embConfig.Timeout = time.Duration(timeout) * time.Second
	}

	if embCfg.Dimensions > 0 {
		embConfig.Dimensions = &embCfg.Dimensions
	}

	embedder, err := dashscope.NewEmbedder(ctx, embConfig)
	if err != nil {
		log.Printf("Warning: failed to create embedder: %v", err)
		return nil
	}

	return embedder
}

// newES8Retriever 创建 ES8 检索器
func newES8Retriever(ctx context.Context, cfg *config.Config, embedder embedding.Embedder) *es8.Retriever {
	esCfg := cfg.Elastic

	if esCfg.Host == "" {
		log.Printf("Warning: elasticsearch host not configured")
		return nil
	}

	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{esCfg.Host},
		Username:  esCfg.Username,
		Password:  esCfg.Password,
	})
	if err != nil {
		log.Printf("Warning: failed to create es client: %v", err)
		return nil
	}

	indexName := esCfg.IndexPrefix + "_chunks"

	retriever, err := es8.NewRetriever(ctx, &es8.RetrieverConfig{
		Client:     esClient,
		Index:      indexName,
		TopK:       10,
		SearchMode: search_mode.SearchModeDenseVectorSimilarity(search_mode.DenseVectorSimilarityTypeCosineSimilarity, "content_vector"),
		Embedding:  embedder,
	})
	if err != nil {
		log.Printf("Warning: failed to create retriever: %v", err)
		return nil
	}

	return retriever
}

// newWebSearchTool 创建网络搜索工具
func newWebSearchTool(ctx context.Context) einotool.InvokableTool {
	searchTool, err := duckduckgov2.NewTextSearchTool(ctx, &duckduckgov2.Config{
		ToolName: "web_search",
		ToolDesc: "Search the web for current information using DuckDuckGo. Use this when you need up-to-date information or the knowledge base doesn't have the answer.",
		MaxResults: 10,
	})
	if err != nil {
		log.Printf("Warning: failed to create web search tool: %v", err)
		return &stubTool{name: "web_search"}
	}

	return searchTool
}

// newTools 初始化所有工具
func newTools(ctx context.Context, cfg *config.Config, retriever *es8.Retriever, repo *repository.Repositories) []einotool.BaseTool {
	tools := []einotool.BaseTool{}

	// 添加网络搜索工具 (eino-ext duckduckgo)
	tools = append(tools, newWebSearchTool(ctx))

	// 添加 HTTP 请求工具 (eino-ext httprequest)
	httpTools, err := httptool.NewToolKit(ctx, &httptool.Config{})
	if err != nil {
		log.Printf("Warning: failed to create http tools: %v", err)
	} else {
		tools = append(tools, httpTools...)
	}

	// 添加 Wikipedia 搜索工具 (eino-ext wikipedia)
	wikiTool, err := wikipediatool.NewTool(ctx, &wikipediatool.Config{
		Language: "zh", // 中文 Wikipedia
		TopK:     3,
	})
	if err != nil {
		log.Printf("Warning: failed to create wikipedia tool: %v", err)
	} else {
		tools = append(tools, wikiTool)
	}

	// 添加顺序思考工具 (eino-ext sequentialthinking)
	thinkTool, err := sequencethinking.NewTool()
	if err != nil {
		log.Printf("Warning: failed to create sequentialthinking tool: %v", err)
	} else {
		tools = append(tools, thinkTool)
	}

	// 添加 todo_write 工具
	tools = append(tools, newTodoWriteTool())

	// 添加知识库搜索工具
	if retriever != nil {
		tools = append(tools, newKnowledgeSearchTool(retriever))
	}

	// 添加文档相关工具
	if repo != nil {
		tools = append(tools, newDocumentInfoTool(repo))
		tools = append(tools, newListChunksTool(repo))
		tools = append(tools, newGrepChunksTool(repo))
	}

	// 添加数据库工具
	// 注意: 数据库工具需要 sessionID 和 tenantID，在 Agent 运行时动态创建
	// 这里使用 stub 占位，实际使用时在 Agent 配置中添加

	return tools
}

// stubTool 占位工具
type stubTool struct {
	name string
}

func (t *stubTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: t.name,
		Desc: t.name + " (unavailable)",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"query": {
				Type:     schema.String,
				Desc:     "The query string",
				Required: true,
			},
		}),
	}, nil
}

func (t *stubTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...einotool.Option) (string, error) {
	return fmt.Sprintf(`{"error":"%s is not available"}`, t.name), nil
}

// ========== todo_write 工具 ==========

// PlanStep 计划步骤
type PlanStep struct {
	ID          string `json:"id" jsonschema_description:"步骤ID"`
	Description string `json:"description" jsonschema_description:"步骤描述"`
	Status      string `json:"status" jsonschema_description:"状态: pending, in_progress, completed"`
}

// TodoWriteInput todo_write 输入参数
type TodoWriteInput struct {
	Task  string     `json:"task" jsonschema_description:"任务描述"`
	Steps []PlanStep `json:"steps" jsonschema_description:"任务步骤列表"`
}

// newTodoWriteTool 创建任务计划工具
// 使用 utils.InferTool 自动生成 ToolInfo
func newTodoWriteTool() einotool.InvokableTool {
	t, err := utils.InferTool(
		"todo_write",
		`创建和管理结构化的检索任务列表。用于跟踪复杂任务的进度。

**使用场景**：
- 复杂多步骤任务（3个或以上步骤）
- 需要仔细规划的操作
- 用户明确请求创建任务列表

**任务状态**：
- pending: 未开始
- in_progress: 进行中（同时只能有一个）
- completed: 已完成

**重要**：
- 仅包含检索/研究任务，不包含总结任务
- 完成所有检索任务后，使用 thinking 工具进行总结`,
		func(ctx context.Context, input *TodoWriteInput) (string, error) {
			if input.Task == "" {
				input.Task = "未提供任务描述"
			}
			return generateTodoOutput(input.Task, input.Steps), nil
		},
	)
	if err != nil {
		log.Printf("Warning: failed to create todo_write tool: %v", err)
		return nil
	}
	return t
}

// ========== 知识库搜索工具 ==========

// KnowledgeSearchInput 知识库搜索输入
type KnowledgeSearchInput struct {
	Query string `json:"query" jsonschema_description:"The search query" jsonschema_required:"true"`
	TopK  int    `json:"top_k" jsonschema_description:"Number of results (default 10)"`
}

// KnowledgeSearchOutput 知识库搜索输出
type KnowledgeSearchOutput struct {
	Query   string                 `json:"query"`
	Total   int                    `json:"total"`
	Results []map[string]interface{} `json:"results"`
}

// newKnowledgeSearchTool 创建知识库搜索工具
func newKnowledgeSearchTool(r *es8.Retriever) einotool.InvokableTool {
	t, err := utils.InferTool(
		"knowledge_search",
		"Searches the knowledge base for relevant information using semantic and keyword search.",
		func(ctx context.Context, input *KnowledgeSearchInput) (*KnowledgeSearchOutput, error) {
			if input.Query == "" {
				return nil, fmt.Errorf("query is required")
			}
			if input.TopK <= 0 {
				input.TopK = 10
			}

			docs, err := r.Retrieve(ctx, input.Query, retriever.WithTopK(input.TopK))
			if err != nil {
				return nil, fmt.Errorf("retriever failed: %w", err)
			}

			results := make([]map[string]interface{}, 0, len(docs))
			for _, doc := range docs {
				result := map[string]interface{}{
					"content": doc.Content,
					"score":   doc.Score(),
				}
				if doc.MetaData != nil {
					if title, ok := doc.MetaData["title"].(string); ok {
						result["title"] = title
					}
				}
				results = append(results, result)
			}

			return &KnowledgeSearchOutput{
				Query:   input.Query,
				Total:   len(results),
				Results: results,
			}, nil
		},
	)
	if err != nil {
		log.Printf("Warning: failed to create knowledge_search tool: %v", err)
		return nil
	}
	return t
}

// ========== 文档相关工具 ==========

// DocumentInfoInput 文档信息输入
type DocumentInfoInput struct {
	DocumentIDs []string `json:"document_ids" jsonschema_description:"文档 ID 列表，最多 10 个" jsonschema_required:"true"`
}

// DocumentInfoOutput 文档信息输出
type DocumentInfoOutput struct {
	Count     int                    `json:"count"`
	Documents []map[string]interface{} `json:"documents"`
}

// newDocumentInfoTool 创建文档信息工具
func newDocumentInfoTool(repo *repository.Repositories) einotool.InvokableTool {
	t, err := utils.InferTool(
		"get_document_info",
		"获取文档的详细元数据信息，包括标题、文件名、大小、分块数量等。用于查询文档基本信息和处理状态。",
		func(ctx context.Context, input *DocumentInfoInput) (*DocumentInfoOutput, error) {
			if len(input.DocumentIDs) == 0 {
				return nil, fmt.Errorf("document_ids is required")
			}
			if len(input.DocumentIDs) > 10 {
				return nil, fmt.Errorf("maximum 10 document IDs allowed")
			}

			results := make([]map[string]interface{}, 0)
			for _, docID := range input.DocumentIDs {
				doc, err := repo.Knowledge.GetDocumentByID(docID)
				if err != nil {
					continue
				}
				chunks, _ := repo.Knowledge.GetChunksByDocumentID(docID)

				results = append(results, map[string]interface{}{
					"id":           doc.ID,
					"title":        doc.Title,
					"file_name":    doc.FileName,
					"file_size":    doc.FileSize,
					"content_type": doc.ContentType,
					"status":       doc.Status,
					"chunk_count":  len(chunks),
					"created_at":   doc.CreatedAt,
				})
			}

			return &DocumentInfoOutput{
				Count:     len(results),
				Documents: results,
			}, nil
		},
	)
	if err != nil {
		log.Printf("Warning: failed to create get_document_info tool: %v", err)
		return nil
	}
	return t
}

// ========== 分块列表工具 ==========

// ListChunksInput 列出分块输入
type ListChunksInput struct {
	DocumentID string `json:"document_id" jsonschema_description:"文档 ID" jsonschema_required:"true"`
}

// ChunkItem 分块项
type ChunkItem struct {
	ID         string `json:"id"`
	ChunkIndex int    `json:"chunk_index"`
	Content    string `json:"content"`
}

// ListChunksOutput 列出分块输出
type ListChunksOutput struct {
	DocumentID string      `json:"document_id"`
	Title      string      `json:"title"`
	Total      int         `json:"total"`
	Chunks     []ChunkItem `json:"chunks"`
}

// newListChunksTool 创建列出分块工具
func newListChunksTool(repo *repository.Repositories) einotool.InvokableTool {
	t, err := utils.InferTool(
		"list_chunks",
		"获取指定文档的所有分块内容。用于查看文档的完整分块列表。",
		func(ctx context.Context, input *ListChunksInput) (*ListChunksOutput, error) {
			if input.DocumentID == "" {
				return nil, fmt.Errorf("document_id is required")
			}

			doc, err := repo.Knowledge.GetDocumentByID(input.DocumentID)
			if err != nil {
				return nil, fmt.Errorf("document not found: %w", err)
			}

			chunks, err := repo.Knowledge.GetChunksByDocumentID(input.DocumentID)
			if err != nil {
				return nil, fmt.Errorf("failed to get chunks: %w", err)
			}

			chunkList := make([]ChunkItem, 0, len(chunks))
			for _, c := range chunks {
				chunkList = append(chunkList, ChunkItem{
					ID:         c.ID,
					ChunkIndex: c.ChunkIndex,
					Content:    c.Content,
				})
			}

			return &ListChunksOutput{
				DocumentID: doc.ID,
				Title:      doc.Title,
				Total:      len(chunks),
				Chunks:     chunkList,
			}, nil
		},
	)
	if err != nil {
		log.Printf("Warning: failed to create list_chunks tool: %v", err)
		return nil
	}
	return t
}

// ========== 分块搜索工具 ==========

// GrepChunksInput 分块搜索输入
type GrepChunksInput struct {
	Pattern    string `json:"pattern" jsonschema_description:"搜索模式（文本）" jsonschema_required:"true"`
	DocumentID string `json:"document_id" jsonschema_description:"可选：限制在特定文档中搜索"`
}

// GrepChunkItem 搜索结果项
type GrepChunkItem struct {
	ID         string `json:"id"`
	DocumentID string `json:"document_id"`
	ChunkIndex int    `json:"chunk_index"`
	Content    string `json:"content"`
}

// GrepChunksOutput 分块搜索输出
type GrepChunksOutput struct {
	Pattern string         `json:"pattern"`
	Count   int            `json:"count"`
	Matches []GrepChunkItem `json:"matches"`
}

// newGrepChunksTool 创建分块搜索工具
func newGrepChunksTool(repo *repository.Repositories) einotool.InvokableTool {
	t, err := utils.InferTool(
		"grep_chunks",
		"在文档分块中搜索包含特定文本的内容。支持精确文本匹配。",
		func(ctx context.Context, input *GrepChunksInput) (*GrepChunksOutput, error) {
			if input.Pattern == "" {
				return nil, fmt.Errorf("pattern is required")
			}

			var chunks []*model.DocumentChunk
			var err error

			if input.DocumentID != "" {
				// 搜索特定文档的分块
				chunks, err = repo.Knowledge.GetChunksByDocumentID(input.DocumentID)
				if err != nil {
					return nil, fmt.Errorf("failed to get chunks: %w", err)
				}
			}

			// 过滤包含匹配内容的分块
			matches := make([]GrepChunkItem, 0)
			for _, c := range chunks {
				if containsIgnoreCase(c.Content, input.Pattern) {
					matches = append(matches, GrepChunkItem{
						ID:         c.ID,
						DocumentID: c.DocumentID,
						ChunkIndex: c.ChunkIndex,
						Content:    c.Content,
					})
				}
			}

			return &GrepChunksOutput{
				Pattern: input.Pattern,
				Count:   len(matches),
				Matches: matches,
			}, nil
		},
	)
	if err != nil {
		log.Printf("Warning: failed to create grep_chunks tool: %v", err)
		return nil
	}
	return t
}

// TodoWriteTool 任务计划工具（已弃用，保留用于兼容）
// 使用 utils.InferTool 重构后不再需要此类型
type TodoWriteTool struct{}

// generateTodoOutput 生成 todo 输出
func generateTodoOutput(task string, steps []PlanStep) string {
	output := "## 计划已创建\n\n"
	output += fmt.Sprintf("**任务**: %s\n\n", task)

	if len(steps) == 0 {
		output += "注意：未提供具体步骤。建议创建3-7个检索任务。\n\n"
		output += "建议的检索流程：\n"
		output += "1. 使用 grep_chunks 搜索关键词定位相关文档\n"
		output += "2. 使用 knowledge_search 进行语义搜索获取相关内容\n"
		output += "3. 使用 list_chunks 获取关键文档的完整内容\n"
		output += "4. 使用 web_search 获取补充信息（如需要）\n"
		return output
	}

	// 统计任务状态
	pendingCount := 0
	inProgressCount := 0
	completedCount := 0
	for _, step := range steps {
		switch step.Status {
		case "pending":
			pendingCount++
		case "in_progress":
			inProgressCount++
		case "completed":
			completedCount++
		}
	}
	totalCount := len(steps)
	remainingCount := pendingCount + inProgressCount

	output += "**任务步骤**:\n\n"
	for i, step := range steps {
		output += formatTodoStep(i+1, step)
	}

	// 添加进度汇总
	output += "\n## 任务进度\n"
	output += fmt.Sprintf("总计: %d 个任务 | ", totalCount)
	output += fmt.Sprintf("✅ 已完成: %d | ", completedCount)
	output += fmt.Sprintf("🔄 进行中: %d | ", inProgressCount)
	output += fmt.Sprintf("⏳ 待处理: %d\n\n", pendingCount)

	// 添加提醒
	output += "## ⚠️ 重要提醒\n"
	if remainingCount > 0 {
		output += fmt.Sprintf("**还有 %d 个任务未完成！**\n\n", remainingCount)
		output += "**必须完成所有任务后才能总结或得出结论。**\n\n"
		output += "下一步操作：\n"
		if inProgressCount > 0 {
			output += "- 继续完成当前进行中的任务\n"
		}
		if pendingCount > 0 {
			output += fmt.Sprintf("- 开始处理 %d 个待处理任务\n", pendingCount)
		}
		output += "- 完成每个任务后，更新 todo_write 标记为 completed\n"
		output += "- 所有任务完成后，使用 thinking 工具生成最终总结\n"
	} else {
		output += "✅ **所有任务已完成！**\n\n"
		output += "现在可以：\n"
		output += "- 综合所有任务的发现\n"
		output += "- 使用 thinking 工具生成完整的最终答案\n"
	}

	return output
}

// formatTodoStep 格式化单个任务步骤
func formatTodoStep(index int, step PlanStep) string {
	statusEmoji := map[string]string{
		"pending":     "⏳",
		"in_progress": "🔄",
		"completed":   "✅",
	}

	emoji, ok := statusEmoji[step.Status]
	if !ok {
		emoji = "⏳"
	}

	return fmt.Sprintf("%d. %s [%s] %s\n", index, emoji, step.Status, step.Description)
}

// GetToolsByName 根据名称获取工具
func GetToolsByName(ctx context.Context, names []string, allTools []einotool.BaseTool) ([]einotool.BaseTool, error) {
	if len(names) == 0 {
		return allTools, nil
	}

	toolMap := make(map[string]einotool.BaseTool)
	for _, t := range allTools {
		info, err := t.Info(ctx)
		if err != nil {
			continue
		}
		toolMap[info.Name] = t
	}

	result := make([]einotool.BaseTool, 0, len(names))
	for _, name := range names {
		t, ok := toolMap[name]
		if !ok {
			return nil, fmt.Errorf("tool not found: %s", name)
		}
		result = append(result, t)
	}

	return result, nil
}

// ListToolNames 列出所有工具名称
func ListToolNames(ctx context.Context, allTools []einotool.BaseTool) []string {
	names := make([]string, 0, len(allTools))
	for _, t := range allTools {
		info, err := t.Info(ctx)
		if err != nil {
			continue
		}
		names = append(names, info.Name)
	}
	return names
}

// newRerankers 创建默认的重排器列表
func newRerankers(ctx context.Context, cfg *config.Config, chatModel ecomodel.ChatModel) []types.Reranker {
	rerankers := []types.Reranker{}

	// 添加分数重排（轻量级，始终启用）
	// 这里直接实现简单重排，避免额外导入
	rerankers = append(rerankers, &scoreReranker{})

	// LLM 重排（如果有 ChatModel）
	if chatModel != nil {
		rerankers = append(rerankers, &llmRerankerWrapper{
			chatModel: chatModel,
			topN:      5,
		})
	}

	return rerankers
}

// scoreReranker 分数重排器（简单实现）
type scoreReranker struct{}

func (r *scoreReranker) Rerank(ctx context.Context, query string, docs []*schema.Document) ([]*schema.Document, error) {
	if len(docs) <= 1 {
		return docs, nil
	}

	// 复制并按分数排序
	sorted := make([]*schema.Document, len(docs))
	copy(sorted, docs)

	// 简单排序
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].Score() > sorted[i].Score() {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted, nil
}

// llmRerankerWrapper LLM 重排器包装
type llmRerankerWrapper struct {
	chatModel ecomodel.ChatModel
	topN      int
}

func (r *llmRerankerWrapper) Rerank(ctx context.Context, query string, docs []*schema.Document) ([]*schema.Document, error) {
	if len(docs) <= r.topN || r.chatModel == nil {
		return docs, nil
	}

	// 构建文档描述
	docDesc := ""
	for i, doc := range docs {
		content := doc.Content
		if len(content) > 200 {
			content = content[:200] + "..."
		}
		docDesc += fmt.Sprintf("%d. %s\n", i+1, content)
	}

	// 调用 LLM
	prompt := fmt.Sprintf(`你是一个检索结果重排专家。请根据查询的相关性，对检索到的文档进行排序。

查询：%s

检索到的文档：
%s

请按照与查询的相关度从高到低排序，输出排序后的文档编号（用逗号分隔，如：1,3,2,4,5）。

排序结果：`, query, docDesc)

	messages := []*schema.Message{
		{Role: schema.System, Content: "你是一个专业的检索结果重排助手。"},
		{Role: schema.User, Content: prompt},
	}

	resp, err := r.chatModel.Generate(ctx, messages)
	if err != nil {
		return docs, nil // 失败时返回原顺序
	}

	// 解析排序结果（简化版）
	indices := extractNumbersFromOutput(resp.Content)
	if len(indices) == 0 {
		return docs[:r.topN], nil
	}

	// 应用排序
	result := make([]*schema.Document, 0, minInt(r.topN, len(indices)))
	for i, idx := range indices {
		if idx >= 0 && idx < len(docs) && i < r.topN {
			result = append(result, docs[idx])
		}
	}

	if len(result) == 0 {
		return docs[:r.topN], nil
	}

	return result, nil
}

func extractNumbersFromOutput(s string) []int {
	nums := make([]int, 0)
	current := 0

	for _, ch := range s {
		if ch >= '0' && ch <= '9' {
			current = current*10 + int(ch-'0')
		} else {
			if current > 0 {
				nums = append(nums, current)
				current = 0
			}
		}
	}
	if current > 0 {
		nums = append(nums, current)
	}

	return nums
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
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

// newFileService 创建文件存储服务
func newFileService(repo *repository.Repositories, cfg *config.Config) *file.Service {
	// 默认使用本地存储
	storageType := file.StorageTypeLocal
	fileCfg := make(map[string]string)

	// 从配置中读取文件存储配置
	if cfg.File != nil {
		switch cfg.File.Type {
		case "minio":
			storageType = file.StorageTypeMinIO
			fileCfg = map[string]string{
				"endpoint":   cfg.File.MinIO.Endpoint,
				"access_key": cfg.File.MinIO.AccessKey,
				"secret_key": cfg.File.MinIO.SecretKey,
				"bucket":     cfg.File.MinIO.Bucket,
				"use_ssl":    cfg.File.MinIO.UseSSL,
				"url_prefix": cfg.File.MinIO.URLPrefix,
			}
		case "local":
			storageType = file.StorageTypeLocal
			fileCfg = map[string]string{
				"base_path":  cfg.File.Local.BasePath,
				"url_prefix": cfg.File.Local.URLPrefix,
			}
		}
	}

	// 使用默认本地配置
	if len(fileCfg) == 0 {
		fileCfg = map[string]string{
			"base_path":  "./data/files",
			"url_prefix": "/files",
		}
	}

	fileSvc, err := file.NewServiceFromConfig(repo, storageType, fileCfg)
	if err != nil {
		log.Printf("Warning: failed to create file service: %v, using nil", err)
		return nil
	}

	log.Printf("File service initialized with type: %s", storageType)
	return fileSvc
}
