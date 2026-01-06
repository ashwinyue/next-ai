---
name: refactor-to-nextai
description: Complete workflow for refactoring WeKnora features to Next-AI following Eino best practices. Step 0: pre-migration-check (search eino-ext). Step 1: lookup-weknora (understand business logic). Step 2: study-eino-examples (learn patterns). Step 3: use-eino-ext (select components). Step 4: implement with 3-layer architecture. Step 5: build and verify. Step 6: update migration progress.
---

WeKnora → Next-AI 重构流程 Skill。完整的重构工作流程，将 WeKnora 的功能按 Eino 最佳实践重构到 Next-Ai。

## 重构流程

### 第 0 步：检查 Eino 组件可用性（pre-migration-check）⭐
```bash
# 1. 在 eino-ext 中搜索现成组件
rg "关键词" old/eino-ext/

# 2. 在 eino-examples 中搜索示例
rg "关键词" old/eino-examples/

# 3. 列出可用组件
ls old/eino-ext/components/
```

**决策**：
- ✅ eino-ext 有组件 → 直接使用
- ✅ eino-examples 有示例 → 参考实现
- ⚠️ 两者都没有 → 参考 WeKnora 逻辑自己实现

---

### 第一步：理解功能（lookup-weknora）
```bash
# 1. 在 WeKnora 中定位功能
rg "func.*XXX" old/WeKnora/

# 2. 找到相关文件
# 3. 理解业务逻辑（数据模型、API 设计、核心流程）
```

**关注**：功能做什么、需要什么输入、产生什么输出

**忽略**：WeKnora 的具体实现方式（LLM 封装、自定义组件等）

---

### 第二步：学习最佳实践（study-eino-examples）
```bash
# 1. 在 eino-examples 中找类似功能
ls old/eino-examples/

# 2. 阅读示例代码
# 3. 学习 Eino 的标准模式
```

**学习**：如何使用 Eino 组件、代码组织方式

---

### 第三步：选择组件（use-eino-ext）
```bash
# 1. 确定需要哪些 eino-ext 组件
# 2. 查看组件文档
# 3. 记录初始化方式
```

**常用组件**：
- ChatModel: `openai.NewChatModel()`
- Embedding: `dashscope.NewEmbedder()`
- Retriever: `es8.NewRetriever()`
- Tool: 实现 `tool.InvokableTool`

---

### 第四步：实现功能
```
在 Next-Ai 中按三层架构实现：

Handler (internal/handler/xxx_handler.go)
    ↓ 接收 HTTP 请求
Service (internal/service/xxx/xxx.go)
    ↓ 业务逻辑 + Eino 组件初始化
Repository (internal/repository/xxx_repository.go)
    ↓ 数据访问
Model (internal/model/xxx.go)
```

**关键原则**：
- ✅ 用 `newXxx()` 函数初始化 Eino 组件
- ✅ 直接使用 eino 类型，不封装
- ✅ 简洁的 New 函数，不用 Wire
- ❌ 不创建 domain/ 目录
- ❌ 不使用工厂模式

---

### 第五步：编译验证
```bash
# 每完成一个小功能就编译
go build ./...

# 发现错误立即修复
```

---

### 第六步：更新迁移进度（migration-progress）⭐

**每次完成功能后，必须更新 `.claude/skills/migration-progress/SKILL.md`：**

1. 更新进度总览（Handler/Service 完成度）
2. 将功能从"待迁移"移到"已完成"
3. 在任务清单中勾选已完成项
4. 添加更新记录

**更新内容：**
```markdown
## 进度总览
| Handler | 8 | 1 | 89% |  # 更新数字

## 一、Handler 层迁移状态
### ✅ 已完成
| **Chunk** | chunk.go | chunk_handler.go | ✅ 完整迁移 (2025-01-07) |  # 添加

## 四、迁移任务清单
- [x] **Chunk Handler + Service** (完成 2025-01-07)  # 勾选

## 五、更新记录
| 2025-01-07 | Chunk 功能迁移完成 | - |  # 添加
```

---

## 快速参考

| 任务 | Skill | 命令 |
|------|-------|------|
| 检查 Eino 组件 | pre-migration-check | `rg "关键词" old/eino-ext/` |
| 查找 WeKnora 实现 | lookup-weknora | `rg "func" old/WeKnora/` |
| 学习 Eino 模式 | study-eino-examples | `ls old/eino-examples/` |
| 使用 Eino 组件 | use-eino-ext | 查看 `old/eino-ext/` |
| 更新迁移进度 | migration-progress | 更新 SKILL.md |

---

## 功能对照表

| WeKnora 组件 | Eino 替换 | 位置 | 状态 |
|--------------|-----------|------|------|
| `weknora/llm/chat.go` | `openai.NewChatModel()` | eino-ext/components/model/openai | ✅ |
| `weknora/embedding/` | `dashscope.NewEmbedder()` | eino-ext/components/embedding/dashscope | ✅ |
| `weknora/vectorstore/` | `es8.NewRetriever()` | eino-ext/components/retriever/es8 | ✅ |
| `weknora/tool/` | `tool.InvokableTool` | eino/components/tool | ✅ |
| `weknora/agent/` | `adk.NewChatModelAgent()` | eino/adk | ✅ |
| `weknora/parser/` | `pdf.NewPDFParser()` | eino-ext/components/document/parser | ✅ |
| `weknora/auth/` | ❌ 无 | 需自己实现 | ✅ 完成 (2025-01-07) |
| `weknora/chunk/` | ❌ 无 | 需自己实现 | ✅ 完成 (2025-01-07) |
