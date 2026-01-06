package handler

import (
	"net/http"
	"strconv"

	"github.com/ashwinyue/next-rag/next-ai/internal/service"
	"github.com/ashwinyue/next-rag/next-ai/internal/service/faq"
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
