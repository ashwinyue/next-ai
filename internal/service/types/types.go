// Package types 定义共享的类型和接口
package types

import (
	"context"

	"github.com/cloudwego/eino/schema"
)

// Reranker 重排器接口
type Reranker interface {
	Rerank(ctx context.Context, query string, docs []*schema.Document) ([]*schema.Document, error)
}

// Document 文档
type Document struct {
	ID       string                 `json:"id"`
	Content  string                 `json:"content"`
	Score    float64                `json:"score"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}
