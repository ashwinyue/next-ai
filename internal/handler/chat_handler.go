package handler

import (
	"strconv"

	"github.com/ashwinyue/next-ai/internal/service"
	"github.com/ashwinyue/next-ai/internal/service/agent"
	"github.com/ashwinyue/next-ai/internal/service/chat"
	"github.com/ashwinyue/next-ai/internal/service/rag"
	"github.com/gin-gonic/gin"
)

// ChatHandler 聊天处理器
type ChatHandler struct {
	svc *service.Services
}

// NewChatHandler 创建聊天处理器
func NewChatHandler(svc *service.Services) *ChatHandler {
	return &ChatHandler{svc: svc}
}

// getPagination 获取分页参数 (WeKnora API: page_size)
func getPagination(c *gin.Context) (page, pageSize int) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "10")) // WeKnora 默认 10
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}
	return
}

// getUserID 获取用户ID
func getUserID(c *gin.Context) string {
	if id, exists := c.Get("user_id"); exists {
		if userID, ok := id.(string); ok {
			return userID
		}
	}
	return ""
}

// CreateSession 创建会话
func (h *ChatHandler) CreateSession(c *gin.Context) {
	var req chat.CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	session, err := h.svc.Chat.CreateSession(c.Request.Context(), &req)
	if err != nil {
		Error(c, err)
		return
	}

	Created(c, session)
}

// GetSession 获取会话
func (h *ChatHandler) GetSession(c *gin.Context) {
	id := c.Param("id")

	session, err := h.svc.Chat.GetSession(c.Request.Context(), id)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, session)
}

// ListSessions 列出会话
func (h *ChatHandler) ListSessions(c *gin.Context) {
	page, pageSize := getPagination(c)

	sessions, total, err := h.svc.Chat.ListSessions(c.Request.Context(), &chat.ListSessionsRequest{
		UserID: getUserID(c),
		Page:   page,
		Size:   pageSize,
	})
	if err != nil {
		Error(c, err)
		return
	}

	SuccessWithPagination(c, sessions, total, page, pageSize)
}

// UpdateSession 更新会话
func (h *ChatHandler) UpdateSession(c *gin.Context) {
	id := c.Param("id")
	var req chat.CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	session, err := h.svc.Chat.UpdateSession(c.Request.Context(), id, &req)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, session)
}

// DeleteSession 删除会话
func (h *ChatHandler) DeleteSession(c *gin.Context) {
	id := c.Param("id")

	if err := h.svc.Chat.DeleteSession(c.Request.Context(), id); err != nil {
		Error(c, err)
		return
	}

	NoContent(c)
}

// SendMessage 发送消息
func (h *ChatHandler) SendMessage(c *gin.Context) {
	id := c.Param("id")

	var req chat.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	message, err := h.svc.Chat.SendMessage(c.Request.Context(), id, &req)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, message)
}

// GetMessages 获取会话消息
func (h *ChatHandler) GetMessages(c *gin.Context) {
	id := c.Param("id")

	messages, err := h.svc.Chat.GetMessages(c.Request.Context(), id)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{"messages": messages})
}

// ========== 独立消息管理 ==========

// LoadMessages 加载消息历史（支持分页和时间筛选）
func (h *ChatHandler) LoadMessages(c *gin.Context) {
	sessionID := c.Param("session_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	beforeTime := c.Query("before_time")

	messages, err := h.svc.Chat.LoadMessages(c.Request.Context(), sessionID, &chat.LoadMessagesRequest{
		Limit:      limit,
		BeforeTime: beforeTime,
	})
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{"messages": messages})
}

// GetMessage 获取单条消息
func (h *ChatHandler) GetMessage(c *gin.Context) {
	messageID := c.Param("id")

	message, err := h.svc.Chat.GetMessage(c.Request.Context(), messageID)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, message)
}

// DeleteMessage 删除消息
func (h *ChatHandler) DeleteMessage(c *gin.Context) {
	sessionID := c.Param("session_id")
	messageID := c.Param("id")

	if err := h.svc.Chat.DeleteMessage(c.Request.Context(), sessionID, messageID); err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{"message": "Message deleted successfully"})
}

// ========== 会话标题生成 ==========

// GenerateTitle 生成会话标题
func (h *ChatHandler) GenerateTitle(c *gin.Context) {
	sessionID := c.Param("id")

	var req chat.GenerateTitleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	title, err := h.svc.Chat.GenerateTitle(c.Request.Context(), sessionID, &req)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{"title": title})
}

// ========== WeKnora API 兼容接口 ==========

// KnowledgeChatRequest WeKnora 知识库聊天请求
type KnowledgeChatRequest struct {
	Query            string                 `json:"query" binding:"required"`
	KnowledgeBaseIDs []string               `json:"knowledge_base_ids"`
	KnowledgeIDs     []string               `json:"knowledge_ids"`
	AgentEnabled     bool                   `json:"agent_enabled"`
	AgentID          string                 `json:"agent_id"`
	WebSearchEnabled bool                   `json:"web_search_enabled"`
	SummaryModelID   string                 `json:"summary_model_id"`
	MentionedItems   []MentionedItem        `json:"mentioned_items"`
	DisableTitle     bool                   `json:"disable_title"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// MentionedItem 提及的项
type MentionedItem struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	KbType string `json:"kb_type,omitempty"`
}

// KnowledgeChat 知识库聊天（WeKnora API 兼容）
// POST /api/v1/knowledge-chat/:session_id
func (h *ChatHandler) KnowledgeChat(c *gin.Context) {
	sessionID := c.Param("session_id")

	var req KnowledgeChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	// TODO: 实现知识库聊天逻辑
	// 这需要调用 RAG 服务进行检索和生成

	// 暂时返回基本响应
	Success(c, gin.H{
		"answer":       "知识库聊天功能正在开发中",
		"session_id":   sessionID,
		"query":        req.Query,
		"sources":      []interface{}{},
		"message_id":   "",
		"stream_event": "",
	})
}

// AgentChatRequest WeKnora 智能体聊天请求
type AgentChatRequest struct {
	Query            string                 `json:"query" binding:"required"`
	AgentID          string                 `json:"agent_id"`
	KnowledgeBaseIDs []string               `json:"knowledge_base_ids"`
	WebSearchEnabled bool                   `json:"web_search_enabled"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// AgentChat 智能体聊天（WeKnora API 兼容）
// POST /api/v1/agent-chat/:session_id
func (h *ChatHandler) AgentChat(c *gin.Context) {
	sessionID := c.Param("session_id")

	var req AgentChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	// 如果指定了 Agent ID，使用 Agent 运行
	if req.AgentID != "" {
		runReq := agent.RunRequest{
			Query:     req.Query,
			SessionID: sessionID,
		}

		resp, err := h.svc.Agent.Run(c.Request.Context(), req.AgentID, &runReq)
		if err != nil {
			Error(c, err)
			return
		}

		Success(c, resp)
		return
	}

	// 暂时返回基本响应
	Success(c, gin.H{
		"answer":     "智能体聊天功能正在开发中",
		"session_id": sessionID,
		"query":      req.Query,
	})
}

// KnowledgeSearchRequest WeKnora 知识搜索请求
type KnowledgeSearchRequest struct {
	Query          string `json:"query" binding:"required"`
	TopK           int    `json:"top_k"`
	EnableOptimize bool   `json:"enable_optimize"`
	EnableRerank   bool   `json:"enable_rerank"`
}

// KnowledgeSearch 知识搜索（WeKnora API 兼容）
// POST /api/v1/knowledge-search
func (h *ChatHandler) KnowledgeSearch(c *gin.Context) {
	var req KnowledgeSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	// 设置默认值
	if req.TopK <= 0 {
		req.TopK = 5
	}

	// 创建 RAG 服务（与 RAGHandler 相同的方式）
	ragSvc := rag.NewService(
		h.svc.ChatModel,
		h.svc.Retriever,
		h.svc.Query,
		h.svc.Rerankers,
	)

	// 调用 RAG 服务进行检索
	result, err := ragSvc.Retrieve(c.Request.Context(), &rag.RetrieveRequest{
		Query:          req.Query,
		TopK:           req.TopK,
		EnableOptimize: req.EnableOptimize,
		EnableRerank:   req.EnableRerank,
	})
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, result)
}
