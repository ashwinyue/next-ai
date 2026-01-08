// Package rerank 提供重排序服务
// 参考 next-ai/docs/eino-integration-guide.md
// 直接实现重排逻辑，避免冗余封装
package rerank

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/ashwinyue/next-ai/internal/config"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// Reranker 重排序器接口
type Reranker interface {
	Rerank(ctx context.Context, query string, docs []*schema.Document) ([]*schema.Document, error)
}

// ========== 分数重排 ==========

// NewScoreReranker 创建分数重排器
func NewScoreReranker() Reranker {
	return &scoreReranker{}
}

type scoreReranker struct{}

func (r *scoreReranker) Rerank(ctx context.Context, query string, docs []*schema.Document) ([]*schema.Document, error) {
	if len(docs) <= 1 {
		return docs, nil
	}

	// 复制并按分数降序排序
	sorted := make([]*schema.Document, len(docs))
	copy(sorted, docs)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Score() > sorted[j].Score()
	})

	return sorted, nil
}

// ========== 多样性重排 (MMR) ==========

// NewDiversityReranker 创建多样性重排器
func NewDiversityReranker(lambda float64, topN int) Reranker {
	if lambda <= 0 {
		lambda = 0.5
	}
	if lambda > 1 {
		lambda = 1
	}
	if topN <= 0 {
		topN = 5
	}

	return &diversityReranker{
		lambda: lambda,
		topN:   topN,
	}
}

type diversityReranker struct {
	lambda float64
	topN   int
}

func (r *diversityReranker) Rerank(ctx context.Context, query string, docs []*schema.Document) ([]*schema.Document, error) {
	if len(docs) <= 1 {
		return docs, nil
	}

	topN := r.topN
	if topN > len(docs) {
		topN = len(docs)
	}

	// MMR 算法: lambda * relevance - (1-lambda) * max_similarity
	selected := make([]*schema.Document, 0, topN)
	remaining := make([]*schema.Document, len(docs))
	copy(remaining, docs)

	for i := 0; i < topN && len(remaining) > 0; i++ {
		bestIdx := 0
		bestScore := -1.0

		for j, doc := range remaining {
			// 相关性分数 (归一化)
			relevance := doc.Score()
			if relevance > 1 {
				relevance = 1
			}

			// 与已选文档的最大相似度
			maxSim := 0.0
			for _, selDoc := range selected {
				sim := contentSimilarity(doc.Content, selDoc.Content)
				if sim > maxSim {
					maxSim = sim
				}
			}

			// MMR 分数
			mmrScore := r.lambda*relevance - (1-r.lambda)*maxSim

			if mmrScore > bestScore {
				bestScore = mmrScore
				bestIdx = j
			}
		}

		selected = append(selected, remaining[bestIdx])
		remaining = append(remaining[:bestIdx], remaining[bestIdx+1:]...)
	}

	return selected, nil
}

// ========== LLM 重排 ==========

// NewLLMReranker 创建 LLM 重排器
func NewLLMReranker(chatModel model.ChatModel, topN int) Reranker {
	if topN <= 0 {
		topN = 5
	}

	return &llmReranker{
		chatModel: chatModel,
		topN:      topN,
		prompt: `你是一个检索结果重排专家。请根据查询的相关性，对检索到的文档进行排序。

查询：{query}

检索到的文档：
{docs}

请按照与查询的相关度从高到低排序，输出排序后的文档编号（用逗号分隔，如：1,3,2,4,5）。

排序结果：`,
	}
}

type llmReranker struct {
	chatModel model.ChatModel
	topN      int
	prompt    string
}

func (r *llmReranker) Rerank(ctx context.Context, query string, docs []*schema.Document) ([]*schema.Document, error) {
	if len(docs) <= r.topN {
		return docs, nil
	}

	// 构建文档描述
	docDesc := r.buildDocDescription(docs)

	// 构建提示词
	prompt := strings.ReplaceAll(r.prompt, "{query}", query)
	prompt = strings.ReplaceAll(prompt, "{docs}", docDesc)

	// 调用 LLM
	messages := []*schema.Message{
		{Role: schema.System, Content: "你是一个专业的检索结果重排助手。"},
		{Role: schema.User, Content: prompt},
	}

	resp, err := r.chatModel.Generate(ctx, messages)
	if err != nil {
		// 重排失败，返回前 topN 个
		return docs[:r.topN], nil
	}

	// 解析排序结果
	indices := r.parseResult(resp.Content, len(docs))
	if len(indices) == 0 {
		return docs[:r.topN], nil
	}

	// 应用排序
	result := make([]*schema.Document, 0, min(r.topN, len(indices)))
	for i, idx := range indices {
		if idx >= 0 && idx < len(docs) && i < r.topN {
			result = append(result, docs[idx])
		}
	}

	return result, nil
}

func (r *llmReranker) buildDocDescription(docs []*schema.Document) string {
	var desc strings.Builder
	for i, doc := range docs {
		content := doc.Content
		if len(content) > 200 {
			content = content[:200] + "..."
		}
		fmt.Fprintf(&desc, "%d. %s\n", i+1, content)
	}
	return desc.String()
}

func (r *llmReranker) parseResult(output string, docCount int) []int {
	indices := make([]int, 0)

	// 提取数字
	nums := extractNumbers(output)
	for _, num := range nums {
		if num >= 1 && num <= docCount {
			indices = append(indices, num-1) // 转换为 0-based
		}
	}

	return indices
}

// ========== 组合重排 ==========

// NewCoReranker 创建组合重排器
func NewCoReranker(topN int) *CoReranker {
	if topN <= 0 {
		topN = 5
	}
	return &CoReranker{
		rerankers: make(map[string]rerankerWithWeight),
		topN:      topN,
	}
}

type CoReranker struct {
	rerankers map[string]rerankerWithWeight
	topN      int
}

type rerankerWithWeight struct {
	reranker Reranker
	weight   float64
}

func (r *CoReranker) Add(name string, reranker Reranker, weight float64) *CoReranker {
	r.rerankers[name] = rerankerWithWeight{
		reranker: reranker,
		weight:   weight,
	}
	return r
}

func (r *CoReranker) Rerank(ctx context.Context, query string, docs []*schema.Document) ([]*schema.Document, error) {
	if len(docs) <= 1 || len(r.rerankers) == 0 {
		return docs, nil
	}

	// 累积每个文档的分数
	scores := make(map[string]float64)
	docMap := make(map[string]*schema.Document)

	for _, doc := range docs {
		scores[doc.ID] = 0
		docMap[doc.ID] = doc
	}

	// 执行每个重排序器
	for _, rw := range r.rerankers {
		reranked, err := rw.reranker.Rerank(ctx, query, docs)
		if err != nil {
			continue
		}

		// 根据排名给予分数
		for rank, doc := range reranked {
			score := 1.0 / float64(rank+1)
			scores[doc.ID] += score * rw.weight
		}
	}

	// 按分数排序
	type docScore struct {
		doc   *schema.Document
		score float64
	}

	docScores := make([]docScore, 0, len(docMap))
	for id, score := range scores {
		docScores = append(docScores, docScore{
			doc:   docMap[id],
			score: score,
		})
	}

	sort.Slice(docScores, func(i, j int) bool {
		return docScores[i].score > docScores[j].score
	})

	// 返回前 topN 个
	result := make([]*schema.Document, 0, min(r.topN, len(docScores)))
	for i := 0; i < len(docScores) && i < r.topN; i++ {
		result = append(result, docScores[i].doc)
	}

	return result, nil
}

// ========== 辅助函数 ==========

func extractNumbers(s string) []int {
	nums := make([]int, 0)
	current := 0

	for _, ch := range s {
		if ch >= '0' && ch <= '9' {
			current = current*10 + int(ch-'0')
		} else {
			if current > 0 {
				nums = append(nums, current)
				current = 0
			}
		}
	}
	if current > 0 {
		nums = append(nums, current)
	}

	return nums
}

func contentSimilarity(a, b string) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}

	// 简单的词重叠计算 (Jaccard)
	extractWords := func(s string) map[string]bool {
		words := make(map[string]bool)
		current := ""

		for _, ch := range s {
			if isAlphanumeric(ch) {
				current += string(ch)
			} else {
				if current != "" {
					words[current] = true
					current = ""
				}
			}
		}
		if current != "" {
			words[current] = true
		}
		return words
	}

	wordsA := extractWords(a)
	wordsB := extractWords(b)

	if len(wordsA) == 0 || len(wordsB) == 0 {
		return 0
	}

	intersection := 0
	for word := range wordsA {
		if wordsB[word] {
			intersection++
		}
	}

	union := len(wordsA) + len(wordsB) - intersection
	if union == 0 {
		return 0
	}

	return float64(intersection) / float64(union)
}

func isAlphanumeric(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9')
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ========== 便捷创建函数 ==========

// CreateRerankers 根据配置创建重排器列表
// 直接在服务中使用，不创建工厂
func CreateRerankers(ctx context.Context, cfg *config.Config, chatModel model.ChatModel) []Reranker {
	rerankers := []Reranker{}

	// 默认添加分数重排
	rerankers = append(rerankers, NewScoreReranker())

	// LLM 重排 (如果有 ChatModel)
	if chatModel != nil {
		rerankers = append(rerankers, NewLLMReranker(chatModel, 5))
	}

	// 多样性重排
	rerankers = append(rerankers, NewDiversityReranker(0.5, 5))

	return rerankers
}

// ========== 预设配置 ==========

// PresetBasicReranker 基础重排（仅分数）
func PresetBasicReranker() Reranker {
	return NewScoreReranker()
}

// PresetDiversityReranker 多样性重排
func PresetDiversityReranker() Reranker {
	return NewDiversityReranker(0.5, 5)
}

// PresetLLMReranker LLM 重排
func PresetLLMReranker(chatModel model.ChatModel) Reranker {
	return NewLLMReranker(chatModel, 5)
}

// PresetCombinedReranker 组合重排（推荐）
func PresetCombinedReranker(chatModel model.ChatModel) Reranker {
	return NewCoReranker(5).
		Add("score", NewScoreReranker(), 1.0).
		Add("llm", NewLLMReranker(chatModel, 10), 2.0).
		Add("diversity", NewDiversityReranker(0.3, 10), 1.5)
}
