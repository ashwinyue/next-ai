package router

import (
	"github.com/ashwinyue/next-ai/internal/handler"
	"github.com/ashwinyue/next-ai/internal/middleware"
	"github.com/ashwinyue/next-ai/internal/service"
	"github.com/gin-gonic/gin"
)

// SetupRouter 设置路由
func SetupRouter(h *handler.Handlers, svc *service.Services) *gin.Engine {
	r := gin.New()

	// 中间件
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.LoggingMiddleware())
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.AuthMiddleware(svc))

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1
	v1 := r.Group("/api/v1")
	{
		// Auth 认证（公开路由）
		auth := v1.Group("/auth")
		{
			auth.POST("/register", h.Auth.Register)
			auth.POST("/login", h.Auth.Login)
			auth.POST("/refresh", h.Auth.RefreshToken)
			auth.GET("/validate", h.Auth.ValidateToken)

			// 需要认证的路由
			auth.POST("/logout", h.Auth.Logout)
			auth.GET("/me", h.Auth.GetCurrentUser)
			auth.POST("/change-password", h.Auth.ChangePassword)
		}

		// Sessions 聊天会话（WeKnora API 兼容）
		sessions := v1.Group("/sessions")
		{
			sessions.POST("", h.Chat.CreateSession)
			sessions.GET("", h.Chat.ListSessions)
			sessions.GET("/:id", h.Chat.GetSession)
			sessions.PUT("/:id", h.Chat.UpdateSession)
			sessions.DELETE("/:id", h.Chat.DeleteSession)
			sessions.POST("/:id/messages", h.Chat.SendMessage)
			sessions.GET("/:id/messages", h.Chat.GetMessages)
			sessions.POST("/:id/title", h.Chat.GenerateTitle)
		}

		// WeKnora API 兼容 - 聊天接口
		v1.POST("/knowledge-chat/:session_id", h.Chat.KnowledgeChat)
		v1.POST("/agent-chat/:session_id", h.Chat.AgentChat)
		v1.POST("/knowledge-search", h.Chat.KnowledgeSearch)

		// Messages 消息管理（独立接口）
		messages := v1.Group("/messages")
		{
			messages.GET("/:session_id/load", h.Chat.LoadMessages)
			messages.GET("/:id", h.Chat.GetMessage)
			messages.DELETE("/:session_id/:id", h.Chat.DeleteMessage)
		}

		// Agent 智能体
		agents := v1.Group("/agents")
		{
			agents.POST("", h.Agent.CreateAgent)
			agents.GET("", h.Agent.ListAgents)
			agents.GET("/active", h.Agent.ListActiveAgents)
			agents.GET("/builtin", h.Agent.ListBuiltinAgents)
			agents.POST("/builtin/init", h.Agent.InitBuiltinAgents)
			agents.GET("/placeholders", h.Agent.GetPlaceholders)
			agents.GET("/:id", h.Agent.GetAgent)
			agents.GET("/:id/config", h.Agent.GetAgentConfig)
			agents.PUT("/:id", h.Agent.UpdateAgent)
			agents.DELETE("/:id", h.Agent.DeleteAgent)
			agents.POST("/:id/copy", h.Agent.CopyAgent)
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
			kb.POST("/:id/documents", h.Knowledge.UploadDocument)
			kb.GET("/:id/documents", h.Knowledge.ListDocuments)

			// 分块管理
			kb.GET("/:id/chunks", h.Chunk.ListChunksByKnowledgeBaseID)
			kb.DELETE("/:id/chunks", h.Chunk.DeleteChunksByKnowledgeBaseID)

			// 标签管理
			kb.POST("/:id/tags", h.Tag.CreateTag)
			kb.GET("/:id/tags", h.Tag.ListTags)
			kb.GET("/:id/tags/all", h.Tag.GetAllTags)
			kb.GET("/:id/tags/:tag_id", h.Tag.GetTag)
			kb.PUT("/:id/tags/:tag_id", h.Tag.UpdateTag)
			kb.DELETE("/:id/tags/:tag_id", h.Tag.DeleteTag)
		}

		// Document 文档
		docs := v1.Group("/documents")
		{
			docs.GET("/:id", h.Knowledge.GetDocument)
			docs.DELETE("/:id", h.Knowledge.DeleteDocument)
			docs.POST("/:id/process", h.Knowledge.ProcessDocument)

			// 分块管理
			docs.DELETE("/:doc_id/chunks", h.Chunk.DeleteChunksByDocumentID)
		}

		// Chunk 分块
		chunks := v1.Group("/chunks")
		{
			chunks.GET("/:id", h.Chunk.GetChunkByID)
			chunks.PUT("/:id", h.Chunk.UpdateChunk)
			chunks.DELETE("/:id", h.Chunk.DeleteChunk)
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

		// FAQ Entry 增强版
		faqEntries := v1.Group("/faq-entries")
		{
			faqEntries.POST("", h.FAQ.CreateEntry)
			faqEntries.GET("", h.FAQ.ListEntries)
			faqEntries.GET("/search", h.FAQ.SearchEntries)
			faqEntries.GET("/export", h.FAQ.ExportEntries)
			faqEntries.POST("/batch", h.FAQ.BatchUpsert)
			faqEntries.GET("/import/:task_id/progress", h.FAQ.GetImportProgress)
			faqEntries.PUT("/categories/batch", h.FAQ.UpdateEntryCategoryBatch)
			faqEntries.PUT("/fields/batch", h.FAQ.UpdateEntryFieldsBatch)
			faqEntries.DELETE("/batch", h.FAQ.DeleteEntries)
			faqEntries.GET("/:id", h.FAQ.GetEntry)
			faqEntries.PUT("/:id", h.FAQ.UpdateEntry)
			faqEntries.DELETE("/:id", h.FAQ.DeleteEntry)
		}

		// RAG 检索
		ragGroup := v1.Group("/rag")
		{
			ragGroup.POST("/retrieve", h.RAG.Retrieve)
			ragGroup.GET("/search", h.RAG.RetrieveSimple)
		}

		// Initialization 初始化
		initGroup := v1.Group("/initialization")
		{
			initGroup.GET("/system/info", h.Initialization.GetSystemInfo)
			initGroup.GET("/ollama/status", h.Initialization.CheckOllamaStatus)
			initGroup.GET("/ollama/models", h.Initialization.ListOllamaModels)
			initGroup.POST("/ollama/models/check", h.Initialization.CheckOllamaModels)
			initGroup.POST("/test/embedding", h.Initialization.TestEmbedding)
			initGroup.POST("/models/remote/check", h.Initialization.CheckRemoteModel)
			initGroup.POST("/models/rerank/check", h.Initialization.CheckRerankModel)
			initGroup.GET("/kb/:kbId/config", h.Initialization.GetKBConfig)
			initGroup.PUT("/kb/:kbId/config", h.Initialization.UpdateKBConfig)
			initGroup.POST("/kb/:kbId", h.Initialization.InitializeByKB)
		}

		// Model 模型管理
		models := v1.Group("/models")
		{
			models.POST("", h.Model.CreateModel)
			models.GET("", h.Model.ListModels)
			models.GET("/providers", h.Model.ListModelProviders)
			models.GET("/:id", h.Model.GetModel)
			models.PUT("/:id", h.Model.UpdateModel)
			models.DELETE("/:id", h.Model.DeleteModel)
		}

		// Evaluation 评估
		eval := v1.Group("/evaluations")
		{
			eval.POST("", h.Evaluation.CreateEvaluation)
			eval.GET("", h.Evaluation.ListEvaluations)
			eval.GET("/result", h.Evaluation.GetEvaluationResult)
			eval.DELETE("/:id", h.Evaluation.DeleteEvaluation)
			eval.POST("/:id/cancel", h.Evaluation.CancelEvaluation)
		}

		// MCP 服务管理
		mcpServices := v1.Group("/mcp-services")
		{
			mcpServices.POST("", h.MCPService.CreateMCPService)
			mcpServices.GET("", h.MCPService.ListMCPServices)
			mcpServices.GET("/:id", h.MCPService.GetMCPService)
			mcpServices.PUT("/:id", h.MCPService.UpdateMCPService)
			mcpServices.DELETE("/:id", h.MCPService.DeleteMCPService)
			mcpServices.POST("/:id/test", h.MCPService.TestMCPService)
			mcpServices.GET("/:id/tools", h.MCPService.GetMCPServiceTools)
			mcpServices.GET("/:id/resources", h.MCPService.GetMCPServiceResources)
		}

		// Tenant 租户管理
		tenants := v1.Group("/tenants")
		{
			tenants.POST("", h.Tenant.CreateTenant)
			tenants.GET("", h.Tenant.ListTenants)
			tenants.GET("/:id", h.Tenant.GetTenant)
			tenants.PUT("/:id", h.Tenant.UpdateTenant)
			tenants.DELETE("/:id", h.Tenant.DeleteTenant)
			tenants.GET("/:id/config", h.Tenant.GetTenantConfig)
			tenants.PUT("/:id/config", h.Tenant.UpdateTenantConfig)
			tenants.GET("/:id/storage", h.Tenant.GetTenantStorage)
		}

		// File 文件管理
		files := v1.Group("/files")
		{
			files.POST("/upload", h.File.UploadFile)
			files.GET("/:id", h.File.GetFile)
			files.GET("/:id/url", h.File.GetFileURL)
			files.DELETE("/:id", h.File.DeleteFile)
			files.GET("/knowledge/:knowledge_id", h.File.ListFilesByKnowledge)
		}

		// Dataset 数据集管理
		datasets := v1.Group("/datasets")
		{
			datasets.POST("", h.Dataset.CreateDataset)
			datasets.GET("", h.Dataset.ListDatasets)
			datasets.GET("/:id", h.Dataset.GetDataset)
			datasets.PUT("/:id", h.Dataset.UpdateDataset)
			datasets.DELETE("/:id", h.Dataset.DeleteDataset)

			// QA 对管理
			datasets.POST("/:dataset_id/qapairs", h.Dataset.CreateQAPair)
			datasets.POST("/:dataset_id/qapairs/batch", h.Dataset.CreateQAPairsBatch)
			datasets.GET("/:dataset_id/qapairs", h.Dataset.GetQAPairs)
			datasets.GET("/qapairs/:id", h.Dataset.GetQAPair)
		}
	}

	return r
}
