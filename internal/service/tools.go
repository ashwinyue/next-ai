package service

import (
	"context"
	"fmt"
	"log"

	"github.com/ashwinyue/next-ai/internal/config"
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
		ToolDesc:   "Search the web for current information using DuckDuckGo. Use this when you need up-to-date information.",
		MaxResults: 10,
	})
	if err != nil {
		log.Printf("Warning: failed to create web search tool: %v", err)
		return &stubTool{name: "web_search"}
	}

	return searchTool
}

// newTools 初始化所有工具（仅通用工具，不依赖知识库）
func newTools(ctx context.Context, cfg *config.Config, repo *repository.Repositories) []tool.BaseTool {
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

	return tools
}

// newTodoWriteTool 创建任务计划工具
func newTodoWriteTool() tool.InvokableTool {
	t, err := utils.InferTool(
		"todo_write",
		`创建和管理结构化的任务列表。用于跟踪复杂任务的进度。

**使用场景**：
- 复杂多步骤任务（3个或以上步骤）
- 需要仔细规划的操作
- 用户明确请求创建任务列表

**任务状态**：
- pending: 未开始
- in_progress: 进行中（同时只能有一个）
- completed: 已完成

**重要**：
- 包含检索/研究任务
- 完成所有任务后，进行总结`,
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
		output += "注意：未提供具体步骤。建议创建3-7个任务。\n\n"
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
		output += "- 所有任务完成后，生成最终总结\n"
	} else {
		output += "✅ **所有任务已完成！**\n\n"
		output += "现在可以：\n"
		output += "- 综合所有任务的发现\n"
		output += "- 生成完整的最终答案\n"
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
