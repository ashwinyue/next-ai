// Package rag 提供 RAG 检索服务
package rag

import (
	"context"
	"fmt"

	"github.com/ashwinyue/next-rag/next-ai/internal/service/query"
	"github.com/ashwinyue/next-rag/next-ai/internal/service/types"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
)

// Service RAG 服务
type Service struct {
	chatModel  model.ChatModel
	retriever  retriever.Retriever
	query      *query.Optimizer
	rerankers  []types.Reranker
}

// NewService 创建 RAG 服务
func NewService(
	chatModel model.ChatModel,
	retriever retriever.Retriever,
	queryOpt *query.Optimizer,
	rerankers []types.Reranker,
) *Service {
	return &Service{
		chatModel: chatModel,
		retriever: retriever,
		query:     queryOpt,
		rerankers: rerankers,
	}
}

// RetrieveRequest 检索请求
type RetrieveRequest struct {
	Query          string `json:"query"`
	TopK           int    `json:"top_k"`
	EnableOptimize bool   `json:"enable_optimize"`
	EnableRerank   bool   `json:"enable_rerank"`
}

// RetrieveResponse 检索响应
type RetrieveResponse struct {
	Query     string              `json:"query"`
	Documents []types.Document    `json:"documents"`
	Total     int                `json:"total"`
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

	// 1. 查询优化（可选）
	var queries []string
	if req.EnableOptimize && s.query != nil {
		optimized, err := s.query.Optimize(ctx, req.Query)
		if err != nil {
			queries = []string{req.Query}
		} else {
			queries = optimized.GetQueries()
		}
	} else {
		queries = []string{req.Query}
	}

	// 2. 多查询检索并去重
	allDocs := make([]*schema.Document, 0)
	seenIDs := make(map[string]bool)

	for _, q := range queries {
		if s.retriever == nil {
			continue
		}

		docs, err := s.retriever.Retrieve(ctx, q)
		if err != nil {
			continue // 某个查询失败不影响其他
		}

		for _, doc := range docs {
			if doc.ID != "" && !seenIDs[doc.ID] {
				seenIDs[doc.ID] = true
				allDocs = append(allDocs, doc)
			} else if doc.ID == "" {
				allDocs = append(allDocs, doc)
			}
		}
	}

	if len(allDocs) == 0 {
		return &RetrieveResponse{
			Query:     req.Query,
			Documents: []types.Document{},
			Total:     0,
		}, nil
	}

	// 3. 重排（可选）
	if req.EnableRerank && len(s.rerankers) > 0 {
		for _, rnk := range s.rerankers {
			reranked, err := rnk.Rerank(ctx, req.Query, allDocs)
			if err != nil {
				continue
			}
			allDocs = reranked
		}
	}

	// 4. 限制返回数量
	if len(allDocs) > topK {
		allDocs = allDocs[:topK]
	}

	// 5. 转换为响应格式
	docs := make([]types.Document, len(allDocs))
	for i, doc := range allDocs {
		metadata := make(map[string]interface{})
		if doc.MetaData != nil {
			for k, v := range doc.MetaData {
				metadata[k] = v
			}
		}

		docs[i] = types.Document{
			ID:       doc.ID,
			Content:  doc.Content,
			Score:    doc.Score(),
			Metadata: metadata,
		}
	}

	return &RetrieveResponse{
		Query:     req.Query,
		Documents: docs,
		Total:     len(docs),
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
