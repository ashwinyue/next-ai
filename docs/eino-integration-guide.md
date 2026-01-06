# Eino 框架集成指南

## 核心原则

**避免冗余封装，直接使用 eino/eino-ext 标准库**

参考 `eino/eino-examples` 的最佳实践，不要创建额外的抽象层。

## 代码组织

### ❌ 不推荐：冗余的工厂封装

```go
// internal/eino/chatmodel.go
package eino

func ChatModelFactory(ctx context.Context, cfg *config.Config) (model.ChatModel, error) {
    // 配置适配...
    return openai.NewChatModel(ctx, &openai.ChatModelConfig{...})
}
```

### ✅ 推荐：直接在服务中初始化

```go
// internal/service/agent/agent.go
package agent

func (s *Service) newChatModel(ctx context.Context) (model.ChatModel, error) {
    return openai.NewChatModel(ctx, &openai.ChatModelConfig{
        APIKey:  apiKey,
        BaseURL: baseURL,
        Model:   modelName,
    })
}
```

## 命名规范

初始化函数使用 `new` 前缀，简单直接：

| 组件 | 函数名 | 位置 |
|------|--------|------|
| ChatModel | `newChatModel()` | 服务包内 |
| Embedding | `newEmbedder()` | service.go |
| Retriever | `newES8Retriever()` | service.go |
| Tools | `newTools()` | service.go |
| Agent | `createAgent()` | agent.go |

## 参考示例

eino-examples 最佳实践：

```go
// eino/eino-examples/flow/agent/manus/manus.go

func newChatModel(ctx context.Context) model.ToolCallingChatModel {
    cm, err = openai.NewChatModel(ctx, &openai.ChatModelConfig{
        APIKey:      openaiAPIKey,
        BaseURL:     openaiBaseURL,
        Model:       openaiModel,
        Temperature: &temp,
    })
    // ...
}

func main() {
    cm := newChatModel(ctx)
    tools := newCommandLineTools(ctx, sb)
    agent := composeAgent(ctx, cm, browserTool, tools)
    // ...
}
```

## 目录结构

```
internal/service/
├── service.go          # 统一初始化 Embedding、Retriever、Tools
├── agent/
│   └── agent.go        # Agent 服务，直接使用 eino ADK
├── chat/
│   └── chat.go         # Chat 服务
└── ...
```

## 直接导入 eino 组件

```go
import (
    // eino 标准库
    "github.com/cloudwego/eino/components/model"
    "github.com/cloudwego/eino/components/tool"
    "github.com/cloudwego/eino/components/embedding"
    "github.com/cloudwego/eino/adk"
    "github.com/cloudwego/eino/compose"
    "github.com/cloudwego/eino/schema"

    // eino-ext 扩展库（直接使用，不封装）
    "github.com/cloudwego/eino-ext/components/model/openai"
    "github.com/cloudwego/eino-ext/components/embedding/dashscope"
    "github.com/cloudwego/eino-ext/components/retriever/es8"
    "github.com/cloudwego/eino-ext/components/tool/duckduckgo/v2"
)
```

## Agent 使用模式

```go
// 1. 创建 ChatModel
chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{...})

// 2. 创建 Agent（使用 ADK）
agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
    Name:        agentName,
    Description: agentDesc,
    Instruction: systemPrompt,
    Model:       chatModel,
    ToolsConfig: adk.ToolsConfig{
        ToolsNodeConfig: compose.ToolsNodeConfig{
            Tools: tools,
        },
    },
})

// 3. 运行 Agent
iter := agent.Run(ctx, &adk.AgentInput{
    Messages:        messages,
    EnableStreaming: false,
})

// 4. 处理事件
for {
    event, ok := iter.Next()
    if !ok {
        break
    }
    // 处理 event...
}
```

## 工具实现

直接实现 `tool.InvokableTool` 或 `tool.StreamableTool` 接口：

```go
type KnowledgeSearchTool struct {
    retriever *es8.Retriever
}

func (t *KnowledgeSearchTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
    return &schema.ToolInfo{
        Name: "knowledge_search",
        Desc: "Search the knowledge base...",
        ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
            "query": {Type: schema.String, Desc: "The search query", Required: true},
        }),
    }, nil
}

func (t *KnowledgeSearchTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
    // 直接使用 retriever...
    docs, err := t.retriever.Retrieve(ctx, query, retriever.WithTopK(topK))
    // ...
}
```

## 注意事项

1. **不创建单独的 eino 包**：初始化函数放在需要的服务包中
2. **使用简单的 newXxx() 函数**：不要用工厂模式
3. **直接调用 eino-ext API**：不要二次封装
4. **参考 eino-examples**：保持与官方示例一致的代码风格
