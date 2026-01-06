package handler

import (
	"github.com/ashwinyue/next-ai/internal/service"
	"github.com/ashwinyue/next-ai/internal/service/tool"
	"github.com/gin-gonic/gin"
)

// ToolHandler 工具处理器
type ToolHandler struct {
	svc *service.Services
}

// NewToolHandler 创建工具处理器
func NewToolHandler(svc *service.Services) *ToolHandler {
	return &ToolHandler{svc: svc}
}

// RegisterTool 注册工具
func (h *ToolHandler) RegisterTool(c *gin.Context) {
	var req tool.RegisterToolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	tool, err := h.svc.Tool.RegisterTool(c.Request.Context(), &req)
	if err != nil {
		Error(c, err)
		return
	}

	Created(c, tool)
}

// GetTool 获取工具
func (h *ToolHandler) GetTool(c *gin.Context) {
	id := c.Param("id")

	tool, err := h.svc.Tool.GetTool(c.Request.Context(), id)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, tool)
}

// ListTools 列出工具
func (h *ToolHandler) ListTools(c *gin.Context) {
	page, pageSize := getPagination(c)

	tools, err := h.svc.Tool.ListTools(c.Request.Context(), &tool.ListToolsRequest{
		Page: page,
		Size: pageSize,
	})
	if err != nil {
		Error(c, err)
		return
	}

	SuccessWithPagination(c, tools, int64(len(tools)), page, pageSize)
}

// ListActiveTools 列出活跃工具
func (h *ToolHandler) ListActiveTools(c *gin.Context) {
	tools, err := h.svc.Tool.ListActiveTools(c.Request.Context())
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, tools)
}

// UpdateTool 更新工具
func (h *ToolHandler) UpdateTool(c *gin.Context) {
	id := c.Param("id")
	var req tool.RegisterToolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	tool, err := h.svc.Tool.UpdateTool(c.Request.Context(), id, &req)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, tool)
}

// UnregisterTool 注销工具
func (h *ToolHandler) UnregisterTool(c *gin.Context) {
	id := c.Param("id")

	if err := h.svc.Tool.UnregisterTool(c.Request.Context(), id); err != nil {
		Error(c, err)
		return
	}

	NoContent(c)
}
