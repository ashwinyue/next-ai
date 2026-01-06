package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/ashwinyue/next-rag/next-ai/internal/config"
	"github.com/ashwinyue/next-rag/next-ai/internal/repository"
	"github.com/ashwinyue/next-rag/next-ai/internal/service/agent"
	"github.com/ashwinyue/next-rag/next-ai/internal/service/chat"
	"github.com/ashwinyue/next-rag/next-ai/internal/service/faq"
	"github.com/ashwinyue/next-rag/next-ai/internal/service/knowledge"
	"github.com/ashwinyue/next-rag/next-ai/internal/service/query"
	"github.com/ashwinyue/next-ai/internal/service/session"
	"github.com/ashwinyue/next-ai/internal/service/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/ashwinyue/next-rag/next-ai/internal/service/types"
	"github.com/cloudwego/eino-ext/components/embedding/dashscope"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino-ext/components/retriever/es8"
	"github.com/cloudwego/eino-ext/components/retriever/es8/search_mode"
	duckduckgov2 "github.com/cloudwego/eino-ext/components/tool/duckduckgo/v2"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/retriever"
	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/redis/go-redis/v9"
)

// Services 服务集合
type Services struct {
	// 业务服务
	Chat      *chat.Service
	Agent     *agent.Service
	Knowledge *knowledge.Service
	Tool      *tool.Service
	FAQ       *faq.Service

	// 配置
	Config     *config.Config
	SessionMgr *session.Manager

	// Eino 组件（直接使用 eino 类型，无封装）
	AllTools []einotool.BaseTool
	Embedder embedding.Embedder

	// RAG 组件
	ChatModel  model.ChatModel             // 用于查询处理和重排的 ChatModel
	Retriever  retriever.Retriever         // ES8 检索器
	Query      *query.Optimizer            // 查询优化器
	Rerankers  []types.Reranker            // 重排器列表（使用 types 包避免循环导入）
}

// NewServices 创建所有服务
// 参考 eino-examples，使用简单的 newXxx() 函数直接初始化 eino 组件
func NewServices(repo *repository.Repositories, cfg *config.Config, redisClient *redis.Client) (*Services, error) {
	ctx := context.Background()

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
	allTools := newTools(ctx, cfg, retriever)
	log.Printf("Initialized %d tools", len(allTools))

	return &Services{
		Chat:      chat.NewService(repo),
		Agent:     agent.NewService(repo, cfg, allTools),
		Knowledge: knowledge.NewService(repo, cfg, embedder),
		Tool:      tool.NewService(repo),
		FAQ:       faq.NewService(repo),

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
func newChatModel(ctx context.Context, cfg *config.Config) (model.ChatModel, error) {
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
func newToolCallingChatModel(ctx context.Context, cfg *config.Config) (model.ToolCallingChatModel, error) {
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
func newTools(ctx context.Context, cfg *config.Config, retriever *es8.Retriever) []einotool.BaseTool {
	tools := []einotool.BaseTool{}

	// 添加网络搜索工具
	tools = append(tools, newWebSearchTool(ctx))

	// 添加知识库搜索工具
	if retriever != nil {
		tools = append(tools, newKnowledgeSearchTool(retriever))
	}

	return tools
}

// KnowledgeSearchTool 知识库搜索工具
type KnowledgeSearchTool struct {
	retriever *es8.Retriever
}

// NewKnowledgeSearchTool 创建知识库搜索工具
func newKnowledgeSearchTool(r *es8.Retriever) einotool.InvokableTool {
	return &KnowledgeSearchTool{retriever: r}
}

func (t *KnowledgeSearchTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "knowledge_search",
		Desc: "Searches the knowledge base for relevant information using semantic and keyword search.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"query": {
				Type:     schema.String,
				Desc:     "The search query",
				Required: true,
			},
			"top_k": {
				Type: schema.Integer,
				Desc: "Number of results (optional, default 10)",
			},
		}),
	}, nil
}

func (t *KnowledgeSearchTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...einotool.Option) (string, error) {
	var input struct {
		Query string `json:"query"`
		TopK  int    `json:"top_k"`
	}
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	if input.Query == "" {
		return "", fmt.Errorf("query is required")
	}

	if input.TopK <= 0 {
		input.TopK = 10
	}

	docs, err := t.retriever.Retrieve(ctx, input.Query, retriever.WithTopK(input.TopK))
	if err != nil {
		return "", fmt.Errorf("retriever failed: %w", err)
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

	output, _ := json.MarshalIndent(map[string]interface{}{
		"results": results,
		"total":   len(results),
		"query":   input.Query,
	}, "", "  ")

	return string(output), nil
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
func newRerankers(ctx context.Context, cfg *config.Config, chatModel model.ChatModel) []types.Reranker {
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
	chatModel model.ChatModel
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
