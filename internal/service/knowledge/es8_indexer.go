// Package knowledge 提供 ES8 Indexer 集成
// 使用 Eino 官方 eino-ext es8.NewIndexer 实现文档索引
package knowledge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/ashwinyue/next-ai/internal/config"
	"github.com/ashwinyue/next-ai/internal/model"
	"github.com/cloudwego/eino-ext/components/indexer/es8"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/schema"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// NewES8Indexer 创建 ES8 Indexer（Eino 官方组件）
// 用于将文档块索引到 Elasticsearch
//
// 使用示例:
//
//	indexer, err := knowledge.NewES8Indexer(ctx, cfg, embedder)
//	if err != nil {
//	    return err
//	}
//
//	// 转换 DocumentChunk 到 Eino Document
//	docs := chunkToEinoDocument(chunks)
//
//	// 索引文档
//	ids, err := indexer.Store(ctx, docs)
func NewES8Indexer(ctx context.Context, cfg *config.Config, embedder embedding.Embedder) (es8Indexer, error) {
	client, err := newES8Client(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create ES client: %w", err)
	}

	indexName := cfg.Elastic.IndexPrefix + "_chunks"

	indexer, err := es8.NewIndexer(ctx, &es8.IndexerConfig{
		Client: client,
		Index:  indexName,
		BatchSize: 10,
		Embedding: embedder,
		DocumentToFields: func(ctx context.Context, doc *schema.Document) (map[string]es8.FieldValue, error) {
			return documentToESFields(doc), nil
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create ES8 indexer: %w", err)
	}

	return &es8IndexerWrapper{
		Indexer:   indexer,
		indexName: indexName,
		client:    client,
		cfg:       cfg,
	}, nil
}

// es8Indexer 是 ES8 Indexer 的接口封装
type es8Indexer interface {
	Store(ctx context.Context, docs []*schema.Document) ([]string, error)
	EnsureIndex(ctx context.Context) error
	Close() error
}

// es8IndexerWrapper 包装 Eino ES8 Indexer，添加索引管理功能
type es8IndexerWrapper struct {
	Indexer   *es8.Indexer
	indexName string
	client    *elasticsearch.Client
	cfg       *config.Config
}

// Store 存储文档到 ES
func (w *es8IndexerWrapper) Store(ctx context.Context, docs []*schema.Document) ([]string, error) {
	return w.Indexer.Store(ctx, docs)
}

// EnsureIndex 确保索引存在（Eino Indexer 不包含此功能，需要手动实现）
func (w *es8IndexerWrapper) EnsureIndex(ctx context.Context) error {
	return ensureESIndex(ctx, w.client, w.indexName, w.cfg.AI.Embedding.Dimensions)
}

// Close 关闭索引器（ES8 Indexer 无需显式关闭）
func (w *es8IndexerWrapper) Close() error {
	return nil
}

// documentToESFields 将 Eino Document 转换为 ES 字段
// 这是 ES8 Indexer 要求的转换函数
func documentToESFields(doc *schema.Document) map[string]es8.FieldValue {
	fields := make(map[string]es8.FieldValue)

	// 内容字段（需要向量化）
	fields["content"] = es8.FieldValue{
		Value:    doc.Content,
		EmbedKey: "content_vector", // 指定向量结果的存储键名
	}

	// 元数据字段（直接存储，不向量化）
	if doc.MetaData != nil {
		for k, v := range doc.MetaData {
			fields[k] = es8.FieldValue{
				Value: v,
			}
		}
	}

	return fields
}

// ChunksToEinoDocuments 将 DocumentChunk 列片转换为 Eino Document 列片
func ChunksToEinoDocuments(chunks []*model.DocumentChunk) []*schema.Document {
	docs := make([]*schema.Document, len(chunks))
	for i, chunk := range chunks {
		metadata := make(map[string]any)
		if chunk.Metadata != nil {
			for k, v := range chunk.Metadata {
				metadata[k] = v
			}
		}
		// 添加额外的元数据
		metadata["document_id"] = chunk.DocumentID
		metadata["chunk_index"] = chunk.ChunkIndex
		metadata["knowledge_base_id"] = chunk.KnowledgeBaseID
		if chunk.ParentChunkID != "" {
			metadata["parent_chunk_id"] = chunk.ParentChunkID
		}

		docs[i] = &schema.Document{
			ID:       chunk.ID,
			Content:  chunk.Content,
			MetaData: metadata,
		}
	}
	return docs
}

// ========== ES 辅助函数 ==========

// newES8Client 创建 ES8 客户端
func newES8Client(cfg *config.Config) (*elasticsearch.Client, error) {
	return elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{cfg.Elastic.Host},
		Username:  cfg.Elastic.Username,
		Password:  cfg.Elastic.Password,
	})
}

// ensureESIndex 确保 ES 索引存在（如不存在则创建）
func ensureESIndex(ctx context.Context, client *elasticsearch.Client, indexName string, dimensions int) error {
	// 检查索引是否存在
	res, err := client.Indices.Exists([]string{indexName})
	if err != nil {
		return fmt.Errorf("failed to check index existence: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode == 200 {
		return nil // 索引已存在
	}

	if dimensions == 0 {
		dimensions = 1536 // 默认 OpenAI 维度
	}

	// 创建索引映射，支持向量字段
	mapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"content": map[string]interface{}{
					"type": "text",
				},
				"content_vector": map[string]interface{}{
					"type":       "dense_vector",
					"dims":       dimensions,
					"index":      true,
					"similarity": "cosine",
				},
				"document_id": map[string]interface{}{
					"type": "keyword",
				},
				"chunk_index": map[string]interface{}{
					"type": "integer",
				},
				"knowledge_base_id": map[string]interface{}{
					"type": "keyword",
				},
				"parent_chunk_id": map[string]interface{}{
					"type": "keyword",
				},
				"metadata": map[string]interface{}{
					"type": "object",
				},
			},
		},
		"settings": map[string]interface{}{
			"number_of_shards":   1,
			"number_of_replicas": 0,
		},
	}

	mappingData, err := json.Marshal(mapping)
	if err != nil {
		return fmt.Errorf("failed to marshal mapping: %w", err)
	}

	req := esapi.IndicesCreateRequest{
		Index: indexName,
		Body:  bytes.NewReader(mappingData),
	}

	res, err = req.Do(ctx, client)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("failed to create index: %s", res.String())
	}

	log.Printf("Index %s created with %d dimensions", indexName, dimensions)
	return nil
}
