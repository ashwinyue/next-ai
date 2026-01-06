// Package rag 提供 RAG 增强功能
// 包括 MMR 多样性优化、复合评分、内容签名去重等
package rag

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math"
	"sort"
	"strings"
	"unicode"

	"github.com/cloudwego/eino/schema"
)

// ========== MMR 多样性优化 ==========

// MMRConfig MMR 配置
type MMRConfig struct {
	K      int     // 选择的结果数量
	Lambda float64 // 相关性权重 (0-1), 默认 0.7
	//   lambda = 1.0: 只看相关性，不考虑多样性
	//   lambda = 0.5: 平衡相关性和多样性
	//   lambda = 0.0: 只看多样性，不考虑相关性
}

// DefaultMMRConfig 默认 MMR 配置
func DefaultMMRConfig() *MMRConfig {
	return &MMRConfig{
		K:      10,
		Lambda: 0.7,
	}
}

// ApplyMMR 应用 MMR (Maximal Marginal Relevance) 算法
// 减少检索结果冗余，提高多样性
//
// 参数:
//   docs - 候选文档列表（已按相关性排序）
//   cfg - MMR 配置
//
// 返回:
//   经过 MMR 处理后的文档列表
func ApplyMMR(docs []*schema.Document, cfg *MMRConfig) []*schema.Document {
	if cfg == nil {
		cfg = DefaultMMRConfig()
	}
	if len(docs) == 0 || cfg.K <= 0 {
		return docs
	}

	k := cfg.K
	if k > len(docs) {
		k = len(docs)
	}

	// 预计算所有文档的 token 集合
	tokenSets := make([]map[string]struct{}, len(docs))
	for i, doc := range docs {
		tokenSets[i] = tokenize(doc.Content)
	}

	selected := make([]*schema.Document, 0, k)
	remaining := make([]*schema.Document, len(docs))
	copy(remaining, docs)
	remainingTokenSets := make([]map[string]struct{}, len(docs))
	copy(remainingTokenSets, tokenSets)

	// MMR 选择循环
	for len(selected) < k && len(remaining) > 0 {
		bestIdx := 0
		bestScore := -1.0

		for i, r := range remaining {
			relevance := r.Score()
			redundancy := 0.0

			// 计算与已选结果的最大冗余度
			for _, s := range selected {
				selectedTokens := tokenize(s.Content)
				redundancy = math.Max(redundancy, jaccard(remainingTokenSets[i], selectedTokens))
			}

			// MMR 分数 = λ * 相关性 - (1-λ) * 冗余度
			mmr := cfg.Lambda*relevance - (1.0-cfg.Lambda)*redundancy
			if mmr > bestScore {
				bestScore = mmr
				bestIdx = i
			}
		}

		// 添加最佳候选到已选列表
		selected = append(selected, remaining[bestIdx])

		// 从剩余列表中移除
		remaining = append(remaining[:bestIdx], remaining[bestIdx+1:]...)
		remainingTokenSets = append(remainingTokenSets[:bestIdx], remainingTokenSets[bestIdx+1:]...)
	}

	return selected
}

// ========== 复合评分 ==========

// CompositeScoreConfig 复合评分配置
type CompositeScoreConfig struct {
	ModelWeight  float64 // 模型分数权重
	BaseWeight   float64 // 基础分数权重
	SourceWeight float64 // 来源权重
}

// DefaultCompositeConfig 默认复合评分配置
func DefaultCompositeConfig() *CompositeScoreConfig {
	return &CompositeScoreConfig{
		ModelWeight:  0.6,
		BaseWeight:   0.3,
		SourceWeight: 0.1,
	}
}

// DocumentWithSource 带来源信息的文档
type DocumentWithSource struct {
	Document    *schema.Document
	Source      string // 来源标识 (如 "es8", "milvus")
	BaseScore   float64
	PositionPrior float64 // 位置优先级
}

// ApplyCompositeScore 应用复合评分
// 综合模型分数、基础分数、来源权重计算最终分数
func ApplyCompositeScore(docs []*DocumentWithSource, cfg *CompositeScoreConfig) []*DocumentWithSource {
	if cfg == nil {
		cfg = DefaultCompositeConfig()
	}

	for _, d := range docs {
		// 复合公式: w1*model + w2*base + w3*source
		modelScore := d.Document.Score()
		sourceWeight := cfg.SourceWeight

		// 根据来源调整权重
		if d.Source == "web_search" {
			sourceWeight = 0.95
		}

		// 计算位置先验
		positionPrior := 1.0
		if d.PositionPrior > 0 {
			positionPrior = d.PositionPrior
		}

		// 复合分数
		composite := cfg.ModelWeight*modelScore +
			cfg.BaseWeight*d.BaseScore +
			sourceWeight

		// 应用位置先验
		composite *= positionPrior

		// 更新文档分数（使用 WithScore 方法）
		d.Document = d.Document.WithScore(clamp(composite, 0, 1))
	}

	// 按新分数排序
	sort.Slice(docs, func(i, j int) bool {
		return docs[i].Document.Score() > docs[j].Document.Score()
	})

	return docs
}

// ========== 内容签名去重 ==========

// ContentDedupConfig 内容去重配置
type ContentDedupConfig struct {
	Threshold float64 // Jaccard 相似度阈值 (0-1)
}

// DefaultContentDedupConfig 默认去重配置
func DefaultContentDedupConfig() *ContentDedupConfig {
	return &ContentDedupConfig{
		Threshold: 0.85, // 85% 相似度视为重复
	}
}

// ContentDedupResult 内容去重结果
type ContentDedupResult struct {
	Unique       []*schema.Document
	DuplicateCount int
	RemovedCount   int
}

// ContentDedupBySignature 基于内容签名去重
// 使用 MD5 签名检测完全相同的内容
func ContentDedupBySignature(docs []*schema.Document) *ContentDedupResult {
	seen := make(map[string]bool)
	unique := make([]*schema.Document, 0)

	for _, doc := range docs {
		sig := buildContentSignature(doc.Content)
		if !seen[sig] {
			seen[sig] = true
			unique = append(unique, doc)
		}
	}

	return &ContentDedupResult{
		Unique:         unique,
		DuplicateCount: len(docs) - len(unique),
		RemovedCount:   len(docs) - len(unique),
	}
}

// ContentDedupBySimilarity 基于相似度去重
// 使用 Jaccard 相似度检测近似重复内容
func ContentDedupBySimilarity(docs []*schema.Document, cfg *ContentDedupConfig) *ContentDedupResult {
	if cfg == nil {
		cfg = DefaultContentDedupConfig()
	}

	unique := make([]*schema.Document, 0)
	removed := 0

	for _, doc := range docs {
		isDuplicate := false
		docTokens := tokenize(doc.Content)

		for _, u := range unique {
			similarity := jaccard(docTokens, tokenize(u.Content))
			if similarity >= cfg.Threshold {
				isDuplicate = true
				removed++
				break
			}
		}

		if !isDuplicate {
			unique = append(unique, doc)
		}
	}

	return &ContentDedupResult{
		Unique:         unique,
		DuplicateCount: removed,
		RemovedCount:   removed,
	}
}

// buildContentSignature 构建内容签名
// 归一化处理后计算 MD5
func buildContentSignature(content string) string {
	// 归一化：转小写、去空格、去标点
	normalized := normalizeContent(content)

	hash := md5.Sum([]byte(normalized))
	return hex.EncodeToString(hash[:])
}

// normalizeContent 归一化内容
func normalizeContent(content string) string {
	// 转小写
	content = strings.ToLower(content)

	// 去除多余空格
	words := strings.Fields(content)

	return strings.Join(words, " ")
}

// tokenize 简单分词
func tokenize(text string) map[string]struct{} {
	// 转小写
	text = strings.ToLower(text)

	// 按空白和标点分词
	words := strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != '_'
	})

	tokenSet := make(map[string]struct{}, len(words))
	for _, w := range words {
		if len(w) > 1 { // 忽略单字符
			tokenSet[w] = struct{}{}
		}
	}

	return tokenSet
}

// jaccard 计算 Jaccard 相似度
// J(A,B) = |A∩B| / |A∪B|
func jaccard(a, b map[string]struct{}) float64 {
	if len(a) == 0 && len(b) == 0 {
		return 1.0
	}
	if len(a) == 0 || len(b) == 0 {
		return 0.0
	}

	intersection := 0
	for k := range a {
		if _, ok := b[k]; ok {
			intersection++
		}
	}

	union := len(a) + len(b) - intersection
	return float64(intersection) / float64(union)
}

// clamp 限制值在范围内
func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// ========== 检索统计输出 ==========

// RetrievalStats 检索统计信息
type RetrievalStats struct {
	Query           string
	TotalResults    int
	UniqueResults   int
	DuplicateCount  int
	AverageScore    float64
	TopScore        float64
	BottomScore     float64
	Sources         map[string]int // 来源计数
}

// CalculateStats 计算检索统计信息
func CalculateStats(docs []*schema.Document) *RetrievalStats {
	if len(docs) == 0 {
		return &RetrievalStats{
			Sources: make(map[string]int),
		}
	}

	stats := &RetrievalStats{
		TotalResults:  len(docs),
		Sources:      make(map[string]int),
	}

	sumScore := 0.0
	stats.TopScore = docs[0].Score()
	stats.BottomScore = docs[0].Score()

	for _, doc := range docs {
		score := doc.Score()
		sumScore += score

		if score > stats.TopScore {
			stats.TopScore = score
		}
		if score < stats.BottomScore {
			stats.BottomScore = score
		}

		// 统计来源
		if doc.MetaData != nil {
			if source, ok := doc.MetaData["source"].(string); ok {
				stats.Sources[source]++
			}
		}
	}

	stats.AverageScore = sumScore / float64(len(docs))
	stats.UniqueResults = len(docs)

	return stats
}

// FormatStats 格式化统计信息为字符串
func FormatStats(stats *RetrievalStats) string {
	if stats == nil {
		return ""
	}

	sb := strings.Builder{}
	sb.WriteString("=== 检索统计 ===\n")
	sb.WriteString(fmtString("总结果数", stats.TotalResults))
	sb.WriteString(fmtString("唯一结果数", stats.UniqueResults))
	sb.WriteString(fmtString("平均分数", formatFloat(stats.AverageScore)))
	sb.WriteString(fmtString("最高分数", formatFloat(stats.TopScore)))
	sb.WriteString(fmtString("最低分数", formatFloat(stats.BottomScore)))

	if len(stats.Sources) > 0 {
		sb.WriteString("\n来源分布:\n")
		for source, count := range stats.Sources {
			sb.WriteString(fmt.Sprintf("  - %s: %d\n", source, count))
		}
	}

	return sb.String()
}

func fmtString(key string, value interface{}) string {
	return fmt.Sprintf("%-20s: %v\n", key, value)
}

func formatFloat(f float64) string {
	return fmt.Sprintf("%.4f", f)
}

// ========== 批量处理 ==========

// BatchConfig 批量处理配置
type BatchConfig struct {
	BatchSize    int
	MaxContentLength int
}

// DefaultBatchConfig 默认批量配置
func DefaultBatchConfig() *BatchConfig {
	return &BatchConfig{
		BatchSize:       15,
		MaxContentLength: 800,
	}
}

// ProcessBatches 批量处理文档
// 将大量文档分批处理，避免一次性处理过多数据
func ProcessBatches(items []interface{}, cfg *BatchConfig, processFunc func([]interface{}) error) error {
	if cfg == nil {
		cfg = DefaultBatchConfig()
	}

	for i := 0; i < len(items); i += cfg.BatchSize {
		end := i + cfg.BatchSize
		if end > len(items) {
			end = len(items)
		}

		batch := items[i:end]
		if err := processFunc(batch); err != nil {
			return fmt.Errorf("batch %d-%d failed: %w", i, end, err)
		}
	}

	return nil
}
