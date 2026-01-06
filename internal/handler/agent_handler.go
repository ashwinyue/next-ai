package handler

import (
	"net/http"

	"github.com/ashwinyue/next-ai/internal/service"
	"github.com/ashwinyue/next-ai/internal/service/agent"
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

// CopyAgent 复制Agent
func (h *AgentHandler) CopyAgent(c *gin.Context) {
	id := c.Param("id")

	copiedAgent, err := h.svc.Agent.CopyAgent(c.Request.Context(), id)
	if err != nil {
		errorResponse(c, err)
		return
	}

	created(c, copiedAgent)
}

// GetPlaceholders 获取占位符定义
func (h *AgentHandler) GetPlaceholders(c *gin.Context) {
	success(c, getAllPlaceholders())
}

// GetAgentConfig 获取Agent配置
func (h *AgentHandler) GetAgentConfig(c *gin.Context) {
	id := c.Param("id")

	agent, err := h.svc.Agent.GetAgent(c.Request.Context(), id)
	if err != nil {
		errorResponse(c, err)
		return
	}

	// 返回 ModelConfig
	success(c, agent.ModelConfig)
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

// InitBuiltinAgents 初始化内置 Agent
func (h *AgentHandler) InitBuiltinAgents(c *gin.Context) {
	if err := h.svc.Agent.InitBuiltinAgents(c.Request.Context()); err != nil {
		errorResponse(c, err)
		return
	}

	success(c, gin.H{"message": "内置 Agent 初始化成功"})
}

// ListBuiltinAgents 列出内置 Agent
func (h *AgentHandler) ListBuiltinAgents(c *gin.Context) {
	agents, err := h.svc.Agent.ListBuiltinAgents(c.Request.Context())
	if err != nil {
		errorResponse(c, err)
		return
	}

	success(c, agents)
}

// ========== 占位符定义 ==========

// Placeholder 占位符定义
type Placeholder struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

// PlaceholderResponse 占位符响应
type PlaceholderResponse struct {
	All                 []Placeholder            `json:"all"`
	SystemPrompt        []Placeholder            `json:"system_prompt"`
	AgentSystemPrompt   []Placeholder            `json:"agent_system_prompt"`
	ContextTemplate     []Placeholder            `json:"context_template"`
	RewriteSystemPrompt []Placeholder            `json:"rewrite_system_prompt"`
	RewritePrompt       []Placeholder            `json:"rewrite_prompt"`
	FallbackPrompt      []Placeholder            `json:"fallback_prompt"`
}

// getAllPlaceholders 获取所有占位符定义
func getAllPlaceholders() *PlaceholderResponse {
	return &PlaceholderResponse{
		All: allPlaceholders(),
		SystemPrompt: []Placeholder{
			{Name: "query", Label: "用户问题", Description: "用户当前的问题或查询内容"},
			{Name: "contexts", Label: "检索内容", Description: "从知识库检索到的相关内容列表"},
			{Name: "current_time", Label: "当前时间", Description: "当前系统时间（格式：2006-01-02 15:04:05）"},
			{Name: "current_week", Label: "当前星期", Description: "当前星期几（如：星期一、Monday）"},
		},
		AgentSystemPrompt: []Placeholder{
			{Name: "knowledge_bases", Label: "知识库列表", Description: "自动格式化的知识库列表，包含名称、描述、文档数量等信息"},
			{Name: "web_search_status", Label: "网络搜索状态", Description: "网络搜索工具是否启用的状态（Enabled 或 Disabled）"},
			{Name: "current_time", Label: "当前时间", Description: "当前系统时间"},
		},
		ContextTemplate: []Placeholder{
			{Name: "query", Label: "用户问题", Description: "用户当前的问题"},
			{Name: "contexts", Label: "检索内容", Description: "从知识库检索到的相关内容"},
			{Name: "current_time", Label: "当前时间", Description: "当前系统时间"},
		},
		RewriteSystemPrompt: []Placeholder{
			{Name: "query", Label: "用户问题", Description: "用户当前的问题"},
			{Name: "conversation", Label: "历史对话", Description: "格式化的历史对话内容"},
			{Name: "current_time", Label: "当前时间", Description: "当前系统时间"},
		},
		RewritePrompt: []Placeholder{
			{Name: "query", Label: "用户问题", Description: "用户当前的问题"},
			{Name: "conversation", Label: "历史对话", Description: "格式化的历史对话内容"},
			{Name: "current_time", Label: "当前时间", Description: "当前系统时间"},
		},
		FallbackPrompt: []Placeholder{
			{Name: "query", Label: "用户问题", Description: "用户当前的问题"},
		},
	}
}

// allPlaceholders 返回所有可用占位符
func allPlaceholders() []Placeholder {
	return []Placeholder{
		{Name: "query", Label: "用户问题", Description: "用户当前的问题或查询内容"},
		{Name: "contexts", Label: "检索内容", Description: "从知识库检索到的相关内容列表"},
		{Name: "current_time", Label: "当前时间", Description: "当前系统时间（格式：2006-01-02 15:04:05）"},
		{Name: "current_week", Label: "当前星期", Description: "当前星期几（如：星期一、Monday）"},
		{Name: "conversation", Label: "历史对话", Description: "格式化的历史对话内容，用于多轮对话改写"},
		{Name: "yesterday", Label: "昨天日期", Description: "昨天的日期（格式：2006-01-02）"},
		{Name: "answer", Label: "助手回答", Description: "助手的回答内容"},
		{Name: "knowledge_bases", Label: "知识库列表", Description: "自动格式化的知识库列表，包含名称、描述、文档数量等信息"},
		{Name: "web_search_status", Label: "网络搜索状态", Description: "网络搜索工具是否启用的状态"},
	}
}
