// Package rag 提供 RAG Service 功能单元测试
package rag

import (
	"context"
	"errors"
	"testing"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/ashwinyue/next-ai/internal/service/types"
)

// ========== mockServiceChatModel ==========

type mockServiceChatModel struct{}

func newMockServiceChatModel() model.ChatModel {
	return &mockServiceChatModel{}
}

func (m *mockServiceChatModel) Generate(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	return &schema.Message{Role: schema.Assistant, Content: "response"}, nil
}

func (m *mockServiceChatModel) Stream(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	return nil, nil
}

func (m *mockServiceChatModel) BindTools(tools []*schema.ToolInfo) error {
	return nil
}

// ========== NewService 测试 ==========

func TestNewService(t *testing.T) {
	chatModel := newMockServiceChatModel()
	retr := newMockRetriever([]*schema.Document{}, nil)

	svc := NewService(chatModel, retr, nil)

	if svc == nil {
		t.Fatal("NewService() returned nil")
	}
	if svc.chatModel != chatModel {
		t.Error("NewService() chatModel not set")
	}
	if svc.baseRetriever != retr {
		t.Error("NewService() baseRetriever not set")
	}
	if svc.rerankers != nil {
		t.Error("NewService() rerankers should be nil")
	}
}

func TestNewService_WithRerankers(t *testing.T) {
	chatModel := newMockServiceChatModel()
	retr := newMockRetriever([]*schema.Document{}, nil)
	rerankers := []types.Reranker{&scoreReranker{}}

	svc := NewService(chatModel, retr, rerankers)

	if len(svc.rerankers) != 1 {
		t.Errorf("NewService() rerankers count = %d, want 1", len(svc.rerankers))
	}
}

// ========== NewServiceWithConfig 测试 ==========

func TestNewServiceWithConfig(t *testing.T) {
	ctx := context.Background()
	chatModel := newMockServiceChatModel()
	retr := newMockRetriever([]*schema.Document{}, nil)

	tests := []struct {
		name        string
		cfg         *ServiceConfig
		wantErr     bool
		errContains string
	}{
		{
			name: "basic config",
			cfg: &ServiceConfig{
				ChatModel:        chatModel,
				Retriever:        retr,
				EnableMultiQuery: false,
			},
			wantErr: false,
		},
		{
			name: "with multiquery enabled",
			cfg: &ServiceConfig{
				ChatModel:        chatModel,
				Retriever:        retr,
				EnableMultiQuery: true,
				MaxQueriesNum:    3,
			},
			wantErr: false,
		},
		{
			name: "with custom rewrite LLM",
			cfg: &ServiceConfig{
				ChatModel:        chatModel,
				Retriever:        retr,
				EnableMultiQuery: true,
				RewriteLLM:       newMockServiceChatModel(),
			},
			wantErr: false,
		},
		{
			name: "nil retriever with multiquery",
			cfg: &ServiceConfig{
				ChatModel:        chatModel,
				Retriever:        nil,
				EnableMultiQuery: true,
			},
			wantErr: true,
			errContains: "Retriever is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, err := NewServiceWithConfig(ctx, tt.cfg)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewServiceWithConfig() expected error, got nil")
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("Error = %v, want contain %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("NewServiceWithConfig() unexpected error: %v", err)
			}
			if svc == nil {
				t.Error("NewServiceWithConfig() returned nil service")
			}
		})
	}
}

func TestNewServiceWithConfig_MultiQueryCreated(t *testing.T) {
	ctx := context.Background()
	chatModel := newMockServiceChatModel()
	retr := newMockRetriever([]*schema.Document{}, nil)

	svc, err := NewServiceWithConfig(ctx, &ServiceConfig{
		ChatModel:        chatModel,
		Retriever:        retr,
		EnableMultiQuery: true,
		MaxQueriesNum:    3,
	})

	if err != nil {
		t.Fatalf("NewServiceWithConfig() unexpected error: %v", err)
	}
	if svc.multiRetriever == nil {
		t.Error("NewServiceWithConfig() multiRetriever should be created")
	}
}

// ========== Service Retrieve 测试 ==========

func TestServiceRetrieve(t *testing.T) {
	ctx := context.Background()
	docs := []*schema.Document{
		{ID: "doc1", Content: "content 1"},
		{ID: "doc2", Content: "content 2"},
	}

	svc := &Service{
		baseRetriever: newMockRetriever(docs, nil),
	}

	tests := []struct {
		name        string
		req         *RetrieveRequest
		wantErr     bool
		errContains string
		docCount    int
	}{
		{
			name: "valid request",
			req: &RetrieveRequest{
				Query: "test query",
			},
			wantErr:  false,
			docCount: 2,
		},
		{
			name: "empty query",
			req: &RetrieveRequest{
				Query: "",
			},
			wantErr:     true,
			errContains: "query is required",
		},
		{
			name: "with topK",
			req: &RetrieveRequest{
				Query: "test",
				TopK:  1,
			},
			wantErr:  false,
			docCount: 1,
		},
		{
			name: "topK zero uses default",
			req: &RetrieveRequest{
				Query: "test",
				TopK:  0,
			},
			wantErr:  false,
			docCount: 2, // default 10, but only 2 docs
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := svc.Retrieve(ctx, tt.req)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Retrieve() expected error, got nil")
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("Error = %v, want contain %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("Retrieve() unexpected error: %v", err)
			}
			if resp == nil {
				t.Fatal("Retrieve() returned nil response")
			}
			if len(resp.Documents) != tt.docCount {
				t.Errorf("Retrieve() returned %d docs, want %d", len(resp.Documents), tt.docCount)
			}
		})
	}
}

func TestServiceRetrieve_EmptyResult(t *testing.T) {
	ctx := context.Background()

	svc := &Service{
		baseRetriever: newMockRetriever([]*schema.Document{}, nil),
	}

	resp, err := svc.Retrieve(ctx, &RetrieveRequest{Query: "test"})

	if err != nil {
		t.Errorf("Retrieve() unexpected error: %v", err)
	}
	if resp.Total != 0 {
		t.Errorf("Retrieve() total = %d, want 0", resp.Total)
	}
	if len(resp.Documents) != 0 {
		t.Errorf("Retrieve() documents count = %d, want 0", len(resp.Documents))
	}
}

func TestServiceRetrieve_NilRetriever(t *testing.T) {
	ctx := context.Background()

	svc := &Service{
		baseRetriever: nil,
		multiRetriever: nil,
	}

	resp, err := svc.Retrieve(ctx, &RetrieveRequest{Query: "test"})

	if err != nil {
		t.Errorf("Retrieve() unexpected error: %v", err)
	}
	if resp.Total != 0 {
		t.Errorf("Retrieve() total = %d, want 0", resp.Total)
	}
}

func TestServiceRetrieve_RetrieverError(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("retrieve error")

	svc := &Service{
		baseRetriever: newMockRetriever([]*schema.Document{}, expectedErr),
	}

	resp, err := svc.Retrieve(ctx, &RetrieveRequest{Query: "test"})

	if err == nil {
		t.Error("Retrieve() expected error, got nil")
	}
	if resp != nil {
		t.Error("Retrieve() should return nil response on error")
	}
}

func TestServiceRetrieve_WithRerank(t *testing.T) {
	ctx := context.Background()
	docs := []*schema.Document{
		newDoc("doc1", 0.5),
		newDoc("doc2", 0.9),
		newDoc("doc3", 0.7),
	}

	svc := &Service{
		baseRetriever: newMockRetriever(docs, nil),
		rerankers:     []types.Reranker{&scoreReranker{}},
	}

	resp, err := svc.Retrieve(ctx, &RetrieveRequest{
		Query:        "test",
		EnableRerank: true,
	})

	if err != nil {
		t.Errorf("Retrieve() unexpected error: %v", err)
	}
	// 重排后，doc2 (0.9) 应该在第一位
	if resp.Documents[0].ID != "doc2" {
		t.Errorf("After rerank, first doc should be doc2, got %s", resp.Documents[0].ID)
	}
}

func TestServiceRetrieve_EnableMultiQuery(t *testing.T) {
	ctx := context.Background()
	chatModel := newMockServiceChatModel()
	docs := []*schema.Document{newDoc("doc1", 0.8)}

	// 创建带 multiRetriever 的服务
	multiRetriever, err := NewMultiQueryRetriever(ctx, &MultiQueryConfig{
		Retriever: newMockRetriever(docs, nil),
		RewriteLLM: chatModel,
	})
	if err != nil {
		t.Fatalf("Failed to create multiquery retriever: %v", err)
	}

	svc := &Service{
		baseRetriever:  newMockRetriever(docs, nil),
		multiRetriever: multiRetriever,
	}

	resp, err := svc.Retrieve(ctx, &RetrieveRequest{
		Query:            "test",
		EnableMultiQuery: true,
	})

	if err != nil {
		t.Errorf("Retrieve() unexpected error: %v", err)
	}
	if resp.Total == 0 {
		t.Error("Retrieve() should return docs with multiquery enabled")
	}
}

func TestServiceRetrieve_MetadataConversion(t *testing.T) {
	ctx := context.Background()
	docs := []*schema.Document{
		{
			ID:      "doc1",
			Content: "content",
			MetaData: map[string]any{
				"title": "Test Title",
				"_score": 0.85,
			},
		},
	}

	svc := &Service{
		baseRetriever: newMockRetriever(docs, nil),
	}

	resp, err := svc.Retrieve(ctx, &RetrieveRequest{Query: "test"})

	if err != nil {
		t.Errorf("Retrieve() unexpected error: %v", err)
	}
	if len(resp.Documents) != 1 {
		t.Fatalf("Retrieve() returned %d docs, want 1", len(resp.Documents))
	}

	doc := resp.Documents[0]
	if doc.ID != "doc1" {
		t.Errorf("Document ID = %q, want 'doc1'", doc.ID)
	}
	if doc.Content != "content" {
		t.Errorf("Document Content = %q, want 'content'", doc.Content)
	}
	if doc.Score != 0.85 {
		t.Errorf("Document Score = %f, want 0.85", doc.Score)
	}
	if doc.Metadata == nil {
		t.Error("Document Metadata is nil")
	} else if doc.Metadata["title"] != "Test Title" {
		t.Errorf("Metadata title = %v, want 'Test Title'", doc.Metadata["title"])
	}
}

// ========== ToContext 测试 ==========

func TestToContext(t *testing.T) {
	tests := []struct {
		name     string
		resp     *RetrieveResponse
		contains []string
	}{
		{
			name: "empty documents",
			resp: &RetrieveResponse{
				Documents: []types.Document{},
			},
			contains: []string{"未找到相关文档"},
		},
		{
			name: "single document",
			resp: &RetrieveResponse{
				Query: "test query",
				Documents: []types.Document{
					{
						ID:      "doc1",
						Content: "test content",
						Score:   0.85,
					},
				},
			},
			contains: []string{"[1]", "test content", "0.85"},
		},
		{
			name: "document with title",
			resp: &RetrieveResponse{
				Documents: []types.Document{
					{
						ID:      "doc1",
						Content: "content",
						Score:   0.9,
						Metadata: map[string]interface{}{
							"title": "Test Title",
						},
					},
				},
			},
			contains: []string{"Test Title"},
		},
		{
			name: "long content truncated",
			resp: &RetrieveResponse{
				Documents: []types.Document{
					{
						ID:      "doc1",
						Content: string(make([]byte, 600)), // 超过 500 字符
					},
				},
			},
			contains: []string{"..."},
		},
		{
			name: "multiple documents",
			resp: &RetrieveResponse{
				Documents: []types.Document{
					{ID: "doc1", Content: "content 1"},
					{ID: "doc2", Content: "content 2"},
				},
			},
			contains: []string{"[1]", "[2]", "content 1", "content 2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToContext(tt.resp)

			for _, substr := range tt.contains {
				if !contains(result, substr) {
					t.Errorf("ToContext() result missing %q", substr)
				}
			}
		})
	}
}
