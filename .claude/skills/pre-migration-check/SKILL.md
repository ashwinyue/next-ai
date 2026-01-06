---
name: pre-migration-check
description: Check if eino/eino-ext has existing components before implementing. Use this skill before migrating WeKnora features to avoid duplicating work. Search eino-ext and eino-examples for ready-made components (ChatModel, Embedding, Retriever, Tools), then decide: use existing component, reference example code, or implement custom.
---

重构前检查：Eino 组件可用性 Skill。在迁移 WeKnora 功能前，先检查 eino/eino-ext/eino-examples 是否已有现成组件可用。

## 检查流程

### 第一步：明确要迁移的功能

```
示例：迁移 Auth 认证功能
```

### 第二步：搜索 eino-ext 是否有现成组件

```bash
# 在 eino-ext 中搜索关键词
rg "auth" old/eino-ext/
rg "jwt" old/eino-ext/
rg "token" old/eino-ext/

# 搜索更通用的关键词
rg "middleware" old/eino-ext/
```

### 第三步：搜索 eino-examples 是否有类似实现

```bash
# 在 eino-examples 中搜索
rg "auth" old/eino-examples/
rg "login" old/eino-examples/
```

### 第四步：决策

| 情况 | 行动 |
|------|------|
| eino-ext 有组件 | ✅ 直接使用，不自己实现 |
| eino-examples 有示例 | ✅ 参考示例代码实现 |
| 两者都没有 | ⚠️ 参考 WeKnora 逻辑，自己实现 |

---

## 常见功能对照表

| 功能类别 | WeKnora | eino-ext/eino | 建议 |
|----------|---------|---------------|------|
| **认证** | 自定义 JWT | ❌ 无 | 需自己实现 |
| **Agent** | 自定义 ReAct | ✅ adk.NewChatModelAgent | 用 eino |
| **ChatModel** | 自定义封装 | ✅ openai.NewChatModel | 用 eino-ext |
| **Embedding** | 自定义封装 | ✅ dashscope.NewEmbedder | 用 eino-ext |
| **Retriever** | 自定义 ES | ✅ es8.NewRetriever | 用 eino-ext |
| **Parser** | 自定义 | ✅ pdf.NewPDFParser | 用 eino-ext |
| **Splitter** | 自定义 | ✅ recursive.NewSplitter | 用 eino-ext |
| **Tool** | 自定义框架 | ✅ tool.InvokableTool | 用 eino |
| **Tool-WebSearch** | 自定义 | ✅ duckduckgo.NewTextSearchTool | 用 eino-ext |
| **Tool-Wikipedia** | - | ✅ wikipedia.NewTool | 用 eino-ext |
| **Tool-SequentialThinking** | 自定义 | ✅ sequentialthinking.NewTool | 用 eino-ext |
| **Context管理** | 自定义 | ⚠️ adk.AddSessionValue | 用 eino |
| **事件** | 自定义 EventBus | ⚠️ 回调机制 | 参考eino |

---

## 检查命令速查

```bash
# 搜索组件
rg "New\w+(" old/eino-ext/components/

# 列出所有可用组件
ls old/eino-ext/components/

# 搜索工具
ls old/eino-ext/components/tool/

# 搜索示例
ls old/eino-examples/
```

---

## 原则

> **永远优先使用 eino 组件，而不是自己实现。**

如果有疑问，先查看：
1. `old/eino-ext/components/` - 扩展组件
2. `old/eino-examples/` - 官方示例
3. `old/eino/components/` - 核心组件接口
