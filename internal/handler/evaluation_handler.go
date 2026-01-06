// Package handler 提供评估相关的 HTTP 处理器
package handler

import (
	"fmt"
	"net/http"

	"github.com/ashwinyue/next-ai/internal/service"
	"github.com/ashwinyue/next-ai/internal/service/evaluation"
	"github.com/gin-gonic/gin"
)

// EvaluationHandler 评估处理器
type EvaluationHandler struct {
	svc *service.Services
}

// NewEvaluationHandler 创建评估处理器
func NewEvaluationHandler(svc *service.Services) *EvaluationHandler {
	return &EvaluationHandler{svc: svc}
}

// CreateEvaluationRequest 创建评估请求
type CreateEvaluationRequest = evaluation.CreateEvaluationRequest

// CreateEvaluation 创建评估任务
// @Summary      创建评估任务
// @Description  创建知识库评估任务
// @Tags         评估
// @Accept       json
// @Produce      json
// @Param        request  body      CreateEvaluationRequest  true  "评估请求"
// @Success      200      {object}  Response
// @Router       /api/v1/evaluations [post]
func (h *EvaluationHandler) CreateEvaluation(c *gin.Context) {
	ctx := c.Request.Context()

	var req CreateEvaluationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: -1, Message: err.Error()})
		return
	}

	task, err := h.svc.Evaluation.CreateEvaluation(ctx, &req)
	if err != nil {
		errorResponse(c, err)
		return
	}

	success(c, task)
}

// GetEvaluationResult 获取评估结果
// @Summary      获取评估结果
// @Description  根据任务ID获取评估结果
// @Tags         评估
// @Accept       json
// @Produce      json
// @Param        task_id  query     string  true  "任务ID"
// @Success      200      {object}  Response
// @Router       /api/v1/evaluations/result [get]
func (h *EvaluationHandler) GetEvaluationResult(c *gin.Context) {
	ctx := c.Request.Context()

	taskID := c.Query("task_id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, Response{Code: -1, Message: "task_id is required"})
		return
	}

	result, err := h.svc.Evaluation.GetEvaluationResult(ctx, taskID)
	if err != nil {
		errorResponse(c, err)
		return
	}

	success(c, result)
}

// ListEvaluations 列出评估任务
// @Summary      列出评估任务
// @Description  获取评估任务列表
// @Tags         评估
// @Accept       json
// @Produce      json
// @Param        kb_id    query     string  false  "知识库ID"
// @Param        limit    query     int     false  "每页数量" default(20)
// @Param        offset   query     int     false  "偏移量" default(0)
// @Success      200      {object}  Response
// @Router       /api/v1/evaluations [get]
func (h *EvaluationHandler) ListEvaluations(c *gin.Context) {
	ctx := c.Request.Context()

	kbID := c.Query("kb_id")
	limit := queryInt(c.Query("limit"), 20)
	offset := queryInt(c.Query("offset"), 0)

	tasks, total, err := h.svc.Evaluation.ListEvaluationTasks(ctx, kbID, limit, offset)
	if err != nil {
		errorResponse(c, err)
		return
	}

	success(c, gin.H{
		"items": tasks,
		"total": total,
		"limit": limit,
		"offset": offset,
	})
}

// DeleteEvaluation 删除评估任务
// @Summary      删除评估任务
// @Description  删除指定的评估任务
// @Tags         评估
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "任务ID"
// @Success      200  {object}  Response
// @Router       /api/v1/evaluations/{id} [delete]
func (h *EvaluationHandler) DeleteEvaluation(c *gin.Context) {
	ctx := c.Request.Context()

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, Response{Code: -1, Message: "id is required"})
		return
	}

	if err := h.svc.Evaluation.DeleteEvaluationTask(ctx, id); err != nil {
		errorResponse(c, err)
		return
	}

	success(c, gin.H{"message": "评估任务已删除"})
}

// CancelEvaluation 取消评估任务
// @Summary      取消评估任务
// @Description  取消正在执行的评估任务
// @Tags         评估
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "任务ID"
// @Success      200  {object}  Response
// @Router       /api/v1/evaluations/{id}/cancel [post]
func (h *EvaluationHandler) CancelEvaluation(c *gin.Context) {
	ctx := c.Request.Context()

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, Response{Code: -1, Message: "id is required"})
		return
	}

	if err := h.svc.Evaluation.CancelEvaluation(ctx, id); err != nil {
		errorResponse(c, err)
		return
	}

	success(c, gin.H{"message": "评估任务已取消"})
}

// queryInt 辅助函数：解析整数参数
func queryInt(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	var val int
	if _, err := fmt.Sscanf(s, "%d", &val); err != nil || val < 0 {
		return defaultVal
	}
	return val
}
