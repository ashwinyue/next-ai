---
name: use-eino-ext
description: Quick reference for Eino-Ext extension components. Use this skill when implementing AI features with ChatModel, Embedding, Retriever, Tool, Parser, Splitter. Directly use eino-ext components: openai.NewChatModel, dashscope.NewEmbedder, es8.NewRetriever, duckduckgo.NewTextSearchTool, pdf.NewPDFParser, recursive.NewSplitter. Initialize in internal/service/service.go, no wrappers, no factories.
---

使用 Eino-Ext 扩展组件 Skill。快速查找和使用 eino-ext 提供的扩展组件（OpenAI、DashScope、ES8 等）。

## 使用场景
需要使用具体的 AI 组件实现时，直接使用 eino-ext，**不自己封装**。

## 可用组件

### 1. ChatModel (对话模型)
```go
import "github.com/cloudwego/eino-ext/components/model/openai"

// 直接使用
chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
    APIKey:      apiKey,
    BaseURL:     baseURL,
    Model:       modelName,
    Temperature: &temperature,
})
```

**路径**：`old/eino-ext/components/model/openai/`

### 2. Embedding (向量化)
```go
import "github.com/cloudwego/eino-ext/components/embedding/dashscope"

// 直接使用
embedder, err := dashscope.NewEmbedder(ctx, &dashscope.EmbeddingConfig{
    APIKey: apiKey,
    Model:  "text-embedding-v3",
})
```

**路径**：`old/eino-ext/components/embedding/dashscope/`

### 3. Retriever (检索器)
```go
import "github.com/cloudwego/eino-ext/components/retriever/es8"
import "github.com/cloudwego/eino-ext/components/retriever/es8/search_mode"

// 直接使用
retriever, err := es8.NewRetriever(ctx, &es8.RetrieverConfig{
    Client:     esClient,
    Index:      indexName,
    TopK:       10,
    SearchMode: search_mode.SearchModeDenseVectorSimilarity(...),
    Embedding:  embedder,
})
```

**路径**：`old/eino-ext/components/retriever/es8/`

### 4. Tool (工具)
```go
import duckduckgov2 "github.com/cloudwego/eino-ext/components/tool/duckduckgo/v2"

// 直接使用
searchTool, err := duckduckgov2.NewTextSearchTool(ctx, &duckduckgov2.Config{
    ToolName:  "web_search",
    ToolDesc:  "Search the web",
    MaxResults: 10,
})
```

**路径**：`old/eino-ext/components/tool/duckduckgo/v2/`

### 5. Parser (文档解析)
```go
import "github.com/cloudwego/eino-ext/components/document/parser/pdf"
import "github.com/cloudwego/eino-ext/components/document/parser/docx"

// PDF 解析
pdfParser, err := pdf.NewPDFParser(ctx, &pdf.Config{ToPages: false})

// DOCX 解析
docxParser, err := docx.NewDocxParser(ctx, &docx.Config{
    IncludeHeaders: true,
    IncludeTables:  true,
})
```

**路径**：`old/eino-ext/components/document/parser/`

### 6. Splitter (文档分块)
```go
import "github.com/cloudwego/eino-ext/components/document/transformer/splitter/recursive"

// 直接使用
splitter, err := recursive.NewSplitter(ctx, &recursive.Config{
    ChunkSize:   512,
    OverlapSize: 50,
    Separators:  []string{"\n\n", "\n", ". "},
})
```

**路径**：`old/eino-ext/components/document/transformer/splitter/recursive/`

## 在 Next-AI 中的使用方式

### 初始化位置
所有 Eino 组件的初始化放在 `internal/service/service.go`：
```go
// internal/service/service.go

func newChatModel(ctx context.Context, cfg *config.Config) (model.ChatModel, error) {
    return openai.NewChatModel(ctx, &openai.ChatModelConfig{...})
}

func newEmbedder(ctx context.Context, cfg *config.Config) embedding.Embedder {
    embedder, _ := dashscope.NewEmbedder(ctx, &dashscope.EmbeddingConfig{...})
    return embedder
}
```

### 传递方式
通过 `*service.Services` 或具体服务传递：
```go
type Services struct {
    ChatModel  model.ChatModel
    Embedder   embedding.Embedder
    Retriever  retriever.Retriever
    AllTools   []einotool.BaseTool
}
```

## 禁止事项
❌ **不要**创建 `internal/eino/` 包
❌ **不要**创建工厂类
❌ **不要**封装 eino-ext 组件
✅ **直接**在服务中初始化和使用
