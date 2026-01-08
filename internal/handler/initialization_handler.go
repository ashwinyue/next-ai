package handler

import (
	"github.com/ashwinyue/next-ai/internal/service"
	"github.com/ashwinyue/next-ai/internal/service/initialization"
	"github.com/gin-gonic/gin"
)

// InitializationHandler 初始化处理器
type InitializationHandler struct {
	svc *service.Services
}

// NewInitializationHandler 创建初始化处理器
func NewInitializationHandler(svc *service.Services) *InitializationHandler {
	return &InitializationHandler{svc: svc}
}

// CheckOllamaStatus 检查 Ollama 服务状态
// @Summary      检查 Ollama 服务状态
// @Description  检查 Ollama 服务是否可用
// @Tags         初始化
// @Accept       json
// @Produce      json
// @Success      200  {object}  Response
// @Router       /api/v1/initialization/ollama/status [get]
func (h *InitializationHandler) CheckOllamaStatus(c *gin.Context) {
	ctx := c.Request.Context()

	status, err := h.svc.Initialization.CheckOllamaStatus(ctx)
	if err != nil {
		status = &initialization.OllamaStatusResponse{
			Available: false,
			Error:     err.Error(),
		}
	}

	Success(c, status)
}

// UpdateKBConfigRequest 更新知识库配置请求
type UpdateKBConfigRequest = initialization.UpdateKBConfigRequest

// UpdateKBConfig 更新知识库配置
// @Summary      更新知识库配置
// @Description  更新知识库的分块配置和问答配置
// @Tags         初始化
// @Accept       json
// @Produce      json
// @Param        kbId   path      string                  true  "知识库ID"
// @Param        request body      UpdateKBConfigRequest  true  "配置请求"
// @Success      200     {object}  Response
// @Router       /api/v1/initialization/kb/{kbId}/config [put]
func (h *InitializationHandler) UpdateKBConfig(c *gin.Context) {
	kbID := c.Param("kbId")

	var req UpdateKBConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	if err := h.svc.Initialization.UpdateKBConfig(c.Request.Context(), kbID, &req); err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{"message": "配置更新成功"})
}

// GetKBConfig 获取知识库配置
// @Summary      获取知识库配置
// @Description  获取知识库的当前配置
// @Tags         初始化
// @Accept       json
// @Produce      json
// @Param        kbId   path      string  true  "知识库ID"
// @Success      200     {object}  Response
// @Router       /api/v1/initialization/kb/{kbId}/config [get]
func (h *InitializationHandler) GetKBConfig(c *gin.Context) {
	kbID := c.Param("kbId")

	config, err := h.svc.Initialization.GetKBConfig(c.Request.Context(), kbID)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, config)
}

// InitializeByKBRequest 初始化知识库请求
type InitializeByKBRequest = initialization.InitializeByKBRequest

// InitializeByKB 初始化知识库
// @Summary      初始化知识库
// @Description  初始化知识库配置
// @Tags         初始化
// @Accept       json
// @Produce      json
// @Param        kbId     path      string                     true  "知识库ID"
// @Param        request  body      InitializeByKBRequest  true  "初始化请求"
// @Success      200      {object}  Response
// @Router       /api/v1/initialization/kb/{kbId} [post]
func (h *InitializationHandler) InitializeByKB(c *gin.Context) {
	kbID := c.Param("kbId")

	var req InitializeByKBRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	if err := h.svc.Initialization.InitializeByKB(c.Request.Context(), kbID, &req); err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{"message": "知识库初始化成功"})
}

// GetSystemInfo 获取系统信息
// @Summary      获取系统信息
// @Description  获取系统版本和运行状态信息
// @Tags         初始化
// @Accept       json
// @Produce      json
// @Success      200  {object}  Response
// @Router       /api/v1/initialization/system/info [get]
func (h *InitializationHandler) GetSystemInfo(c *gin.Context) {
	ctx := c.Request.Context()

	info, err := h.svc.Initialization.GetSystemInfo(ctx)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, info)
}

// TestEmbeddingRequest 测试 Embedding 请求
type TestEmbeddingRequest = initialization.TestEmbeddingRequest

// TestEmbeddingResponse 测试 Embedding 响应
type TestEmbeddingResponse = initialization.TestEmbeddingResponse

// TestEmbedding 测试 Embedding 模型
// @Summary      测试 Embedding 模型
// @Description  测试 Embedding 接口是否可用
// @Tags         初始化
// @Accept       json
// @Produce      json
// @Param        request  body      TestEmbeddingRequest  true  "测试请求"
// @Success      200      {object}  Response
// @Router       /api/v1/initialization/test/embedding [post]
func (h *InitializationHandler) TestEmbedding(c *gin.Context) {
	var req TestEmbeddingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	result, err := h.svc.Initialization.TestEmbedding(c.Request.Context(), &req)
	if err != nil {
		result = &TestEmbeddingResponse{
			Available: false,
			Message:   err.Error(),
		}
	}

	Success(c, result)
}

// ListOllamaModels 列出已安装的 Ollama 模型
// @Summary      列出 Ollama 模型
// @Description  列出已安装的 Ollama 模型
// @Tags         初始化
// @Accept       json
// @Produce      json
// @Success      200  {object}  Response
// @Router       /api/v1/initialization/ollama/models [get]
func (h *InitializationHandler) ListOllamaModels(c *gin.Context) {
	ctx := c.Request.Context()

	models, err := h.svc.Initialization.ListOllamaModels(ctx)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{"models": models})
}

// CheckOllamaModelsRequest 检查 Ollama 模型请求
type CheckOllamaModelsRequest = initialization.CheckOllamaModelsRequest

// CheckOllamaModels 检查指定的 Ollama 模型是否已安装
// @Summary      检查 Ollama 模型
// @Description  检查指定的 Ollama 模型是否已安装
// @Tags         初始化
// @Accept       json
// @Produce      json
// @Param        request  body      CheckOllamaModelsRequest  true  "检查请求"
// @Success      200      {object}  Response
// @Router       /api/v1/initialization/ollama/models/check [post]
func (h *InitializationHandler) CheckOllamaModels(c *gin.Context) {
	var req CheckOllamaModelsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	result, err := h.svc.Initialization.CheckOllamaModels(c.Request.Context(), &req)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{"models": result})
}

// CheckRemoteModelRequest 检查远程模型请求
type CheckRemoteModelRequest = initialization.CheckRemoteModelRequest

// CheckRemoteModelResponse 检查远程模型响应
type CheckRemoteModelResponse = initialization.CheckRemoteModelResponse

// CheckRemoteModel 检查远程 LLM 模型连接
// @Summary      检查远程模型
// @Description  检查远程 API 模型连接是否正常
// @Tags         初始化
// @Accept       json
// @Produce      json
// @Param        request  body      CheckRemoteModelRequest  true  "检查请求"
// @Success      200      {object}  Response
// @Router       /api/v1/initialization/models/remote/check [post]
func (h *InitializationHandler) CheckRemoteModel(c *gin.Context) {
	var req CheckRemoteModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	result, err := h.svc.Initialization.CheckRemoteModel(c.Request.Context(), &req)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, result)
}

// CheckRerankModelRequest 检查 Rerank 模型请求
type CheckRerankModelRequest = initialization.CheckRerankModelRequest

// CheckRerankModelResponse 检查 Rerank 模型响应
type CheckRerankModelResponse = initialization.CheckRerankModelResponse

// CheckRerankModel 检查 Rerank 模型连接和功能
// @Summary      检查 Rerank 模型
// @Description  检查 Rerank 模型连接和功能是否正常
// @Tags         初始化
// @Accept       json
// @Produce      json
// @Param        request  body      CheckRerankModelRequest  true  "检查请求"
// @Success      200      {object}  Response
// @Router       /api/v1/initialization/models/rerank/check [post]
func (h *InitializationHandler) CheckRerankModel(c *gin.Context) {
	var req CheckRerankModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	result, err := h.svc.Initialization.CheckRerankModel(c.Request.Context(), &req)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, result)
}

// ========== 文本处理功能（WeKnora API 兼容）==========

// ExtractTextRelationsRequest 提取文本关系请求
type ExtractTextRelationsRequest = initialization.ExtractTextRelationsRequest

// ExtractTextRelations 提取文本关系
// POST /api/v1/initialization/extract/text-relation
func (h *InitializationHandler) ExtractTextRelations(c *gin.Context) {
	var req ExtractTextRelationsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	result, err := h.svc.Initialization.ExtractTextRelations(c.Request.Context(), &req)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, result)
}

// FabriTagRequest 生成标签请求
type FabriTagRequest = initialization.FabriTagRequest

// FabriTag 生成文本标签
// POST /api/v1/initialization/extract/fabri-tag
func (h *InitializationHandler) FabriTag(c *gin.Context) {
	var req FabriTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	result, err := h.svc.Initialization.FabriTag(c.Request.Context(), &req)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, result)
}

// FabriTextRequest 生成文本请求
type FabriTextRequest = initialization.FabriTextRequest

// FabriText 根据标签生成文本
// POST /api/v1/initialization/extract/fabri-text
func (h *InitializationHandler) FabriText(c *gin.Context) {
	var req FabriTextRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	result, err := h.svc.Initialization.FabriText(c.Request.Context(), &req)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, result)
}

// TestMultimodalRequest 测试多模态请求
type TestMultimodalRequest = initialization.TestMultimodalRequest

// TestMultimodal 测试多模态功能
// POST /api/v1/initialization/multimodal/test
func (h *InitializationHandler) TestMultimodal(c *gin.Context) {
	var req TestMultimodalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	result, err := h.svc.Initialization.TestMultimodal(c.Request.Context(), &req)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, result)
}

// ========== Ollama 模型下载 ==========

// DownloadModelRequest 下载模型请求
type DownloadModelRequest = initialization.DownloadModelRequest

// DownloadModel 下载 Ollama 模型
// @Summary      下载 Ollama 模型
// @Description  启动 Ollama 模型下载任务（异步）
// @Tags         初始化
// @Accept       json
// @Produce      json
// @Param        request  body      DownloadModelRequest  true  "下载请求"
// @Success      200      {object}  Response
// @Router       /api/v1/initialization/ollama/models/download [post]
func (h *InitializationHandler) DownloadModel(c *gin.Context) {
	var req DownloadModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	result, err := h.svc.Initialization.DownloadModel(c.Request.Context(), &req)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, result)
}

// GetDownloadProgress 获取下载进度
// @Summary      获取下载进度
// @Description  获取 Ollama 模型下载任务进度
// @Tags         初始化
// @Accept       json
// @Produce      json
// @Param        task_id  path      string  true  "任务 ID"
// @Success      200      {object}  Response
// @Router       /api/v1/initialization/ollama/download/progress/{task_id} [get]
func (h *InitializationHandler) GetDownloadProgress(c *gin.Context) {
	taskID := c.Param("task_id")

	progress, err := h.svc.Initialization.GetDownloadProgress(c.Request.Context(), taskID)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, progress)
}

// ListDownloadTasks 列出所有下载任务
// @Summary      列出下载任务
// @Description  列出所有 Ollama 模型下载任务
// @Tags         初始化
// @Accept       json
// @Produce      json
// @Success      200  {object}  Response
// @Router       /api/v1/initialization/ollama/download/tasks [get]
func (h *InitializationHandler) ListDownloadTasks(c *gin.Context) {
	tasks, err := h.svc.Initialization.ListDownloadTasks(c.Request.Context())
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{"tasks": tasks})
}

// CancelDownloadRequest 取消下载请求
type CancelDownloadRequest struct {
	TaskID string `json:"task_id" binding:"required"`
}

// CancelDownload 取消下载任务
// @Summary      取消下载任务
// @Description  取消正在进行的 Ollama 模型下载任务
// @Tags         初始化
// @Accept       json
// @Produce      json
// @Param        task_id  path      string  true  "任务 ID"
// @Success      200      {object}  Response
// @Router       /api/v1/initialization/ollama/download/cancel/{task_id} [post]
func (h *InitializationHandler) CancelDownload(c *gin.Context) {
	taskID := c.Param("task_id")

	err := h.svc.Initialization.CancelDownload(c.Request.Context(), taskID)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{
		"success": true,
		"message": "下载任务已取消",
		"task_id": taskID,
	})
}
