package service

import (
	"context"
	"fmt"
	"log"

	"github.com/ashwinyue/next-ai/internal/config"
	"github.com/ashwinyue/next-ai/internal/model"
	"github.com/ashwinyue/next-ai/internal/repository"
	"github.com/cloudwego/eino-ext/components/tool/duckduckgo/v2"
	httptool "github.com/cloudwego/eino-ext/components/tool/httprequest"
	sequencethinking "github.com/cloudwego/eino-ext/components/tool/sequentialthinking"
	wikipediatool "github.com/cloudwego/eino-ext/components/tool/wikipedia"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
)

// PlanStep 计划步骤
type PlanStep struct {
	ID          string `json:"id" jsonschema_description:"步骤ID"`
	Description string `json:"description" jsonschema_description:"步骤描述"`
	Status      string `json:"status" jsonschema_description:"状态: pending, in_progress, completed"`
}

// TodoWriteInput todo_write 输入参数
type TodoWriteInput struct {
	Task  string     `json:"task" jsonschema_description:"任务描述"`
	Steps []PlanStep `json:"steps" jsonschema_description:"任务步骤列表"`
}

// KnowledgeSearchInput 知识库搜索输入
type KnowledgeSearchInput struct {
	Query string `json:"query" jsonschema_description:"The search query" jsonschema_required:"true"`
	TopK  int    `json:"top_k" jsonschema_description:"Number of results (default 10)"`
}

// KnowledgeSearchOutput 知识库搜索输出
type KnowledgeSearchOutput struct {
	Query   string                   `json:"query"`
	Total   int                      `json:"total"`
	Results []map[string]interface{} `json:"results"`
}

// DocumentInfoInput 文档信息输入
type DocumentInfoInput struct {
	DocumentIDs []string `json:"document_ids" jsonschema_description:"文档 ID 列表，最多 10 个" jsonschema_required:"true"`
}

// DocumentInfoOutput 文档信息输出
type DocumentInfoOutput struct {
	Count     int                      `json:"count"`
	Documents []map[string]interface{} `json:"documents"`
}

// ListChunksInput 列出分块输入
type ListChunksInput struct {
	DocumentID string `json:"document_id" jsonschema_description:"文档 ID" jsonschema_required:"true"`
}

// ChunkItem 分块项
type ChunkItem struct {
	ID         string `json:"id"`
	ChunkIndex int    `json:"chunk_index"`
	Content    string `json:"content"`
}

// ListChunksOutput 列出分块输出
type ListChunksOutput struct {
	DocumentID string      `json:"document_id"`
	Title      string      `json:"title"`
	Total      int         `json:"total"`
	Chunks     []ChunkItem `json:"chunks"`
}

// GrepChunksInput 分块搜索输入
type GrepChunksInput struct {
	Pattern    string `json:"pattern" jsonschema_description:"搜索模式（文本）" jsonschema_required:"true"`
	DocumentID string `json:"document_id" jsonschema_description:"可选：限制在特定文档中搜索"`
}

// GrepChunkItem 搜索结果项
type GrepChunkItem struct {
	ID         string `json:"id"`
	DocumentID string `json:"document_id"`
	ChunkIndex int    `json:"chunk_index"`
	Content    string `json:"content"`
}

// GrepChunksOutput 分块搜索输出
type GrepChunksOutput struct {
	Pattern string          `json:"pattern"`
	Count   int             `json:"count"`
	Matches []GrepChunkItem `json:"matches"`
}

// stubTool 占位工具
type stubTool struct {
	name string
}

func (t *stubTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: t.name,
		Desc: t.name + " (unavailable)",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"query": {
				Type:     schema.String,
				Desc:     "The query string",
				Required: true,
			},
		}),
	}, nil
}

func (t *stubTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	return fmt.Sprintf(`{"error":"%s is not available"}`, t.name), nil
}

// newWebSearchTool 创建网络搜索工具
func newWebSearchTool(ctx context.Context) tool.InvokableTool {
	searchTool, err := duckduckgo.NewTextSearchTool(ctx, &duckduckgo.Config{
		ToolName:   "web_search",
		ToolDesc:   "Search the web for current information using DuckDuckGo. Use this when you need up-to-date information or the knowledge base doesn't have the answer.",
		MaxResults: 10,
	})
	if err != nil {
		log.Printf("Warning: failed to create web search tool: %v", err)
		return &stubTool{name: "web_search"}
	}

	return searchTool
}

// newTools 初始化所有工具
func newTools(ctx context.Context, cfg *config.Config, retriever interface{}, repo *repository.Repositories) []tool.BaseTool {
	tools := []tool.BaseTool{}

	// 添加网络搜索工具 (eino-ext duckduckgo)
	tools = append(tools, newWebSearchTool(ctx))

	// 添加 HTTP 请求工具 (eino-ext httprequest)
	httpTools, err := httptool.NewToolKit(ctx, &httptool.Config{})
	if err != nil {
		log.Printf("Warning: failed to create http tools: %v", err)
	} else {
		tools = append(tools, httpTools...)
	}

	// 添加 Wikipedia 搜索工具 (eino-ext wikipedia)
	wikiTool, err := wikipediatool.NewTool(ctx, &wikipediatool.Config{
		Language: "zh", // 中文 Wikipedia
		TopK:     3,
	})
	if err != nil {
		log.Printf("Warning: failed to create wikipedia tool: %v", err)
	} else {
		tools = append(tools, wikiTool)
	}

	// 添加顺序思考工具 (eino-ext sequentialthinking)
	thinkTool, err := sequencethinking.NewTool()
	if err != nil {
		log.Printf("Warning: failed to create sequentialthinking tool: %v", err)
	} else {
		tools = append(tools, thinkTool)
	}

	// 添加 todo_write 工具
	tools = append(tools, newTodoWriteTool())

	// 添加知识库搜索工具
	if retriever != nil {
		tools = append(tools, newKnowledgeSearchTool(retriever))
	}

	// 添加文档相关工具
	if repo != nil {
		tools = append(tools, newDocumentInfoTool(repo))
		tools = append(tools, newListChunksTool(repo))
		tools = append(tools, newGrepChunksTool(repo))
	}

	return tools
}

// newTodoWriteTool 创建任务计划工具
func newTodoWriteTool() tool.InvokableTool {
	t, err := utils.InferTool(
		"todo_write",
		`创建和管理结构化的检索任务列表。用于跟踪复杂任务的进度。

**使用场景**：
- 复杂多步骤任务（3个或以上步骤）
- 需要仔细规划的操作
- 用户明确请求创建任务列表

**任务状态**：
- pending: 未开始
- in_progress: 进行中（同时只能有一个）
- completed: 已完成

**重要**：
- 仅包含检索/研究任务，不包含总结任务
- 完成所有检索任务后，使用 thinking 工具进行总结`,
		func(ctx context.Context, input *TodoWriteInput) (string, error) {
			if input.Task == "" {
				input.Task = "未提供任务描述"
			}
			return generateTodoOutput(input.Task, input.Steps), nil
		},
	)
	if err != nil {
		log.Printf("Warning: failed to create todo_write tool: %v", err)
		return nil
	}
	return t
}

// newKnowledgeSearchTool 创建知识库搜索工具
func newKnowledgeSearchTool(r interface{}) tool.InvokableTool {
	t, err := utils.InferTool(
		"knowledge_search",
		"Searches the knowledge base for relevant information using semantic and keyword search.",
		func(ctx context.Context, input *KnowledgeSearchInput) (*KnowledgeSearchOutput, error) {
			if input.Query == "" {
				return nil, fmt.Errorf("query is required")
			}
			if input.TopK <= 0 {
				input.TopK = 10
			}

			// 使用 retriever.Retrieve 接口
			docs, err := retrieveInterface(r, ctx, input.Query, input.TopK)
			if err != nil {
				return nil, fmt.Errorf("retriever failed: %w", err)
			}

			results := make([]map[string]interface{}, 0, len(docs))
			for _, doc := range docs {
				result := map[string]interface{}{
					"content": doc.Content,
					"score":   doc.Score(),
				}
				if doc.MetaData != nil {
					if title, ok := doc.MetaData["title"].(string); ok {
						result["title"] = title
					}
				}
				results = append(results, result)
			}

			return &KnowledgeSearchOutput{
				Query:   input.Query,
				Total:   len(results),
				Results: results,
			}, nil
		},
	)
	if err != nil {
		log.Printf("Warning: failed to create knowledge_search tool: %v", err)
		return nil
	}
	return t
}

// newDocumentInfoTool 创建文档信息工具
func newDocumentInfoTool(repo *repository.Repositories) tool.InvokableTool {
	t, err := utils.InferTool(
		"get_document_info",
		"获取文档的详细元数据信息，包括标题、文件名、大小、分块数量等。用于查询文档基本信息和处理状态。",
		func(ctx context.Context, input *DocumentInfoInput) (*DocumentInfoOutput, error) {
			if len(input.DocumentIDs) == 0 {
				return nil, fmt.Errorf("document_ids is required")
			}
			if len(input.DocumentIDs) > 10 {
				return nil, fmt.Errorf("maximum 10 document IDs allowed")
			}

			results := make([]map[string]interface{}, 0)
			for _, docID := range input.DocumentIDs {
				doc, err := repo.Knowledge.GetDocumentByID(docID)
				if err != nil {
					continue
				}
				chunks, _ := repo.Knowledge.GetChunksByDocumentID(docID)

				results = append(results, map[string]interface{}{
					"id":           doc.ID,
					"title":        doc.Title,
					"file_name":    doc.FileName,
					"file_size":    doc.FileSize,
					"content_type": doc.ContentType,
					"status":       doc.Status,
					"chunk_count":  len(chunks),
					"created_at":   doc.CreatedAt,
				})
			}

			return &DocumentInfoOutput{
				Count:     len(results),
				Documents: results,
			}, nil
		},
	)
	if err != nil {
		log.Printf("Warning: failed to create get_document_info tool: %v", err)
		return nil
	}
	return t
}

// newListChunksTool 创建列出分块工具
func newListChunksTool(repo *repository.Repositories) tool.InvokableTool {
	t, err := utils.InferTool(
		"list_chunks",
		"获取指定文档的所有分块内容。用于查看文档的完整分块列表。",
		func(ctx context.Context, input *ListChunksInput) (*ListChunksOutput, error) {
			if input.DocumentID == "" {
				return nil, fmt.Errorf("document_id is required")
			}

			doc, err := repo.Knowledge.GetDocumentByID(input.DocumentID)
			if err != nil {
				return nil, fmt.Errorf("document not found: %w", err)
			}

			chunks, err := repo.Knowledge.GetChunksByDocumentID(input.DocumentID)
			if err != nil {
				return nil, fmt.Errorf("failed to get chunks: %w", err)
			}

			chunkList := make([]ChunkItem, 0, len(chunks))
			for _, c := range chunks {
				chunkList = append(chunkList, ChunkItem{
					ID:         c.ID,
					ChunkIndex: c.ChunkIndex,
					Content:    c.Content,
				})
			}

			return &ListChunksOutput{
				DocumentID: doc.ID,
				Title:      doc.Title,
				Total:      len(chunks),
				Chunks:     chunkList,
			}, nil
		},
	)
	if err != nil {
		log.Printf("Warning: failed to create list_chunks tool: %v", err)
		return nil
	}
	return t
}

// newGrepChunksTool 创建分块搜索工具
func newGrepChunksTool(repo *repository.Repositories) tool.InvokableTool {
	t, err := utils.InferTool(
		"grep_chunks",
		"在文档分块中搜索包含特定文本的内容。支持精确文本匹配。",
		func(ctx context.Context, input *GrepChunksInput) (*GrepChunksOutput, error) {
			if input.Pattern == "" {
				return nil, fmt.Errorf("pattern is required")
			}

			var chunks []*model.DocumentChunk
			var err error

			if input.DocumentID != "" {
				// 搜索特定文档的分块
				chunks, err = repo.Knowledge.GetChunksByDocumentID(input.DocumentID)
				if err != nil {
					return nil, fmt.Errorf("failed to get chunks: %w", err)
				}
			}

			// 过滤包含匹配内容的分块
			matches := make([]GrepChunkItem, 0)
			for _, c := range chunks {
				if containsIgnoreCase(c.Content, input.Pattern) {
					matches = append(matches, GrepChunkItem{
						ID:         c.ID,
						DocumentID: c.DocumentID,
						ChunkIndex: c.ChunkIndex,
						Content:    c.Content,
					})
				}
			}

			return &GrepChunksOutput{
				Pattern: input.Pattern,
				Count:   len(matches),
				Matches: matches,
			}, nil
		},
	)
	if err != nil {
		log.Printf("Warning: failed to create grep_chunks tool: %v", err)
		return nil
	}
	return t
}

// GetToolsByName 根据名称获取工具
func GetToolsByName(ctx context.Context, names []string, allTools []tool.BaseTool) ([]tool.BaseTool, error) {
	if len(names) == 0 {
		return allTools, nil
	}

	toolMap := make(map[string]tool.BaseTool)
	for _, t := range allTools {
		info, err := t.Info(ctx)
		if err != nil {
			continue
		}
		toolMap[info.Name] = t
	}

	result := make([]tool.BaseTool, 0, len(names))
	for _, name := range names {
		t, ok := toolMap[name]
		if !ok {
			return nil, fmt.Errorf("tool not found: %s", name)
		}
		result = append(result, t)
	}

	return result, nil
}

// ListToolNames 列出所有工具名称
func ListToolNames(ctx context.Context, allTools []tool.BaseTool) []string {
	names := make([]string, 0, len(allTools))
	for _, t := range allTools {
		info, err := t.Info(ctx)
		if err != nil {
			continue
		}
		names = append(names, info.Name)
	}
	return names
}

// generateTodoOutput 生成 todo 输出
func generateTodoOutput(task string, steps []PlanStep) string {
	output := "## 计划已创建\n\n"
	output += fmt.Sprintf("**任务**: %s\n\n", task)

	if len(steps) == 0 {
		output += "注意：未提供具体步骤。建议创建3-7个检索任务。\n\n"
		output += "建议的检索流程：\n"
		output += "1. 使用 grep_chunks 搜索关键词定位相关文档\n"
		output += "2. 使用 knowledge_search 进行语义搜索获取相关内容\n"
		output += "3. 使用 list_chunks 获取关键文档的完整内容\n"
		output += "4. 使用 web_search 获取补充信息（如需要）\n"
		return output
	}

	// 统计任务状态
	pendingCount := 0
	inProgressCount := 0
	completedCount := 0
	for _, step := range steps {
		switch step.Status {
		case "pending":
			pendingCount++
		case "in_progress":
			inProgressCount++
		case "completed":
			completedCount++
		}
	}
	totalCount := len(steps)
	remainingCount := pendingCount + inProgressCount

	output += "**任务步骤**:\n\n"
	for i, step := range steps {
		output += formatTodoStep(i+1, step)
	}

	// 添加进度汇总
	output += "\n## 任务进度\n"
	output += fmt.Sprintf("总计: %d 个任务 | ", totalCount)
	output += fmt.Sprintf("✅ 已完成: %d | ", completedCount)
	output += fmt.Sprintf("🔄 进行中: %d | ", inProgressCount)
	output += fmt.Sprintf("⏳ 待处理: %d\n\n", pendingCount)

	// 添加提醒
	output += "## ⚠️ 重要提醒\n"
	if remainingCount > 0 {
		output += fmt.Sprintf("**还有 %d 个任务未完成！**\n\n", remainingCount)
		output += "**必须完成所有任务后才能总结或得出结论。**\n\n"
		output += "下一步操作：\n"
		if inProgressCount > 0 {
			output += "- 继续完成当前进行中的任务\n"
		}
		if pendingCount > 0 {
			output += fmt.Sprintf("- 开始处理 %d 个待处理任务\n", pendingCount)
		}
		output += "- 完成每个任务后，更新 todo_write 标记为 completed\n"
		output += "- 所有任务完成后，使用 thinking 工具生成最终总结\n"
	} else {
		output += "✅ **所有任务已完成！**\n\n"
		output += "现在可以：\n"
		output += "- 综合所有任务的发现\n"
		output += "- 使用 thinking 工具生成完整的最终答案\n"
	}

	return output
}

// formatTodoStep 格式化单个任务步骤
func formatTodoStep(index int, step PlanStep) string {
	statusEmoji := map[string]string{
		"pending":     "⏳",
		"in_progress": "🔄",
		"completed":   "✅",
	}

	emoji, ok := statusEmoji[step.Status]
	if !ok {
		emoji = "⏳"
	}

	return fmt.Sprintf("%d. %s [%s] %s\n", index, emoji, step.Status, step.Description)
}

// retrieveInterface 调用 retriever 的接口
func retrieveInterface(r interface{}, ctx context.Context, query string, topK int) ([]*schema.Document, error) {
	// 这里使用类型断言来处理不同的 retriever 类型
	// 实际实现取决于 retriever 的具体类型
	type retrieverInterface interface {
		Retrieve(ctx context.Context, query string, opts ...interface{}) ([]*schema.Document, error)
	}

	if retriever, ok := r.(retrieverInterface); ok {
		return retriever.Retrieve(ctx, query)
	}

	return nil, fmt.Errorf("unsupported retriever type")
}
