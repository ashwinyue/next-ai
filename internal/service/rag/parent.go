// Package rag 提供 Parent 检索器集成
// 使用 Eino 官方 parent.NewRetriever 实现子文档检索后返回父文档
package rag

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/retriever"
	einoParent "github.com/cloudwego/eino/flow/retriever/parent"
	"github.com/cloudwego/eino/schema"
)

// ParentConfig Parent 检索器配置
// 参考 eino/flow/retriever/parent/parent.go
type ParentConfig struct {
	// 子文档检索器（通常是向量检索器）
	Retriever retriever.Retriever

	// ParentIDKey 子文档元数据中存储父文档 ID 的键名
	// 例如: "parent_id", "source_doc_id", "knowledge_id"
	ParentIDKey string

	// OrigDocGetter 根据父文档 ID 获取完整父文档的函数
	// 参数:
	//   - ctx: 上下文
	//   - ids: 父文档 ID 列表
	// 返回:
	//   - []*schema.Document: 父文档列表
	//   - error: 错误信息
	OrigDocGetter func(ctx context.Context, ids []string) ([]*schema.Document, error)
}

// NewParentRetriever 创建父文档检索器
// 使用 Eino 官方实现，先检索子文档，然后返回完整的父文档
//
// 适用场景：
// - 文档被切分成小块（chunks）进行向量化
// - 检索到相关块后，需要返回完整文档作为上下文
//
// 使用示例:
//
//	parentRetriever, err := rag.NewParentRetriever(ctx, &rag.ParentConfig{
//	    Retriever: es8Retriever,  // 检索 chunks
//	    ParentIDKey: "parent_id", // chunk 元数据中的父文档 ID 键
//	    OrigDocGetter: func(ctx context.Context, ids []string) ([]*schema.Document, error) {
//	        // 从数据库/存储中获取完整文档
//	        return getDocumentsByIDs(ctx, ids)
//	    },
//	})
func NewParentRetriever(ctx context.Context, cfg *ParentConfig) (retriever.Retriever, error) {
	if cfg.Retriever == nil {
		return nil, fmt.Errorf("Retriever is required")
	}
	if cfg.OrigDocGetter == nil {
		return nil, fmt.Errorf("OrigDocGetter is required")
	}
	if cfg.ParentIDKey == "" {
		return nil, fmt.Errorf("ParentIDKey is required")
	}

	// 转换为 Eino 配置
	einoCfg := &einoParent.Config{
		Retriever:     cfg.Retriever,
		ParentIDKey:   cfg.ParentIDKey,
		OrigDocGetter: cfg.OrigDocGetter,
	}

	return einoParent.NewRetriever(ctx, einoCfg)
}

// ========== 预定义父文档获取函数 ==========

// MapGetter 从内存 map 中获取父文档
// 适用于测试或小规模文档集
func MapGetter(docs map[string]*schema.Document) func(ctx context.Context, ids []string) ([]*schema.Document, error) {
	return func(ctx context.Context, ids []string) ([]*schema.Document, error) {
		result := make([]*schema.Document, 0, len(ids))
		for _, id := range ids {
			if doc, ok := docs[id]; ok {
				result = append(result, doc)
			}
		}
		return result, nil
	}
}

// SliceGetter 从文档切片中获取父文档
// 通过 ID 匹配文档
func SliceGetter(allDocs []*schema.Document) func(ctx context.Context, ids []string) ([]*schema.Document, error) {
	docMap := make(map[string]*schema.Document, len(allDocs))
	for _, doc := range allDocs {
		docMap[doc.ID] = doc
	}

	return func(ctx context.Context, ids []string) ([]*schema.Document, error) {
		result := make([]*schema.Document, 0, len(ids))
		for _, id := range ids {
			if doc, ok := docMap[id]; ok {
				result = append(result, doc)
			}
		}
		return result, nil
	}
}

// ========== 辅助函数 ==========

// AddParentID 为子文档添加父文档 ID 元数据
// 创建子文档时使用，确保 Parent 检索器能找到父文档
func AddParentID(doc *schema.Document, parentID string) *schema.Document {
	if doc.MetaData == nil {
		doc.MetaData = make(map[string]any)
	}
	doc.MetaData["parent_id"] = parentID
	return doc
}

// WithParentMetadata 创建带有父文档信息的子文档
// 完整创建子文档，包含所有必要的元数据
func WithParentMetadata(
	id string,
	content string,
	parentID string,
	additionalMeta map[string]any,
) *schema.Document {
	meta := make(map[string]any)
	meta["parent_id"] = parentID
	for k, v := range additionalMeta {
		meta[k] = v
	}

	return &schema.Document{
		ID:       id,
		Content:  content,
		MetaData: meta,
	}
}
