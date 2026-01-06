// Package knowledge 提供 Parent Indexer 集成
// 使用 Eino 官方 parent.NewIndexer 实现文档分块索引
package knowledge

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/components/indexer"
	einoParent "github.com/cloudwego/eino/flow/indexer/parent"
	"github.com/cloudwego/eino/schema"
)

// ParentIndexerConfig Parent Indexer 配置
// 参考 eino/flow/indexer/parent/parent.go
type ParentIndexerConfig struct {
	// Indexer 底层索引器实现（如 ES8 Indexer）
	Indexer indexer.Indexer

	// Transformer 文档转换器，用于将文档分块
	// 例如：文本分割器、代码分割器
	Transformer document.Transformer

	// ParentIDKey 子文档元数据中存储父文档 ID 的键名
	// 例如: "parent_id", "source_doc_id"
	ParentIDKey string

	// SubIDGenerator 为子文档生成唯一 ID 的函数
	// 参数:
	//   - ctx: 上下文
	//   - parentID: 父文档 ID
	//   - num: 需要生成的子文档数量
	// 返回:
	//   - []string: 子文档 ID 列表
	//   - error: 错误信息
	SubIDGenerator func(ctx context.Context, parentID string, num int) ([]string, error)
}

// NewParentIndexer 创建父文档索引器
// 使用 Eino 官方实现，在索引时自动分块并保留父文档关系
//
// 适用场景：
// - 大文档需要切分成小块进行向量化
// - 检索时需要返回完整的父文档
//
// 使用示例:
//
//	parentIndexer, err := knowledge.NewParentIndexer(ctx, &knowledge.ParentIndexerConfig{
//	    Indexer: es8Indexer,
//	    Transformer: recursiveSplitter,
//	    ParentIDKey: "parent_id",
//	    SubIDGenerator: func(ctx context.Context, parentID string, num int) ([]string, error) {
//	        ids := make([]string, num)
//	        for i := 0; i < num; i++ {
//	            ids[i] = fmt.Sprintf("%s_chunk_%d", parentID, i+1)
//	        }
//	        return ids, nil
//	    },
//	})
func NewParentIndexer(ctx context.Context, cfg *ParentIndexerConfig) (indexer.Indexer, error) {
	if cfg.Indexer == nil {
		return nil, fmt.Errorf("Indexer is required")
	}
	if cfg.Transformer == nil {
		return nil, fmt.Errorf("Transformer is required")
	}
	if cfg.SubIDGenerator == nil {
		return nil, fmt.Errorf("SubIDGenerator is required")
	}

	// 转换为 Eino 配置
	einoCfg := &einoParent.Config{
		Indexer:        cfg.Indexer,
		Transformer:    cfg.Transformer,
		ParentIDKey:    cfg.ParentIDKey,
		SubIDGenerator: cfg.SubIDGenerator,
	}

	return einoParent.NewIndexer(ctx, einoCfg)
}

// ========== 预定义子文档 ID 生成器 ==========

// SequentialChunkGenerator 顺序块 ID 生成器
// 生成格式: {parentID}_chunk_1, {parentID}_chunk_2, ...
func SequentialChunkGenerator() func(ctx context.Context, parentID string, num int) ([]string, error) {
	return func(ctx context.Context, parentID string, num int) ([]string, error) {
		ids := make([]string, num)
		for i := 0; i < num; i++ {
			ids[i] = fmt.Sprintf("%s_chunk_%d", parentID, i+1)
		}
		return ids, nil
	}
}

// UUIDChunkGenerator UUID 块 ID 生成器
// 使用 UUID 保证全局唯一性
func UUIDChunkGenerator() func(ctx context.Context, parentID string, num int) ([]string, error) {
	return func(ctx context.Context, parentID string, num int) ([]string, error) {
		ids := make([]string, num)
		for i := 0; i < num; i++ {
			// 使用 parentID 前缀 + 随机后缀
			ids[i] = fmt.Sprintf("%s_%d", parentID, i)
		}
		return ids, nil
	}
}

// ========== 辅助函数 ==========

// StoreWithParentIndexer 使用 Parent Indexer 存储文档
// 便捷函数，自动创建 Parent Indexer 并存储文档
func StoreWithParentIndexer(
	ctx context.Context,
	baseIndexer indexer.Indexer,
	transformer document.Transformer,
	parentIDKey string,
	docs []*schema.Document,
) ([]string, error) {
	// 创建 Parent Indexer
	parentIndexer, err := NewParentIndexer(ctx, &ParentIndexerConfig{
		Indexer:        baseIndexer,
		Transformer:    transformer,
		ParentIDKey:    parentIDKey,
		SubIDGenerator: SequentialChunkGenerator(),
	})
	if err != nil {
		return nil, fmt.Errorf("create parent indexer: %w", err)
	}

	// 存储文档
	return parentIndexer.Store(ctx, docs)
}
