// Package handler 提供租户相关的 HTTP 处理器
package handler

import (
	"net/http"
	"strconv"

	"github.com/ashwinyue/next-ai/internal/model"
	"github.com/ashwinyue/next-ai/internal/service"
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
// @Param        request  body      model.Tenant  true  "租户信息"
// @Success      200      {object}  Response
// @Router       /api/v1/tenants [post]
func (h *TenantHandler) CreateTenant(c *gin.Context) {
	ctx := c.Request.Context()

	var tenant model.Tenant
	if err := c.ShouldBindJSON(&tenant); err != nil {
		c.JSON(http.StatusBadRequest, BadRequest(c, err.Error()))
		return
	}

	result, err := h.svc.Tenant.CreateTenant(ctx, &tenant)
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

	tenant, err := h.svc.Tenant.GetTenant(ctx, id)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, tenant)
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
// @Param        id       path      string       true  "租户 ID"
// @Param        request  body      model.Tenant  true  "租户信息"
// @Success      200      {object}  Response
// @Router       /api/v1/tenants/{id} [put]
func (h *TenantHandler) UpdateTenant(c *gin.Context) {
	ctx := c.Request.Context()

	id := c.Param("id")
	if id == "" {
		BadRequest(c, "id is required")
		return
	}

	var tenant model.Tenant
	if err := c.ShouldBindJSON(&tenant); err != nil {
		c.JSON(http.StatusBadRequest, BadRequest(c, err.Error()))
		return
	}

	tenant.ID = id
	result, err := h.svc.Tenant.UpdateTenant(ctx, &tenant)
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
// @Param        request body      object  true  "配置内容"
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
		c.JSON(http.StatusBadRequest, BadRequest(c, err.Error()))
		return
	}

	if err := h.svc.Tenant.UpdateTenantConfig(ctx, id, configType, config); err != nil {
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

	tenant, err := h.svc.Tenant.GetTenant(ctx, id)
	if err != nil {
		Error(c, err)
		return
	}

	// 计算存储使用百分比
	usedPercent := 0.0
	if tenant.StorageQuota > 0 {
		usedPercent = float64(tenant.StorageUsed) / float64(tenant.StorageQuota) * 100
	}

	Success(c, gin.H{
		"storage_used":  tenant.StorageUsed,
		"storage_quota": tenant.StorageQuota,
		"used_percent":  usedPercent,
		"available":     tenant.StorageQuota - tenant.StorageUsed,
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
