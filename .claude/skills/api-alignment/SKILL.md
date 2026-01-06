---
name: api-alignment
description: 对齐 WeKnora 接口协议，复用前端。用于确保 Next-AI API 与 WeKnora 前端兼容。
---

# API 接口对齐 Skill

## 目的

对齐 Next-AI 后端 API 与 WeKnora 前端的接口协议，实现前端直接复用。

---

## 响应格式规范

### 成功响应

**WeKnora 格式**（需要兼容）：
```json
{
  "success": true,
  "data": { ... }
}
```

**Next-AI 当前格式**：
```json
{
  "code": 0,
  "message": "success",
  "data": { ... }
}
```

### 错误响应

**WeKnora 格式**：
```json
{
  "code": 400,
  "msg": "error message"
}
```

**Next-AI 当前格式**（已在 `errorResponse` 中实现）：
```json
{
  "code": -1,
  "message": "error message"
}
```

---

## 分页响应格式

### WeKnora 格式
```json
{
  "success": true,
  "data": {
    "items": [...],
    "total": 100,
    "page": 1,
    "size": 20
  }
}
```

### Next-AI 当前格式（已一致）
```go
c.JSON(http.StatusOK, gin.H{
    "code":    0,
    "message": "success",
    "data": gin.H{
        "items": agents,
        "total": total,
        "page":  page,
        "size":  size,
    },
})
```

---

## API 路由对照表

### Agent 智能体

| WeKnora | Next-AI | 状态 |
|---------|---------|------|
| `GET /api/v1/agents` | `GET /api/v1/agents` | ✅ 一致 |
| `POST /api/v1/agents` | `POST /api/v1/agents` | ✅ 一致 |
| `GET /api/v1/agents/:id` | `GET /api/v1/agents/:id` | ✅ 一致 |
| `PUT /api/v1/agents/:id` | `PUT /api/v1/agents/:id` | ✅ 一致 |
| `DELETE /api/v1/agents/:id` | `DELETE /api/v1/agents/:id` | ✅ 一致 |
| - | `GET /api/v1/agents/builtin` | ✅ 新增 |
| - | `POST /api/v1/agents/builtin/init` | ✅ 新增 |

### Chat 聊天

| WeKnora | Next-AI | 状态 |
|---------|---------|------|
| `POST /api/v1/sessions` | `POST /api/v1/chats` | ⚠️ 路径不同 |
| `GET /api/v1/sessions` | `GET /api/v1/chats` | ⚠️ 路径不同 |
| `POST /api/v1/sessions/:id/messages` | `POST /api/v1/chats/:id/messages` | ⚠️ 路径不同 |
| `GET /api/v1/sessions/:id/messages` | `GET /api/v1/chats/:id/messages` | ⚠️ 路径不同 |

### Knowledge 知识库

| WeKnora | Next-AI | 状态 |
|---------|---------|------|
| `GET /api/v1/knowledge-bases` | `GET /api/v1/knowledge-bases` | ✅ 一致 |
| `POST /api/v1/knowledge-bases` | `POST /api/v1/knowledge-bases` | ✅ 一致 |
| `GET /api/v1/knowledge-bases/:id` | `GET /api/v1/knowledge-bases/:id` | ✅ 一致 |
| `PUT /api/v1/knowledge-bases/:id` | `PUT /api/v1/knowledge-bases/:id` | ✅ 一致 |
| `DELETE /api/v1/knowledge-bases/:id` | `DELETE /api/v1/knowledge-bases/:id` | ✅ 一致 |
| `POST /api/v1/knowledge-bases/:id/documents` | `POST /api/v1/knowledge-bases/:kb_id/documents` | ⚠️ 参数名不同 |
| `GET /api/v1/knowledge-bases/:id/documents` | `GET /api/v1/knowledge-bases/:kb_id/documents` | ⚠️ 参数名不同 |

### FAQ 常见问题

| WeKnora | Next-AI | 状态 |
|---------|---------|------|
| `GET /api/v1/faqs` | `GET /api/v1/faqs` | ✅ 一致 |
| `POST /api/v1/faqs` | `POST /api/v1/faqs` | ✅ 一致 |
| `GET /api/v1/faqs/:id` | `GET /api/v1/faqs/:id` | ✅ 一致 |
| `PUT /api/v1/faqs/:id` | `PUT /api/v1/faqs/:id` | ✅ 一致 |
| `DELETE /api/v1/faqs/:id` | `DELETE /api/v1/faqs/:id` | ✅ 一致 |

---

## 需要对齐的接口差异

### 1. 路径差异

**Chat/Sessions**: WeKnora 使用 `/sessions`，Next-AI 使用 `/chats`
- 解决方案：添加路由别名 `sessions` → `chats`

### 2. 参数名差异

**Knowledge Base 文档上传**：
- WeKnora: `:id` (knowledge-base id)
- Next-AI: `:kb_id`
- 解决方案：统一为 `:id` 或添加别名

### 3. 响应格式差异

**成功响应**：WeKnora 使用 `success` 字段，Next-AI 使用 `code` 字段
- 解决方案：统一响应格式（推荐使用 `code` 格式，或在前端适配）

---

## 对齐步骤

### Step 1: 统一响应格式

修改 `internal/handler/handler.go` 中的响应辅助函数：

```go
// 成功响应（WeKnora 兼容格式）
func successWeKnora(c *gin.Context, data interface{}) {
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data":    data,
    })
}

// 错误响应（WeKnora 兼容格式）
func errorWeKnora(c *gin.Context, code int, msg string) {
    c.JSON(code, gin.H{
        "code": code,
        "msg":  msg,
    })
}
```

### Step 2: 添加路由别名

在 `internal/router/router.go` 中添加别名路由：

```go
// Chat 路由（兼容 WeKnora 的 sessions 路径）
sessions := v1.Group("/sessions")
{
    sessions.POST("", h.Chat.CreateSession)
    sessions.GET("", h.Chat.ListSessions)
    sessions.GET("/:id", h.Chat.GetSession)
    sessions.PUT("/:id", h.Chat.UpdateSession)
    sessions.DELETE("/:id", h.Chat.DeleteSession)
    sessions.POST("/:id/messages", h.Chat.SendMessage)
    sessions.GET("/:id/messages", h.Chat.GetMessages)
}
```

### Step 3: 参数名统一

修改 Handler 参数名以匹配 WeKnora：

```go
// WeKnora: POST /api/v1/knowledge-bases/:id/documents
// Next-AI: POST /api/v1/knowledge-bases/:kb_id/documents
// 解决：统一使用 :id
func (h *KnowledgeHandler) UploadDocument(c *gin.Context) {
    kbID := c.Param("id")  // 改为 id 而不是 kb_id
    // ...
}
```

---

## 前端适配建议

如果后端修改困难，可以在前端做适配：

### 方案 A：请求拦截器适配

```typescript
// src/utils/request.ts
const apiPathMap: Record<string, string> = {
  '/api/v1/chats': '/api/v1/sessions',
  '/api/v1/chats/': '/api/v1/sessions/',
  // 其他路径映射...
}

instance.interceptors.request.use((config) => {
  // 路径替换
  for (const [nextAI, weknora] of Object.entries(apiPathMap)) {
    if (config.url?.startsWith(nextAI)) {
      config.url = config.url.replace(nextAI, weknora)
    }
  }
  return config
})
```

### 方案 B：响应拦截器适配

```typescript
instance.interceptors.response.use((response) => {
  // 统一处理 Next-AI 的 code 格式为 WeKnora 的 success 格式
  const { data } = response
  if (data && typeof data === 'object') {
    if ('code' in data && data.code === 0) {
      return { success: true, data: data.data }
    }
  }
  return data
})
```

---

## 快速检查命令

```bash
# 检查 WeKnora 的 API 路由
rg 'Router\.(GET|POST|PUT|DELETE)' old/WeKnora/internal/handler/ -A 1

# 检查 Next-AI 的 API 路由
rg '\.(GET|POST|PUT|DELETE)\("' internal/router/router.go -A 1

# 对比特定 Handler
diff old/WeKnora/internal/handler/custom_agent.go internal/handler/agent_handler.go
```

---

## 实现优先级

| 优先级 | 任务 | 说明 |
|--------|------|------|
| 🔴 高 | 统一响应格式 | `success` vs `code` |
| 🔴 高 | 路由别名 | `/sessions` → `/chats` |
| 🟡 中 | 参数名统一 | `:id` vs `:kb_id` |
| 🟢 低 | 字段名适配 | 其他细微差异 |

---

## 注意事项

1. **向后兼容**：保持 Next-AI 现有 API 不变，添加别名路由
2. **渐进迁移**：优先对齐核心接口（Chat、Agent、Knowledge）
3. **前端优先**：尽量在后端适配，减少前端修改
