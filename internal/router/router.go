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

			// 会话流控制（WeKnora API 兼容）
			sessions.POST("/:id/stop", h.Chat.StopSession)
			sessions.GET("/continue-stream/:id", h.Chat.ContinueStream)
		}

		// WeKnora API 兼容 - 聊天接口
		v1.POST("/agent-chat/:session_id", h.Chat.AgentChat)

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

		// Initialization 初始化
		initGroup := v1.Group("/initialization")
		{
			initGroup.GET("/system/info", h.Initialization.GetSystemInfo)
			initGroup.GET("/ollama/status", h.Initialization.CheckOllamaStatus)
			initGroup.GET("/ollama/models", h.Initialization.ListOllamaModels)
			initGroup.POST("/ollama/models/check", h.Initialization.CheckOllamaModels)
			// Ollama 模型下载（WeKnora API 兼容）
			initGroup.POST("/ollama/models/download", h.Initialization.DownloadModel)
			initGroup.GET("/ollama/download/progress/:task_id", h.Initialization.GetDownloadProgress)
			initGroup.GET("/ollama/download/tasks", h.Initialization.ListDownloadTasks)
			initGroup.POST("/ollama/download/cancel/:task_id", h.Initialization.CancelDownload)
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

			// 租户 KV 配置（WeKnora API 兼容）
			tenants.GET("/kv/:key", h.Tenant.GetTenantKV)
			tenants.PUT("/kv/:key", h.Tenant.UpdateTenantKV)
			tenants.GET("/all", h.Tenant.ListAllTenants)
			tenants.GET("/search", h.Tenant.SearchTenants)
		}

		// File 文件管理
		files := v1.Group("/files")
		{
			files.POST("/upload", h.File.UploadFile)
			files.GET("/:id", h.File.GetFile)
			files.GET("/:id/url", h.File.GetFileURL)
			files.DELETE("/:id", h.File.DeleteFile)
		}

		// System 系统管理（WeKnora API 兼容）
		system := v1.Group("/system")
		{
			system.GET("/info", h.System.GetSystemInfo)
			system.GET("/minio/buckets", h.System.ListMinioBuckets)
		}

		// WebSearch 网络搜索（WeKnora API 兼容）
		v1.GET("/web-search/providers", h.System.GetWebSearchProviders)
		v1.POST("/web-search/search", h.WebSearch.Search)
	}

	return r
}
