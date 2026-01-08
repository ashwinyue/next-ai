// Package rag 提供 RAG 测试辅助函数
package rag

import (
	"context"

	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
)

// newDoc 创建测试文档
func newDoc(content string, score float64) *schema.Document {
	return &schema.Document{
		ID:       content,
		Content:  content,
		MetaData: map[string]interface{}{"_score": score},
	}
}

// newDocWithSource 创建带来源的测试文档
func newDocWithSource(content string, score float64, source string) *schema.Document {
	return &schema.Document{
		ID:       content,
		Content:  content,
		MetaData: map[string]interface{}{"_score": score, "source": source},
	}
}

// contains 检查字符串包含
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// almostEqual 浮点数近似比较
func almostEqual(a, b, epsilon float64) bool {
	if a < b {
		a, b = b, a
	}
	return a-b <= epsilon
}

// ========== Mock Retriever ==========

type mockRetriever struct {
	documents []*schema.Document
	err       error
}

func newMockRetriever(docs []*schema.Document, err error) retriever.Retriever {
	return &mockRetriever{documents: docs, err: err}
}

func (m *mockRetriever) Retrieve(ctx context.Context, query string, opts ...retriever.Option) ([]*schema.Document, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.documents, nil
}
