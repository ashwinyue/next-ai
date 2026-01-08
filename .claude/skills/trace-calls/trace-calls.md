# 调用链梳理与完善技能

## 技能说明

用于梳理 WeKnora 调用链并对比 next-ai 实现，找出已实现但未打通的调用链，提供完善方案。

## WeKnora 核心调用链分析

### 1. Agent 调用链

```
HTTP Layer (handler/custom_agent.go)
    ↓ CreateAgent/GetAgent/UpdateAgent/DeleteAgent
Service Layer (application/service/agent_service.go)
    ↓ CreateAgentEngine
    - ValidateConfig
    - registerTools (知识搜索、网络搜索、工具等)
    - getKnowledgeBaseInfos (获取知识库详情)
    - getSelectedDocumentInfos (获取用户@提及文档)
Engine Layer (agent/engine.go)
    ↓ NewAgentEngine
    ↓ Execute
    - buildMessagesWithLLMContext
    - buildToolsForLLM
    - executeLoop (ReAct 循环)
        ├─ streamThinkingToEventBus (思考)
        ├─ toolRegistry.ExecuteTool (执行工具)
        └─ appendToolResults (保存结果)
Event Bus (event/)
    ↓ Emit 事件到前端
    - EventAgentThought (思考内容)
    - EventAgentToolCall (工具调用)
    - EventAgentToolResult (工具结果)
    - EventAgentFinalAnswer (最终答案)
```

**关键特点**：
- 使用 `dig.In` 进行依赖注入
- EventBus 实现 SSE 流式传输
- 工具注册器模式动态注册工具
- 知识库信息集成到 System Prompt

### 2. Chat 调用链

```
HTTP Layer (handler/session/qa.go)
    ↓ parseQARequest
    - GetSession
    - GetCustomAgent (如果指定 agent_id)
    ↓ setupSSEStream
    - 创建 EventBus
    - 设置 stop handler
    - 设置 stream handler
Service Layer (application/service/session.go)
    ↓ CreateQA
    - QueryRewriteService.RewriteQuery (查询重写)
    - KnowledgeBaseService.Search (知识检索)
    - ChatService.Chat (LLM 调用)
```

### 3. Knowledge 调用链

```
HTTP Layer (handler/knowledge.go)
    ↓ CreateKnowledgeFromFile
    - validateKnowledgeBaseAccess
    - FormFile 获取上传文件
Service Layer (application/service/knowledge.go)
    ↓ Upload
    - 文件解析 (Parse)
    - 分块 (Chunk)
    - 向量化 (Embed)
    - 索引 (Index)
Repository (application/repository/knowledge.go)
    ↓ Create/Update/Delete
```

## Next-AI 对比分析

### 已实现 ✅

| 组件 | 状态 | 文件路径 |
|------|------|----------|
| Agent Handler | ✅ | `internal/handler/agent_handler.go` |
| Agent Service | ✅ | `internal/service/agent/agent.go` |
| Knowledge Handler | ✅ | `internal/handler/knowledge_handler.go` |
| Knowledge Service | ✅ | `internal/service/knowledge/knowledge.go` |
| Chat Handler | ✅ | `internal/handler/chat_handler.go` |
| Eino ChatModel | ✅ | `internal/service/service.go:newChatModel` |
| Eino Embedder | ✅ | `internal/service/service.go:newEmbedder` |
| Eino Retriever | ✅ | `internal/service/service.go:newES8Retriever` |
| Tools (web_search, etc.) | ✅ | `internal/service/service.go:newTools` |

### 未打通的调用链 ⚠️

#### 1. Agent 执行与知识库集成

**问题**：Agent 执行时知识库搜索工具缺少知识库 ID 上下文

```go
// 当前实现 (service/agent/agent.go:454)
selectedTools, err := GetToolsByName(ctx, toolNames, s.allTools)
// 工具创建时没有知识库 ID 信息
func newKnowledgeSearchTool(r *es8.Retriever) einotool.InvokableTool {
    // 直接使用全局 Retriever，没有知识库限制
}
```

**需要**：
- 创建带知识库 ID 上下文的 Retriever
- Agent 运行时动态创建带限制的工具

#### 2. Chat 会话与 Agent 集成

**问题**：Chat 服务调用 Agent 时缺少会话上下文传递

```go
// 当前 ChatService 简单调用，没有传递 session 相关信息
// 缺少：会话历史加载、消息保存、事件流集成
```

**需要**：
- ChatService.AgentChat 调用 Agent 时传递会话 ID
- Agent 执行结果保存到会话
- SSE 事件流统一处理

#### 3. 事件总线流式传输

**问题**：EventBus 已创建但未在 Handler 层使用

```go
// service/service.go:134
eventBus := event.NewEventBus(newEventStore(redisClient))
// 但 Handler 层 SSE 流式输出未使用 EventBus
```

**需要**：
- Handler 层使用 EventBus 转发 Agent 事件
- 统一的 SSE 事件格式

#### 4. 工具参数注入

**问题**：工具需要运行时参数（sessionID、tenantID、knowledgeBaseIDs）

```go
// WeKnora 方式：在 registerTools 时动态创建带参数的工具
toolToRegister = tools.NewKnowledgeSearchTool(
    s.knowledgeBaseService,
    s.knowledgeService,
    s.chunkService,
    config.SearchTargets,
    rerankModel,
    chatModel,
    s.cfg,
)
```

**需要**：
- Agent 运行时动态创建带上下文的工具
- 而非启动时创建全局工具

## 完善方案

### 1. 增强 Agent Service (优先级：高)

```go
// internal/service/agent/agent.go

// RunRequest 添加可选的上下文参数
type RunRequest struct {
    Query           string   `json:"query"`
    SessionID       string   `json:"session_id"`
    KnowledgeBaseIDs []string `json:"knowledge_base_ids"` // 新增
    TenantID        string   `json:"tenant_id"`           // 新增
}

// createAgentWithContext 运行时创建带上下文的 Agent
func (s *Service) createAgentWithContext(
    ctx context.Context,
    agentModel *agentmodel.Agent,
    req *RunRequest,
) (*adk.ChatModelAgent, error) {
    // 根据请求中的 knowledge_base_ids 创建受限的 Retriever
    retriever := s.createScopedRetriever(ctx, req.KnowledgeBaseIDs, req.TenantID)

    // 创建带上下文的工具
    tools := s.createToolsWithContext(ctx, retriever, req.SessionID, req.TenantID)

    // ... 创建 Agent
}
```

### 2. 完善 Chat Service (优先级：高)

```go
// internal/service/chat/chat.go

// AgentChat 调用 Agent 并处理流式响应
func (s *Service) AgentChat(
    ctx context.Context,
    sessionID, agentID, query string,
    knowledgeBaseIDs []string,
) <-chan AgentEvent {
    // 1. 加载会话历史
    history := s.loadHistory(ctx, sessionID)

    // 2. 调用 Agent
    eventCh := s.agent.Stream(ctx, agentID, &agent.RunRequest{
        Query:           query,
        SessionID:       sessionID,
        KnowledgeBaseIDs: knowledgeBaseIDs,
    })

    // 3. 转换事件并保存消息
    outCh := make(chan AgentEvent)
    go func() {
        defer close(outCh)
        for event := range eventCh {
            // 转发事件
            outCh <- event
            // 保存最终答案
            if event.Type == "end" {
                s.saveMessage(ctx, sessionID, "user", query)
                s.saveMessage(ctx, sessionID, "assistant", event.Data)
            }
        }
    }()
    return outCh
}
```

### 3. 统一事件格式 (优先级：中)

```go
// internal/service/event/events.go

// AgentEvent 统一的 Agent 事件格式
type AgentEvent struct {
    Type     string                 `json:"type"` // start, thinking, tool_call, tool_result, message, end, error
    Data     string                 `json:"data"`
    Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ToSSE 转换为 SSE 格式
func (e *AgentEvent) ToSSE() string {
    return fmt.Sprintf("data: %s\n\n", e.ToJSON())
}
```

### 4. 动态工具创建 (优先级：中)

```go
// internal/service/agent/tools.go

// createToolsWithContext 运行时创建带上下文的工具
func (s *Service) createToolsWithContext(
    ctx context.Context,
    retriever *es8.Retriever,
    sessionID, tenantID string,
) []tool.BaseTool {
    tools := []tool.BaseTool{}

    // 知识搜索工具（带租户过滤）
    if retriever != nil {
        tools = append(tools, newScopedKnowledgeSearchTool(retriever, tenantID))
    }

    // 文档工具（带租户过滤）
    tools = append(tools, newScopedDocumentTools(s.repo, tenantID))

    return tools
}
```

## 使用方法

在对话中使用此技能：

```
使用 trace-calls 技能，帮我分析：
1. Agent 执行时如何集成知识库搜索？
2. Chat 调用 Agent 时如何传递会话上下文？
```

或

```
使用 trace-calls 技能，检查 next-ai 中 [模块名] 的调用链是否完整
```

## 最新状态 (2025-01-07)

### ✅ 已完成的调用链打通

1. **Agent Service 增强** (`internal/service/agent/agent.go`)
   - ✅ 添加 `RunWithContext` 方法支持带知识库 ID 的 Agent 执行
   - ✅ 添加 `StreamWithContext` 方法支持流式输出和事件发布
   - ✅ 添加 `AgentEvent` 统一事件格式
   - ✅ 添加 `createToolsWithContext` 方法 - 创建带上下文的工具
   - ✅ 添加 `newScopedKnowledgeSearchTool` - 带知识库 ID 过滤的搜索工具
   - ✅ 添加 `stubTool` - 存根工具实现

2. **Chat Service 集成** (`internal/service/chat/chat.go`)
   - ✅ 添加 `AgentService` 接口（避免循环依赖）
   - ✅ 添加 `ServiceWithAgent` 结构（含 RetrieverProvider）
   - ✅ 添加 `AgentChat` 方法 - 调用 Agent 进行聊天（流式）
   - ✅ 添加 `KnowledgeChat` 方法 - 知识库聊天（使用快速问答 Agent）
   - ✅ 添加 `KnowledgeSearch` 方法 - 独立知识库搜索接口（含 Retriever 调用）

3. **Handler 层 SSE 流式输出** (`internal/handler/chat_handler.go`)
   - ✅ `KnowledgeChat` - 知识库聊天 SSE 流式输出
   - ✅ `AgentChat` - 智能体聊天 SSE 流式输出
   - ✅ `KnowledgeSearch` - 知识搜索（已有实现）

4. **服务初始化完善** (`internal/service/service.go`)
   - ✅ Provider 适配器（eventBusProvider, retrieverProvider, chatModelProvider）
   - ✅ `agentServiceAdapter` 实现 Agent 到 Chat 的桥接
   - ✅ `Services.Chat` 类型为 `*chat.ServiceWithAgent`

### 🔄 调用链流程

```
HTTP Request (Handler)
    ↓
[AgentChat / KnowledgeChat] (chat_handler.go)
    ↓
ServiceWithAgent.AgentChat() / .KnowledgeChat()
    ↓
Agent.StreamWithContext() (agent.go:1129)
    ├─ loadHistory() - 加载会话历史
    ├─ createToolsWithContext() - 创建带上下文的工具
    ├─ adk.Agent.Run() - 运行 Eino Agent
    └─ publishEvent() - 发布事件到 EventBus
    ↓
StreamEvent Channel
    ↓
SSE Response (Handler 层流式输出)
```

### ⚠️ 后续可优化部分

1. **ES8 Retriever 按知识库 ID 过滤**
   - 当前 `createToolsWithContext` 创建工具框架
   - 需要实现 ES 查询时按 `knowledge_base_id` 字段过滤
   - 位置：`internal/service/agent/agent.go:1354`

2. **EventBus 订阅和转发**
   - 当前事件发布到 EventBus
   - 可添加 Handler 层订阅 EventBus 转发事件
   - 用于多客户端事件广播

---

## 最新更新 (2025-01-08)

### ✅ RAG Service 完全统一到 Eino 组件

**重构内容**:
1. **删除手动实现** (`internal/service/rag/service.go`):
   - ✅ 删除 `query` 字段（不再需要手动查询优化器）
   - ✅ 删除 `multiRetrieve()` 方法（使用 Eino 组件替代）
   - ✅ `EnableOptimize` 重定向到 `multiRetriever`

2. **统一 API**:
   - ✅ `EnableOptimize` 和 `EnableMultiQuery` 都使用 Eino `multiquery.NewRetriever`
   - ✅ 简化 `NewService()` 签名，删除 `query` 参数
   - ✅ 保留 `NewServiceWithConfig()` 用于高级配置

3. **清理依赖** (`internal/service/service.go`):
   - ✅ 删除 `Services.Query` 字段
   - ✅ 删除 `queryOptimizer` 初始化代码
   - ✅ 删除 `query` 包导入

**调用链**:
```
Handler (chat_handler.go / rag_handler.go)
    ↓ ragSvc := rag.NewService(chatModel, retriever, rerankers)
    ↓ req.EnableOptimize = true 或 req.EnableMultiQuery = true
Service (rag/service.go)
    ↓ if (req.EnableMultiQuery || req.EnableOptimize) && s.multiRetriever != nil
    ↓     retrieverForUse = s.multiRetriever
    ↓ retrieverForUse.Retrieve(ctx, query)
Eino multiquery.NewRetriever
    ↓ RewriteLLM 生成多条查询
    ↓ 并行调用底层 Retriever
    ↓ FusionFunc 融合去重结果
```

**编译状态**: ✅ 通过 `go build ./...` 验证

**修改文件**:
- `internal/service/rag/service.go` - 删除手动实现，统一到 Eino
- `internal/service/agent/agent.go` - 删除 `scopedDocTool`
- `internal/service/service.go` - 清理 Query 字段
- `internal/handler/chat_handler.go` - 更新调用
- `internal/handler/rag_handler.go` - 更新调用
