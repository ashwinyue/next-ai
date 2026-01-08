// Package rag 提供 RAG Graph 功能单元测试
package rag

import (
	"context"
	"errors"
	"testing"

	"github.com/ashwinyue/next-ai/internal/service/rerank"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// ========== mockChatModel ==========

type mockGraphChatModel struct{}

func newMockGraphChatModel() model.ChatModel {
	return &mockGraphChatModel{}
}

func (m *mockGraphChatModel) Generate(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	return &schema.Message{Role: schema.Assistant, Content: "rewritten query"}, nil
}

func (m *mockGraphChatModel) Stream(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	return nil, nil
}

func (m *mockGraphChatModel) BindTools(tools []*schema.ToolInfo) error {
	return nil
}

// ========== State 测试 ==========

func TestState_ToContext(t *testing.T) {
	tests := []struct {
		name     string
		state    *State
		contains []string
	}{
		{
			name:     "empty docs",
			state:    &State{RerankedDocs: []*schema.Document{}},
			contains: []string{"未找到相关文档"},
		},
		{
			name: "nil docs",
			state: &State{
				RerankedDocs: nil,
			},
			contains: []string{"未找到相关文档"},
		},
		{
			name: "single doc with title",
			state: &State{
				RerankedDocs: []*schema.Document{
					{
						ID:      "doc1",
						Content: "This is a test document with some content",
						MetaData: map[string]any{
							"title": "Test Document",
						},
					},
				},
			},
			contains: []string{"Test Document", "test document"},
		},
		{
			name: "doc with score",
			state: &State{
				RerankedDocs: []*schema.Document{
					{
						ID:      "doc1",
						Content: "Content here",
						MetaData: map[string]any{
							"_score": 0.85,
						},
					},
				},
			},
			contains: []string{"Content here", "0.85"},
		},
		{
			name: "long content truncated",
			state: &State{
				RerankedDocs: []*schema.Document{
					{
						ID:      "doc1",
						Content: string(make([]byte, 600)), // 超过 500 字符
					},
				},
			},
			contains: []string{"..."},
		},
		{
			name: "multiple docs",
			state: &State{
				RerankedDocs: []*schema.Document{
					{ID: "doc1", Content: "Content 1"},
					{ID: "doc2", Content: "Content 2"},
				},
			},
			contains: []string{"[1]", "[2]", "Content 1", "Content 2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.state.ToContext()

			for _, substr := range tt.contains {
				if !contains(result, substr) {
					t.Errorf("ToContext() result missing %q", substr)
				}
			}
		})
	}
}

// ========== RAG New 测试 ==========

func TestNewRAG(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		cfg         *Config
		wantErr     bool
		errContains string
	}{
		{
			name: "valid basic config",
			cfg: &Config{
				ChatModel: newMockGraphChatModel(),
				Retriever: newMockRetriever([]*schema.Document{}, nil),
			},
			wantErr: false,
		},
		{
			name: "nil chat model",
			cfg: &Config{
				ChatModel: nil,
				Retriever: newMockRetriever([]*schema.Document{}, nil),
			},
			wantErr:     true,
			errContains: "chat model is required",
		},
		{
			name: "nil retriever",
			cfg: &Config{
				ChatModel: newMockGraphChatModel(),
				Retriever: nil,
			},
			wantErr:     true,
			errContains: "retriever is required",
		},
		{
			name: "with query rewrite enabled",
			cfg: &Config{
				ChatModel:   newMockGraphChatModel(),
				Retriever:   newMockRetriever([]*schema.Document{}, nil),
				EnableRewrite: true,
			},
			wantErr: false,
		},
		{
			name: "with query expand enabled",
			cfg: &Config{
				ChatModel: newMockGraphChatModel(),
				Retriever:   newMockRetriever([]*schema.Document{}, nil),
				EnableExpand: true,
				NumVariants:  3,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := New(ctx, tt.cfg)

			if tt.wantErr {
				if err == nil {
					t.Errorf("New() expected error, got nil")
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("Error = %v, want contain %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("New() unexpected error: %v", err)
			}
			if r == nil {
				t.Error("New() returned nil RAG")
			}
		})
	}
}

// ========== Builder 测试 ==========

func TestNewBuilder(t *testing.T) {
	chatModel := newMockGraphChatModel()
	retr := newMockRetriever([]*schema.Document{}, nil)

	builder := NewBuilder(chatModel, retr)

	if builder == nil {
		t.Fatal("NewBuilder() returned nil")
	}
	if builder.cfg == nil {
		t.Fatal("NewBuilder() config is nil")
	}
	if builder.cfg.ChatModel != chatModel {
		t.Error("NewBuilder() ChatModel not set")
	}
	if builder.cfg.Retriever != retr {
		t.Error("NewBuilder() Retriever not set")
	}
	if builder.cfg.NumVariants != 3 {
		t.Errorf("NumVariants = %d, want 3", builder.cfg.NumVariants)
	}
}

func TestBuilder_WithQueryRewrite(t *testing.T) {
	chatModel := newMockGraphChatModel()
	retr := newMockRetriever([]*schema.Document{}, nil)

	builder := NewBuilder(chatModel, retr).WithQueryRewrite()

	if !builder.cfg.EnableRewrite {
		t.Error("WithQueryRewrite() did not set EnableRewrite")
	}
}

func TestBuilder_WithQueryExpand(t *testing.T) {
	chatModel := newMockGraphChatModel()
	retr := newMockRetriever([]*schema.Document{}, nil)

	builder := NewBuilder(chatModel, retr).WithQueryExpand(5)

	if !builder.cfg.EnableExpand {
		t.Error("WithQueryExpand() did not set EnableExpand")
	}
	if builder.cfg.NumVariants != 5 {
		t.Errorf("NumVariants = %d, want 5", builder.cfg.NumVariants)
	}
}

func TestBuilder_WithQueryExpand_DefaultVariants(t *testing.T) {
	chatModel := newMockGraphChatModel()
	retr := newMockRetriever([]*schema.Document{}, nil)

	builder := NewBuilder(chatModel, retr).WithQueryExpand(0) // 0 应该保持默认值

	if builder.cfg.NumVariants != 3 {
		t.Errorf("NumVariants = %d, want 3 (default)", builder.cfg.NumVariants)
	}
}

func TestBuilder_Chaining(t *testing.T) {
	chatModel := newMockGraphChatModel()
	retr := newMockRetriever([]*schema.Document{}, nil)

	builder := NewBuilder(chatModel, retr).
		WithQueryRewrite().
		WithQueryExpand(5).
		WithScoreReranker()

	if !builder.cfg.EnableRewrite {
		t.Error("Chaining: EnableRewrite not set")
	}
	if !builder.cfg.EnableExpand {
		t.Error("Chaining: EnableExpand not set")
	}
	if builder.cfg.NumVariants != 5 {
		t.Error("Chaining: NumVariants not set")
	}
	if len(builder.cfg.Rerankers) == 0 {
		t.Error("Chaining: Rerankers not added")
	}
}

func TestBuilder_Build(t *testing.T) {
	ctx := context.Background()
	chatModel := newMockGraphChatModel()
	retr := newMockRetriever([]*schema.Document{}, nil)

	builder := NewBuilder(chatModel, retr)
	rag, err := builder.Build(ctx)

	if err != nil {
		t.Errorf("Build() unexpected error: %v", err)
	}
	if rag == nil {
		t.Error("Build() returned nil RAG")
	}
}

// ========== Preset 配置测试 ==========

func TestPresetBasic(t *testing.T) {
	ctx := context.Background()
	chatModel := newMockGraphChatModel()
	retr := newMockRetriever([]*schema.Document{}, nil)

	rag, err := PresetBasic(ctx, chatModel, retr)

	if err != nil {
		t.Errorf("PresetBasic() unexpected error: %v", err)
	}
	if rag == nil {
		t.Error("PresetBasic() returned nil")
	}
	if rag.config.EnableRewrite {
		t.Error("PresetBasic() should not enable rewrite")
	}
	if rag.config.EnableExpand {
		t.Error("PresetBasic() should not enable expand")
	}
}

func TestPresetAdvanced(t *testing.T) {
	ctx := context.Background()
	chatModel := newMockGraphChatModel()
	retr := newMockRetriever([]*schema.Document{}, nil)

	rag, err := PresetAdvanced(ctx, chatModel, retr)

	if err != nil {
		t.Errorf("PresetAdvanced() unexpected error: %v", err)
	}
	if rag == nil {
		t.Error("PresetAdvanced() returned nil")
	}
	if !rag.config.EnableRewrite {
		t.Error("PresetAdvanced() should enable rewrite")
	}
	if !rag.config.EnableExpand {
		t.Error("PresetAdvanced() should enable expand")
	}
	if len(rag.config.Rerankers) == 0 {
		t.Error("PresetAdvanced() should add rerankers")
	}
}

func TestPresetSearchOptimized(t *testing.T) {
	ctx := context.Background()
	chatModel := newMockGraphChatModel()
	retr := newMockRetriever([]*schema.Document{}, nil)

	rag, err := PresetSearchOptimized(ctx, chatModel, retr)

	if err != nil {
		t.Errorf("PresetSearchOptimized() unexpected error: %v", err)
	}
	if rag == nil {
		t.Error("PresetSearchOptimized() returned nil")
	}
	if !rag.config.EnableRewrite {
		t.Error("PresetSearchOptimized() should enable rewrite")
	}
	if len(rag.config.Rerankers) == 0 {
		t.Error("PresetSearchOptimized() should add rerankers")
	}
}

func TestPresetRecallOptimized(t *testing.T) {
	ctx := context.Background()
	chatModel := newMockGraphChatModel()
	retr := newMockRetriever([]*schema.Document{}, nil)

	rag, err := PresetRecallOptimized(ctx, chatModel, retr)

	if err != nil {
		t.Errorf("PresetRecallOptimized() unexpected error: %v", err)
	}
	if rag == nil {
		t.Error("PresetRecallOptimized() returned nil")
	}
	if !rag.config.EnableExpand {
		t.Error("PresetRecallOptimized() should enable expand")
	}
	if rag.config.NumVariants != 5 {
		t.Errorf("NumVariants = %d, want 5", rag.config.NumVariants)
	}
	if len(rag.config.Rerankers) == 0 {
		t.Error("PresetRecallOptimized() should add rerankers")
	}
}

// ========== RAG 方法测试 ==========

func TestRAG_GetGraph(t *testing.T) {
	ctx := context.Background()
	chatModel := newMockGraphChatModel()
	retr := newMockRetriever([]*schema.Document{}, nil)

	rag, _ := New(ctx, &Config{
		ChatModel: chatModel,
		Retriever: retr,
	})

	graph := rag.GetGraph()
	if graph == nil {
		t.Error("GetGraph() returned nil")
	}
}

func TestRAG_Invoke(t *testing.T) {
	ctx := context.Background()
	doc := newDoc("test content", 0.8)
	chatModel := newMockGraphChatModel()

	rag, _ := New(ctx, &Config{
		ChatModel: chatModel,
		Retriever: newMockRetriever([]*schema.Document{doc}, nil),
	})

	state, err := rag.Invoke(ctx, "test query")

	if err != nil {
		t.Errorf("Invoke() unexpected error: %v", err)
	}
	if state == nil {
		t.Fatal("Invoke() returned nil state")
	}
	if state.Query != "test query" {
		t.Errorf("Query = %q, want 'test query'", state.Query)
	}
}

func TestRAG_Invoke_RetrieveSuccess(t *testing.T) {
	ctx := context.Background()
	docs := []*schema.Document{
		newDoc("doc1", 0.9),
		newDoc("doc2", 0.7),
	}
	chatModel := newMockGraphChatModel()

	rag, _ := New(ctx, &Config{
		ChatModel: chatModel,
		Retriever: newMockRetriever(docs, nil),
	})

	state, err := rag.Invoke(ctx, "test query")

	if err != nil {
		t.Errorf("Invoke() unexpected error: %v", err)
	}
	if len(state.RetrievedDocs) != 2 {
		t.Errorf("RetrievedDocs count = %d, want 2", len(state.RetrievedDocs))
	}
	// 没有重排器，RerankedDocs 应该和 RetrievedDocs 相同
	if len(state.RerankedDocs) != 2 {
		t.Errorf("RerankedDocs count = %d, want 2", len(state.RerankedDocs))
	}
}

func TestRAG_Retrieve(t *testing.T) {
	ctx := context.Background()
	docs := []*schema.Document{
		newDoc("doc1", 0.9),
	}
	chatModel := newMockGraphChatModel()

	rag, _ := New(ctx, &Config{
		ChatModel: chatModel,
		Retriever: newMockRetriever(docs, nil),
	})

	result, err := rag.Retrieve(ctx, "test query")

	if err != nil {
		t.Errorf("Retrieve() unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("Retrieve() returned %d docs, want 1", len(result))
	}
	if result[0].ID != "doc1" {
		t.Errorf("result[0].ID = %q, want 'doc1'", result[0].ID)
	}
}

func TestRAG_Retrieve_EmptyResult(t *testing.T) {
	ctx := context.Background()
	chatModel := newMockGraphChatModel()

	rag, _ := New(ctx, &Config{
		ChatModel: chatModel,
		Retriever: newMockRetriever([]*schema.Document{}, nil),
	})

	result, err := rag.Retrieve(ctx, "test query")

	if err != nil {
		t.Errorf("Retrieve() unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("Retrieve() returned %d docs, want 0", len(result))
	}
}

func TestRAG_Retrieve_RetrieverError(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("retriever error")
	chatModel := newMockGraphChatModel()

	rag, _ := New(ctx, &Config{
		ChatModel: chatModel,
		Retriever: newMockRetriever([]*schema.Document{}, expectedErr),
	})

	// retriever 错误应该在 processRetrieve 中被忽略，返回空结果
	result, err := rag.Retrieve(ctx, "test query")

	if err != nil {
		t.Errorf("Retrieve() should not error on retriever failure, got: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("Retrieve() should return empty on retriever error, got %d docs", len(result))
	}
}

func TestRAG_StreamRetrieve(t *testing.T) {
	ctx := context.Background()
	docs := []*schema.Document{
		newDoc("doc1", 0.9),
	}
	chatModel := newMockGraphChatModel()

	rag, _ := New(ctx, &Config{
		ChatModel: chatModel,
		Retriever: newMockRetriever(docs, nil),
	})

	state, err := rag.StreamRetrieve(ctx, "test query")

	if err != nil {
		t.Errorf("StreamRetrieve() unexpected error: %v", err)
	}
	if state == nil {
		t.Fatal("StreamRetrieve() returned nil state")
	}
	if len(state.RerankedDocs) != 1 {
		t.Errorf("RerankedDocs count = %d, want 1", len(state.RerankedDocs))
	}
}

// ========== Config 测试 ==========

func TestConfig_DefaultValues(t *testing.T) {
	cfg := &Config{
		ChatModel: newMockGraphChatModel(),
		Retriever: newMockRetriever([]*schema.Document{}, nil),
	}

	if cfg.EnableRewrite {
		t.Error("EnableRewrite should be false by default")
	}
	if cfg.EnableExpand {
		t.Error("EnableExpand should be false by default")
	}
	if cfg.NumVariants != 0 {
		// 0 表示使用默认值
		t.Errorf("NumVariants = %d, want 0 (default)", cfg.NumVariants)
	}
	if len(cfg.Rerankers) != 0 {
		t.Error("Rerankers should be empty by default")
	}
}

// ========== Builder Reranker 方法测试 ==========

func TestBuilder_WithReranker(t *testing.T) {
	chatModel := newMockGraphChatModel()
	retr := newMockRetriever([]*schema.Document{}, nil)

	builder := NewBuilder(chatModel, retr).WithReranker(&scoreReranker{})

	if len(builder.cfg.Rerankers) != 1 {
		t.Errorf("WithReranker() rerankers count = %d, want 1", len(builder.cfg.Rerankers))
	}
}

func TestBuilder_WithScoreReranker(t *testing.T) {
	chatModel := newMockGraphChatModel()
	retr := newMockRetriever([]*schema.Document{}, nil)

	builder := NewBuilder(chatModel, retr).WithScoreReranker()

	if len(builder.cfg.Rerankers) != 1 {
		t.Errorf("WithScoreReranker() rerankers count = %d, want 1", len(builder.cfg.Rerankers))
	}
}

func TestBuilder_WithDiversityReranker(t *testing.T) {
	chatModel := newMockGraphChatModel()
	retr := newMockRetriever([]*schema.Document{}, nil)

	lambda := 0.5
	topN := 5
	builder := NewBuilder(chatModel, retr).WithDiversityReranker(lambda, topN)

	if len(builder.cfg.Rerankers) != 1 {
		t.Errorf("WithDiversityReranker() rerankers count = %d, want 1", len(builder.cfg.Rerankers))
	}
}

func TestBuilder_WithLLMReranker(t *testing.T) {
	chatModel := newMockGraphChatModel()
	retr := newMockRetriever([]*schema.Document{}, nil)

	builder := NewBuilder(chatModel, retr).WithLLMReranker(3)

	if len(builder.cfg.Rerankers) != 1 {
		t.Errorf("WithLLMReranker() rerankers count = %d, want 1", len(builder.cfg.Rerankers))
	}
}

func TestBuilder_MultipleRerankers(t *testing.T) {
	chatModel := newMockGraphChatModel()
	retr := newMockRetriever([]*schema.Document{}, nil)

	builder := NewBuilder(chatModel, retr).
		WithScoreReranker().
		WithDiversityReranker(0.5, 5).
		WithLLMReranker(3)

	if len(builder.cfg.Rerankers) != 3 {
		t.Errorf("Multiple rerankers count = %d, want 3", len(builder.cfg.Rerankers))
	}
}

// ========== State 测试 ==========

func TestState_Struct(t *testing.T) {
	state := &State{
		Query: "test query",
		RetrievedDocs: []*schema.Document{
			newDoc("doc1", 0.8),
		},
		RerankedDocs: []*schema.Document{
			newDoc("doc1", 0.9),
		},
		Error: errors.New("test error"),
	}

	if state.Query != "test query" {
		t.Errorf("Query = %q, want 'test query'", state.Query)
	}
	if state.Error == nil {
		t.Error("Error should not be nil")
	}
	if len(state.RetrievedDocs) != 1 {
		t.Errorf("RetrievedDocs count = %d, want 1", len(state.RetrievedDocs))
	}
}

// ========== processQuery 测试 ==========

func TestRAG_processQuery_WithOptimization(t *testing.T) {
	ctx := context.Background()
	chatModel := newMockGraphChatModel()
	retr := newMockRetriever([]*schema.Document{}, nil)

	// 启用查询重写，会调用 processQuery
	rag, _ := New(ctx, &Config{
		ChatModel:    chatModel,
		Retriever:    retr,
		EnableRewrite: true,
	})

	state, err := rag.Invoke(ctx, "test query")

	if err != nil {
		t.Errorf("Invoke() unexpected error: %v", err)
	}
	if state.OptimizedQuery == nil {
		t.Error("OptimizedQuery should be set when EnableRewrite is true")
	}
}

func TestRAG_processQuery_ErrorHandling(t *testing.T) {
	ctx := context.Background()
	// 使用会返回错误的 mock chat model
	errorChatModel := &errorGraphChatModel{}
	retr := newMockRetriever([]*schema.Document{}, nil)

	// 启用查询重写
	rag, _ := New(ctx, &Config{
		ChatModel:    errorChatModel,
		Retriever:    retr,
		EnableRewrite: true,
	})

	state, err := rag.Invoke(ctx, "test query")

	if err != nil {
		t.Errorf("Invoke() should not error even with query optimization failure, got: %v", err)
	}
	// 失败时应该使用原查询
	if state.OptimizedQuery == nil {
		t.Error("OptimizedQuery should have fallback value")
	}
}

// ========== processRerank 测试 ==========

func TestRAG_processRerank_WithReranker(t *testing.T) {
	ctx := context.Background()
	chatModel := newMockGraphChatModel()
	docs := []*schema.Document{
		newDoc("doc1", 0.5),
		newDoc("doc2", 0.9),
	}

	// 使用自定义重排器
	mockReranker := &mockReranker{
		result: docs,
	}

	rag, _ := New(ctx, &Config{
		ChatModel:  chatModel,
		Retriever:  newMockRetriever(docs, nil),
		Rerankers: []rerank.Reranker{mockReranker},
	})

	state, err := rag.Invoke(ctx, "test query")

	if err != nil {
		t.Errorf("Invoke() unexpected error: %v", err)
	}
	// 应该调用了重排器
	if len(state.RerankedDocs) != 2 {
		t.Errorf("RerankedDocs count = %d, want 2", len(state.RerankedDocs))
	}
}

// ========== Error Cases ==========

func TestNewRAG_QueryComponents(t *testing.T) {
	ctx := context.Background()
	chatModel := newMockGraphChatModel()
	retr := newMockRetriever([]*schema.Document{}, nil)

	// 启用查询重写和扩展
	rag, err := New(ctx, &Config{
		ChatModel:    chatModel,
		Retriever:    retr,
		EnableRewrite: true,
		EnableExpand:  true,
		NumVariants:   3,
	})

	if err != nil {
		t.Errorf("New() unexpected error: %v", err)
	}
	if rag.rewriter == nil {
		t.Error("New() should create rewriter when EnableRewrite is true")
	}
	if rag.expander == nil {
		t.Error("New() should create expander when EnableExpand is true")
	}
}

// ========== Mock Types ==========

type errorGraphChatModel struct{}

func (m *errorGraphChatModel) Generate(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	return nil, errors.New("generate error")
}

func (m *errorGraphChatModel) Stream(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	return nil, nil
}

func (m *errorGraphChatModel) BindTools(tools []*schema.ToolInfo) error {
	return nil
}

type mockReranker struct {
	result []*schema.Document
	err    error
	callCount int
}

func (m *mockReranker) Rerank(ctx context.Context, query string, docs []*schema.Document) ([]*schema.Document, error) {
	m.callCount++
	if m.err != nil {
		return nil, m.err
	}
	if m.result != nil {
		return m.result, nil
	}
	return docs, nil
}
