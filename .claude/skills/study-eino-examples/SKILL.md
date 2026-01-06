---
name: study-eino-examples
description: Learn Eino best practices from official examples. Use this skill when implementing Eino components (ChatModel, Agent, Tool, Retriever) to understand correct usage patterns. Reference agent/quick_start, agent/tool_use, compose examples. Key patterns: adk.NewChatModelAgent, openai.NewChatModel, tool.InvokableTool. Use direct eino types, newXxx() initialization, no factories.
---

参考 Eino 官方示例 Skill。查找和学习 eino-examples 中的最佳实践，确保代码风格符合 Eino 标准。

## 使用场景
当你需要实现某个 Eino 组件（ChatModel、Agent、Tool、Retriever），想了解正确的使用方式。

## 操作步骤

### 1. 定位相关示例
eino-examples 目录结构：
```
old/eino-examples/
├── agent/               # Agent 示例
│   ├── quick_start/        # 快速开始
│   └── tool_use/           # 工具使用
├── callback/            # 回调使用
├── compose/             # 组合模式
├── embedding/           # 向量使用
├── indexing/            # 索引示例
└── retriever/           # 检索示例
```

### 2. 搜索示例代码
```bash
# 查找 Agent 示例
rg "adk.NewChatModelAgent" old/eino-examples/

# 查找 ChatModel 使用
rg "openai.NewChatModel" old/eino-examples/

# 查找 Tool 实现
rg "tool.InvokableTool" old/eino-examples/
```

### 3. 学习代码模式
关注以下模式：

#### Agent 创建模式
```go
// 参考 eino-examples/agent/quick_start/main.go
chatModel, _ := openai.NewChatModel(ctx, &openai.ChatModelConfig{...})

agent, _ := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
    Name:          "my_agent",
    Description:   "Agent description",
    Instruction:   systemPrompt,
    Model:         chatModel,
    MaxIterations: 10,
    ToolsConfig: adk.ToolsConfig{
        ToolsNodeConfig: compose.ToolsNodeConfig{
            Tools: []einotool.BaseTool{...},
        },
    },
})
```

#### Tool 实现模式
```go
// 直接实现 tool.InvokableTool 接口
type MyTool struct{}

func (t *MyTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
    return &schema.ToolInfo{
        Name: "my_tool",
        Desc: "Tool description",
        ParamsOneOf: schema.NewParamsOneOfByParams(...),
    }, nil
}

func (t *MyTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
    // 实现逻辑
}
```

### 4. 应用到 Next-AI
在 internal/service/ 中按相同模式实现：
- 直接使用 eino 类型，不创建额外封装
- 初始化函数使用 `newXxx()` 命名
- 不使用工厂模式

## 示例

### 任务：创建带工具的 Agent
```bash
# 1. 查找示例
ls old/eino-examples/agent/tool_use/

# 2. 阅读示例代码
cat old/eino-examples/agent/tool_use/main.go

# 3. 在 Next-AI 中实现
# internal/service/agent/agent.go
```

## 核心原则
1. **直接使用**：不创建抽象层，直接使用 eino API
2. **简单初始化**：用 `newXxx()` 函数，不用工厂
3. **接口实现**：直接实现 Eino 接口（如 tool.InvokableTool）
