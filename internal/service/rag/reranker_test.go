// Package rag 提供 Reranker 功能单元测试
package rag

import (
	"context"
	"testing"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// ========== scoreReranker 测试 ==========

func TestScoreReranker_Rerank(t *testing.T) {
	ctx := context.Background()
	r := &scoreReranker{}

	tests := []struct {
		name     string
		query    string
		docs     []*schema.Document
		wantLen  int
		firstID  string // 排序后第一个文档的 ID
	}{
		{
			name:  "sort by score descending",
			query: "test query",
			docs: []*schema.Document{
				newDoc("doc1", 0.5),
				newDoc("doc2", 0.9),
				newDoc("doc3", 0.7),
			},
			wantLen: 3,
			firstID: "doc2", // 最高分
		},
		{
			name:  "single doc",
			query: "test",
			docs: []*schema.Document{
				newDoc("doc1", 0.5),
			},
			wantLen: 1,
			firstID: "doc1",
		},
		{
			name:  "empty docs",
			query: "test",
			docs:  []*schema.Document{},
			wantLen: 0,
		},
		{
			name:  "zero scores",
			query: "test",
			docs: []*schema.Document{
				newDoc("doc1", 0),
				newDoc("doc2", 0),
			},
			wantLen: 2,
			firstID: "", // 不检查顺序，因为分数相同
		},
		{
			name:  "negative scores",
			query: "test",
			docs: []*schema.Document{
				newDoc("doc1", -0.5),
				newDoc("doc2", 0.3),
			},
			wantLen: 2,
			firstID: "doc2", // 0.3 > -0.5
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := r.Rerank(ctx, tt.query, tt.docs)

			if err != nil {
				t.Errorf("Rerank() unexpected error: %v", err)
			}
			if len(result) != tt.wantLen {
				t.Errorf("Rerank() returned %d docs, want %d", len(result), tt.wantLen)
			}
			if tt.firstID != "" && tt.wantLen > 0 && result[0].ID != tt.firstID {
				t.Errorf("Rerank()[0].ID = %q, want %q", result[0].ID, tt.firstID)
			}
		})
	}
}

func TestScoreReranker_Rerank_PreservesOriginal(t *testing.T) {
	ctx := context.Background()
	r := &scoreReranker{}

	originalDocs := []*schema.Document{
		newDoc("doc1", 0.5),
		newDoc("doc2", 0.9),
	}

	// 保存原始分数
	originalScores := map[string]float64{
		originalDocs[0].ID: originalDocs[0].Score(),
		originalDocs[1].ID: originalDocs[1].Score(),
	}

	result, _ := r.Rerank(ctx, "test", originalDocs)

	// 验证原始文档没被修改
	if originalDocs[0].Score() != originalScores["doc1"] {
		t.Error("Original docs were modified")
	}
	if originalDocs[1].Score() != originalScores["doc2"] {
		t.Error("Original docs were modified")
	}

	// 结果应该是新的切片
	if &result == &originalDocs {
		t.Error("Rerank() should return new slice")
	}
}

// ========== llmRerankerWrapper 测试 ==========

func TestLLMRerankerWrapper_Rerank(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		topN        int
		docCount    int
		chatModel   model.ChatModel // nil 或 mock
		expectedLen int
	}{
		{
			name:        "nil chat model returns original",
			topN:        5,
			docCount:    3,
			chatModel:   nil,
			expectedLen: 3,
		},
		{
			name:        "doc count less than topN",
			topN:        10,
			docCount:    5,
			chatModel:   &mockRerankChatModel{},
			expectedLen: 5, // 不触发重排
		},
		{
			name:        "doc count equals topN",
			topN:        5,
			docCount:    5,
			chatModel:   &mockRerankChatModel{},
			expectedLen: 5, // 不触发重排
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &llmRerankerWrapper{
				chatModel: tt.chatModel,
				topN:      tt.topN,
			}

			docs := make([]*schema.Document, tt.docCount)
			for i := 0; i < tt.docCount; i++ {
				docs[i] = newDoc(string(rune('a'+i)), 0.5)
			}

			result, err := r.Rerank(ctx, "test query", docs)

			if err != nil {
				t.Errorf("Rerank() unexpected error: %v", err)
			}
			if len(result) != tt.expectedLen {
				t.Errorf("Rerank() returned %d docs, want %d", len(result), tt.expectedLen)
			}
		})
	}
}

// ========== extractNumbersFromOutput 测试 ==========

func TestExtractNumbersFromOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []int
	}{
		{
			name:     "comma separated numbers",
			input:    "1,2,3,4,5",
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "spaces between numbers",
			input:    "1 2 3 4 5",
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "mixed separators",
			input:    "1, 2, 3, 4, 5",
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "with text",
			input:    "排序结果：1,3,2,4,5",
			expected: []int{1, 3, 2, 4, 5},
		},
		{
			name:     "single number",
			input:    "42",
			expected: []int{42},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []int{},
		},
		{
			name:     "no numbers",
			input:    "hello world",
			expected: []int{},
		},
		{
			name:     "number at end",
			input:    "text 123",
			expected: []int{123},
		},
		{
			name:     "multiple zeros",
			input:    "0,0,0",
			expected: []int{},
		},
		{
			name:     "zero and positive",
			input:    "0,1,2",
			expected: []int{1, 2}, // 0 被过滤
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractNumbersFromOutput(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("extractNumbersFromOutput() = %v, want %v", result, tt.expected)
			}
			for i, exp := range tt.expected {
				if i >= len(result) || result[i] != exp {
					t.Errorf("extractNumbersFromOutput()[%d] = %d, want %d", i, result[i], exp)
				}
			}
		})
	}
}

// ========== minInt 测试 ==========

func TestMinInt(t *testing.T) {
	tests := []struct {
		name     string
		a        int
		b        int
		expected int
	}{
		{
			name:     "a less than b",
			a:        3,
			b:        5,
			expected: 3,
		},
		{
			name:     "a greater than b",
			a:        7,
			b:        4,
			expected: 4,
		},
		{
			name:     "equal",
			a:        5,
			b:        5,
			expected: 5,
		},
		{
			name:     "negative a",
			a:        -3,
			b:        5,
			expected: -3,
		},
		{
			name:     "negative b",
			a:        3,
			b:        -5,
			expected: -5,
		},
		{
			name:     "both negative",
			a:        -3,
			b:        -7,
			expected: -7,
		},
		{
			name:     "zero",
			a:        0,
			b:        5,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := minInt(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("minInt(%d, %d) = %d, want %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// ========== Mock ChatModel for Reranker ==========

type mockRerankChatModel struct {
	response string
	err      error
}

func (m *mockRerankChatModel) Generate(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.response == "" {
		m.response = "1,2,3,4,5"
	}
	return &schema.Message{
		Role:    schema.Assistant,
		Content: m.response,
	}, nil
}

func (m *mockRerankChatModel) Stream(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	return nil, nil
}

func (m *mockRerankChatModel) BindTools(tools []*schema.ToolInfo) error {
	return nil
}
