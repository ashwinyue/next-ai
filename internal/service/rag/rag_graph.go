// Package rag 提供基于 Eino Graph 的 RAG 编排服务
// 参考 next-ai/docs/eino-integration-guide.md
// 直接使用 eino compose.Graph，避免冗余封装
package rag

import (
	"context"
	"fmt"

	"github.com/ashwinyue/next-ai/internal/service/query"
	"github.com/ashwinyue/next-ai/internal/service/rerank"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// ========== RAG 状态 ==========

// State RAG 流程状态
type State struct {
	// 输入
	Query string

	// 查询处理
	OptimizedQuery *query.OptimizedQuery

	// 检索结果
	RetrievedDocs []*schema.Document
	RerankedDocs  []*schema.Document

	// 错误
	Error error
}

// ToContext 转换为 LLM 上下文
func (s *State) ToContext() string {
	if len(s.RerankedDocs) == 0 {
		return "未找到相关文档。"
	}

	contextStr := "以下是与查询相关的文档内容：\n\n"
	for i, doc := range s.RerankedDocs {
		content := doc.Content
		if len(content) > 500 {
			content = content[:500] + "..."
		}

		contextStr += fmt.Sprintf("[%d] %s\n", i+1, content)

		if doc.MetaData != nil {
			if title, ok := doc.MetaData["title"].(string); ok {
				contextStr += fmt.Sprintf("    标题: %s\n", title)
			}
		}

		if doc.Score() > 0 {
			contextStr += fmt.Sprintf("    相关度: %.2f\n", doc.Score())
		}

		contextStr += "\n"
	}

	return contextStr
}

// ========== RAG Graph 配置 ==========

// Config RAG Graph 配置
type Config struct {
	// 核心组件
	ChatModel model.ChatModel
	Retriever retriever.Retriever

	// 查询处理
	EnableRewrite bool
	EnableExpand  bool
	NumVariants   int

	// 重排
	Rerankers []rerank.Reranker
}

// ========== RAG Graph ==========

// RAG 基于 Eino Graph 的 RAG 编排器
type RAG struct {
	graph    compose.Runnable[string, *State]
	config   *Config
	rewriter *query.Rewriter
	expander *query.Expander
}

// New 创建 RAG Graph
func New(ctx context.Context, cfg *Config) (*RAG, error) {
	if cfg.ChatModel == nil {
		return nil, fmt.Errorf("chat model is required")
	}
	if cfg.Retriever == nil {
		return nil, fmt.Errorf("retriever is required")
	}

	rag := &RAG{
		config: cfg,
	}

	// 初始化查询处理组件
	if cfg.EnableRewrite {
		rag.rewriter = query.NewRewriter(cfg.ChatModel)
	}
	if cfg.EnableExpand {
		rag.expander = query.NewExpander(cfg.ChatModel, cfg.NumVariants)
	}

	// 构建 Graph
	graph, err := rag.buildGraph(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build graph: %w", err)
	}

	rag.graph = graph
	return rag, nil
}

// buildGraph 构建 Eino Graph
func (r *RAG) buildGraph(ctx context.Context) (compose.Runnable[string, *State], error) {
	// 创建 Graph，使用 State 作为本地状态
	g := compose.NewGraph[string, *State](
		compose.WithGenLocalState(func(ctx context.Context) *State {
			return &State{}
		}),
	)

	// 1. 初始化节点
	initNode := compose.InvokableLambda(func(ctx context.Context, query string) (*State, error) {
		return &State{Query: query}, nil
	})

	// 2. 查询优化节点
	optimizeNode := compose.InvokableLambda(func(ctx context.Context, state *State) (*State, error) {
		return r.processQuery(ctx, state)
	})

	// 3. 检索节点
	retrieveNode := compose.InvokableLambda(func(ctx context.Context, state *State) (*State, error) {
		return r.processRetrieve(ctx, state)
	})

	// 4. 重排节点
	rerankNode := compose.InvokableLambda(func(ctx context.Context, state *State) (*State, error) {
		return r.processRerank(ctx, state)
	})

	// 添加节点
	if err := g.AddLambdaNode("init", initNode); err != nil {
		return nil, err
	}
	if err := g.AddLambdaNode("optimize", optimizeNode); err != nil {
		return nil, err
	}
	if err := g.AddLambdaNode("retrieve", retrieveNode); err != nil {
		return nil, err
	}
	if err := g.AddLambdaNode("rerank", rerankNode); err != nil {
		return nil, err
	}

	// 构建边
	currentNode := compose.START

	// init
	if err := g.AddEdge(currentNode, "init"); err != nil {
		return nil, err
	}
	currentNode = "init"

	// optimize (如果启用)
	if r.config.EnableRewrite || r.config.EnableExpand {
		if err := g.AddEdge(currentNode, "optimize"); err != nil {
			return nil, err
		}
		currentNode = "optimize"
	}

	// retrieve (必须)
	if err := g.AddEdge(currentNode, "retrieve"); err != nil {
		return nil, err
	}
	currentNode = "retrieve"

	// rerank (如果有重排器)
	if len(r.config.Rerankers) > 0 {
		if err := g.AddEdge(currentNode, "rerank"); err != nil {
			return nil, err
		}
		currentNode = "rerank"
	}

	// 连接到 END
	if err := g.AddEdge(currentNode, compose.END); err != nil {
		return nil, err
	}

	// 编译 Graph
	return g.Compile(ctx)
}

// processQuery 处理查询优化
func (r *RAG) processQuery(ctx context.Context, state *State) (*State, error) {
	// 创建优化器
	optimizer := query.NewOptimizer(r.config.ChatModel, r.config.NumVariants)

	// 执行优化
	optimized, err := optimizer.Optimize(ctx, state.Query)
	if err != nil {
		state.Error = fmt.Errorf("query optimization failed: %w", err)
		// 失败时使用原查询
		optimized = &query.OptimizedQuery{
			Original:  state.Query,
			Rewritten: state.Query,
			Expanded:  []string{state.Query},
		}
	}

	state.OptimizedQuery = optimized
	return state, nil
}

// processRetrieve 处理检索
func (r *RAG) processRetrieve(ctx context.Context, state *State) (*State, error) {
	var queries []string
	if state.OptimizedQuery != nil {
		queries = state.OptimizedQuery.GetQueries()
	}
	if len(queries) == 0 {
		queries = []string{state.Query}
	}

	// 多查询检索并去重
	allDocs := make([]*schema.Document, 0)
	seenIDs := make(map[string]bool)

	for _, q := range queries {
		docs, err := r.config.Retriever.Retrieve(ctx, q)
		if err != nil {
			continue // 某个查询失败不影响其他查询
		}

		// 去重合并
		for _, doc := range docs {
			if doc.ID != "" && !seenIDs[doc.ID] {
				seenIDs[doc.ID] = true
				allDocs = append(allDocs, doc)
			} else if doc.ID == "" {
				allDocs = append(allDocs, doc)
			}
		}
	}

	state.RetrievedDocs = allDocs
	state.RerankedDocs = allDocs // 初始时两者相同
	return state, nil
}

// processRerank 处理重排
func (r *RAG) processRerank(ctx context.Context, state *State) (*State, error) {
	docs := state.RetrievedDocs
	if len(docs) == 0 || len(r.config.Rerankers) == 0 {
		return state, nil
	}

	// 应用所有重排器
	for _, rnk := range r.config.Rerankers {
		reranked, err := rnk.Rerank(ctx, state.Query, docs)
		if err != nil {
			continue // 某个重排失败不影响其他
		}
		docs = reranked
	}

	state.RerankedDocs = docs
	return state, nil
}

// ========== 公开方法 ==========

// Retrieve 执行 RAG 检索
func (r *RAG) Retrieve(ctx context.Context, query string) ([]*schema.Document, error) {
	state, err := r.graph.Invoke(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("rag graph invoke failed: %w", err)
	}

	return state.RerankedDocs, nil
}

// StreamRetrieve 流式执行 RAG 检索
func (r *RAG) StreamRetrieve(ctx context.Context, query string) (*State, error) {
	return r.graph.Invoke(ctx, query)
}

// Invoke 执行 RAG 并返回完整状态
func (r *RAG) Invoke(ctx context.Context, query string) (*State, error) {
	return r.graph.Invoke(ctx, query)
}

// GetGraph 获取底层 Runnable（用于高级用法）
func (r *RAG) GetGraph() compose.Runnable[string, *State] {
	return r.graph
}

// ========== RAG 构建器 ==========

// Builder RAG 构建器
type Builder struct {
	cfg *Config
}

// NewBuilder 创建 RAG 构建器
func NewBuilder(chatModel model.ChatModel, retriever retriever.Retriever) *Builder {
	return &Builder{
		cfg: &Config{
			ChatModel:   chatModel,
			Retriever:   retriever,
			NumVariants: 3,
		},
	}
}

// WithQueryRewrite 启用查询重写
func (b *Builder) WithQueryRewrite() *Builder {
	b.cfg.EnableRewrite = true
	return b
}

// WithQueryExpand 启用查询扩展
func (b *Builder) WithQueryExpand(numVariants int) *Builder {
	b.cfg.EnableExpand = true
	if numVariants > 0 {
		b.cfg.NumVariants = numVariants
	}
	return b
}

// WithReranker 添加重排器
func (b *Builder) WithReranker(rnk rerank.Reranker) *Builder {
	b.cfg.Rerankers = append(b.cfg.Rerankers, rnk)
	return b
}

// WithScoreReranker 添加分数重排器
func (b *Builder) WithScoreReranker() *Builder {
	b.cfg.Rerankers = append(b.cfg.Rerankers, rerank.NewScoreReranker())
	return b
}

// WithDiversityReranker 添加多样性重排器
func (b *Builder) WithDiversityReranker(lambda float64, topN int) *Builder {
	b.cfg.Rerankers = append(b.cfg.Rerankers, rerank.NewDiversityReranker(lambda, topN))
	return b
}

// WithLLMReranker 添加 LLM 重排器
func (b *Builder) WithLLMReranker(topN int) *Builder {
	b.cfg.Rerankers = append(b.cfg.Rerankers, rerank.NewLLMReranker(b.cfg.ChatModel, topN))
	return b
}

// Build 构建 RAG Graph
func (b *Builder) Build(ctx context.Context) (*RAG, error) {
	return New(ctx, b.cfg)
}

// ========== 预设配置 ==========

// PresetBasic 基础 RAG（无额外处理）
func PresetBasic(ctx context.Context, chatModel model.ChatModel, retriever retriever.Retriever) (*RAG, error) {
	return New(ctx, &Config{
		ChatModel: chatModel,
		Retriever: retriever,
	})
}

// PresetAdvanced 高级 RAG（包含所有功能）
func PresetAdvanced(ctx context.Context, chatModel model.ChatModel, retriever retriever.Retriever) (*RAG, error) {
	return NewBuilder(chatModel, retriever).
		WithQueryRewrite().
		WithQueryExpand(3).
		WithScoreReranker().
		WithDiversityReranker(0.5, 5).
		Build(ctx)
}

// PresetSearchOptimized 搜索优化配置（重写 + 重排）
func PresetSearchOptimized(ctx context.Context, chatModel model.ChatModel, retriever retriever.Retriever) (*RAG, error) {
	return NewBuilder(chatModel, retriever).
		WithQueryRewrite().
		WithScoreReranker().
		Build(ctx)
}

// PresetRecallOptimized 召回优化配置（查询扩展 + 多样性重排）
func PresetRecallOptimized(ctx context.Context, chatModel model.ChatModel, retriever retriever.Retriever) (*RAG, error) {
	return NewBuilder(chatModel, retriever).
		WithQueryExpand(5).
		WithScoreReranker().
		WithDiversityReranker(0.3, 5).
		Build(ctx)
}
