package handler

import (
	"net/http"
	"strconv"

	"github.com/ashwinyue/next-rag/next-ai/internal/service"
	"github.com/ashwinyue/next-rag/next-ai/internal/service/chat"
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
