---
name: lookup-weknora
description: Find WeKnora business implementation as refactoring reference. Use this skill when implementing features (Agent, Chat, Knowledge) to understand how WeKnora handles them. Search with ripgrep, extract business logic (data models, API design, core flow), but ignore WeKnora's custom LLM/Embedding/Tool wrappers - replace with Eino components.
---

查找 WeKnora 业务实现 Skill。在 old/WeKnora/ 目录中查找特定功能的实现，作为重构参考。

## 使用场景
当你需要实现某个功能（如 Agent、Chat、Knowledge），想了解 WeKnora 中是如何实现的。

## 操作步骤

### 1. 定位功能模块
WeKnora 目录结构：
```
old/WeKnora/
├── internal/
│   ├── agent/          # Agent 服务
│   ├── chat/           # Chat 服务
│   ├── knowledge/      # 知识库服务
│   ├── document/       # 文档处理
│   ├── retrieval/      # 检索服务
│   └── llm/            # LLM 调用
```

### 2. 搜索相关代码
使用 ripgrep 搜索关键词：
```bash
# 在 WeKnora 中搜索函数名
rg "func.*CreateAgent" old/WeKnora/

# 搜索结构体定义
rg "type Agent struct" old/WeKnora/

# 搜索特定功能
rg "ToolCall" old/WeKnora/
```

### 3. 提取业务逻辑
阅读 WeKnora 代码时，关注：
- **业务流程**：核心逻辑是什么
- **数据模型**：如何组织数据
- **API 设计**：接口定义

### 4. 忽略的实现细节
以下 WeKnora 的实现方式**不要**复制：
- 自定义的 LLM 封装 → 用 Eino ChatModel 替换
- 自定义的 Embedding → 用 eino-ext 替换
- 自定义的 Vector Store → 用 eino-ext ES8 替换
- 自定义的 Tool 框架 → 用 Eino Tool 接口替换

## 示例

### 任务：实现 Agent 创建功能
```bash
# 1. 查找 WeKnora 中的 Agent 创建
rg "CreateAgent" old/WeKnora/internal/agent/

# 2. 找到相关文件
old/WeKnora/internal/agent/service.go
old/WeKnora/internal/agent/types.go

# 3. 阅读：了解需要哪些字段、验证逻辑

# 4. 在 Next-AI 中按 Eino 方式重写
```

## 注意事项
- WeKnora 代码仅作为**功能参考**，不要直接复制
- 所有 AI 组件必须用 Eino 标准实现
- 架构采用简洁的三层结构，不使用 DDD
