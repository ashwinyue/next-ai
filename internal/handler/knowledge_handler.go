package handler

import (
	"strings"

	"github.com/ashwinyue/next-ai/internal/service"
	"github.com/ashwinyue/next-ai/internal/service/knowledge"
	"github.com/ashwinyue/next-ai/internal/service/tag"
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
		BadRequest(c, err.Error())
		return
	}

	kb, err := h.svc.Knowledge.CreateKnowledgeBase(c.Request.Context(), &req)
	if err != nil {
		Error(c, err)
		return
	}

	Created(c, kb)
}

// GetKnowledgeBase 获取知识库
func (h *KnowledgeHandler) GetKnowledgeBase(c *gin.Context) {
	id := c.Param("id")

	kb, err := h.svc.Knowledge.GetKnowledgeBase(c.Request.Context(), id)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, kb)
}

// ListKnowledgeBases 列出知识库
func (h *KnowledgeHandler) ListKnowledgeBases(c *gin.Context) {
	page, pageSize := getPagination(c)

	kbs, err := h.svc.Knowledge.ListKnowledgeBases(c.Request.Context(), &knowledge.ListKnowledgeBasesRequest{
		Page: page,
		Size: pageSize,
	})
	if err != nil {
		Error(c, err)
		return
	}

	SuccessWithPagination(c, kbs, int64(len(kbs)), page, pageSize)
}

// UpdateKnowledgeBase 更新知识库
func (h *KnowledgeHandler) UpdateKnowledgeBase(c *gin.Context) {
	id := c.Param("id")
	var req knowledge.CreateKnowledgeBaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	kb, err := h.svc.Knowledge.UpdateKnowledgeBase(c.Request.Context(), id, &req)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, kb)
}

// DeleteKnowledgeBase 删除知识库
func (h *KnowledgeHandler) DeleteKnowledgeBase(c *gin.Context) {
	id := c.Param("id")

	if err := h.svc.Knowledge.DeleteKnowledgeBase(c.Request.Context(), id); err != nil {
		Error(c, err)
		return
	}

	NoContent(c)
}

// UploadDocument 上传文档
func (h *KnowledgeHandler) UploadDocument(c *gin.Context) {
	var req knowledge.UploadDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	doc, err := h.svc.Knowledge.UploadDocument(c.Request.Context(), &req)
	if err != nil {
		Error(c, err)
		return
	}

	Created(c, doc)
}

// GetDocument 获取文档
func (h *KnowledgeHandler) GetDocument(c *gin.Context) {
	id := c.Param("id")

	doc, err := h.svc.Knowledge.GetDocument(c.Request.Context(), id)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, doc)
}

// ListDocuments 列出文档
func (h *KnowledgeHandler) ListDocuments(c *gin.Context) {
	kbID := c.Param("id")
	page, pageSize := getPagination(c)

	docs, err := h.svc.Knowledge.ListDocuments(c.Request.Context(), &knowledge.ListDocumentsRequest{
		KnowledgeBaseID: kbID,
		Page:            page,
		Size:            pageSize,
	})
	if err != nil {
		Error(c, err)
		return
	}

	SuccessWithPagination(c, docs, int64(len(docs)), page, pageSize)
}

// DeleteDocument 删除文档
func (h *KnowledgeHandler) DeleteDocument(c *gin.Context) {
	id := c.Param("id")

	if err := h.svc.Knowledge.DeleteDocument(c.Request.Context(), id); err != nil {
		Error(c, err)
		return
	}

	NoContent(c)
}

// ProcessDocument 处理文档（解析、分块、向量化、索引）
func (h *KnowledgeHandler) ProcessDocument(c *gin.Context) {
	docID := c.Param("id")
	kbID := c.Query("kb_id")

	if kbID == "" {
		BadRequest(c, "kb_id is required")
		return
	}

	result, err := h.svc.Knowledge.ProcessDocument(c.Request.Context(), docID, kbID)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, result)
}

// CreateChunkIndex 创建文档块索引
func (h *KnowledgeHandler) CreateChunkIndex(c *gin.Context) {
	if err := knowledge.CreateChunkIndex(c.Request.Context(), h.svc.Config, h.svc.Embedder); err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{
		"message": "chunk index created successfully",
	})
}

// HybridSearch 混合搜索（向量 + 关键词）
// 兼容 WeKnora API 格式
func (h *KnowledgeHandler) HybridSearch(c *gin.Context) {
	id := c.Param("id")

	var req knowledge.HybridSearchParams
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	results, err := h.svc.Knowledge.HybridSearch(c.Request.Context(), id, &req)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{
		"success": true,
		"data":    results,
	})
}

// CopyKnowledgeBase 复制知识库（WeKnora API 兼容）
func (h *KnowledgeHandler) CopyKnowledgeBase(c *gin.Context) {
	var req knowledge.CopyKnowledgeBaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	result, err := h.svc.Knowledge.CopyKnowledgeBase(c.Request.Context(), &req)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{
		"success": true,
		"data":    result,
	})
}

// GetKBCloneProgress 获取知识库复制进度（WeKnora API 兼容）
func (h *KnowledgeHandler) GetKBCloneProgress(c *gin.Context) {
	taskID := c.Param("task_id")

	progress, err := h.svc.Knowledge.GetKBCloneProgress(c.Request.Context(), taskID)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{
		"success": true,
		"data":    progress,
	})
}

// ========== 知识批量管理（WeKnora API 兼容）==========

// GetKnowledgeBatch 批量获取知识
// GET /api/v1/knowledge/batch?ids=id1,id2,id3
func (h *KnowledgeHandler) GetKnowledgeBatch(c *gin.Context) {
	var ids []string
	if idsParam := c.Query("ids"); idsParam != "" {
		// 解析逗号分隔的 ID 列表
		ids = strings.Split(idsParam, ",")
		// 去除空白
		for i, id := range ids {
			ids[i] = strings.TrimSpace(id)
		}
	} else {
		// 尝试从 JSON body 解析
		var req struct {
			IDs []string `json:"ids"`
		}
		if err := c.ShouldBindJSON(&req); err == nil {
			ids = req.IDs
		}
	}

	if len(ids) == 0 {
		BadRequest(c, "ids is required")
		return
	}

	// 获取知识列表
	documents := make([]interface{}, 0)
	for _, id := range ids {
		doc, err := h.svc.Knowledge.GetDocument(c.Request.Context(), id)
		if err == nil {
			documents = append(documents, doc)
		}
	}

	Success(c, gin.H{
		"success": true,
		"data":    documents,
	})
}

// UpdateKnowledgeTagBatchRequest 批量更新标签请求
type UpdateKnowledgeTagBatchRequest struct {
	Updates []TagUpdate `json:"updates"`
}

// TagUpdate 标签更新
type TagUpdate struct {
	KnowledgeID string   `json:"knowledge_id"`
	TagIDs      []string `json:"tag_ids"`
}

// UpdateKnowledgeTags 批量更新知识标签
// PUT /api/v1/knowledge/tags
func (h *KnowledgeHandler) UpdateKnowledgeTags(c *gin.Context) {
	var req UpdateKnowledgeTagBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	// 转换为 service 格式
	updates := make([]tag.TagUpdate, 0, len(req.Updates))
	for _, u := range req.Updates {
		updates = append(updates, tag.TagUpdate{
			KnowledgeID: u.KnowledgeID,
			TagIDs:      u.TagIDs,
		})
	}

	if err := h.svc.Tag.BatchUpdateDocumentTags(c.Request.Context(), updates); err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{
		"success": true,
		"message": "Tags updated successfully",
	})
}

// UpdateImageInfoRequest 更新图像信息请求
type UpdateImageInfoRequest struct {
	ImageInfo string `json:"image_info"`
}

// UpdateImageInfo 更新分块图像信息
// PUT /api/v1/knowledge/image/:id/:chunk_id
func (h *KnowledgeHandler) UpdateImageInfo(c *gin.Context) {
	id := c.Param("id")
	chunkID := c.Param("chunk_id")

	var req UpdateImageInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	// 验证 chunk 属于指定的 document
	chunk, err := h.svc.Knowledge.GetChunk(c.Request.Context(), chunkID)
	if err != nil {
		Error(c, err)
		return
	}

	if chunk.DocumentID != id {
		BadRequest(c, "chunk does not belong to the specified document")
		return
	}

	// 调用 service 更新图像信息
	if err := h.svc.Knowledge.UpdateChunkImageInfo(c.Request.Context(), &knowledge.UpdateChunkImageInfoRequest{
		ChunkID:   chunkID,
		ImageInfo: req.ImageInfo,
	}); err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{
		"success":  true,
		"message":  "Image info updated",
		"id":       id,
		"chunk_id": chunkID,
	})
}

// DownloadKnowledge 下载知识文件
// GET /api/v1/knowledge/:id/download
func (h *KnowledgeHandler) DownloadKnowledge(c *gin.Context) {
	id := c.Param("id")

	// 获取文档和文件内容
	doc, reader, err := h.svc.File.DownloadKnowledge(c.Request.Context(), id)
	if err != nil {
		Error(c, err)
		return
	}
	defer reader.Close()

	// 设置响应头
	c.Header("Content-Disposition", `attachment; filename="`+doc.FileName+`"`)
	c.Header("Content-Type", doc.ContentType)

	// 流式传输文件内容
	c.DataFromReader(200, doc.FileSize, doc.ContentType, reader, nil)
}

// DeleteQuestionsByChunk 删除分块关联的问题
// DELETE /api/v1/chunks/questions/:chunk_id
func (h *KnowledgeHandler) DeleteQuestionsByChunk(c *gin.Context) {
	chunkID := c.Param("chunk_id")

	if err := h.svc.Knowledge.DeleteQuestionsByChunk(c.Request.Context(), chunkID); err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{
		"success":  true,
		"message":  "Questions deleted",
		"chunk_id": chunkID,
	})
}
