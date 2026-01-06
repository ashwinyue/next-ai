// Package handler 提供模型管理的 HTTP 处理器
package handler

import (
	dataModel "github.com/ashwinyue/next-ai/internal/model"
	"github.com/ashwinyue/next-ai/internal/service/model"
	"github.com/gin-gonic/gin"
)

// ModelHandler 模型处理器
type ModelHandler struct {
	svc *model.Service
}

// NewModelHandler 创建模型处理器
func NewModelHandler(svc *model.Service) *ModelHandler {
	return &ModelHandler{svc: svc}
}

// CreateModelRequest 创建模型请求
type CreateModelRequest struct {
	Name        string                    `json:"name" binding:"required"`
	Type        dataModel.ModelType       `json:"type" binding:"required"`
	Source      dataModel.ModelSource     `json:"source" binding:"required"`
	Description string                    `json:"description"`
	Parameters  dataModel.ModelParameters `json:"parameters" binding:"required"`
	IsDefault   bool                      `json:"is_default"`
}

// UpdateModelRequest 更新模型请求
type UpdateModelRequest struct {
	Name        *string                    `json:"name"`
	Type        *dataModel.ModelType       `json:"type"`
	Source      *dataModel.ModelSource     `json:"source"`
	Description *string                    `json:"description"`
	Parameters  *dataModel.ModelParameters `json:"parameters"`
	IsDefault   *bool                      `json:"is_default"`
	Status      *dataModel.ModelStatus     `json:"status"`
}

// CreateModel 创建模型
// @Summary 创建模型
// @Tags 模型管理
// @Accept json
// @Produce json
// @Param request body CreateModelRequest true "模型信息"
// @Success 201 {object} dataModel.Model
// @Router /api/v1/models [post]
func (h *ModelHandler) CreateModel(c *gin.Context) {
	ctx := c.Request.Context()

	var req CreateModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	m := &dataModel.Model{
		Name:        req.Name,
		Type:        req.Type,
		Source:      req.Source,
		Description: req.Description,
		Parameters:  req.Parameters,
		IsDefault:   req.IsDefault,
		Status:      dataModel.ModelStatusActive,
	}

	if err := h.svc.CreateModel(ctx, m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, m)
}

// GetModel 获取模型详情
// @Summary 获取模型详情
// @Tags 模型管理
// @Accept json
// @Produce json
// @Param id path string true "模型ID"
// @Success 200 {object} dataModel.Model
// @Router /api/v1/models/{id} [get]
func (h *ModelHandler) GetModel(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	m, err := h.svc.GetModelByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "model not found"})
		return
	}

	// 隐藏内置模型的敏感信息
	if m.IsBuiltin {
		m.Parameters.APIKey = ""
		m.Parameters.BaseURL = ""
	}

	c.JSON(http.StatusOK, m)
}

// ListModels 列出模型
// @Summary 列出模型
// @Tags 模型管理
// @Accept json
// @Produce json
// @Param type query string false "模型类型"
// @Param source query string false "模型来源"
// @Success 200 {array} dataModel.Model
// @Router /api/v1/models [get]
func (h *ModelHandler) ListModels(c *gin.Context) {
	ctx := c.Request.Context()

	var modelType *dataModel.ModelType
	if typeStr := c.Query("type"); typeStr != "" {
		t := dataModel.ModelType(typeStr)
		modelType = &t
	}

	var source *dataModel.ModelSource
	if sourceStr := c.Query("source"); sourceStr != "" {
		s := dataModel.ModelSource(sourceStr)
		source = &s
	}

	models, err := h.svc.ListModels(ctx, modelType, source)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 隐藏内置模型的敏感信息
	for _, m := range models {
		if m.IsBuiltin {
			m.Parameters.APIKey = ""
			m.Parameters.BaseURL = ""
		}
	}

	c.JSON(http.StatusOK, gin.H{"models": models})
}

// UpdateModel 更新模型
// @Summary 更新模型
// @Tags 模型管理
// @Accept json
// @Produce json
// @Param id path string true "模型ID"
// @Param request body UpdateModelRequest true "更新内容"
// @Success 200 {object} dataModel.Model
// @Router /api/v1/models/{id} [put]
func (h *ModelHandler) UpdateModel(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	// 获取现有模型
	m, err := h.svc.GetModelByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "model not found"})
		return
	}

	var req UpdateModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 更新字段
	if req.Name != nil {
		m.Name = *req.Name
	}
	if req.Type != nil {
		m.Type = *req.Type
	}
	if req.Source != nil {
		m.Source = *req.Source
	}
	if req.Description != nil {
		m.Description = *req.Description
	}
	if req.Parameters != nil {
		m.Parameters = *req.Parameters
	}
	if req.IsDefault != nil {
		m.IsDefault = *req.IsDefault
	}
	if req.Status != nil {
		m.Status = *req.Status
	}

	if err := h.svc.UpdateModel(ctx, m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, m)
}

// DeleteModel 删除模型
// @Summary 删除模型
// @Tags 模型管理
// @Accept json
// @Produce json
// @Param id path string true "模型ID"
// @Success 204
// @Router /api/v1/models/{id} [delete]
func (h *ModelHandler) DeleteModel(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	if err := h.svc.DeleteModel(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	NoContent(c)
}

// ListModelProviders 列出支持的模型提供商
// @Summary 列出模型提供商
// @Tags 模型管理
// @Accept json
// @Produce json
// @Success 200 {array} dataModel.ModelProvider
// @Router /api/v1/models/providers [get]
func (h *ModelHandler) ListModelProviders(c *gin.Context) {
	ctx := c.Request.Context()

	providers := h.svc.ListModelProviders(ctx)
	c.JSON(http.StatusOK, gin.H{"providers": providers})
}
