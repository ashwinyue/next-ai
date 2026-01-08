// Package handler 提供租户相关的 HTTP 处理器
package handler

import (
	"strconv"
	"strings"

	"github.com/ashwinyue/next-ai/internal/model"
	"github.com/ashwinyue/next-ai/internal/service"
	"github.com/ashwinyue/next-ai/internal/service/tenant"
	"github.com/gin-gonic/gin"
)

// TenantHandler 租户处理器
type TenantHandler struct {
	svc *service.Services
}

// NewTenantHandler 创建租户处理器
func NewTenantHandler(svc *service.Services) *TenantHandler {
	return &TenantHandler{svc: svc}
}

// CreateTenant 创建租户
// @Summary      创建租户
// @Description  创建新的租户
// @Tags         租户管理
// @Accept       json
// @Produce      json
// @Param        request  body      tenant.CreateTenantRequest  true  "租户信息"
// @Success      200      {object}  Response
// @Router       /api/v1/tenants [post]
func (h *TenantHandler) CreateTenant(c *gin.Context) {
	ctx := c.Request.Context()

	var req tenant.CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	result, err := h.svc.Tenant.CreateTenant(ctx, &req)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, result)
}

// GetTenant 获取租户详情
// @Summary      获取租户
// @Description  根据 ID 获取租户详情
// @Tags         租户管理
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "租户 ID"
// @Success      200  {object}  Response
// @Router       /api/v1/tenants/{id} [get]
func (h *TenantHandler) GetTenant(c *gin.Context) {
	ctx := c.Request.Context()

	id := c.Param("id")
	if id == "" {
		BadRequest(c, "id is required")
		return
	}

	tenantModel, err := h.svc.Tenant.GetTenant(ctx, id)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, tenantModel)
}

// ListTenants 列出租户
// @Summary      列出租户
// @Description  获取租户列表
// @Tags         租户管理
// @Accept       json
// @Produce      json
// @Success      200  {object}  Response
// @Router       /api/v1/tenants [get]
func (h *TenantHandler) ListTenants(c *gin.Context) {
	ctx := c.Request.Context()

	tenants, err := h.svc.Tenant.ListTenants(ctx)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, tenants)
}

// UpdateTenant 更新租户
// @Summary      更新租户
// @Description  更新租户信息
// @Tags         租户管理
// @Accept       json
// @Produce      json
// @Param        id       path      string  true  "租户 ID"
// @Param        request  body      tenant.UpdateTenantRequest  true  "租户信息"
// @Success      200      {object}  Response
// @Router       /api/v1/tenants/{id} [put]
func (h *TenantHandler) UpdateTenant(c *gin.Context) {
	ctx := c.Request.Context()

	id := c.Param("id")
	if id == "" {
		BadRequest(c, "id is required")
		return
	}

	var req tenant.UpdateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	result, err := h.svc.Tenant.UpdateTenant(ctx, id, &req)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, result)
}

// DeleteTenant 删除租户
// @Summary      删除租户
// @Description  删除指定的租户
// @Tags         租户管理
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "租户 ID"
// @Success      200  {object}  Response
// @Router       /api/v1/tenants/{id} [delete]
func (h *TenantHandler) DeleteTenant(c *gin.Context) {
	ctx := c.Request.Context()

	id := c.Param("id")
	if id == "" {
		BadRequest(c, "id is required")
		return
	}

	if err := h.svc.Tenant.DeleteTenant(ctx, id); err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{"message": "租户已删除"})
}

// GetTenantConfig 获取租户配置
// @Summary      获取租户配置
// @Description  获取租户的指定配置
// @Tags         租户管理
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "租户 ID"
// @Param        type query     string  true  "配置类型: agent, context, web_search, conversation"
// @Success      200  {object}  Response
// @Router       /api/v1/tenants/{id}/config [get]
func (h *TenantHandler) GetTenantConfig(c *gin.Context) {
	ctx := c.Request.Context()

	id := c.Param("id")
	if id == "" {
		BadRequest(c, "id is required")
		return
	}

	configType := c.Query("type")
	if configType == "" {
		BadRequest(c, "type is required")
		return
	}

	config, err := h.svc.Tenant.GetTenantConfig(ctx, id, configType)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, config)
}

// UpdateTenantConfig 更新租户配置
// @Summary      更新租户配置
// @Description  更新租户的指定配置
// @Tags         租户管理
// @Accept       json
// @Produce      json
// @Param        id     path      string  true  "租户 ID"
// @Param        type   query     string  true  "配置类型: agent, context, web_search, conversation"
// @Param        request body      tenant.UpdateTenantConfigRequest  true  "配置内容"
// @Success      200      {object}  Response
// @Router       /api/v1/tenants/{id}/config [put]
func (h *TenantHandler) UpdateTenantConfig(c *gin.Context) {
	ctx := c.Request.Context()

	id := c.Param("id")
	if id == "" {
		BadRequest(c, "id is required")
		return
	}

	configType := c.Query("type")
	if configType == "" {
		BadRequest(c, "type is required")
		return
	}

	var config interface{}
	if err := c.ShouldBindJSON(&config); err != nil {
		BadRequest(c, err.Error())
		return
	}

	req := &tenant.UpdateTenantConfigRequest{
		ConfigType: configType,
		Config:     config,
	}

	if err := h.svc.Tenant.UpdateTenantConfig(ctx, id, req); err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{"message": "配置已更新"})
}

// GetTenantStorage 获取租户存储信息
// @Summary      获取租户存储
// @Description  获取租户存储使用情况
// @Tags         租户管理
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "租户 ID"
// @Success      200  {object}  Response
// @Router       /api/v1/tenants/{id}/storage [get]
func (h *TenantHandler) GetTenantStorage(c *gin.Context) {
	ctx := c.Request.Context()

	id := c.Param("id")
	if id == "" {
		BadRequest(c, "id is required")
		return
	}

	used, quota, err := h.svc.Tenant.GetStorageStats(ctx, id)
	if err != nil {
		Error(c, err)
		return
	}

	// 计算存储使用百分比
	usedPercent := 0.0
	if quota > 0 {
		usedPercent = float64(used) / float64(quota) * 100
	}

	Success(c, gin.H{
		"storage_used":  used,
		"storage_quota": quota,
		"used_percent":  usedPercent,
		"available":     quota - used,
	})
}

// parseInt 辅助函数：解析整数参数
func parseInt(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(s)
	if err != nil || val < 0 {
		return defaultVal
	}
	return val
}

// ========== 租户 KV 配置（WeKnora API 兼容）==========

// GetTenantKV 获取租户 KV 配置
// GET /api/v1/tenants/kv/:key
// 支持: agent, context, web_search, conversation
func (h *TenantHandler) GetTenantKV(c *gin.Context) {
	key := c.Param("key")

	// 获取当前租户 ID（从上下文）
	tenantID := "default" // 简化版：使用默认租户

	config, err := h.svc.Tenant.GetTenantConfig(c.Request.Context(), tenantID, key)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{
		"success": true,
		"data":    config,
	})
}

// UpdateTenantKV 更新租户 KV 配置
// PUT /api/v1/tenants/kv/:key
func (h *TenantHandler) UpdateTenantKV(c *gin.Context) {
	key := c.Param("key")

	// 获取当前租户 ID（从上下文）
	tenantID := "default" // 简化版：使用默认租户

	var config interface{}
	if err := c.ShouldBindJSON(&config); err != nil {
		BadRequest(c, err.Error())
		return
	}

	req := &tenant.UpdateTenantConfigRequest{
		ConfigType: key,
		Config:     config,
	}

	if err := h.svc.Tenant.UpdateTenantConfig(c.Request.Context(), tenantID, req); err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{
		"success": true,
		"message": "Configuration updated successfully",
	})
}

// ListAllTenants 列出所有租户（WeKnora API 兼容）
// GET /api/v1/tenants/all
func (h *TenantHandler) ListAllTenants(c *gin.Context) {
	tenants, err := h.svc.Tenant.ListTenants(c.Request.Context())
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{
		"success": true,
		"data":    tenants,
	})
}

// SearchTenants 搜索租户（WeKnora API 兼容）
// GET /api/v1/tenants/search?q=keyword
func (h *TenantHandler) SearchTenants(c *gin.Context) {
	query := c.Query("q")

	tenants, err := h.svc.Tenant.ListTenants(c.Request.Context())
	if err != nil {
		Error(c, err)
		return
	}

	// 简单过滤（生产环境应在数据库层实现）
	filtered := make([]*model.Tenant, 0)
	if query != "" {
		for _, t := range tenants {
			// 实现简单搜索逻辑：匹配名称或描述
			if strings.Contains(strings.ToLower(t.Name), strings.ToLower(query)) ||
				strings.Contains(strings.ToLower(t.Description), strings.ToLower(query)) {
				filtered = append(filtered, t)
			}
		}
	} else {
		filtered = tenants
	}

	Success(c, gin.H{
		"success": true,
		"data":    filtered,
	})
}
