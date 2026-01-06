package router

import (
	"github.com/ashwinyue/next-rag/next-ai/internal/handler"
	"github.com/ashwinyue/next-rag/next-ai/internal/middleware"
	"github.com/gin-gonic/gin"
)

// SetupRouter 设置路由
func SetupRouter(h *handler.Handlers) *gin.Engine {
	r := gin.New()

	// 中间件
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.LoggingMiddleware())
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.AuthMiddleware())

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1
	v1 := r.Group("/api/v1")
	{
		// Chat 聊天
		chats := v1.Group("/chats")
		{
			chats.POST("", h.Chat.CreateSession)
			chats.GET("", h.Chat.ListSessions)
			chats.GET("/:id", h.Chat.GetSession)
			chats.PUT("/:id", h.Chat.UpdateSession)
			chats.DELETE("/:id", h.Chat.DeleteSession)
			chats.POST("/:id/messages", h.Chat.SendMessage)
			chats.GET("/:id/messages", h.Chat.GetMessages)
		}

		// Agent 智能体
		agents := v1.Group("/agents")
		{
			agents.POST("", h.Agent.CreateAgent)
			agents.GET("", h.Agent.ListAgents)
			agents.GET("/active", h.Agent.ListActiveAgents)
			agents.GET("/:id", h.Agent.GetAgent)
			agents.GET("/:id/config", h.Agent.GetAgentConfig)
			agents.PUT("/:id", h.Agent.UpdateAgent)
			agents.DELETE("/:id", h.Agent.DeleteAgent)
			agents.POST("/:id/run", h.Agent.RunAgent)
			agents.POST("/:id/stream", h.Agent.StreamAgent)
		}

		// Knowledge 知识库
		kb := v1.Group("/knowledge-bases")
		{
			kb.POST("", h.Knowledge.CreateKnowledgeBase)
			kb.GET("", h.Knowledge.ListKnowledgeBases)
			kb.GET("/:id", h.Knowledge.GetKnowledgeBase)
			kb.PUT("/:id", h.Knowledge.UpdateKnowledgeBase)
			kb.DELETE("/:id", h.Knowledge.DeleteKnowledgeBase)
			kb.POST("/:kb_id/documents", h.Knowledge.UploadDocument)
			kb.GET("/:kb_id/documents", h.Knowledge.ListDocuments)
		}

		// Document 文档
		docs := v1.Group("/documents")
		{
			docs.GET("/:id", h.Knowledge.GetDocument)
			docs.DELETE("/:id", h.Knowledge.DeleteDocument)
			docs.POST("/:id/process", h.Knowledge.ProcessDocument)
		}

		// Index 索引管理
		index := v1.Group("/index")
		{
			index.POST("/chunks", h.Knowledge.CreateChunkIndex)
		}

		// Tool 工具
		tools := v1.Group("/tools")
		{
			tools.POST("", h.Tool.RegisterTool)
			tools.GET("", h.Tool.ListTools)
			tools.GET("/active", h.Tool.ListActiveTools)
			tools.GET("/:id", h.Tool.GetTool)
			tools.PUT("/:id", h.Tool.UpdateTool)
			tools.DELETE("/:id", h.Tool.UnregisterTool)
		}

		// FAQ 常见问题
		faqs := v1.Group("/faqs")
		{
			faqs.POST("", h.FAQ.CreateFAQ)
			faqs.GET("", h.FAQ.ListFAQs)
			faqs.GET("/active", h.FAQ.ListActiveFAQs)
			faqs.GET("/search", h.FAQ.SearchFAQs)
			faqs.GET("/:id", h.FAQ.GetFAQ)
			faqs.PUT("/:id", h.FAQ.UpdateFAQ)
			faqs.DELETE("/:id", h.FAQ.DeleteFAQ)
		}

		// RAG 检索
		ragGroup := v1.Group("/rag")
		{
			ragGroup.POST("/retrieve", h.RAG.Retrieve)
			ragGroup.GET("/search", h.RAG.RetrieveSimple)
		}
	}

	return r
}
