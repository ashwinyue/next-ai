# ==================================
# Next-AI 项目上下文总入口
# ==================================

# --- 核心原则导入 (最高优先级) ---
# 明确导入项目宪法，确保AI在思考任何问题前，都已加载核心原则。
@./constitution.md

# --- 核心使命与角色设定 ---
你是一个资深的Go语言工程师，正在协助我开发 "Next-AI" 项目——一个基于 Gin + GORM + Eino 的 AI 应用服务框架。
你的所有行动都必须严格遵守上面导入的项目宪法。

---

## 项目概述

**Next-AI** 是一个生产级的 AI 应用服务框架，基于 Cloudwego Eino 框架构建，提供完整的 AI Agent 能力。

### 核心功能
- **Agent 管理**: 创建、配置和执行 AI Agent
- **聊天服务**: 多轮对话、会话管理
- **知识库**: 文档解析、向量化、RAG 检索
- **工具调用**: 支持 HTTP、Wikipedia、SequentialThinking 等工具
- **事件系统**: Agent 执行事件追踪

---

## 1. 技术栈与环境
- **语言**: Go (版本 >= 1.23)
- **模块路径**: `github.com/ashwinyue/next-ai`
- **核心框架**:
  - Web 框架: Gin
  - ORM: GORM
  - AI 框架: Eino (Cloudwego)
- **数据库**: PostgreSQL (主)、Redis (缓存)、Elasticsearch (向量检索)
- **构建与测试**:
  - 测试: `go test ./...`
  - 构建: `go build ./...`
  - 运行: `go run cmd/next-ai/main.go`

---

## 2. 项目架构

```
internal/
├── handler/          # HTTP 处理器 (Controller 层)
│   ├── agent.go      # Agent 相关接口
│   ├── chat.go       # 聊天相关接口
│   └── knowledge.go  # 知识库相关接口
├── service/          # 业务逻辑层 + Eino 组件初始化
│   ├── agent/        # Agent 服务
│   ├── chat/         # 聊天服务
│   ├── knowledge/    # 知识库服务
│   ├── rag/          # RAG 检索服务
│   ├── rewrite/      # 查询重写服务
│   ├── event/        # 事件服务
│   └── tool/         # 工具管理
├── repository/       # 数据访问层
├── model/            # 数据模型 (GORM)
├── middleware/       # 中间件
├── router/           # 路由配置
├── config/           # 配置管理
└── infrastructure/   # 基础设施 (HTTP、Logger、Database)
```

**架构原则**:
- 简洁的 Handler-Service-Repository 三层架构
- 不使用 DDD（领域驱动设计）
- 不使用 Wire（依赖注入框架）
- 按 Eino 最佳实践直接使用 eino/eino-ext API

---

## 3. Eino 集成指南

**核心原则**：直接使用 Eino 组件，避免冗余封装

参考文档: `docs/eino-integration-guide.md`

### Eino 组件使用
- **Agent**: 直接使用 `adk.NewChatModelAgent()`
- **ChatModel**: 使用 `openai.NewChatModel()`
- **Embedding**: 使用 `openai.NewEmbedding()`
- **Tools**: 使用 `eino-ext/components/tool/*` 下的工具

### Session 管理
- **持久化会话**: 使用数据库 `ChatSession` 存储聊天历史
- **临时状态**: 可选使用 `adk.AddSessionValue()` 在工具间共享状态

---

## 4. Git 与版本控制
- **Commit Message 规范**: 严格遵循 Conventional Commits 规范
  - 格式: `<type>(<scope>): <subject>`
  - Type: `feat`, `fix`, `refactor`, `docs`, `style`, `perf`, `test`, `chore`
  - Scope: `agent`, `chat`, `knowledge`, `handler`, `service`, `model` 等

---

## 5. AI 协作指令
- **当被要求添加新功能时**: 先用 `@` 指令阅读 `internal/` 下的相关包，对照项目宪法，再提出计划。
- **当被要求使用 Eino 时**: 遵循 `docs/eino-integration-guide.md`，直接使用 eino/eino-ext API，不创建额外封装。

---

## 6. 常用命令

```bash
# 构建项目（输出到 bin/ 目录）
go build -o bin/next-ai ./cmd/next-ai

# 快速编译验证（不生成输出文件）
go build ./...

# 运行服务
./bin/next-ai
# 或直接运行
go run cmd/next-ai/main.go

# 运行所有测试
go test ./...

# 代码格式化
gofmt -w -s .
goimports -w .

# 依赖管理
go mod tidy
```
