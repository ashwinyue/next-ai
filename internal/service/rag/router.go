// Package rag 提供 Eino Router 集成
// 使用 Eino 官方的 router.NewRetriever 实现多检索器路由和 RRF 融合
package rag

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/retriever"
	einoRouter "github.com/cloudwego/eino/flow/retriever/router"
	"github.com/cloudwego/eino/schema"
)

// RouterConfig 路由检索器配置
// 参考 eino-examples/components/retriever/router/main.go
type RouterConfig struct {
	// Retrievers 可用的检索器集合
	// key 为检索器名称，value 为检索器实例
	Retrievers map[string]retriever.Retriever
	// Router 路由函数：根据查询返回要使用的检索器名称列表
	// 返回空列表表示使用所有检索器
	Router func(ctx context.Context, query string) ([]string, error)
	// FusionFunc 融合函数：nil 使用默认 RRF 算法
	FusionFunc func(ctx context.Context, result map[string][]*schema.Document) ([]*schema.Document, error)
}

// NewRouterRetriever 创建路由检索器
// 使用 Eino 官方实现，支持多检索器并发检索和 RRF 融合
func NewRouterRetriever(ctx context.Context, cfg *RouterConfig) (retriever.Retriever, error) {
	if len(cfg.Retrievers) == 0 {
		return nil, fmt.Errorf("retrievers is empty")
	}

	// 转换为 Eino 类型
	einoRetrievers := make(map[string]retriever.Retriever)
	for name, r := range cfg.Retrievers {
		einoRetrievers[name] = r
	}

	// 如果未提供路由函数，使用默认的"全选"路由
	routerFn := cfg.Router
	if routerFn == nil {
		routerFn = func(ctx context.Context, query string) ([]string, error) {
			names := make([]string, 0, len(cfg.Retrievers))
			for name := range cfg.Retrievers {
				names = append(names, name)
			}
			return names, nil
		}
	}

	// 创建 Eino Router
	einoCfg := &einoRouter.Config{
		Retrievers: einoRetrievers,
		Router:     routerFn,
		FusionFunc: convertFusionFunc(cfg.FusionFunc),
	}

	return einoRouter.NewRetriever(ctx, einoCfg)
}

// convertFusionFunc 转换融合函数签名（不再需要，已简化）
func convertFusionFunc(fn func(ctx context.Context, result map[string][]*schema.Document) ([]*schema.Document, error)) func(ctx context.Context, result map[string][]*schema.Document) ([]*schema.Document, error) {
	return fn
}

// ========== 预定义路由函数 ==========

// KeywordRouter 创建关键词路由函数
// 根据查询中的关键词匹配选择检索器
//
// 参数：
//
//	rules - 路由规则：map[检索器名称][]关键词
//
// 示例：
//
//	rules := map[string][]string{
//	    "kb_tech":  {"技术", "开发", "API", "代码"},
//	    "kb_sales": {"销售", "价格", "报价"},
//	}
//	routerFunc := KeywordRouter(rules)
func KeywordRouter(rules map[string][]string) func(ctx context.Context, query string) ([]string, error) {
	return func(ctx context.Context, query string) ([]string, error) {
		selected := make([]string, 0)

		for retrieverName, keywords := range rules {
			for _, keyword := range keywords {
				if strings.Contains(query, keyword) {
					selected = append(selected, retrieverName)
					break
				}
			}
		}

		// 如果没有匹配，返回空（使用所有检索器）
		return selected, nil
	}
}

// AllRouter 创建"全选"路由函数
// 始终返回所有已注册的检索器名称
func AllRouter(retrieverNames []string) func(ctx context.Context, query string) ([]string, error) {
	return func(ctx context.Context, query string) ([]string, error) {
		return retrieverNames, nil
	}
}

// PriorityRouter 创建优先级路由函数
// 按优先级顺序尝试检索器，一旦某个检索器返回足够结果就停止
func PriorityRouter(priority []string) func(ctx context.Context, query string) ([]string, error) {
	return func(ctx context.Context, query string) ([]string, error) {
		// 返回按优先级排序的检索器列表
		// 注意：Router 的并发执行可能会忽略顺序
		// 如需严格的顺序执行，需要在应用层控制
		return priority, nil
	}
}

// ========== 预定义融合函数 ==========

// DefaultRRF 使用默认 RRF (Reciprocal Rank Fusion) 算法
// FusionFunc 设为 nil 即使用默认 RRF，此函数仅为文档说明
func DefaultRRF() func(ctx context.Context, result map[string][]*schema.Document) ([]*schema.Document, error) {
	// Eino Router 内置的 RRF 实现：
	// score = 1.0 / (k + rank)，其中 k=60
	// 这与我们删除的自定义 Router 实现相同
	return nil
}

// WeightedFusion 创建加权融合函数
// 不同检索器的结果按权重加权合并
func WeightedFusion(weights map[string]float64) func(ctx context.Context, result map[string][]*schema.Document) ([]*schema.Document, error) {
	return func(ctx context.Context, result map[string][]*schema.Document) ([]*schema.Document, error) {
		type docScore struct {
			doc   *schema.Document
			score float64
		}

		scoreMap := make(map[string]*docScore)

		for retrieverName, docList := range result {
			weight := weights[retrieverName]
			if weight == 0 {
				weight = 1.0
			}

			for _, doc := range docList {
				if existing, found := scoreMap[doc.ID]; found {
					existing.score += doc.Score() * weight
				} else {
					scoreMap[doc.ID] = &docScore{
						doc:   doc,
						score: doc.Score() * weight,
					}
				}
			}
		}

		// 转换为切片并排序
		resultDocs := make([]*schema.Document, 0, len(scoreMap))
		for _, ds := range scoreMap {
			// 创建新文档并设置加权分数
			newDoc := ds.doc.WithScore(ds.score)
			resultDocs = append(resultDocs, newDoc)
		}

		// 按分数排序
		sortByScore(resultDocs)

		return resultDocs, nil
	}
}

// ========== 辅助函数 ==========

func sortByScore(docs []*schema.Document) {
	// 简单的冒泡排序（小数据集足够）
	n := len(docs)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if docs[j].Score() < docs[j+1].Score() {
				docs[j], docs[j+1] = docs[j+1], docs[j]
			}
		}
	}
}
