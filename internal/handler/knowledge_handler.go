package handler

import (
	"net/http"

	"github.com/ashwinyue/next-rag/next-ai/internal/service"
	"github.com/ashwinyue/next-rag/next-ai/internal/service/knowledge"
	"github.com/gin-gonic/gin"
)

// KnowledgeHandler 知识库处理器
type KnowledgeHandler struct {
	svc *service.Services
}

// NewKnowledgeHandler 创建知识库处理器
func NewKnowledgeHandler(svc *service.Services) *KnowledgeHandler {
	return &KnowledgeHandler{svc: svc}
}

// CreateKnowledgeBase 创建知识库
func (h *KnowledgeHandler) CreateKnowledgeBase(c *gin.Context) {
	var req knowledge.CreateKnowledgeBaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: -1, Message: err.Error()})
		return
	}

	kb, err := h.svc.Knowledge.CreateKnowledgeBase(c.Request.Context(), &req)
	if err != nil {
		errorResponse(c, err)
		return
	}

	created(c, kb)
}

// GetKnowledgeBase 获取知识库
func (h *KnowledgeHandler) GetKnowledgeBase(c *gin.Context) {
	id := c.Param("id")

	kb, err := h.svc.Knowledge.GetKnowledgeBase(c.Request.Context(), id)
	if err != nil {
		errorResponse(c, err)
		return
	}

	success(c, kb)
}

// ListKnowledgeBases 列出知识库
func (h *KnowledgeHandler) ListKnowledgeBases(c *gin.Context) {
	page, size := getPagination(c)

	kbs, err := h.svc.Knowledge.ListKnowledgeBases(c.Request.Context(), &knowledge.ListKnowledgeBasesRequest{
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
			"items": kbs,
			"total": int64(len(kbs)),
			"page":  page,
			"size":  size,
		},
	})
}

// UpdateKnowledgeBase 更新知识库
func (h *KnowledgeHandler) UpdateKnowledgeBase(c *gin.Context) {
	id := c.Param("id")
	var req knowledge.CreateKnowledgeBaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: -1, Message: err.Error()})
		return
	}

	kb, err := h.svc.Knowledge.UpdateKnowledgeBase(c.Request.Context(), id, &req)
	if err != nil {
		errorResponse(c, err)
		return
	}

	success(c, kb)
}

// DeleteKnowledgeBase 删除知识库
func (h *KnowledgeHandler) DeleteKnowledgeBase(c *gin.Context) {
	id := c.Param("id")

	if err := h.svc.Knowledge.DeleteKnowledgeBase(c.Request.Context(), id); err != nil {
		errorResponse(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// UploadDocument 上传文档
func (h *KnowledgeHandler) UploadDocument(c *gin.Context) {
	var req knowledge.UploadDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: -1, Message: err.Error()})
		return
	}

	doc, err := h.svc.Knowledge.UploadDocument(c.Request.Context(), &req)
	if err != nil {
		errorResponse(c, err)
		return
	}

	created(c, doc)
}

// GetDocument 获取文档
func (h *KnowledgeHandler) GetDocument(c *gin.Context) {
	id := c.Param("id")

	doc, err := h.svc.Knowledge.GetDocument(c.Request.Context(), id)
	if err != nil {
		errorResponse(c, err)
		return
	}

	success(c, doc)
}

// ListDocuments 列出文档
func (h *KnowledgeHandler) ListDocuments(c *gin.Context) {
	kbID := c.Param("kb_id")
	page, size := getPagination(c)

	docs, err := h.svc.Knowledge.ListDocuments(c.Request.Context(), &knowledge.ListDocumentsRequest{
		KnowledgeBaseID: kbID,
		Page:            page,
		Size:            size,
	})
	if err != nil {
		errorResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"items": docs,
			"total": int64(len(docs)),
			"page":  page,
			"size":  size,
		},
	})
}

// DeleteDocument 删除文档
func (h *KnowledgeHandler) DeleteDocument(c *gin.Context) {
	id := c.Param("id")

	if err := h.svc.Knowledge.DeleteDocument(c.Request.Context(), id); err != nil {
		errorResponse(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// ProcessDocument 处理文档（解析、分块、向量化、索引）
func (h *KnowledgeHandler) ProcessDocument(c *gin.Context) {
	docID := c.Param("id")
	kbID := c.Query("kb_id")

	if kbID == "" {
		c.JSON(http.StatusBadRequest, Response{Code: -1, Message: "kb_id is required"})
		return
	}

	result, err := h.svc.Knowledge.ProcessDocument(c.Request.Context(), docID, kbID)
	if err != nil {
		errorResponse(c, err)
		return
	}

	success(c, result)
}

// CreateChunkIndex 创建文档块索引
func (h *KnowledgeHandler) CreateChunkIndex(c *gin.Context) {
	if err := knowledge.CreateChunkIndex(c.Request.Context(), h.svc.Config); err != nil {
		errorResponse(c, err)
		return
	}

	success(c, gin.H{
		"message": "chunk index created successfully",
	})
}
