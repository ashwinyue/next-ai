// Package handler 提供消息相关的 HTTP 处理器
// 对齐 WeKnora 的 message.go handler
package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/ashwinyue/next-ai/internal/service/chat"
	"github.com/gin-gonic/gin"
)

// MessageHandler 消息处理器
type MessageHandler struct {
	chatService *chat.ServiceWithAgent
}

// NewMessageHandler 创建消息处理器
func NewMessageHandler(chatService *chat.ServiceWithAgent) *MessageHandler {
	return &MessageHandler{
		chatService: chatService,
	}
}

// LoadMessages godoc
// @Summary      加载消息历史
// @Description  加载会话的消息历史，支持分页和时间筛选
// @Tags         消息
// @Accept       json
// @Produce      json
// @Param        session_id   path      string  true   "会话ID"
// @Param        limit        query     int     false  "返回数量"  default(20)
// @Param        before_time  query     string  false  "在此时间之前的消息（RFC3339Nano格式）"
// @Success      200          {object}  map[string]interface{}  "消息列表"
// @Failure      400          {object}  map[string]interface{}  "请求参数错误"
// @Security     Bearer
// @Router       /messages/{session_id}/load [get]
func (h *MessageHandler) LoadMessages(c *gin.Context) {
	ctx := c.Request.Context()

	// 获取会话 ID
	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "会话ID不能为空",
		})
		return
	}

	// 解析 limit 参数
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100 // 最多返回 100 条
	}

	// 解析 before_time 参数
	beforeTimeStr := c.Query("before_time")

	if beforeTimeStr == "" {
		// 获取最近的消息
		messages, err := h.chatService.Service.LoadMessages(ctx, sessionID, &chat.LoadMessagesRequest{
			Limit: limit,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "获取消息失败",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    messages,
		})
		return
	}

	// 验证时间格式（不进行实际转换，直接传递字符串）
	messages, err := h.chatService.Service.LoadMessages(ctx, sessionID, &chat.LoadMessagesRequest{
		BeforeTime: beforeTimeStr,
		Limit:      limit,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取消息失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    messages,
	})
}

// DeleteMessage godoc
// @Summary      删除消息
// @Description  从会话中删除指定消息
// @Tags         消息
// @Accept       json
// @Produce      json
// @Param        session_id  path      string  true  "会话ID"
// @Param        id          path      string  true  "消息ID"
// @Success      200         {object}  map[string]interface{}  "删除成功"
// @Failure      400         {object}  map[string]interface{}  "请求参数错误"
// @Failure      404         {object}  map[string]interface{}  "消息不存在"
// @Security     Bearer
// @Router       /messages/{session_id}/{id} [delete]
func (h *MessageHandler) DeleteMessage(c *gin.Context) {
	ctx := c.Request.Context()

	// 获取路径参数
	sessionID := c.Param("session_id")
	messageID := c.Param("id")

	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "会话ID不能为空",
		})
		return
	}

	if messageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "消息ID不能为空",
		})
		return
	}

	// 删除消息
	err := h.chatService.Service.DeleteMessage(ctx, sessionID, messageID)
	if err != nil {
		// 检查错误类型
		errMsg := err.Error()
		if strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "不存在") {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "消息不存在",
			})
			return
		}
		if strings.Contains(errMsg, "does not belong") || strings.Contains(errMsg, "不属于") {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "消息不属于该会话",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "删除消息失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "消息删除成功",
	})
}
