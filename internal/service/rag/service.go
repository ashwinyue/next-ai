// Package rag 提供 RAG 检索服务
package rag

import (
	"context"
	"fmt"

	"github.com/ashwinyue/next-ai/internal/service/types"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/retriever"
)

// Service RAG 服务
type Service struct {
	chatModel      model.ChatModel
	baseRetriever  retriever.Retriever // 基础检索器
	multiRetriever retriever.Retriever // 多查询检索器（Eino 组件）
	rerankers      []types.Reranker
}

// ServiceConfig RAG 服务配置
type ServiceConfig struct {
	ChatModel model.ChatModel
	Retriever retriever.Retriever
	Rerankers []types.Reranker

	// 多查询配置
	EnableMultiQuery bool            // 是否启用多查询检索
	MaxQueriesNum    int             // 最大查询数量（使用 LLM 生成）
	RewriteLLM       model.ChatModel // 用于生成查询变体的 LLM（可选，不传则用 ChatModel）
}

// NewService 创建 RAG 服务（简单模式，无多查询）
func NewService(
	chatModel model.ChatModel,
	retriever retriever.Retriever,
	rerankers []types.Reranker,
) *Service {
	return &Service{
		chatModel:     chatModel,
		baseRetriever: retriever,
		rerankers:     rerankers,
	}
}

// NewServiceWithConfig 创建 RAG 服务（带完整配置）
func NewServiceWithConfig(ctx context.Context, cfg *ServiceConfig) (*Service, error) {
	svc := &Service{
		chatModel:     cfg.ChatModel,
		baseRetriever: cfg.Retriever,
		rerankers:     cfg.Rerankers,
	}

	// 如果启用多查询检索，创建多查询检索器
	if cfg.EnableMultiQuery {
		rewriteLLM := cfg.RewriteLLM
		if rewriteLLM == nil {
			rewriteLLM = cfg.ChatModel
		}

		var err error
		// 融合函数不需要预计算，直接使用闭包
		svc.multiRetriever, err = NewMultiQueryRetriever(ctx, &MultiQueryConfig{
			Retriever:     cfg.Retriever,
			RewriteLLM:    rewriteLLM,
			MaxQueriesNum: cfg.MaxQueriesNum,
			FusionFunc:    MultiQueryWeightedFusion(nil), // 使用加权融合
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create multiquery retriever: %w", err)
		}
	}

	return svc, nil
}

// RetrieveRequest 检索请求
type RetrieveRequest struct {
	Query            string `json:"query"`
	TopK             int    `json:"top_k"`
	EnableOptimize   bool   `json:"enable_optimize"`
	EnableRerank     bool   `json:"enable_rerank"`
	EnableMultiQuery bool   `json:"enable_multi_query"` // 是否使用 Eino 多查询检索器
}

// RetrieveResponse 检索响应
type RetrieveResponse struct {
	Query     string           `json:"query"`
	Documents []types.Document `json:"documents"`
	Total     int              `json:"total"`
}

// Retrieve 执行 RAG 检索
func (s *Service) Retrieve(ctx context.Context, req *RetrieveRequest) (*RetrieveResponse, error) {
	if req.Query == "" {
		return nil, fmt.Errorf("query is required")
	}

	topK := req.TopK
	if topK <= 0 {
		topK = 10
	}

	// 选择检索器：统一使用 Eino 多查询检索器
	var retrieverForUse retriever.Retriever
	if (req.EnableMultiQuery || req.EnableOptimize) && s.multiRetriever != nil {
		// 使用 Eino NewMultiQueryRetriever（EnableOptimize 重定向到 Eino 组件）
		retrieverForUse = s.multiRetriever
	} else {
		// 使用基础检索器
		retrieverForUse = s.baseRetriever
	}

	if retrieverForUse == nil {
		return &RetrieveResponse{
			Query:     req.Query,
			Documents: []types.Document{},
			Total:     0,
		}, nil
	}

	// 执行检索
	docs, err := retrieverForUse.Retrieve(ctx, req.Query)
	if err != nil {
		return nil, fmt.Errorf("retrieve failed: %w", err)
	}

	if len(docs) == 0 {
		return &RetrieveResponse{
			Query:     req.Query,
			Documents: []types.Document{},
			Total:     0,
		}, nil
	}

	// 重排（可选）
	if req.EnableRerank && len(s.rerankers) > 0 {
		for _, rnk := range s.rerankers {
			reranked, err := rnk.Rerank(ctx, req.Query, docs)
			if err != nil {
				continue
			}
			docs = reranked
		}
	}

	// 限制返回数量
	if len(docs) > topK {
		docs = docs[:topK]
	}

	// 转换为响应格式
	result := make([]types.Document, len(docs))
	for i, doc := range docs {
		metadata := make(map[string]interface{})
		if doc.MetaData != nil {
			for k, v := range doc.MetaData {
				metadata[k] = v
			}
		}

		result[i] = types.Document{
			ID:       doc.ID,
			Content:  doc.Content,
			Score:    doc.Score(),
			Metadata: metadata,
		}
	}

	return &RetrieveResponse{
		Query:     req.Query,
		Documents: result,
		Total:     len(result),
	}, nil
}

// ToContext 将检索结果转换为 LLM 上下文
func ToContext(resp *RetrieveResponse) string {
	if len(resp.Documents) == 0 {
		return "未找到相关文档。"
	}

	contextStr := "以下是与查询相关的文档内容：\n\n"
	for i, doc := range resp.Documents {
		content := doc.Content
		if len(content) > 500 {
			content = content[:500] + "..."
		}

		contextStr += fmt.Sprintf("[%d] %s\n", i+1, content)

		if doc.Score > 0 {
			contextStr += fmt.Sprintf("    相关度: %.2f\n", doc.Score)
		}

		if title, ok := doc.Metadata["title"].(string); ok {
			contextStr += fmt.Sprintf("    标题: %s\n", title)
		}

		contextStr += "\n"
	}

	return contextStr
}
