package handler

import (
	"strconv"

	"github.com/ashwinyue/next-ai/internal/service"
	"github.com/ashwinyue/next-ai/internal/service/chat"
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

// AgentChatRequest 智能体聊天请求
type AgentChatRequest struct {
	Query    string                 `json:"query" binding:"required"`
	AgentID  string                 `json:"agent_id"`
	Metadata map[string]interface{} `json:"metadata"`
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

	// h.svc.Chat 已经是 *chat.ServiceWithAgent 类型
	chatSvc := h.svc.Chat

	// 构建请求
	agentReq := &chat.AgentChatRequest{
		SessionID: sessionID,
		AgentID:   req.AgentID,
		Query:     req.Query,
	}

	// 调用 Agent 聊天（流式）
	eventCh, err := chatSvc.AgentChat(c.Request.Context(), agentReq)
	if err != nil {
		Error(c, err)
		return
	}

	// 设置 SSE 流式响应
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	// 发送事件
	for evt := range eventCh {
		// 格式化为 SSE 格式
		c.SSEvent("message", evt)
		c.Writer.Flush()
	}

	// 发送结束事件
	c.SSEvent("end", gin.H{"session_id": sessionID})
	c.Writer.Flush()
}

// ========== 会话流控制（WeKnora API 兼容）==========

// StopSessionRequest 停止会话请求
type StopSessionRequest struct {
	MessageID string `json:"message_id" binding:"required"`
}

// StopSession 停止会话生成
// POST /api/v1/sessions/:session_id/stop
func (h *ChatHandler) StopSession(c *gin.Context) {
	sessionID := c.Param("session_id")

	var req StopSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	// 使用会话管理器停止流
	stopped := h.svc.SessionMgr.StopStream(sessionID, req.MessageID)

	if !stopped {
		// 流不存在或已完成
		Success(c, gin.H{
			"success":    false,
			"message":    "Stream not found or already completed",
			"session_id": sessionID,
			"message_id": req.MessageID,
		})
		return
	}

	Success(c, gin.H{
		"success":    true,
		"message":    "Session stopped",
		"session_id": sessionID,
		"message_id": req.MessageID,
	})
}

// ContinueStream 继续接收流
// GET /api/v1/sessions/continue-stream/:session_id?message_id=xxx
func (h *ChatHandler) ContinueStream(c *gin.Context) {
	sessionID := c.Param("session_id")
	messageID := c.Query("message_id")

	if messageID == "" {
		BadRequest(c, "message_id is required")
		return
	}

	// 获取流状态
	stream := h.svc.SessionMgr.GetStream(sessionID, messageID)

	if stream == nil {
		// 流不存在，可能已完成或从未存在
		Success(c, gin.H{
			"success":    true,
			"session_id": sessionID,
			"message_id": messageID,
			"done":       true,
			"content":    "",
			"found":      false,
		})
		return
	}

	// 获取当前内容
	content := stream.GetContent()
	done := stream.IsDone()

	// 获取新增的块（用于增量更新）
	chunks := stream.GetPartialChunks()

	Success(c, gin.H{
		"success":    true,
		"session_id": sessionID,
		"message_id": messageID,
		"done":       done,
		"content":    content,
		"chunks":     chunks,
		"found":      true,
	})
}
