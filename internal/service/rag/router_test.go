// Package rag 提供 RAG 路由功能单元测试
package rag

import (
	"context"
	"testing"

	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
)

// ========== NewRouterRetriever 测试 ==========

func TestNewRouterRetriever(t *testing.T) {
	ctx := context.Background()
	baseDoc := newDoc("test content", 0)

	tests := []struct {
		name        string
		cfg         *RouterConfig
		wantErr     bool
		errContains string
	}{
		{
			name: "valid config",
			cfg: &RouterConfig{
				Retrievers: map[string]retriever.Retriever{
					"r1": newMockRetriever([]*schema.Document{baseDoc}, nil),
				},
			},
			wantErr: false,
		},
		{
			name: "empty retrievers",
			cfg: &RouterConfig{
				Retrievers: map[string]retriever.Retriever{},
			},
			wantErr:     true,
			errContains: "retrievers is empty",
		},
		{
			name: "nil retrievers map",
			cfg: &RouterConfig{
				Retrievers: nil,
			},
			wantErr:     true,
			errContains: "retrievers is empty",
		},
		{
			name: "multiple retrievers",
			cfg: &RouterConfig{
				Retrievers: map[string]retriever.Retriever{
					"r1": newMockRetriever([]*schema.Document{baseDoc}, nil),
					"r2": newMockRetriever([]*schema.Document{baseDoc}, nil),
				},
			},
			wantErr: false,
		},
		{
			name: "with custom router",
			cfg: &RouterConfig{
				Retrievers: map[string]retriever.Retriever{
					"r1": newMockRetriever([]*schema.Document{baseDoc}, nil),
				},
				Router: func(ctx context.Context, query string) ([]string, error) {
					return []string{"r1"}, nil
				},
			},
			wantErr: false,
		},
		{
			name: "with custom fusion func",
			cfg: &RouterConfig{
				Retrievers: map[string]retriever.Retriever{
					"r1": newMockRetriever([]*schema.Document{baseDoc}, nil),
				},
				FusionFunc: func(ctx context.Context, result map[string][]*schema.Document) ([]*schema.Document, error) {
					return result["r1"], nil
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := NewRouterRetriever(ctx, tt.cfg)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewRouterRetriever() expected error, got nil")
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("Error = %v, want contain %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("NewRouterRetriever() unexpected error: %v", err)
			}
			if r == nil {
				t.Error("NewRouterRetriever() returned nil retriever")
			}
		})
	}
}

// ========== KeywordRouter 测试 ==========

func TestKeywordRouter(t *testing.T) {
	ctx := context.Background()

	rules := map[string][]string{
		"tech_kb":  {"技术", "开发", "API", "代码", "算法"},
		"sales_kb": {"销售", "价格", "报价", "采购", "订单"},
		"hr_kb":    {"招聘", "薪资", "面试", "离职"},
	}

	router := KeywordRouter(rules)

	tests := []struct {
		name          string
		query         string
		expectedCount int
		expected      []string
	}{
		{
			name:          "single match",
			query:         "如何使用 API 开发",
			expectedCount: 1,
			expected:      []string{"tech_kb"},
		},
		{
			name:          "multiple matches",
			query:         "技术开发的薪资",
			expectedCount: 2,
			// tech_kb: 技术, 开发 | hr_kb: 薪资
		},
		{
			name:          "no match",
			query:         "今天天气怎么样",
			expectedCount: 0,
			expected:      nil,
		},
		{
			name:          "exact keyword match",
			query:         "销售报价",
			expectedCount: 1,
			expected:      []string{"sales_kb"},
		},
		{
			name:          "empty query",
			query:         "",
			expectedCount: 0,
		},
		{
			name:          "case sensitive",
			query:         "TECHNOLOGY",
			expectedCount: 0, // 中文匹配，英文大小写敏感
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := router(ctx, tt.query)

			if err != nil {
				t.Errorf("KeywordRouter() unexpected error: %v", err)
			}
			if len(result) != tt.expectedCount {
				t.Errorf("KeywordRouter() returned %d retrievers, want %d", len(result), tt.expectedCount)
			}
			if tt.expected != nil {
				for _, exp := range tt.expected {
					found := false
					for _, r := range result {
						if r == exp {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("KeywordRouter() missing expected retriever %q", exp)
					}
				}
			}
		})
	}
}

func TestKeywordRouter_EmptyRules(t *testing.T) {
	ctx := context.Background()
	router := KeywordRouter(map[string][]string{})

	result, err := router(ctx, "test query")
	if err != nil {
		t.Errorf("KeywordRouter() unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("KeywordRouter() with empty rules returned %d, want 0", len(result))
	}
}

func TestKeywordRouter_NilRules(t *testing.T) {
	ctx := context.Background()
	router := KeywordRouter(nil)

	result, err := router(ctx, "test query")
	if err != nil {
		t.Errorf("KeywordRouter() unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("KeywordRouter() with nil rules returned %d, want 0", len(result))
	}
}

// ========== AllRouter 测试 ==========

func TestAllRouter(t *testing.T) {
	ctx := context.Background()

	retrieverNames := []string{"r1", "r2", "r3"}
	router := AllRouter(retrieverNames)

	result, err := router(ctx, "test query")

	if err != nil {
		t.Errorf("AllRouter() unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Errorf("AllRouter() returned %d retrievers, want 3", len(result))
	}
	if result[0] != "r1" {
		t.Errorf("AllRouter()[0] = %q, want 'r1'", result[0])
	}
}

func TestAllRouter_Empty(t *testing.T) {
	ctx := context.Background()

	router := AllRouter([]string{})

	result, err := router(ctx, "test query")

	if err != nil {
		t.Errorf("AllRouter() unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("AllRouter() with empty names returned %d, want 0", len(result))
	}
}

// ========== PriorityRouter 测试 ==========

func TestPriorityRouter(t *testing.T) {
	ctx := context.Background()

	priority := []string{"r1", "r2", "r3"}
	router := PriorityRouter(priority)

	result, err := router(ctx, "test query")

	if err != nil {
		t.Errorf("PriorityRouter() unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Errorf("PriorityRouter() returned %d retrievers, want 3", len(result))
	}
	// 应该保持优先级顺序
	if result[0] != "r1" {
		t.Errorf("PriorityRouter()[0] = %q, want 'r1'", result[0])
	}
	if result[1] != "r2" {
		t.Errorf("PriorityRouter()[1] = %q, want 'r2'", result[1])
	}
}

func TestPriorityRouter_Single(t *testing.T) {
	ctx := context.Background()

	router := PriorityRouter([]string{"only"})

	result, err := router(ctx, "test query")

	if err != nil {
		t.Errorf("PriorityRouter() unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("PriorityRouter() returned %d retrievers, want 1", len(result))
	}
}

// ========== WeightedFusion 测试 ==========

func TestWeightedFusion(t *testing.T) {
	ctx := context.Background()

	weights := map[string]float64{
		"r1": 2.0, // 高权重
		"r2": 1.0, // 默认权重
		"r3": 0.5, // 低权重
	}

	fusion := WeightedFusion(weights)

	docs := map[string][]*schema.Document{
		"r1": {newDoc("doc1", 0.8), newDoc("doc2", 0.6)},
		"r2": {newDoc("doc1", 0.4)}, // doc1 重复，权重不同
		"r3": {newDoc("doc3", 0.9)},
	}

	results, err := fusion(ctx, docs)

	if err != nil {
		t.Errorf("WeightedFusion() unexpected error: %v", err)
	}

	// doc1 在 r1 和 r2 中，权重累积
	// r1: 0.8 * 2.0 = 1.6, r2: 0.4 * 1.0 = 0.4, total = 2.0
	// doc2: 0.6 * 2.0 = 1.2
	// doc3: 0.9 * 0.5 = 0.45

	// 排序应该是：doc1 (2.0) > doc2 (1.2) > doc3 (0.45)
	if len(results) != 3 {
		t.Errorf("WeightedFusion() returned %d docs, want 3", len(results))
	}
	if results[0].ID != "doc1" {
		t.Errorf("results[0].ID = %q, want 'doc1'", results[0].ID)
	}
}

func TestWeightedFusion_Empty(t *testing.T) {
	ctx := context.Background()

	fusion := WeightedFusion(map[string]float64{})

	results, err := fusion(ctx, map[string][]*schema.Document{})

	if err != nil {
		t.Errorf("WeightedFusion() unexpected error: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("WeightedFusion() returned %d docs, want 0", len(results))
	}
}

func TestWeightedFusion_DefaultWeight(t *testing.T) {
	ctx := context.Background()

	fusion := WeightedFusion(map[string]float64{})

	docs := map[string][]*schema.Document{
		"r1": {newDoc("doc1", 0.8)},
		"r2": {newDoc("doc2", 0.9)},
	}

	results, err := fusion(ctx, docs)

	if err != nil {
		t.Errorf("WeightedFusion() unexpected error: %v", err)
	}

	// 没有指定权重时使用默认权重 1.0
	if len(results) != 2 {
		t.Errorf("WeightedFusion() returned %d docs, want 2", len(results))
	}
}

// ========== sortByScore 测试 ==========

func TestSortByScore(t *testing.T) {
	docs := []*schema.Document{
		newDoc("doc3", 0.5),
		newDoc("doc1", 0.9),
		newDoc("doc2", 0.7),
	}

	sortByScore(docs)

	// 降序排序
	if docs[0].ID != "doc1" {
		t.Errorf("docs[0].ID = %q, want 'doc1'", docs[0].ID)
	}
	if docs[1].ID != "doc2" {
		t.Errorf("docs[1].ID = %q, want 'doc2'", docs[1].ID)
	}
	if docs[2].ID != "doc3" {
		t.Errorf("docs[2].ID = %q, want 'doc3'", docs[2].ID)
	}
}

func TestSortByScore_Empty(t *testing.T) {
	docs := []*schema.Document{}

	sortByScore(docs)

	// 不应该 panic
	if len(docs) != 0 {
		t.Errorf("sortByScore() modified empty slice")
	}
}

func TestSortByScore_Single(t *testing.T) {
	docs := []*schema.Document{newDoc("doc1", 0.5)}

	sortByScore(docs)

	if len(docs) != 1 {
		t.Errorf("sortByScore() returned %d docs, want 1", len(docs))
	}
}

func TestSortByScore_SameScores(t *testing.T) {
	docs := []*schema.Document{
		newDoc("doc1", 0.5),
		newDoc("doc2", 0.5),
		newDoc("doc3", 0.5),
	}

	sortByScore(docs)

	// 分数相同，顺序可能改变
	if len(docs) != 3 {
		t.Errorf("sortByScore() returned %d docs, want 3", len(docs))
	}
}
