// Package query 提供查询处理服务
// 参考 next-ai/docs/eino-integration-guide.md
// 直接使用 eino ChatModel，避免冗余封装
package query

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// ========== 查询重写 ==========

// NewRewriter 创建查询重写器
func NewRewriter(chatModel model.ChatModel) *Rewriter {
	prompt := `你是一个查询优化专家。请将用户的查询重写为更清晰、更完整的表达，以便于检索。

原始查询：{query}

请输出优化后的查询（仅输出查询文本，不要解释）：`

	return &Rewriter{
		chatModel: chatModel,
		prompt:    prompt,
	}
}

type Rewriter struct {
	chatModel model.ChatModel
	prompt    string
}

// Rewrite 重写查询
func (r *Rewriter) Rewrite(ctx context.Context, query string) (string, error) {
	if r.chatModel == nil {
		return query, nil // 没有 ChatModel 时返回原查询
	}

	prompt := strings.ReplaceAll(r.prompt, "{query}", query)

	messages := []*schema.Message{
		{Role: schema.System, Content: "你是一个专业的查询优化助手。"},
		{Role: schema.User, Content: prompt},
	}

	resp, err := r.chatModel.Generate(ctx, messages)
	if err != nil {
		return query, err // 失败时返回原查询
	}

	rewritten := strings.TrimSpace(resp.Content)
	if rewritten == "" {
		return query, nil
	}

	return rewritten, nil
}

// ========== 查询扩展 ==========

// NewExpander 创建查询扩展器
func NewExpander(chatModel model.ChatModel, numVariants int) *Expander {
	if numVariants <= 0 {
		numVariants = 3
	}

	prompt := `你是一个查询扩展专家。请为用户的查询生成多个语义相似的变体，以提升检索召回率。

原始查询：{query}

请生成 %d 个查询变体，每行一个，不要编号：`

	return &Expander{
		chatModel:   chatModel,
		numVariants: numVariants,
		prompt:      prompt,
	}
}

type Expander struct {
	chatModel   model.ChatModel
	numVariants int
	prompt      string
}

// Expand 扩展查询
func (e *Expander) Expand(ctx context.Context, query string) ([]string, error) {
	if e.chatModel == nil {
		return []string{query}, nil // 没有 ChatModel 时返回原查询
	}

	prompt := fmt.Sprintf(e.prompt, query, e.numVariants)

	messages := []*schema.Message{
		{Role: schema.System, Content: "你是一个专业的查询扩展助手。"},
		{Role: schema.User, Content: prompt},
	}

	resp, err := e.chatModel.Generate(ctx, messages)
	if err != nil {
		return []string{query}, nil
	}

	// 解析扩展的查询（按行分割）
	lines := strings.Split(resp.Content, "\n")
	queries := make([]string, 0, e.numVariants+1)
	queries = append(queries, query) // 包含原查询

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// 跳过空行和带编号的行
		if line == "" || isNumberedLine(line) {
			continue
		}
		// 移除可能的编号前缀
		line = removeNumberPrefix(line)
		if line != "" && line != query {
			queries = append(queries, line)
			if len(queries) >= e.numVariants+1 {
				break
			}
		}
	}

	return queries, nil
}

// ========== 查询分解 ==========

// NewDecomposer 创建查询分解器
func NewDecomposer(chatModel model.ChatModel) *Decomposer {
	prompt := `你是一个查询分解专家。请将复杂查询分解为多个简单的子查询，每个子查询可以独立检索。

复杂查询：{query}

请输出分解后的子查询，每行一个，不要编号：`

	return &Decomposer{
		chatModel: chatModel,
		prompt:    prompt,
	}
}

type Decomposer struct {
	chatModel model.ChatModel
	prompt    string
}

// Decompose 分解查询
func (d *Decomposer) Decompose(ctx context.Context, query string) ([]string, error) {
	if d.chatModel == nil {
		return []string{query}, nil
	}

	prompt := strings.ReplaceAll(d.prompt, "{query}", query)

	messages := []*schema.Message{
		{Role: schema.System, Content: "你是一个专业的查询分解助手。"},
		{Role: schema.User, Content: prompt},
	}

	resp, err := d.chatModel.Generate(ctx, messages)
	if err != nil {
		return []string{query}, nil
	}

	// 解析子查询
	lines := strings.Split(resp.Content, "\n")
	queries := make([]string, 0)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || isNumberedLine(line) {
			continue
		}
		line = removeNumberPrefix(line)
		if line != "" {
			queries = append(queries, line)
		}
	}

	if len(queries) == 0 {
		return []string{query}, nil
	}

	return queries, nil
}

// ========== 查询优化器（组合）==========

// NewOptimizer 创建查询优化器（重写 + 扩展）
func NewOptimizer(chatModel model.ChatModel, numVariants int) *Optimizer {
	return &Optimizer{
		rewriter: NewRewriter(chatModel),
		expander: NewExpander(chatModel, numVariants),
	}
}

type Optimizer struct {
	rewriter *Rewriter
	expander *Expander
}

// Optimize 优化查询（重写 + 扩展）
func (o *Optimizer) Optimize(ctx context.Context, query string) (*OptimizedQuery, error) {
	result := &OptimizedQuery{
		Original: query,
	}

	// 1. 重写查询
	rewritten, err := o.rewriter.Rewrite(ctx, query)
	if err != nil {
		rewritten = query
	}
	result.Rewritten = rewritten

	// 2. 扩展查询（基于重写后的查询）
	expanded, err := o.expander.Expand(ctx, rewritten)
	if err != nil {
		expanded = []string{rewritten}
	}
	result.Expanded = expanded

	return result, nil
}

// OptimizedQuery 优化后的查询结果
type OptimizedQuery struct {
	Original  string   // 原始查询
	Rewritten string   // 重写后的查询
	Expanded  []string // 扩展的查询列表（包含重写查询）
}

// GetQueries 获取所有用于检索的查询
func (o *OptimizedQuery) GetQueries() []string {
	if len(o.Expanded) > 0 {
		return o.Expanded
	}
	return []string{o.Rewritten}
}

// ========== 辅助函数 ==========

func isNumberedLine(line string) bool {
	line = strings.TrimSpace(line)
	if len(line) == 0 {
		return false
	}
	// 检查是否以数字开头 (如 "1.", "1)", "- 1.")
	runes := []rune(line)
	if len(runes) > 0 && runes[0] >= '1' && runes[0] <= '9' {
		return true
	}
	return false
}

func removeNumberPrefix(line string) string {
	line = strings.TrimSpace(line)
	// 移除常见的数字前缀
	// "1. xxx" -> "xxx"
	// "1) xxx" -> "xxx"
	// "- 1. xxx" -> "xxx"
	prefixes := []string{"- ", "— ", "• "}

	for _, prefix := range prefixes {
		if strings.HasPrefix(line, prefix) {
			line = strings.TrimPrefix(line, prefix)
			break
		}
	}

	// 移除数字前缀
	runes := []rune(line)
	for i, r := range runes {
		if r >= '0' && r <= '9' {
			continue
		}
		if r == '.' || r == ')' || r == '、' || r == ' ' {
			return strings.TrimSpace(string(runes[i+1:]))
		}
		break
	}

	return line
}

// ========== 预设配置 ==========

// PresetBasicQuery 基础查询处理（无变化）
func PresetBasicQuery() []string {
	return nil // 不处理，使用原查询
}

// PresetRewriteOnly 仅重写
func PresetRewriteOnly(chatModel model.ChatModel) func(ctx context.Context, query string) ([]string, error) {
	rewriter := NewRewriter(chatModel)
	return func(ctx context.Context, query string) ([]string, error) {
		rewritten, err := rewriter.Rewrite(ctx, query)
		if err != nil {
			return []string{query}, nil
		}
		return []string{rewritten}, nil
	}
}

// PresetExpandOnly 仅扩展
func PresetExpandOnly(chatModel model.ChatModel) func(ctx context.Context, query string) ([]string, error) {
	expander := NewExpander(chatModel, 3)
	return func(ctx context.Context, query string) ([]string, error) {
		expanded, err := expander.Expand(ctx, query)
		if err != nil {
			return []string{query}, nil
		}
		return expanded, nil
	}
}

// PresetRewriteAndExpand 重写 + 扩展（推荐）
func PresetRewriteAndExpand(chatModel model.ChatModel) func(ctx context.Context, query string) (*OptimizedQuery, error) {
	optimizer := NewOptimizer(chatModel, 3)
	return optimizer.Optimize
}
