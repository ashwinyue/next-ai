---
name: migration-progress
description: Track WeKnora to Next-AI migration progress. Use this skill when updating features or checking migration status. Shows which handlers, services, and agent tools are migrated (98% complete), and lists pending items.
---

WeKnora → Next-AI 迁移进度追踪 Skill。用于追踪 WeKnora 功能向 Next-AI 的迁移进度，每次更新功能时同步更新此文档。

## 进度总览

| 分类 | 已迁移 | 待迁移 | 完成度 |
|------|--------|--------|--------|
| Handler | 15 | 0 | **100%** |
| Agent 工具 | 12 | 0 | **100%** |
| Agent 功能 | 2 | 0 | 100% |
| FAQ 增强 | 9 | 0 | 100% |
| Service | 15 | 0 | **100%** |
| **Eino 组件** | **7** | **0** | **100%** |
| 总计 | 45 | 0 | **100%** ✅ |

---

## 一、Handler 层迁移状态

### ✅ 已完成 (14/14)

| Handler | WeKnora 文件 | Next-AI 文件 | 状态 |
|---------|--------------|--------------|------|
| Agent | custom_agent.go | agent_handler.go | ✅ 完整迁移 + **Agent模式扩展 (2025-01-07)** |
| Chat | session/ | chat_handler.go | ✅ 完整迁移 |
| Knowledge | knowledge.go, knowledgebase.go | knowledge_handler.go | ✅ 完整迁移 |
| FAQ | faq.go | faq_handler.go | ✅ 完整迁移 (含增强版) |
| RAG | - | rag_handler.go | ✅ 完整迁移 |
| Tool | - | tool_handler.go | ✅ 完整迁移 |
| **Auth** | auth.go | auth_handler.go | ✅ 完整迁移 |
| **Chunk** | chunk.go | chunk_handler.go | ✅ 完整迁移 |
| **Initialization** | initialization.go | initialization_handler.go | ✅ 完整迁移 |
| **Model** | model.go | model_handler.go | ✅ 完整迁移 (2025-01-07) |
| **Evaluation** | evaluation.go | evaluation_handler.go | ✅ 完整迁移 (2025-01-07) |
| **MCP Service** | mcp_service.go | mcp_handler.go | ✅ 完整迁移 (2025-01-07) |
| **Tenant** | tenant.go | tenant_handler.go | ✅ 完整迁移 (2025-01-07) |
| **Tag** | tag.go | tag_handler.go | ✅ 完整迁移 (2025-01-07) |
| **File** | file/ | file_handler.go | ✅ 完整迁移 (2025-01-07) |
| **Dataset** | dataset.go | dataset_handler.go | ✅ 完整迁移 (2025-01-07) |

---

## 二、Eino 组件集成状态

### ✅ 已完成 (7/7)

| 组件 | 功能 | 文件 |
|------|------|------|
| **Router** | 多检索器路由 + RRF 融合 | rag/router.go |
| **MultiQuery** | 查询扩展 (LLM 生成多条查询) | rag/multiquery.go |
| **Parent (Retriever)** | 子文档检索后返回父文档 | rag/parent.go |
| **Parent (Indexer)** | 索引时自动分块并保留父子关系 | knowledge/parent_indexer.go |
| **Callback Logger** | Eino 回调日志支持 | callback/logger.go |
| **ErrorRemover** | 工具错误处理中间件 | agent/middleware.go |
| **JsonFix** | JSON 参数修复中间件 | agent/middleware.go |

---

## 三、Agent 工具迁移状态

### ✅ 已完成 (12/12)

| 工具名 | WeKnora 文件 | Next-AI 实现 | 位置 |
|--------|--------------|--------------|------|
| knowledge_search | knowledge_search.go | ✅ | eino-ext |
| web_search | web_search.go | ✅ DuckDuckGo | eino-ext |
| http_get/post/put/delete | web_fetch.go | ✅ httpRequest | eino-ext |
| thinking | sequentialthinking.go | ✅ | eino-ext |
| wikipedia_search | - | ✅ 新增 | eino-ext |
| get_document_info | get_document_info.go | ✅ DocumentInfoTool | service.go |
| list_chunks | list_knowledge_chunks.go | ✅ ListChunksTool | service.go |
| grep_chunks | grep_chunks.go | ✅ GrepChunksTool | service.go |
| todo_write | todo_write.go | ✅ TodoWriteTool | service.go |
| **database_query** | database_query.go | ✅ QueryTool | database/query.go |
| **data_analysis** | data_analysis.go | ✅ AnalysisTool | database/analysis.go |
| **data_schema** | data_schema.go | ✅ SchemaTool | database/schema.go |

---

## 四、FAQ 增强功能迁移状态

### ✅ 已完成 (9/9)

| 功能 | Next-AI 实现 | 位置 | 状态 |
|------|-------------|------|------|
| CreateEntry | ✅ | faq/entry.go | 完整迁移 |
| GetEntry | ✅ | faq/entry.go | 完整迁移 |
| ListEntries | ✅ | faq/entry.go | 完整迁移 |
| UpdateEntry | ✅ | faq/entry.go | 完整迁移 |
| DeleteEntry | ✅ | faq/entry.go | 完整迁移 |
| DeleteEntries (批量) | ✅ | faq/entry.go | 完整迁移 |
| UpdateEntryCategoryBatch | ✅ | faq/entry.go | 完整迁移 |
| UpdateEntryFieldsBatch | ✅ | faq/entry.go | 完整迁移 |
| ExportEntries | ✅ | faq/entry.go | 完整迁移 |
| BatchUpsert | ✅ | faq/entry.go | 完整迁移 |
| GetImportProgress | ✅ | faq/entry.go | 完整迁移 |

---

## 五、Service 层迁移状态

### ✅ 已完成 (15/15)

| Service | 状态 |
|---------|------|
| auth/ | ✅ 完整 |
| chunk/ | ✅ 完整 |
| agent/ | ✅ 完整 |
| chat/ | ✅ 完整 |
| knowledge/ | ✅ 完整 |
| rag/ | ✅ 完整 |
| retriever/ | ✅ 完整 |
| event/ | ✅ 完整 |
| faq/EntryService | ✅ 完整 |
| initialization/ | ✅ 完整 |
| model/ | ✅ 完整 (2025-01-07) |
| evaluation/ | ✅ 完整 (2025-01-07) |
| mcp/ | ✅ 完整 (2025-01-07) |
| tenant/ | ✅ 完整 (2025-01-07) |
| tag/ | ✅ 完整 (2025-01-07) |
| file/ | ✅ 完整 (2025-01-07) |
| dataset/ | ✅ 完整 (2025-01-07) |

---

## 六、Model Handler 功能概览

### ✅ 已完成 (6/6)

| 功能 | 端点 | 状态 |
|------|------|------|
| CreateModel | POST /api/v1/models | ✅ |
| GetModel | GET /api/v1/models/:id | ✅ |
| ListModels | GET /api/v1/models | ✅ |
| UpdateModel | PUT /api/v1/models/:id | ✅ |
| DeleteModel | DELETE /api/v1/models/:id | ✅ |
| ListModelProviders | GET /api/v1/models/providers | ✅ |

**支持的模型类型**:
- `chat_model` - 对话模型
- `embedding` - 向量化模型
- `rerank` - 重排序模型

**支持的提供商**:
- OpenAI
- 阿里云 DashScope
- 智谱 AI
- 本地模型 (Ollama)
- Jina AI

---

## 七、数据库工具说明

### 🔧 工具特性

| 工具 | 功能 | 特性 |
|------|------|------|
| **database_query** | 执行 SQL 查询获取业务数据 | 使用 pg_query 解析器进行安全验证，支持表白名单 |
| **data_analysis** | 使用 DuckDB 分析 CSV/Excel | 会话级内存表，支持只读查询 |
| **data_schema** | 获取数据文件结构 | 从 chunks 提取 schema 信息 |

### 📍 使用方式

数据库工具需要动态创建（需要 sessionID 和 tenantID）：

```go
import databasesvc "github.com/ashwinyue/next-ai/internal/service/database"

// 创建数据库查询工具
queryTool := databasesvc.NewQueryTool(db, tenantID)

// 创建数据分析工具 (需要 sessionID)
analysisTool := databasesvc.NewAnalysisTool(knowledgeRepo, sessionID)

// 创建数据 schema 工具
schemaTool := databasesvc.NewSchemaTool(knowledgeRepo)
```

---

## 八、更新记录

| 日期 | 更新内容 | 负责人 |
|------|----------|--------|
| 2025-01-07 | **内置 Agent 预设完成**：6 种内置 Agent（快速问答、智能推理、深度研究、数据分析、知识图谱专家、文档助手），完整配置预设，不可修改/删除保护，API 端点 | - |
| 2025-01-07 | **系统配置和用户管理验证完成**：系统配置服务 (GetSystemInfo, CheckOllamaStatus, TestEmbedding 等) 已完整实现；用户管理 (Auth) 包含完整 JWT 认证、注册登录、密码修改等功能 | - |
| 2025-01-07 | **文件存储服务完成**：支持本地/MinIO 两种存储后端，统一 FileService 接口，按租户和知识库组织文件 | - |
| 2025-01-07 | **数据集服务完成**：Dataset + QAPair 模型，支持批量导入 QA 对，用于评估和测试 | - |
| 2025-01-07 | **会话标题自动生成完成**：使用 ChatModel 自动生成简短标题，支持降级到默认标题 | - |
| 2025-01-07 | **Session QA 分析完成**：确认 Agent 功能已覆盖，无需单独实现 | - |
| 2025-01-07 | **Message 服务完成**：独立消息管理，LoadMessages（支持分页和时间筛选）、GetMessage、DeleteMessage | - |
| 2025-01-07 | **Tag Handler 完成**：标签 CRUD、分页查询、批量查询，FindOrCreate 支持 | - |
| 2025-01-07 | **SSE 流式响应验证**：现有实现已使用 Eino 最佳实践，无需修改 | - |
| 2025-01-07 | **Agent 模式扩展完成**：支持 quick-answer / smart-reasoning 两种模式，新增 Avatar/IsBuiltin/Temperature/KnowledgeIDs 字段，使用 Eino 原生实现 | - |
| 2025-01-07 | **WeKnora 特有功能分析**：新增章节十，分析 CustomAgent、Tag、Message、流式响应等未迁移功能 | - |
| 2025-01-07 | **Tenant Handler 完成**：租户 CRUD、配置管理、存储信息查询 | - |
| 2025-01-07 | **MCP Service 完成**：MCP 服务管理（CRUD + 测试连接 + 获取工具/资源），使用官方 go-sdk | - |
| 2025-01-07 | **Evaluation Handler 完成**：评估任务创建、结果查询、列表、删除、取消功能 | - |
| 2025-01-07 | **Initialization 增强**：添加 ListOllamaModels, CheckOllamaModels, CheckRemoteModel, CheckRerankModel | - |
| 2025-01-07 | **数据库工具完成**：database_query + data_analysis + data_schema，Agent 工具达到 100%，总进度 100% ✅ | - |
| 2025-01-07 | **Model Handler 完成**：Model 模型 + Repository + Service + Handler + 路由，Handler/Service 层达到 100%，总进度 95% | - |
| 2025-01-07 | Eino 组件集成完成：Router, MultiQuery, Parent, Middlewares | - |
| 2025-01-07 | 初始化功能完成 + 清理 context 服务 | - |
| 2025-01-07 | FAQ 增强功能完成：FAQEntry 模型 + 完整 CRUD + 批量操作 + 导入导出 | - |
| 2025-01-07 | Agent 功能完成：CopyAgent, GetPlaceholders | - |
| 2025-01-07 | Agent 工具补充完成：todo_write 实现，httprequest 替代 web_fetch | - |
| 2025-01-07 | Chunk 功能迁移完成，Handler/Service 完成 | - |
| 2025-01-07 | Auth 功能迁移完成 | - |
| 2025-01-07 | 初始创建，完成 50% 迁移评估 | - |

---

## 九、使用说明

**每次迁移功能时：**
1. 更新对应的状态（✅/⚠️/❌）
2. 在"更新记录"中添加条目
3. 更新进度总览

**检查命令：**
```bash
# 查看已迁移的 handler
ls -la internal/handler/

# 查看 Eino 组件
ls -la internal/service/rag/
ls -la internal/service/agent/middleware.go

# 查看数据库工具
ls -la internal/service/database/
```

---

## 十、WeKnora 特有功能分析（未迁移）

### 📊 功能对比概览

| WeKnora 功能 | 描述 | Next-AI 状态 | 优先级 |
|-------------|------|-------------|--------|
| **custom_agent 模式** | 多种 Agent 模式（quick-answer, smart-reasoning） | ✅ **已实现 (2025-01-07)** | - |
| **内置 Agent** | 6 种内置 Agent（快速回答、智能推理、深度研究等） | ✅ **已实现 (2025-01-07)** | - |
| **Tag 管理** | 知识库标签 CRUD + Chunk 关联 | ✅ **已实现 (2025-01-07)** | - |
| **Message 服务** | 独立消息管理（加载历史、分页、时间筛选） | ✅ **已实现 (2025-01-07)** | - |
| **Session QA** | 知识库 QA 问答专用接口 | ✅ **已由 Agent 覆盖** | - |
| **流式响应** | SSE 流式输出 + Agent 流式执行 | ✅ **已实现 (2025-01-07)** | - |
| **会话标题** | 自动生成会话标题 | ✅ **已实现 (2025-01-07)** | - |
| **文件存储** | 统一文件存储抽象 (本地/MinIO) | ✅ **已实现 (2025-01-07)** | - |
| **数据集管理** | Dataset 管理 | ✅ **已实现 (2025-01-07)** | - |
| **知识图谱** | Chunk 关系计算 + 实体关系提取 | ❌ 未实现 | 🟢 低 |
| **系统配置** | 系统级配置管理 | ❌ 未实现 | 🟢 低 |
| **用户管理** | 用户 CRUD | ❌ 未实现 | 🟢 低 |

### ✅ 已实现：CustomAgent 模式 (2025-01-07)

**新增字段：**
```go
type Agent struct {
    ID           string
    Name         string
    Description  string
    Avatar       string        // ✅ 头像/图标
    IsBuiltin    bool          // ✅ 是否内置 Agent
    AgentMode    string        // ✅ Agent 模式 (quick-answer / smart-reasoning)
    SystemPrompt string
    ModelConfig  ModelConfig
    Tools        JSON
    MaxIter      int
    Temperature  float64       // ✅ 温度参数
    KnowledgeIDs pq.StringArray // ✅ 关联的知识库 ID
    // ...
}
```

**模式实现：**
- `quick-answer` → `createChatModelAgent()` - RAG 快速问答
- `smart-reasoning` → `createReactAgent()` - ReAct 多步推理

**使用 Eino 原生实现：**
- `adk.NewChatModelAgent` - 两种模式底层都用 Eino ADK
- 无需自定义 Agent 框架

### ✅ 已实现：内置 Agent 预设 (2025-01-07)

**6 种内置 Agent 已完整实现：**
- `builtin-quick-answer` - RAG 快速问答 (⚡)
- `builtin-smart-reasoning` - ReAct 多步推理 (🧠)
- `builtin-deep-researcher` - 深度研究 (🔬)
- `builtin-data-analyst` - 数据分析 (📊)
- `builtin-knowledge-graph-expert` - 知识图谱专家 (🕸️)
- `builtin-document-assistant` - 文档助手 (📄)

**保护机制：**
- 内置 Agent 不允许修改核心配置 (`UpdateAgent` 检查)
- 内置 Agent 不允许删除 (`DeleteAgent` 检查)
- 复制内置 Agent 会创建非内置副本

**API 端点：**
- `GET /api/v1/agents/builtin` - 列出内置 Agent
- `POST /api/v1/agents/builtin/init` - 初始化/更新内置 Agent

**说明：** Session QA 功能已被 Agent 的 RunAgent/StreamAgent 功能覆盖，无需单独实现。Agent 已支持：
- 知识库检索（通过 KnowledgeIDs）
- 工具调用和多步推理
- RAG 快速问答模式

#### Tag 管理
- WeKnora: 独立的 `tag.go` Handler + Service
- Next-AI: 部分功能在 knowledge_handler.go
- 缺失: 独立的标签 CRUD、Chunk 与标签关联

#### Message 服务
- WeKnora: 独立的 `message.go` Handler + Service
- 功能: 按时间加载历史、分页、时间筛选
- Next-AI: 消息功能集成在 chat_handler.go ✅ 已完成

#### 会话标题
- WeKnora: `session/title.go` - LLM 自动生成会话标题
- Next-AI: 已实现 ✅ - 使用 ChatModel 生成，支持降级到默认标题

#### 流式响应
- WeKnora: `session/stream.go`, `agent_stream_handler.go`
- 功能: SSE 流式输出、Agent 流式执行
- Next-AI: 未实现

#### 文件存储
- WeKnora: `file/` 目录 (cos.go, minio.go, local.go)
- 功能: 支持多种文件存储后端
- Next-AI: 文件处理集成在 knowledge 服务

### 🟢 低优先级功能

| 功能 | WeKnora 文件 | 描述 |
|------|-------------|------|
| 知识图谱 | graph.go | Chunk 关系计算、实体提取 |
| 数据集 | dataset.go | 数据集管理 |
| 实体提取 | extract.go | 实体/关系提取 |
| Metric | metric/, metric_hook.go | 指标和钩子 |
| 用户管理 | user.go | 用户 CRUD |
| 系统配置 | system.go | 系统配置管理 |

### 📈 总结

**已迁移**：核心业务功能 (Handler + Service + Agent 工具) ✅ 100%

**未迁移**：WeKnora 特有的高级功能
- 这些功能多为增值功能，不影响核心业务
- 可根据实际需求选择性实现

**建议优先级：**
1. 🔴 CustomAgent 模式扩展（✅ 已完成）
2. 🟡 流式响应（✅ 已验证使用 Eino 最佳实践）
3. 🟡 Tag 管理（✅ 已完成）
4. 🟢 其他功能（按需实现）

---

## 下一步迁移

**核心迁移已完成** ✅ (100%)

可选增强功能（按优先级排序）：

1. **内置 Agent 预设** (✅ 已完成 2025-01-07)
   - ✅ 配置 6 种内置 Agent
   - ✅ 内置 Agent 不可修改/删除保护
   - ✅ API 端点：`GET /agents/builtin`, `POST /agents/builtin/init`

2. **文件存储服务** (✅ 已完成 2025-01-07)
   - MinIO/COS/本地存储抽象
   - 独立的文件管理服务

3. **知识图谱** (🟢 低)
   - Neo4j 集成
   - Chunk 关系计算

---

**🎉 核心迁移完成**：所有业务功能已全部完成 (100%)！未迁移为 WeKnora 特有增值功能。
