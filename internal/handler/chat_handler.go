package handler

import (
	"net/http"
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

// Response 统一响应
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// success 成功响应
func success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{Code: 0, Message: "success", Data: data})
}

// created 创建成功响应
func created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{Code: 0, Message: "created", Data: data})
}

// errorResponse 错误响应
func errorResponse(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, Response{Code: -1, Message: err.Error()})
}

// getPagination 获取分页参数
func getPagination(c *gin.Context) (page, size int) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ = strconv.Atoi(c.DefaultQuery("size", "20"))
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 20
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
		c.JSON(http.StatusBadRequest, Response{Code: -1, Message: err.Error()})
		return
	}

	session, err := h.svc.Chat.CreateSession(c.Request.Context(), &req)
	if err != nil {
		errorResponse(c, err)
		return
	}

	created(c, session)
}

// GetSession 获取会话
func (h *ChatHandler) GetSession(c *gin.Context) {
	id := c.Param("id")

	session, err := h.svc.Chat.GetSession(c.Request.Context(), id)
	if err != nil {
		errorResponse(c, err)
		return
	}

	success(c, session)
}

// ListSessions 列出会话
func (h *ChatHandler) ListSessions(c *gin.Context) {
	page, size := getPagination(c)

	sessions, total, err := h.svc.Chat.ListSessions(c.Request.Context(), &chat.ListSessionsRequest{
		UserID: getUserID(c),
		Page:   page,
		Size:   size,
	})
	if err != nil {
		errorResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"items": sessions,
			"total": total,
			"page":  page,
			"size":  size,
		},
	})
}

// UpdateSession 更新会话
func (h *ChatHandler) UpdateSession(c *gin.Context) {
	id := c.Param("id")
	var req chat.CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: -1, Message: err.Error()})
		return
	}

	session, err := h.svc.Chat.UpdateSession(c.Request.Context(), id, &req)
	if err != nil {
		errorResponse(c, err)
		return
	}

	success(c, session)
}

// DeleteSession 删除会话
func (h *ChatHandler) DeleteSession(c *gin.Context) {
	id := c.Param("id")

	if err := h.svc.Chat.DeleteSession(c.Request.Context(), id); err != nil {
		errorResponse(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// SendMessage 发送消息
func (h *ChatHandler) SendMessage(c *gin.Context) {
	id := c.Param("id")

	var req chat.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: -1, Message: err.Error()})
		return
	}

	message, err := h.svc.Chat.SendMessage(c.Request.Context(), id, &req)
	if err != nil {
		errorResponse(c, err)
		return
	}

	success(c, message)
}

// GetMessages 获取会话消息
func (h *ChatHandler) GetMessages(c *gin.Context) {
	id := c.Param("id")

	messages, err := h.svc.Chat.GetMessages(c.Request.Context(), id)
	if err != nil {
		errorResponse(c, err)
		return
	}

	success(c, gin.H{"messages": messages})
}

// ========== 独立消息管理 ==========

// LoadMessages 加载消息历史（支持分页和时间筛选）
// @Summary      加载消息历史
// @Description  加载会话的消息历史，支持分页和时间筛选
// @Tags         消息管理
// @Accept       json
// @Produce      json
// @Param        session_id  path      string  true   "会话ID"
// @Param        limit       query     int     false  "返回数量"  default(20)
// @Param        before_time query     string  false  "在此时间之前的消息（RFC3339Nano格式）"
// @Success      200         {object}  Response  "消息列表"
// @Failure      500         {object}  Response  "服务器错误"
// @Router       /messages/{session_id}/load [get]
func (h *ChatHandler) LoadMessages(c *gin.Context) {
	sessionID := c.Param("session_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	beforeTime := c.Query("before_time")

	messages, err := h.svc.Chat.LoadMessages(c.Request.Context(), sessionID, &chat.LoadMessagesRequest{
		Limit:      limit,
		BeforeTime: beforeTime,
	})
	if err != nil {
		errorResponse(c, err)
		return
	}

	success(c, gin.H{"messages": messages})
}

// GetMessage 获取单条消息
// @Summary      获取消息
// @Description  获取单条消息详情
// @Tags         消息管理
// @Accept       json
// @Produce      json
// @Param        id         path      string      true   "消息ID"
// @Success      200        {object}  Response    "消息详情"
// @Failure      404        {object}  Response    "消息不存在"
// @Router       /messages/{id} [get]
func (h *ChatHandler) GetMessage(c *gin.Context) {
	messageID := c.Param("id")

	message, err := h.svc.Chat.GetMessage(c.Request.Context(), messageID)
	if err != nil {
		errorResponse(c, err)
		return
	}

	success(c, message)
}

// DeleteMessage 删除消息
// @Summary      删除消息
// @Description  从会话中删除指定消息
// @Tags         消息管理
// @Accept       json
// @Produce      json
// @Param        session_id  path      string  true  "会话ID"
// @Param        id         path      string  true  "消息ID"
// @Success      200        {object}  Response  "删除成功"
// @Failure      500        {object}  Response  "服务器错误"
// @Router       /messages/{session_id}/{id} [delete]
func (h *ChatHandler) DeleteMessage(c *gin.Context) {
	sessionID := c.Param("session_id")
	messageID := c.Param("id")

	if err := h.svc.Chat.DeleteMessage(c.Request.Context(), sessionID, messageID); err != nil {
		errorResponse(c, err)
		return
	}

	success(c, gin.H{"message": "Message deleted successfully"})
}

// ========== 会话标题生成 ==========

// GenerateTitle 生成会话标题
// @Summary      生成会话标题
// @Description  根据首条用户消息自动生成会话标题
// @Tags         聊天管理
// @Accept       json
// @Produce      json
// @Param        id            path      string                 true   "会话ID"
// @Param        request       body      chat.GenerateTitleRequest  true   "生成标题请求"
// @Success      200           {object}  Response               "标题生成成功"
// @Failure      500           {object}  Response               "服务器错误"
// @Router       /chats/{id}/title [post]
func (h *ChatHandler) GenerateTitle(c *gin.Context) {
	sessionID := c.Param("id")

	var req chat.GenerateTitleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: -1, Message: err.Error()})
		return
	}

	title, err := h.svc.Chat.GenerateTitle(c.Request.Context(), sessionID, &req)
	if err != nil {
		errorResponse(c, err)
		return
	}

	success(c, gin.H{"title": title})
}
