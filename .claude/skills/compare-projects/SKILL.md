---
name: compare-projects
description: 对照学习 WeKnora 和 Next-AI 两个项目，了解重构前后的差异。Next-AI 是 WeKnora 的重构版本，采用更简洁的架构和 Eino 框架。
---

对照学习 WeKnora 和 Next-AI 技能。

## 使用场景
当你想了解 WeKnora 如何重构为 Next-AI，或需要比较两个项目在特定功能上的实现差异。

## 文档结构
对照学习文档位于 `compare/` 目录，按章节组织：

- `01-architecture.md` - 架构设计对比
- `02-directory.md` - 目录结构对比
- `03-handler.md` - Handler 层对比
- `04-service.md` - Service 层对比
- `05-model.md` - Model 层对比
- `06-agent.md` - Agent 功能实现对比
- `07-chat.md` - Chat 功能实现对比
- `08-knowledge.md` - 知识库功能实现对比
- `09-dependency.md` - 依赖注入方式对比
- `10-error.md` - 错误处理对比

## 快速查阅

### 查看特定章节
```bash
# 查看架构对比
cat compare/01-architecture.md

# 查看具体功能实现对比
cat compare/06-agent.md
```

### 对比两个文件
```bash
# WeKnora 源文件
old/WeKnora/internal/agent/service.go

# Next-AI 重构后
internal/service/agent/agent.go
```

## 核心差异速查

| 方面 | WeKnora | Next-AI |
|------|---------|---------|
| 架构 | DDD + 分层架构 | 简洁三层架构 |
| 依赖注入 | Wire | 手动初始化 |
| AI框架 | 自研封装 | Eino 标准组件 |
| 目录组织 | 按领域分包 | 按层次分包 |
| 错误处理 | 自定义错误包 | fmt.Errorf + %w |
