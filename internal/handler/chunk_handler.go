package handler

import (
	"net/http"
	"strconv"

	chunksvc "github.com/ashwinyue/next-ai/internal/service/chunk"
	"github.com/gin-gonic/gin"
)

// ChunkHandler 分块处理器
type ChunkHandler struct {
	svc *chunksvc.Service
}

// NewChunkHandler 创建分块处理器
func NewChunkHandler(svc *chunksvc.Service) *ChunkHandler {
	return &ChunkHandler{svc: svc}
}

// GetChunkByID 获取单个分块
// GET /api/v1/chunks/:id
func (h *ChunkHandler) GetChunkByID(c *gin.Context) {
	chunkID := c.Param("id")
	if chunkID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "Chunk ID is required",
		})
		return
	}

	chunk, err := h.svc.GetChunkByID(c.Request.Context(), chunkID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    -1,
			"message": "Chunk not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    chunk,
	})
}

// ListChunksByKnowledgeBaseID 获取知识库的所有分块
// GET /api/v1/knowledge-bases/:kb_id/chunks
func (h *ChunkHandler) ListChunksByKnowledgeBaseID(c *gin.Context) {
	kbID := c.Param("kb_id")
	if kbID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "Knowledge base ID is required",
		})
		return
	}

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	chunks, total, err := h.svc.ListChunksByKnowledgeBaseID(c.Request.Context(), kbID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":      0,
		"message":   "success",
		"data":      chunks,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// UpdateChunk 更新分块
// PUT /api/v1/chunks/:id
func (h *ChunkHandler) UpdateChunk(c *gin.Context) {
	chunkID := c.Param("id")
	if chunkID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "Chunk ID is required",
		})
		return
	}

	var req chunksvc.UpdateChunkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "Invalid parameters",
		})
		return
	}

	updatedChunk, err := h.svc.UpdateChunk(c.Request.Context(), chunkID, &req)
	if err != nil {
		if err == chunksvc.ErrChunkNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    -1,
				"message": "Chunk not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "Chunk updated successfully",
		"data":    updatedChunk,
	})
}

// DeleteChunk 删除单个分块
// DELETE /api/v1/chunks/:id
func (h *ChunkHandler) DeleteChunk(c *gin.Context) {
	chunkID := c.Param("id")
	if chunkID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "Chunk ID is required",
		})
		return
	}

	if err := h.svc.DeleteChunk(c.Request.Context(), chunkID); err != nil {
		if err == chunksvc.ErrChunkNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    -1,
				"message": "Chunk not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "Chunk deleted successfully",
	})
}

// DeleteChunksByDocumentID 删除文档的所有分块
// DELETE /api/v1/documents/:doc_id/chunks
func (h *ChunkHandler) DeleteChunksByDocumentID(c *gin.Context) {
	docID := c.Param("doc_id")
	if docID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "Document ID is required",
		})
		return
	}

	if err := h.svc.DeleteChunksByDocumentID(c.Request.Context(), docID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "Chunks deleted successfully",
	})
}

// DeleteChunksByKnowledgeBaseID 删除知识库的所有分块
// DELETE /api/v1/knowledge-bases/:kb_id/chunks
func (h *ChunkHandler) DeleteChunksByKnowledgeBaseID(c *gin.Context) {
	kbID := c.Param("kb_id")
	if kbID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "Knowledge base ID is required",
		})
		return
	}

	if err := h.svc.DeleteChunksByKnowledgeBaseID(c.Request.Context(), kbID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "All chunks deleted successfully",
	})
}
