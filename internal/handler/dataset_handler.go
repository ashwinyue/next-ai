package handler

import (
	"strconv"

	"github.com/ashwinyue/next-ai/internal/service/dataset"
	"github.com/gin-gonic/gin"
)

// DatasetHandler 数据集处理器
type DatasetHandler struct {
	svc *dataset.Service
}

// NewDatasetHandler 创建数据集处理器
func NewDatasetHandler(svc *dataset.Service) *DatasetHandler {
	return &DatasetHandler{svc: svc}
}

// CreateDataset 创建数据集
func (h *DatasetHandler) CreateDataset(c *gin.Context) {
	var req dataset.CreateDatasetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	data, err := h.svc.CreateDataset(c.Request.Context(), &req)
	if err != nil {
		Error(c, err)
		return
	}

	Created(c, data)
}

// GetDataset 获取数据集
func (h *DatasetHandler) GetDataset(c *gin.Context) {
	id := c.Param("id")

	data, err := h.svc.GetDataset(c.Request.Context(), id)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, data)
}

// ListDatasets 列出数据集
func (h *DatasetHandler) ListDatasets(c *gin.Context) {
	tenantID := c.Query("tenant_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	datasets, total, err := h.svc.ListDatasets(c.Request.Context(), tenantID, page, pageSize)
	if err != nil {
		Error(c, err)
		return
	}

	SuccessWithPagination(c, datasets, total, page, pageSize)
}

// UpdateDataset 更新数据集
func (h *DatasetHandler) UpdateDataset(c *gin.Context) {
	id := c.Param("id")
	var req dataset.CreateDatasetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	data, err := h.svc.UpdateDataset(c.Request.Context(), id, &req)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, data)
}

// DeleteDataset 删除数据集
func (h *DatasetHandler) DeleteDataset(c *gin.Context) {
	id := c.Param("id")

	if err := h.svc.DeleteDataset(c.Request.Context(), id); err != nil {
		Error(c, err)
		return
	}

	NoContent(c)
}

// ========== QA 对操作 ==========

// CreateQAPair 创建 QA 对
func (h *DatasetHandler) CreateQAPair(c *gin.Context) {
	var req dataset.QAPairRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	pair, err := h.svc.CreateQAPair(c.Request.Context(), &req)
	if err != nil {
		Error(c, err)
		return
	}

	Created(c, pair)
}

// CreateQAPairsBatch 批量创建 QA 对
func (h *DatasetHandler) CreateQAPairsBatch(c *gin.Context) {
	var req struct {
		DatasetID string               `json:"dataset_id" binding:"required"`
		Pairs     []dataset.JSONQAPair `json:"pairs" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	count, err := h.svc.ImportFromJSON(c.Request.Context(), req.DatasetID, req.Pairs)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{"count": count, "message": "QA pairs imported successfully"})
}

// GetQAPairs 获取数据集的 QA 对
func (h *DatasetHandler) GetQAPairs(c *gin.Context) {
	datasetID := c.Param("dataset_id")

	pairs, err := h.svc.GetQAPairs(c.Request.Context(), datasetID)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{"pairs": pairs, "total": len(pairs)})
}

// GetQAPair 获取单个 QA 对
func (h *DatasetHandler) GetQAPair(c *gin.Context) {
	id := c.Param("id")

	pair, err := h.svc.GetQAPair(c.Request.Context(), id)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, pair)
}
