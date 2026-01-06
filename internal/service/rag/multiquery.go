// Package rag 提供 MultiQuery 检索器集成
// 使用 Eino 官方 multiquery.NewRetriever 实现查询扩展
package rag

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/retriever"
	einoMultiquery "github.com/cloudwego/eino/flow/retriever/multiquery"
	"github.com/cloudwego/eino/schema"
)

// MultiQueryConfig MultiQuery 配置
// 参考 eino-examples/components/retriever/multiquery/main.go
type MultiQueryConfig struct {
	// 底层检索器
	Retriever retriever.Retriever

	// ========== 方式1: 使用 LLM 生成多条查询 ==========
	// RewriteLLM 用于生成查询变体的 ChatModel
	RewriteLLM model.ChatModel
	// MaxQueriesNum 最大查询数量，默认 5
	MaxQueriesNum int

	// ========== 方式2: 自定义查询生成逻辑 ==========
	// RewriteHandler 自定义查询生成函数，优先级高于 RewriteLLM
	RewriteHandler func(ctx context.Context, query string) ([]string, error)

	// ========== 结果融合 ==========
	// FusionFunc 融合函数：nil 使用默认去重融合
	FusionFunc func(ctx context.Context, docs [][]*schema.Document) ([]*schema.Document, error)
}

// NewMultiQueryRetriever 创建多查询检索器
// 使用 Eino 官方实现，支持查询扩展提高召回率
//
// 使用方式1 - LLM 生成查询:
//
//	multiRetriever, err := rag.NewMultiQueryRetriever(ctx, &rag.MultiQueryConfig{
//	    Retriever:     es8Retriever,
//	    RewriteLLM:    chatModel,
//	    MaxQueriesNum: 3,
//	})
//
// 使用方式2 - 自定义查询生成:
//
//	multiRetriever, err := rag.NewMultiQueryRetriever(ctx, &rag.MultiQueryConfig{
//	    Retriever: es8Retriever,
//	    RewriteHandler: func(ctx context.Context, query string) ([]string, error) {
//	        // 自定义逻辑：如分词、同义词扩展等
//	        return []string{query, strings.ToLower(query)}, nil
//	    },
//	})
func NewMultiQueryRetriever(ctx context.Context, cfg *MultiQueryConfig) (retriever.Retriever, error) {
	if cfg.Retriever == nil {
		return nil, fmt.Errorf("Retriever is required")
	}
	if cfg.RewriteHandler == nil && cfg.RewriteLLM == nil {
		return nil, fmt.Errorf("at least one of RewriteHandler and RewriteLLM must be provided")
	}

	// 转换为 Eino 配置
	einoCfg := &einoMultiquery.Config{
		OrigRetriever:  cfg.Retriever,
		RewriteHandler: cfg.RewriteHandler,
		RewriteLLM:     cfg.RewriteLLM,
		MaxQueriesNum:  cfg.MaxQueriesNum,
		FusionFunc:     cfg.FusionFunc,
	}

	return einoMultiquery.NewRetriever(ctx, einoCfg)
}

// ========== 预定义查询生成函数 ==========

// SimpleSplitter 简单分词查询生成器
// 按空格分割查询，生成多个子查询
func SimpleSplitter() func(ctx context.Context, query string) ([]string, error) {
	return func(ctx context.Context, query string) ([]string, error) {
		words := strings.Fields(query)
		if len(words) <= 1 {
			return []string{query}, nil
		}
		return []string{query, strings.Join(words, " ")}, nil
	}
}

// LowercaseVariants 小写变体查询生成器
// 生成原查询和小写版本
func LowercaseVariants() func(ctx context.Context, query string) ([]string, error) {
	return func(ctx context.Context, query string) ([]string, error) {
		lower := strings.ToLower(query)
		if lower == query {
			return []string{query}, nil
		}
		return []string{query, lower}, nil
	}
}

// PrefixSuffixAdder 前后缀添加器
// 为查询添加常见前缀和后缀
func PrefixSuffixAdder(prefixes, suffixes []string) func(ctx context.Context, query string) ([]string, error) {
	return func(ctx context.Context, query string) ([]string, error) {
		queries := []string{query}

		// 添加前缀变体
		for _, prefix := range prefixes {
			queries = append(queries, prefix+" "+query)
		}

		// 添加后缀变体
		for _, suffix := range suffixes {
			queries = append(queries, query+" "+suffix)
		}

		return queries, nil
	}
}

// ========== 预定义融合函数 ==========

// MultiQueryWeightedFusion 加权融合函数
// 根据文档在多个查询结果中的出现次数和位置进行加权
func MultiQueryWeightedFusion(docs [][]*schema.Document) func(ctx context.Context, docs [][]*schema.Document) ([]*schema.Document, error) {
	return func(ctx context.Context, results [][]*schema.Document) ([]*schema.Document, error) {
		// 统计每个文档的出现次数和最佳位置
		type docInfo struct {
			doc     *schema.Document
			count   int
			bestPos int
		}

		docMap := make(map[string]*docInfo)

		for _, queryDocs := range results {
			for pos, doc := range queryDocs {
				if info, exists := docMap[doc.ID]; exists {
					info.count++
					if pos < info.bestPos {
						info.bestPos = pos
					}
				} else {
					docMap[doc.ID] = &docInfo{
						doc:     doc,
						count:   1,
						bestPos: pos,
					}
				}
			}
		}

		// 按出现次数和位置排序
		type rankItem struct {
			doc   *schema.Document
			score float64
		}

		ranked := make([]*rankItem, 0, len(docMap))
		for _, info := range docMap {
			// 评分公式：出现次数权重 0.7 + 位置权重 0.3
			posScore := 1.0 / float64(info.bestPos+1)
			totalScore := float64(info.count)*0.7 + posScore*0.3
			ranked = append(ranked, &rankItem{
				doc:   info.doc,
				score: totalScore,
			})
		}

		// 排序并更新分数
		for i := 0; i < len(ranked)-1; i++ {
			for j := 0; j < len(ranked)-i-1; j++ {
				if ranked[j].score < ranked[j+1].score {
					ranked[j], ranked[j+1] = ranked[j+1], ranked[j]
				}
			}
		}

		result := make([]*schema.Document, len(ranked))
		for i, item := range ranked {
			result[i] = item.doc.WithScore(item.score)
		}

		return result, nil
	}
}

// RoundRobinFusion 轮询融合函数
// 从每个查询结果中轮流选取文档，保证多样性
func RoundRobinFusion() func(ctx context.Context, docs [][]*schema.Document) ([]*schema.Document, error) {
	return func(ctx context.Context, results [][]*schema.Document) ([]*schema.Document, error) {
		seen := make(map[string]bool)
		var fused []*schema.Document

		maxLen := 0
		for _, docs := range results {
			if len(docs) > maxLen {
				maxLen = len(docs)
			}
		}

		for i := 0; i < maxLen; i++ {
			for _, docs := range results {
				if i < len(docs) {
					doc := docs[i]
					if !seen[doc.ID] {
						seen[doc.ID] = true
						fused = append(fused, doc)
					}
				}
			}
		}

		return fused, nil
	}
}

// ConcatFusion 简单拼接融合函数
// 拼接所有结果，保留顺序（适合已排序的结果）
func ConcatFusion() func(ctx context.Context, docs [][]*schema.Document) ([]*schema.Document, error) {
	return func(ctx context.Context, results [][]*schema.Document) ([]*schema.Document, error) {
		seen := make(map[string]bool)
		var fused []*schema.Document

		for _, docs := range results {
			for _, doc := range docs {
				if !seen[doc.ID] {
					seen[doc.ID] = true
					fused = append(fused, doc)
				}
			}
		}

		return fused, nil
	}
}
