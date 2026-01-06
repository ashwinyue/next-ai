package handler

import (
	"net/http"

	"github.com/ashwinyue/next-rag/next-ai/internal/service"
	"github.com/ashwinyue/next-rag/next-ai/internal/service/agent"
	"github.com/gin-gonic/gin"
)

// AgentHandler Agent处理器
type AgentHandler struct {
	svc *service.Services
}

// NewAgentHandler 创建Agent处理器
func NewAgentHandler(svc *service.Services) *AgentHandler {
	return &AgentHandler{svc: svc}
}

// CreateAgent 创建Agent
func (h *AgentHandler) CreateAgent(c *gin.Context) {
	var req agent.CreateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: -1, Message: err.Error()})
		return
	}

	agent, err := h.svc.Agent.CreateAgent(c.Request.Context(), &req)
	if err != nil {
		errorResponse(c, err)
		return
	}

	created(c, agent)
}

// GetAgent 获取Agent
func (h *AgentHandler) GetAgent(c *gin.Context) {
	id := c.Param("id")

	agent, err := h.svc.Agent.GetAgent(c.Request.Context(), id)
	if err != nil {
		errorResponse(c, err)
		return
	}

	success(c, agent)
}

// ListAgents 列出Agent
func (h *AgentHandler) ListAgents(c *gin.Context) {
	page, size := getPagination(c)

	agents, err := h.svc.Agent.ListAgents(c.Request.Context(), &agent.ListAgentsRequest{
		Page: page,
		Size: size,
	})
	if err != nil {
		errorResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"items": agents,
			"total": int64(len(agents)),
			"page":  page,
			"size":  size,
		},
	})
}

// ListActiveAgents 列出活跃Agent
func (h *AgentHandler) ListActiveAgents(c *gin.Context) {
	agents, err := h.svc.Agent.ListActiveAgents(c.Request.Context())
	if err != nil {
		errorResponse(c, err)
		return
	}

	success(c, agents)
}

// UpdateAgent 更新Agent
func (h *AgentHandler) UpdateAgent(c *gin.Context) {
	id := c.Param("id")
	var req agent.CreateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: -1, Message: err.Error()})
		return
	}

	agent, err := h.svc.Agent.UpdateAgent(c.Request.Context(), id, &req)
	if err != nil {
		errorResponse(c, err)
		return
	}

	success(c, agent)
}

// DeleteAgent 删除Agent
func (h *AgentHandler) DeleteAgent(c *gin.Context) {
	id := c.Param("id")

	if err := h.svc.Agent.DeleteAgent(c.Request.Context(), id); err != nil {
		errorResponse(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// GetAgentConfig 获取Agent配置
func (h *AgentHandler) GetAgentConfig(c *gin.Context) {
	id := c.Param("id")

	config, err := h.svc.Agent.GetAgentConfig(c.Request.Context(), id)
	if err != nil {
		errorResponse(c, err)
		return
	}

	success(c, config)
}

// RunAgent 运行Agent（同步）
func (h *AgentHandler) RunAgent(c *gin.Context) {
	id := c.Param("id")
	var req agent.RunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: -1, Message: err.Error()})
		return
	}

	resp, err := h.svc.Agent.Run(c.Request.Context(), id, &req)
	if err != nil {
		errorResponse(c, err)
		return
	}

	success(c, resp)
}

// StreamAgent 运行Agent（流式）
func (h *AgentHandler) StreamAgent(c *gin.Context) {
	id := c.Param("id")
	var req agent.RunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: -1, Message: err.Error()})
		return
	}

	eventCh, err := h.svc.Agent.Stream(c.Request.Context(), id, &req)
	if err != nil {
		errorResponse(c, err)
		return
	}

	// 设置 SSE 响应头
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	// 发送流式事件
	for event := range eventCh {
		select {
		case <-c.Request.Context().Done():
			return
		default:
			c.SSEvent("", event)
			c.Writer.Flush()
		}

		if event.Type == "end" || event.Type == "error" {
			return
		}
	}
}
