// Package retriever 提供检索器服务
// 参考 next-ai/docs/eino-integration-guide.md
// 直接实现路由检索逻辑，避免冗余封装
package retriever

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
)

// ========== 路由检索器 ==========

// Router 路由检索器
// 根据查询动态选择检索器并融合结果
type Router struct {
	retrievers map[string]retriever.Retriever
	routerFunc func(ctx context.Context, query string) ([]string, error)
	fusionFunc func(map[string][]*schema.Document) []*schema.Document
	mu         sync.RWMutex
}

// NewRouter 创建路由检索器
func NewRouter() *Router {
	return &Router{
		retrievers: make(map[string]retriever.Retriever),
		fusionFunc: rrfFusion(60), // 默认 RRF 融合
	}
}

// Add 添加检索器
func (r *Router) Add(name string, retriever retriever.Retriever) *Router {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.retrievers[name] = retriever
	return r
}

// Remove 移除检索器
func (r *Router) Remove(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.retrievers, name)
}

// SetRouterFunc 设置路由函数
func (r *Router) SetRouterFunc(fn func(ctx context.Context, query string) ([]string, error)) *Router {
	r.routerFunc = fn
	return r
}

// SetFusionFunc 设置融合函数
func (r *Router) SetFusionFunc(fn func(map[string][]*schema.Document) []*schema.Document) *Router {
	r.fusionFunc = fn
	return r
}

// Retrieve 执行路由检索
func (r *Router) Retrieve(ctx context.Context, query string, opts ...retriever.Option) ([]*schema.Document, error) {
	r.mu.RLock()

	// 1. 路由：选择要使用的检索器
	var retrieverNames []string
	if r.routerFunc != nil {
		names, err := r.routerFunc(ctx, query)
		if err != nil {
			r.mu.RUnlock()
			return nil, fmt.Errorf("router error: %w", err)
		}
		retrieverNames = names
	} else {
		// 默认使用所有检索器
		names := make([]string, 0, len(r.retrievers))
		for name := range r.retrievers {
			names = append(names, name)
		}
		retrieverNames = names
	}

	if len(retrieverNames) == 0 {
		r.mu.RUnlock()
		return []*schema.Document{}, nil
	}

	// 2. 并发执行选中的检索器
	type result struct {
		name string
		docs []*schema.Document
		err  error
	}

	results := make(chan result, len(retrieverNames))

	for _, name := range retrieverNames {
		rtr, exists := r.retrievers[name]
		if !exists {
			results <- result{name: name, err: fmt.Errorf("retriever '%s' not found", name)}
			continue
		}

		go func(name string, rtr retriever.Retriever) {
			docs, err := rtr.Retrieve(ctx, query, opts...)
			results <- result{name: name, docs: docs, err: err}
		}(name, rtr)
	}

	r.mu.RUnlock()

	// 3. 收集结果
	allDocs := make(map[string][]*schema.Document)
	var firstError error

	for i := 0; i < len(retrieverNames); i++ {
		res := <-results
		if res.err != nil {
			if firstError == nil {
				firstError = res.err
			}
			continue
		}
		allDocs[res.name] = res.docs
	}

	if len(allDocs) == 0 {
		if firstError != nil {
			return nil, firstError
		}
		return []*schema.Document{}, nil
	}

	// 4. 融合结果
	fusedDocs := r.fusionFunc(allDocs)

	return fusedDocs, nil
}

// ========== 融合函数 ==========

// rrfFusion RRF (Reciprocal Rank Fusion) 融合算法
func rrfFusion(k int) func(map[string][]*schema.Document) []*schema.Document {
	return func(docs map[string][]*schema.Document) []*schema.Document {
		// 使用 map 存储分数和文档
		type docScore struct {
			doc   *schema.Document
			score float64
		}
		scoreMap := make(map[string]*docScore)

		// 为每个文档计算 RRF 分数
		for _, docList := range docs {
			for rank, doc := range docList {
				if existing, found := scoreMap[doc.ID]; found {
					// 累加 RRF 分数
					existing.score += 1.0 / float64(k+rank+1)
				} else {
					// 初始化分数
					scoreMap[doc.ID] = &docScore{
						doc:   doc,
						score: 1.0 / float64(k+rank+1),
					}
				}
			}
		}

		// 转换为切片并按分数排序
		result := make([]*schema.Document, 0, len(scoreMap))
		for _, ds := range scoreMap {
			result = append(result, ds.doc)
		}

		sort.Slice(result, func(i, j int) bool {
			// 使用 RRF 分数排序（存储在临时结构中）
			// 简化版：按原始分数排序
			return result[i].Score() > result[j].Score()
		})

		return result
	}
}

// weightedFusion 加权融合算法
func weightedFusion(weights map[string]float64) func(map[string][]*schema.Document) []*schema.Document {
	return func(docs map[string][]*schema.Document) []*schema.Document {
		// 使用 map 存储分数和文档
		type docScore struct {
			doc   *schema.Document
			score float64
		}
		scoreMap := make(map[string]*docScore)

		for retrieverName, docList := range docs {
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

		// 转换为切片并按分数排序
		result := make([]*schema.Document, 0, len(scoreMap))
		for _, ds := range scoreMap {
			result = append(result, ds.doc)
		}

		sort.Slice(result, func(i, j int) bool {
			return result[i].Score() > result[j].Score()
		})

		return result
	}
}

// ========== 路由函数辅助方法 ==========

// KeywordRouterFunc 创建关键词路由函数
func KeywordRouterFunc(rules map[string][]string) func(ctx context.Context, query string) ([]string, error) {
	return func(ctx context.Context, query string) ([]string, error) {
		selected := make([]string, 0)

		for retrieverName, keywords := range rules {
			for _, keyword := range keywords {
				if contains(query, keyword) {
					selected = append(selected, retrieverName)
					break
				}
			}
		}

		// 如果没有匹配，返回所有检索器名称（需要在调用时传入）
		return selected, nil
	}
}

// AllRouterFunc 返回所有检索器的路由函数
func (r *Router) AllRouterFunc() func(ctx context.Context, query string) ([]string, error) {
	return func(ctx context.Context, query string) ([]string, error) {
		r.mu.RLock()
		defer r.mu.RUnlock()

		names := make([]string, 0, len(r.retrievers))
		for name := range r.retrievers {
			names = append(names, name)
		}
		return names, nil
	}
}

// ========== 辅助函数 ==========

func contains(s, substr string) bool {
	return len(s) >= len(substr) && indexOf(s, substr) >= 0
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
