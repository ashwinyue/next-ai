module github.com/ashwinyue/next-rag/next-ai

go 1.23.3

require (
	github.com/cloudwego/eino v0.7.17
	github.com/cloudwego/eino-ext/components/model/openai v0.1.7
	github.com/cloudwego/eino-ext/components/retriever/es8 v0.0.0-20260106124928-46864ab11d94
	github.com/cloudwego/eino-ext/components/tool/duckduckgo/v2 v2.0.0-20260106124928-46864ab11d94
	github.com/elastic/go-elasticsearch/v8 v8.16.0
	github.com/gin-gonic/gin v1.11.0
	github.com/google/uuid v1.6.0
	github.com/redis/go-redis/v9 v9.17.2
	github.com/spf13/viper v1.21.0
	gorm.io/driver/postgres v1.6.0
	gorm.io/gorm v1.31.1

	github.com/cloudwego/eino-ext/components/document/parser/docx v0.0.0-20260106124928-46864ab11d94
	github.com/cloudwego/eino-ext/components/document/parser/pdf v0.0.0-20260106124928-46864ab11d94
	github.com/cloudwego/eino-ext/components/document/transformer/splitter/recursive v0.0.0-20260106124928-46864ab11d94
	github.com/cloudwego/eino-ext/components/embedding/dashscope v0.0.0-20260106124928-46864ab11d94
)
