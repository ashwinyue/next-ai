// Package handler 提供 MCP 服务相关的 HTTP 处理器
package handler

import (
	"github.com/ashwinyue/next-ai/internal/service"
	"github.com/ashwinyue/next-ai/internal/service/mcp"
	"github.com/gin-gonic/gin"
)

// MCPServiceHandler MCP 服务处理器
type MCPServiceHandler struct {
	svc *service.Services
}

// NewMCPServiceHandler 创建 MCP 服务处理器
func NewMCPServiceHandler(svc *service.Services) *MCPServiceHandler {
	return &MCPServiceHandler{svc: svc}
}

// CreateMCPServiceRequest 创建 MCP 服务请求
type CreateMCPServiceRequest = mcp.CreateMCPServiceRequest

// CreateMCPService 创建 MCP 服务
// @Summary      创建 MCP 服务
// @Description  创建新的 MCP 服务配置
// @Tags         MCP 服务
// @Accept       json
// @Produce      json
// @Param        request  body      CreateMCPServiceRequest  true  "MCP 服务配置"
// @Success      200      {object}  Response
// @Router       /api/v1/mcp-services [post]
func (h *MCPServiceHandler) CreateMCPService(c *gin.Context) {
	ctx := c.Request.Context()

	var req CreateMCPServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	svc, err := h.svc.MCP.CreateMCPService(ctx, &req)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, svc)
}

// ListMCPServices 列出 MCP 服务
// @Summary      列出 MCP 服务
// @Description  获取所有 MCP 服务
// @Tags         MCP 服务
// @Accept       json
// @Produce      json
// @Success      200  {object}  Response
// @Router       /api/v1/mcp-services [get]
func (h *MCPServiceHandler) ListMCPServices(c *gin.Context) {
	ctx := c.Request.Context()

	services, err := h.svc.MCP.ListMCPServices(ctx)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, services)
}

// GetMCPService 获取 MCP 服务详情
// @Summary      获取 MCP 服务
// @Description  根据 ID 获取 MCP 服务详情
// @Tags         MCP 服务
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "服务 ID"
// @Success      200  {object}  Response
// @Router       /api/v1/mcp-services/{id} [get]
func (h *MCPServiceHandler) GetMCPService(c *gin.Context) {
	ctx := c.Request.Context()

	id := c.Param("id")
	if id == "" {
		BadRequest(c, "id is required")
		return
	}

	svc, err := h.svc.MCP.GetMCPService(ctx, id)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, svc)
}

// UpdateMCPServiceRequest 更新 MCP 服务请求
type UpdateMCPServiceRequest = mcp.UpdateMCPServiceRequest

// UpdateMCPService 更新 MCP 服务
// @Summary      更新 MCP 服务
// @Description  更新 MCP 服务配置
// @Tags         MCP 服务
// @Accept       json
// @Produce      json
// @Param        id       path      string                   true  "服务 ID"
// @Param        request  body      UpdateMCPServiceRequest  true  "更新字段"
// @Success      200      {object}  Response
// @Router       /api/v1/mcp-services/{id} [put]
func (h *MCPServiceHandler) UpdateMCPService(c *gin.Context) {
	ctx := c.Request.Context()

	id := c.Param("id")
	if id == "" {
		BadRequest(c, "id is required")
		return
	}

	var req UpdateMCPServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	svc, err := h.svc.MCP.UpdateMCPService(ctx, id, &req)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, svc)
}

// DeleteMCPService 删除 MCP 服务
// @Summary      删除 MCP 服务
// @Description  删除指定的 MCP 服务
// @Tags         MCP 服务
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "服务 ID"
// @Success      200  {object}  Response
// @Router       /api/v1/mcp-services/{id} [delete]
func (h *MCPServiceHandler) DeleteMCPService(c *gin.Context) {
	ctx := c.Request.Context()

	id := c.Param("id")
	if id == "" {
		BadRequest(c, "id is required")
		return
	}

	if err := h.svc.MCP.DeleteMCPService(ctx, id); err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{"message": "MCP 服务已删除"})
}

// TestMCPService 测试 MCP 服务连接
// @Summary      测试 MCP 服务
// @Description  测试 MCP 服务是否可以正常连接
// @Tags         MCP 服务
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "服务 ID"
// @Success      200  {object}  Response
// @Router       /api/v1/mcp-services/{id}/test [post]
func (h *MCPServiceHandler) TestMCPService(c *gin.Context) {
	ctx := c.Request.Context()

	id := c.Param("id")
	if id == "" {
		BadRequest(c, "id is required")
		return
	}

	result, err := h.svc.MCP.TestMCPService(ctx, id)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, result)
}

// GetMCPServiceTools 获取 MCP 服务工具列表
// @Summary      获取 MCP 工具
// @Description  获取 MCP 服务提供的工具列表
// @Tags         MCP 服务
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "服务 ID"
// @Success      200  {object}  Response
// @Router       /api/v1/mcp-services/{id}/tools [get]
func (h *MCPServiceHandler) GetMCPServiceTools(c *gin.Context) {
	ctx := c.Request.Context()

	id := c.Param("id")
	if id == "" {
		BadRequest(c, "id is required")
		return
	}

	tools, err := h.svc.MCP.GetMCPServiceTools(ctx, id)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, tools)
}

// GetMCPServiceResources 获取 MCP 服务资源列表
// @Summary      获取 MCP 资源
// @Description  获取 MCP 服务提供的资源列表
// @Tags         MCP 服务
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "服务 ID"
// @Success      200  {object}  Response
// @Router       /api/v1/mcp-services/{id}/resources [get]
func (h *MCPServiceHandler) GetMCPServiceResources(c *gin.Context) {
	ctx := c.Request.Context()

	id := c.Param("id")
	if id == "" {
		BadRequest(c, "id is required")
		return
	}

	resources, err := h.svc.MCP.GetMCPServiceResources(ctx, id)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, resources)
}
