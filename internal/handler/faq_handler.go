package handler

import (
	"net/http"
	"strconv"

	"github.com/ashwinyue/next-ai/internal/model"
	"github.com/ashwinyue/next-ai/internal/service"
	"github.com/ashwinyue/next-ai/internal/service/faq"
	"github.com/gin-gonic/gin"
)

// FAQHandler FAQ处理器
type FAQHandler struct {
	svc *service.Services
}

// NewFAQHandler 创建FAQ处理器
func NewFAQHandler(svc *service.Services) *FAQHandler {
	return &FAQHandler{svc: svc}
}

// CreateFAQ 创建FAQ
func (h *FAQHandler) CreateFAQ(c *gin.Context) {
	var req faq.CreateFAQRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: -1, Message: err.Error()})
		return
	}

	faq, err := h.svc.FAQ.CreateFAQ(c.Request.Context(), &req)
	if err != nil {
		errorResponse(c, err)
		return
	}

	created(c, faq)
}

// GetFAQ 获取FAQ
func (h *FAQHandler) GetFAQ(c *gin.Context) {
	id := c.Param("id")

	faq, err := h.svc.FAQ.GetFAQ(c.Request.Context(), id)
	if err != nil {
		errorResponse(c, err)
		return
	}

	success(c, faq)
}

// ListFAQs 列出FAQ
func (h *FAQHandler) ListFAQs(c *gin.Context) {
	category := c.Query("category")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 20
	}

	faqs, err := h.svc.FAQ.ListFAQs(c.Request.Context(), &faq.ListFAQsRequest{
		Category: category,
		Page:     page,
		Size:     size,
	})
	if err != nil {
		errorResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"items": faqs,
			"total": int64(len(faqs)),
			"page":  page,
			"size":  size,
		},
	})
}

// ListActiveFAQs 列出活跃FAQ
func (h *FAQHandler) ListActiveFAQs(c *gin.Context) {
	category := c.Query("category")

	faqs, err := h.svc.FAQ.ListActiveFAQs(c.Request.Context(), category)
	if err != nil {
		errorResponse(c, err)
		return
	}

	success(c, faqs)
}

// UpdateFAQ 更新FAQ
func (h *FAQHandler) UpdateFAQ(c *gin.Context) {
	id := c.Param("id")
	var req faq.CreateFAQRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: -1, Message: err.Error()})
		return
	}

	faq, err := h.svc.FAQ.UpdateFAQ(c.Request.Context(), id, &req)
	if err != nil {
		errorResponse(c, err)
		return
	}

	success(c, faq)
}

// DeleteFAQ 删除FAQ
func (h *FAQHandler) DeleteFAQ(c *gin.Context) {
	id := c.Param("id")

	if err := h.svc.FAQ.DeleteFAQ(c.Request.Context(), id); err != nil {
		errorResponse(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// SearchFAQs 搜索FAQ
func (h *FAQHandler) SearchFAQs(c *gin.Context) {
	keyword := c.Query("q")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, Response{Code: -1, Message: "search keyword is required"})
		return
	}

	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	faqs, err := h.svc.FAQ.SearchFAQs(c.Request.Context(), keyword, limit)
	if err != nil {
		errorResponse(c, err)
		return
	}

	success(c, faqs)
}

// ========== FAQEntry 增强版方法 ==========

// CreateEntry 创建FAQ条目
func (h *FAQHandler) CreateEntry(c *gin.Context) {
	var req faq.CreateEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: -1, Message: err.Error()})
		return
	}

	entry, err := h.svc.FAQEntry.CreateEntry(c.Request.Context(), &req)
	if err != nil {
		errorResponse(c, err)
		return
	}

	created(c, entry)
}

// GetEntry 获取FAQ条目
func (h *FAQHandler) GetEntry(c *gin.Context) {
	id := c.Param("id")

	entry, err := h.svc.FAQEntry.GetEntry(c.Request.Context(), id)
	if err != nil {
		errorResponse(c, err)
		return
	}

	success(c, entry)
}

// ListEntries 列出FAQ条目
func (h *FAQHandler) ListEntries(c *gin.Context) {
	category := c.Query("category")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	var isEnabled *bool
	if enabled := c.Query("is_enabled"); enabled != "" {
		if b, err := strconv.ParseBool(enabled); err == nil {
			isEnabled = &b
		}
	}

	entries, err := h.svc.FAQEntry.ListEntries(c.Request.Context(), &faq.ListEntriesRequest{
		Category:  category,
		IsEnabled: isEnabled,
		Page:      page,
		Size:      size,
	})
	if err != nil {
		errorResponse(c, err)
		return
	}

	success(c, entries)
}

// UpdateEntry 更新FAQ条目
func (h *FAQHandler) UpdateEntry(c *gin.Context) {
	id := c.Param("id")
	var req faq.UpdateEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: -1, Message: err.Error()})
		return
	}

	entry, err := h.svc.FAQEntry.UpdateEntry(c.Request.Context(), id, &req)
	if err != nil {
		errorResponse(c, err)
		return
	}

	success(c, entry)
}

// DeleteEntry 删除FAQ条目
func (h *FAQHandler) DeleteEntry(c *gin.Context) {
	id := c.Param("id")

	if err := h.svc.FAQEntry.DeleteEntry(c.Request.Context(), id); err != nil {
		errorResponse(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// DeleteEntries 批量删除FAQ条目
func (h *FAQHandler) DeleteEntries(c *gin.Context) {
	var req struct {
		IDs []string `json:"ids" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: -1, Message: err.Error()})
		return
	}

	if err := h.svc.FAQEntry.DeleteEntries(c.Request.Context(), req.IDs); err != nil {
		errorResponse(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// SearchEntries 搜索FAQ条目
func (h *FAQHandler) SearchEntries(c *gin.Context) {
	keyword := c.Query("q")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, Response{Code: -1, Message: "search keyword is required"})
		return
	}

	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	entries, err := h.svc.FAQEntry.SearchEntries(c.Request.Context(), keyword, limit)
	if err != nil {
		errorResponse(c, err)
		return
	}

	success(c, entries)
}

// UpdateEntryCategoryBatch 批量更新FAQ条目分类
func (h *FAQHandler) UpdateEntryCategoryBatch(c *gin.Context) {
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: -1, Message: err.Error()})
		return
	}

	if err := h.svc.FAQEntry.UpdateEntryCategoryBatch(c.Request.Context(), req); err != nil {
		errorResponse(c, err)
		return
	}

	success(c, gin.H{"message": "分类更新成功"})
}

// UpdateEntryFieldsBatch 批量更新FAQ条目字段
func (h *FAQHandler) UpdateEntryFieldsBatch(c *gin.Context) {
	var req model.FAQEntryFieldsBatchUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: -1, Message: err.Error()})
		return
	}

	if err := h.svc.FAQEntry.UpdateEntryFieldsBatch(c.Request.Context(), &req); err != nil {
		errorResponse(c, err)
		return
	}

	success(c, gin.H{"message": "字段更新成功"})
}

// ExportEntries 导出FAQ条目
func (h *FAQHandler) ExportEntries(c *gin.Context) {
	category := c.Query("category")

	data, err := h.svc.FAQEntry.ExportEntries(c.Request.Context(), category)
	if err != nil {
		errorResponse(c, err)
		return
	}

	c.Header("Content-Type", "application/json; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=faq_export.json")
	c.Data(http.StatusOK, "application/json", data)
}

// BatchUpsert 批量导入FAQ条目
func (h *FAQHandler) BatchUpsert(c *gin.Context) {
	var req faq.BatchUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: -1, Message: err.Error()})
		return
	}

	resp, err := h.svc.FAQEntry.BatchUpsert(c.Request.Context(), &req)
	if err != nil {
		errorResponse(c, err)
		return
	}

	success(c, resp)
}

// GetImportProgress 获取导入进度
func (h *FAQHandler) GetImportProgress(c *gin.Context) {
	taskID := c.Param("task_id")

	progress, err := h.svc.FAQEntry.GetImportProgress(c.Request.Context(), taskID)
	if err != nil {
		errorResponse(c, err)
		return
	}

	success(c, progress)
}
