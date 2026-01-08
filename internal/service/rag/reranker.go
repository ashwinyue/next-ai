package rag

import (
	"context"
	"fmt"

	"github.com/ashwinyue/next-ai/internal/config"
	"github.com/ashwinyue/next-ai/internal/service/types"
	ecomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// NewRerankers 创建默认的重排器列表（导出供 service 使用）
func NewRerankers(ctx context.Context, cfg *config.Config, chatModel ecomodel.ChatModel) []types.Reranker {
	rerankers := []types.Reranker{}

	// 添加分数重排（轻量级，始终启用）
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
