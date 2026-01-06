package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ========== WeKnora API 响应格式 ==========

// WeKnoraSuccessResponse WeKnora 成功响应
type WeKnoraSuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
}

// WeKnoraErrorResponse WeKnora 错误响应
type WeKnoraErrorResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// Success WeKnora 成功响应 (200)
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, WeKnoraSuccessResponse{Success: true, Data: data})
}

// Created WeKnora 创建成功响应 (201)
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, WeKnoraSuccessResponse{Success: true, Data: data})
}

// NoContent 无内容响应 (204)
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// BadRequest WeKnora 400 错误响应
func BadRequest(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, WeKnoraErrorResponse{Code: 400, Msg: msg})
}

// Unauthorized WeKnora 401 错误响应
func Unauthorized(c *gin.Context, msg string) {
	c.JSON(http.StatusUnauthorized, WeKnoraErrorResponse{Code: 401, Msg: msg})
}

// Forbidden WeKnora 403 错误响应
func Forbidden(c *gin.Context, msg string) {
	c.JSON(http.StatusForbidden, WeKnoraErrorResponse{Code: 403, Msg: msg})
}

// NotFound WeKnora 404 错误响应
func NotFound(c *gin.Context, msg string) {
	c.JSON(http.StatusNotFound, WeKnoraErrorResponse{Code: 404, Msg: msg})
}

// Conflict WeKnora 409 错误响应
func Conflict(c *gin.Context, msg string) {
	c.JSON(http.StatusConflict, WeKnoraErrorResponse{Code: 409, Msg: msg})
}

// InternalServerError WeKnora 500 错误响应
func InternalServerError(c *gin.Context, msg string) {
	c.JSON(http.StatusInternalServerError, WeKnoraErrorResponse{Code: 500, Msg: msg})
}

// Error 根据错误类型返回相应的错误响应
func Error(c *gin.Context, err error) {
	if err == nil {
		return
	}
	InternalServerError(c, err.Error())
}

// Pagination 分页响应数据结构
type PaginationData struct {
	Items      interface{} `json:"items"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages,omitempty"`
}

// SuccessWithPagination WeKnora 分页成功响应
func SuccessWithPagination(c *gin.Context, items interface{}, total int64, page, pageSize int) {
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}
	c.JSON(http.StatusOK, WeKnoraSuccessResponse{
		Success: true,
		Data: PaginationData{
			Items:      items,
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
		},
	})
}
