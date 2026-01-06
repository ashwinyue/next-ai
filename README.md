# Next-AI

基于 **Gin + GORM** 的简化版 Clean Architecture 实现。

## 架构

```
next-ai/
├── cmd/
│   └── next-ai/          # 应用入口
│       └── main.go
├── internal/
│   ├── handler/          # HTTP 处理器 (Controller 层)
│   ├── service/          # 业务逻辑层
│   ├── repository/       # 数据访问层
│   ├── model/            # 数据模型
│   ├── middleware/       # 中间件
│   ├── router/           # 路由配置
│   ├── config/           # 配置
│   └── database/         # 数据库初始化
└── configs/
    └── config.yaml       # 配置文件
```

## 分层说明

| 层次 | 职责 | 示例 |
|------|------|------|
| **handler** | HTTP 请求/响应处理 | 绑定参数、调用 service、返回 JSON |
| **service** | 业务逻辑 | 数据校验、事务协调、调用 repository |
| **repository** | 数据访问 | CRUD 操作、数据库查询 |
| **model** | 数据模型 | GORM 模型定义 |

## API 接口

### 聊天 (Chat)
- `POST /api/v1/chats` - 创建会话
- `GET /api/v1/chats` - 列出会话
- `GET /api/v1/chats/:id` - 获取会话
- `PUT /api/v1/chats/:id` - 更新会话
- `DELETE /api/v1/chats/:id` - 删除会话
- `POST /api/v1/chats/:id/messages` - 发送消息
- `GET /api/v1/chats/:id/messages` - 获取消息

### Agent
- `POST /api/v1/agents` - 创建Agent
- `GET /api/v1/agents` - 列出Agent
- `GET /api/v1/agents/active` - 列出活跃Agent
- `GET /api/v1/agents/:id` - 获取Agent
- `PUT /api/v1/agents/:id` - 更新Agent
- `DELETE /api/v1/agents/:id` - 删除Agent
- `GET /api/v1/agents/:id/config` - 获取Agent配置

### 知识库 (Knowledge)
- `POST /api/v1/knowledge-bases` - 创建知识库
- `GET /api/v1/knowledge-bases` - 列出知识库
- `GET /api/v1/knowledge-bases/:id` - 获取知识库
- `PUT /api/v1/knowledge-bases/:id` - 更新知识库
- `DELETE /api/v1/knowledge-bases/:id` - 删除知识库
- `POST /api/v1/knowledge-bases/:kb_id/documents` - 上传文档
- `GET /api/v1/knowledge-bases/:kb_id/documents` - 列出文档

### 文档 (Document)
- `GET /api/v1/documents/:id` - 获取文档
- `DELETE /api/v1/documents/:id` - 删除文档

### 工具 (Tool)
- `POST /api/v1/tools` - 注册工具
- `GET /api/v1/tools` - 列出工具
- `GET /api/v1/tools/active` - 列出活跃工具
- `GET /api/v1/tools/:id` - 获取工具
- `PUT /api/v1/tools/:id` - 更新工具
- `DELETE /api/v1/tools/:id` - 注销工具

### FAQ
- `POST /api/v1/faqs` - 创建FAQ
- `GET /api/v1/faqs` - 列出FAQ
- `GET /api/v1/faqs/active` - 列出活跃FAQ
- `GET /api/v1/faqs/search` - 搜索FAQ
- `GET /api/v1/faqs/:id` - 获取FAQ
- `PUT /api/v1/faqs/:id` - 更新FAQ
- `DELETE /api/v1/faqs/:id` - 删除FAQ

## 运行

```bash
cd next-ai
go mod tidy
go run cmd/next-ai/main.go
```

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `CONFIG_PATH` | 配置文件路径 | `./configs/config.yaml` |
| `NEXT_AI_DATABASE_HOST` | 数据库主机 | `localhost` |
| `NEXT_AI_DATABASE_PORT` | 数据库端口 | `5432` |
| `NEXT_AI_DATABASE_USER` | 数据库用户 | `postgres` |
| `NEXT_AI_DATABASE_PASSWORD` | 数据库密码 | `` |
| `NEXT_AI_DATABASE_DBNAME` | 数据库名 | `next_ai` |

## 响应格式

```json
{
  "code": 0,
  "message": "success",
  "data": { ... }
}
```

分页响应：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [ ... ],
    "total": 100,
    "page": 1,
    "size": 20
  }
}
```
