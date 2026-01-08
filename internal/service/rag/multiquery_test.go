// Package rag 提供 MultiQuery 功能单元测试
package rag

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// ========== Mock ChatModel ==========

type mockChatModel struct {
	responses []string
	err       error
	callCount int
}

func newMockChatModel(responses []string, err error) *mockChatModel {
	return &mockChatModel{responses: responses, err: err}
}

func (m *mockChatModel) Generate(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	m.callCount++
	if m.err != nil {
		return nil, m.err
	}
	if len(m.responses) == 0 {
		return &schema.Message{Role: schema.Assistant, Content: "default response"}, nil
	}
	idx := (m.callCount - 1) % len(m.responses)
	return &schema.Message{Role: schema.Assistant, Content: m.responses[idx]}, nil
}

func (m *mockChatModel) Stream(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	return nil, nil
}

func (m *mockChatModel) BindTools(tools []*schema.ToolInfo) error {
	return nil
}

// ========== NewMultiQueryRetriever 测试 ==========

func TestNewMultiQueryRetriever(t *testing.T) {
	ctx := context.Background()
	baseDoc := newDoc("test content", 0)

	tests := []struct {
		name        string
		cfg         *MultiQueryConfig
		wantErr     bool
		errContains string
	}{
		{
			name: "valid with RewriteLLM",
			cfg: &MultiQueryConfig{
				Retriever:  newMockRetriever([]*schema.Document{baseDoc}, nil),
				RewriteLLM: newMockChatModel([]string{"query1", "query2"}, nil),
			},
			wantErr: false,
		},
		{
			name: "valid with RewriteHandler",
			cfg: &MultiQueryConfig{
				Retriever: newMockRetriever([]*schema.Document{baseDoc}, nil),
				RewriteHandler: func(ctx context.Context, query string) ([]string, error) {
					return []string{query, strings.ToLower(query)}, nil
				},
			},
			wantErr: false,
		},
		{
			name: "nil retriever",
			cfg: &MultiQueryConfig{
				Retriever: nil,
				RewriteHandler: func(ctx context.Context, query string) ([]string, error) {
					return []string{query}, nil
				},
			},
			wantErr:     true,
			errContains: "Retriever is required",
		},
		{
			name: "both nil",
			cfg: &MultiQueryConfig{
				Retriever:     newMockRetriever([]*schema.Document{baseDoc}, nil),
				RewriteLLM:    nil,
				RewriteHandler: nil,
			},
			wantErr:     true,
			errContains: "at least one",
		},
		{
			name: "with MaxQueriesNum",
			cfg: &MultiQueryConfig{
				Retriever:     newMockRetriever([]*schema.Document{baseDoc}, nil),
				RewriteLLM:    newMockChatModel([]string{"q1", "q2", "q3"}, nil),
				MaxQueriesNum: 2,
			},
			wantErr: false,
		},
		{
			name: "with FusionFunc",
			cfg: &MultiQueryConfig{
				Retriever: newMockRetriever([]*schema.Document{baseDoc}, nil),
				RewriteHandler: func(ctx context.Context, query string) ([]string, error) {
					return []string{query}, nil
				},
				FusionFunc: func(ctx context.Context, docs [][]*schema.Document) ([]*schema.Document, error) {
					return docs[0], nil
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := NewMultiQueryRetriever(ctx, tt.cfg)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewMultiQueryRetriever() expected error, got nil")
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("Error = %v, want contain %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("NewMultiQueryRetriever() unexpected error: %v", err)
			}
			if r == nil {
				t.Error("NewMultiQueryRetriever() returned nil retriever")
			}
		})
	}
}

// ========== SimpleSplitter 测试 ==========

func TestSimpleSplitter(t *testing.T) {
	ctx := context.Background()

	splitter := SimpleSplitter()

	tests := []struct {
		name     string
		query    string
		minCount int
	}{
		{
			name:     "single word",
			query:    "hello",
			minCount: 1,
		},
		{
			name:     "multiple words",
			query:    "hello world test",
			minCount: 2,
		},
		{
			name:     "empty query",
			query:    "",
			minCount: 1,
		},
		{
			name:     "multiple spaces",
			query:    "hello    world",
			minCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := splitter(ctx, tt.query)

			if err != nil {
				t.Errorf("SimpleSplitter() unexpected error: %v", err)
			}
			if len(result) < tt.minCount {
				t.Errorf("SimpleSplitter() returned %d queries, want at least %d", len(result), tt.minCount)
			}

			// 第一个结果应该是原查询
			if result[0] != tt.query {
				t.Errorf("SimpleSplitter()[0] = %q, want %q", result[0], tt.query)
			}
		})
	}
}

func TestSimpleSplitter_PreservesOriginal(t *testing.T) {
	ctx := context.Background()
	splitter := SimpleSplitter()

	query := "machine learning algorithms"
	result, _ := splitter(ctx, query)

	// 第一个应该是原查询
	if result[0] != query {
		t.Errorf("First query = %q, want %q", result[0], query)
	}
}

// ========== LowercaseVariants 测试 ==========

func TestLowercaseVariants(t *testing.T) {
	ctx := context.Background()

	generator := LowercaseVariants()

	tests := []struct {
		name          string
		query         string
		expectedCount int
		hasLowercase  bool
	}{
		{
			name:          "uppercase query",
			query:         "HELLO WORLD",
			expectedCount: 2,
			hasLowercase:  true,
		},
		{
			name:          "mixed case query",
			query:         "Hello World",
			expectedCount: 2,
			hasLowercase:  true,
		},
		{
			name:          "already lowercase",
			query:         "hello world",
			expectedCount: 1,
			hasLowercase:  false,
		},
		{
			name:          "empty query",
			query:         "",
			expectedCount: 1,
			hasLowercase:  false,
		},
		{
			name:          "numbers and symbols",
			query:         "Test123_API",
			expectedCount: 2,
			hasLowercase:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := generator(ctx, tt.query)

			if err != nil {
				t.Errorf("LowercaseVariants() unexpected error: %v", err)
			}
			if len(result) != tt.expectedCount {
				t.Errorf("LowercaseVariants() returned %d queries, want %d", len(result), tt.expectedCount)
			}

			// 第一个应该是原查询
			if result[0] != tt.query {
				t.Errorf("LowercaseVariants()[0] = %q, want %q", result[0], tt.query)
			}

			// 检查小写版本
			if tt.hasLowercase {
				hasLower := false
				for _, q := range result {
					if q == strings.ToLower(tt.query) {
						hasLower = true
						break
					}
				}
				if !hasLower {
					t.Error("LowercaseVariants() missing lowercase version")
				}
			}
		})
	}
}

func TestLowercaseVariants_PreservesOriginal(t *testing.T) {
	ctx := context.Background()
	generator := LowercaseVariants()

	query := "Machine Learning"
	result, _ := generator(ctx, query)

	// 第一个应该是原查询
	if result[0] != query {
		t.Errorf("First query = %q, want %q", result[0], query)
	}

	// 第二个应该是小写版本
	if len(result) == 2 && result[1] != strings.ToLower(query) {
		t.Errorf("Second query = %q, want %q", result[1], strings.ToLower(query))
	}
}

// ========== 错误处理测试 ==========

func TestRewriteHandler_Error(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("handler error")

	cfg := &MultiQueryConfig{
		Retriever: newMockRetriever([]*schema.Document{}, nil),
		RewriteHandler: func(ctx context.Context, query string) ([]string, error) {
			return nil, expectedErr
		},
	}

	r, err := NewMultiQueryRetriever(ctx, cfg)
	if err != nil {
		t.Fatalf("NewMultiQueryRetriever() unexpected error: %v", err)
	}

	// Retrieve 应该返回错误
	_, err = r.Retrieve(ctx, "test query")
	if err == nil {
		t.Error("Retrieve() expected error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("Retrieve() error = %v, want %v", err, expectedErr)
	}
}

// ========== PrefixSuffixAdder 测试 ==========

func TestPrefixSuffixAdder(t *testing.T) {
	ctx := context.Background()

	prefixes := []string{"please explain", "describe"}
	suffixes := []string{"in detail", "with examples"}

	adder := PrefixSuffixAdder(prefixes, suffixes)

	queries, err := adder(ctx, "machine learning")

	if err != nil {
		t.Errorf("PrefixSuffixAdder() unexpected error: %v", err)
	}

	// 应该返回：原查询 + 前缀变体 + 后缀变体
	// 1 + 2 + 2 = 5 个查询
	if len(queries) != 5 {
		t.Errorf("PrefixSuffixAdder() returned %d queries, want 5", len(queries))
	}

	// 第一个应该是原查询
	if queries[0] != "machine learning" {
		t.Errorf("queries[0] = %q, want 'machine learning'", queries[0])
	}
}

func TestPrefixSuffixAdder_Empty(t *testing.T) {
	ctx := context.Background()

	adder := PrefixSuffixAdder([]string{}, []string{})

	queries, err := adder(ctx, "test query")

	if err != nil {
		t.Errorf("PrefixSuffixAdder() unexpected error: %v", err)
	}

	// 没有前缀后缀，只返回原查询
	if len(queries) != 1 {
		t.Errorf("PrefixSuffixAdder() returned %d queries, want 1", len(queries))
	}
	if queries[0] != "test query" {
		t.Errorf("queries[0] = %q, want 'test query'", queries[0])
	}
}

func TestPrefixSuffixAdder_OnlyPrefixes(t *testing.T) {
	ctx := context.Background()

	adder := PrefixSuffixAdder([]string{"explain"}, []string{})

	queries, err := adder(ctx, "test")

	if err != nil {
		t.Errorf("PrefixSuffixAdder() unexpected error: %v", err)
	}

	// 原查询 + 前缀变体
	if len(queries) != 2 {
		t.Errorf("PrefixSuffixAdder() returned %d queries, want 2", len(queries))
	}
}

func TestPrefixSuffixAdder_OnlySuffixes(t *testing.T) {
	ctx := context.Background()

	adder := PrefixSuffixAdder([]string{}, []string{"detail"})

	queries, err := adder(ctx, "test")

	if err != nil {
		t.Errorf("PrefixSuffixAdder() unexpected error: %v", err)
	}

	// 原查询 + 后缀变体
	if len(queries) != 2 {
		t.Errorf("PrefixSuffixAdder() returned %d queries, want 2", len(queries))
	}
}

// ========== MultiQueryWeightedFusion 测试 ==========

func TestMultiQueryWeightedFusion(t *testing.T) {
	ctx := context.Background()

	fusion := MultiQueryWeightedFusion(nil)

	docs1 := []*schema.Document{
		newDoc("doc1", 0.8),
		newDoc("doc2", 0.7),
	}
	docs2 := []*schema.Document{
		newDoc("doc1", 0.9), // doc1 重复出现
		newDoc("doc3", 0.6),
	}

	results, err := fusion(ctx, [][]*schema.Document{docs1, docs2})

	if err != nil {
		t.Errorf("MultiQueryWeightedFusion() unexpected error: %v", err)
	}

	// doc1 出现两次，应该排在前面
	if len(results) != 3 {
		t.Errorf("MultiQueryWeightedFusion() returned %d docs, want 3", len(results))
	}

	// doc1 应该排在第一位（出现次数最多）
	if results[0].ID != "doc1" {
		t.Errorf("results[0].ID = %q, want 'doc1'", results[0].ID)
	}
}

func TestMultiQueryWeightedFusion_Empty(t *testing.T) {
	ctx := context.Background()

	fusion := MultiQueryWeightedFusion(nil)

	results, err := fusion(ctx, [][]*schema.Document{})

	if err != nil {
		t.Errorf("MultiQueryWeightedFusion() unexpected error: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("MultiQueryWeightedFusion() returned %d docs, want 0", len(results))
	}
}

// ========== RoundRobinFusion 测试 ==========

func TestRoundRobinFusion(t *testing.T) {
	ctx := context.Background()

	fusion := RoundRobinFusion()

	docs1 := []*schema.Document{
		newDoc("doc1", 0.8),
		newDoc("doc2", 0.7),
	}
	docs2 := []*schema.Document{
		newDoc("doc3", 0.9),
		newDoc("doc4", 0.6),
	}
	docs3 := []*schema.Document{
		newDoc("doc5", 0.5),
	}

	results, err := fusion(ctx, [][]*schema.Document{docs1, docs2, docs3})

	if err != nil {
		t.Errorf("RoundRobinFusion() unexpected error: %v", err)
	}

	// 轮询顺序：doc1, doc3, doc5, doc2, doc4
	if len(results) != 5 {
		t.Errorf("RoundRobinFusion() returned %d docs, want 5", len(results))
	}

	// 检查轮询顺序
	if results[0].ID != "doc1" {
		t.Errorf("results[0].ID = %q, want 'doc1'", results[0].ID)
	}
	if results[1].ID != "doc3" {
		t.Errorf("results[1].ID = %q, want 'doc3'", results[1].ID)
	}
	if results[2].ID != "doc5" {
		t.Errorf("results[2].ID = %q, want 'doc5'", results[2].ID)
	}
}

func TestRoundRobinFusion_Dedup(t *testing.T) {
	ctx := context.Background()

	fusion := RoundRobinFusion()

	docs1 := []*schema.Document{
		newDoc("doc1", 0.8),
		newDoc("doc2", 0.7),
	}
	docs2 := []*schema.Document{
		newDoc("doc1", 0.9), // 重复
		newDoc("doc3", 0.6),
	}

	results, err := fusion(ctx, [][]*schema.Document{docs1, docs2})

	if err != nil {
		t.Errorf("RoundRobinFusion() unexpected error: %v", err)
	}

	// doc1 只应该出现一次
	doc1Count := 0
	for _, doc := range results {
		if doc.ID == "doc1" {
			doc1Count++
		}
	}

	if doc1Count != 1 {
		t.Errorf("doc1 appears %d times, want 1", doc1Count)
	}
}

// ========== ConcatFusion 测试 ==========

func TestConcatFusion(t *testing.T) {
	ctx := context.Background()

	fusion := ConcatFusion()

	docs1 := []*schema.Document{
		newDoc("doc1", 0.8),
		newDoc("doc2", 0.7),
	}
	docs2 := []*schema.Document{
		newDoc("doc3", 0.9),
		newDoc("doc1", 0.6), // 重复
	}

	results, err := fusion(ctx, [][]*schema.Document{docs1, docs2})

	if err != nil {
		t.Errorf("ConcatFusion() unexpected error: %v", err)
	}

	// 拼接并去重：doc1, doc2, doc3
	if len(results) != 3 {
		t.Errorf("ConcatFusion() returned %d docs, want 3", len(results))
	}

	// 保持顺序：先 docs1 的，再 docs2 的
	if results[0].ID != "doc1" {
		t.Errorf("results[0].ID = %q, want 'doc1'", results[0].ID)
	}
	if results[1].ID != "doc2" {
		t.Errorf("results[1].ID = %q, want 'doc2'", results[1].ID)
	}
	if results[2].ID != "doc3" {
		t.Errorf("results[2].ID = %q, want 'doc3'", results[2].ID)
	}
}

func TestConcatFusion_Empty(t *testing.T) {
	ctx := context.Background()

	fusion := ConcatFusion()

	results, err := fusion(ctx, [][]*schema.Document{})

	if err != nil {
		t.Errorf("ConcatFusion() unexpected error: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("ConcatFusion() returned %d docs, want 0", len(results))
	}
}

func TestConcatFusion_AllDuplicates(t *testing.T) {
	ctx := context.Background()

	fusion := ConcatFusion()

	docs1 := []*schema.Document{newDoc("doc1", 0.8)}
	docs2 := []*schema.Document{newDoc("doc1", 0.9)} // 相同 ID

	results, err := fusion(ctx, [][]*schema.Document{docs1, docs2})

	if err != nil {
		t.Errorf("ConcatFusion() unexpected error: %v", err)
	}

	// 只保留一个
	if len(results) != 1 {
		t.Errorf("ConcatFusion() returned %d docs, want 1", len(results))
	}
}
