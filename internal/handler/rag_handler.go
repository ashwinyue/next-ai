package handler

import (
	"net/http"

	"github.com/ashwinyue/next-rag/next-ai/internal/service"
	"github.com/ashwinyue/next-rag/next-ai/internal/service/rag"
	"github.com/gin-gonic/gin"
)

// RAGHandler RAG 处理器
type RAGHandler struct {
	svc *service.Services
}

// NewRAGHandler 创建 RAG 处理器
func NewRAGHandler(svc *service.Services) *RAGHandler {
	return &RAGHandler{svc: svc}
}

// RetrieveRequest 检索请求
type RetrieveRequest struct {
	Query          string `json:"query" binding:"required"`
	TopK           int    `json:"top_k"`
	EnableOptimize bool   `json:"enable_optimize"`
	EnableRerank   bool   `json:"enable_rerank"`
}

// Retrieve 执行 RAG 检索
func (h *RAGHandler) Retrieve(c *gin.Context) {
	var req RetrieveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: -1, Message: err.Error()})
		return
	}

	// 创建 RAG 服务
	ragSvc := rag.NewService(
		h.svc.ChatModel,
		h.svc.Retriever,
		h.svc.Query,
		h.svc.Rerankers,
	)

	// 执行检索
	resp, err := ragSvc.Retrieve(c.Request.Context(), &rag.RetrieveRequest{
		Query:          req.Query,
		TopK:           req.TopK,
		EnableOptimize: req.EnableOptimize,
		EnableRerank:   req.EnableRerank,
	})
	if err != nil {
		errorResponse(c, err)
		return
	}

	success(c, resp)
}

// RetrieveSimple 简化检索（默认参数）
func (h *RAGHandler) RetrieveSimple(c *gin.Context) {
	query := c.Query("query")
	if query == "" {
		c.JSON(http.StatusBadRequest, Response{Code: -1, Message: "query is required"})
		return
	}

	// 创建 RAG 服务
	ragSvc := rag.NewService(
		h.svc.ChatModel,
		h.svc.Retriever,
		h.svc.Query,
		h.svc.Rerankers,
	)

	// 执行检索（启用优化和重排）
	resp, err := ragSvc.Retrieve(c.Request.Context(), &rag.RetrieveRequest{
		Query:          query,
		TopK:           10,
		EnableOptimize: true,
		EnableRerank:   true,
	})
	if err != nil {
		errorResponse(c, err)
		return
	}

	success(c, resp)
}
