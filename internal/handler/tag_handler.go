package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	svctag "github.com/ashwinyue/next-ai/internal/service/tag"
)

// TagHandler 标签处理器
type TagHandler struct {
	svc *svctag.Service
}

// NewTagHandler 创建标签处理器
func NewTagHandler(svc *svctag.Service) *TagHandler {
	return &TagHandler{svc: svc}
}

// CreateTagRequest 创建标签请求
type CreateTagRequest struct {
	Name      string `json:"name" binding:"required"`
	Color     string `json:"color"`
	SortOrder int    `json:"sort_order"`
}

// UpdateTagRequest 更新标签请求
type UpdateTagRequest struct {
	Name      *string `json:"name"`
	Color     *string `json:"color"`
	SortOrder *int    `json:"sort_order"`
}

// CreateTag 创建标签
// @Summary      创建标签
// @Description  在知识库下创建新标签
// @Tags         标签管理
// @Accept       json
// @Produce      json
// @Param        id       path      string              true  "知识库ID"
// @Param        request  body      CreateTagRequest    true  "标签信息"
// @Success      200      {object}  Response            "创建的标签"
// @Failure      400      {object}  Response            "请求参数错误"
// @Router       /knowledge-bases/{id}/tags [post]
func (h *TagHandler) CreateTag(c *gin.Context) {
	kbID := c.Param("id")

	var req CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: -1, Message: "请求参数不合法: " + err.Error()})
		return
	}

	tag, err := h.svc.CreateTag(c.Request.Context(), kbID, &svctag.CreateTagRequest{
		Name:      req.Name,
		Color:     req.Color,
		SortOrder: req.SortOrder,
	})
	if err != nil {
		errorResponse(c, err)
		return
	}

	success(c, tag)
}

// GetTag 获取标签
// @Summary      获取标签
// @Description  根据ID获取标签详情
// @Tags         标签管理
// @Accept       json
// @Produce      json
// @Param        id       path      string      true  "知识库ID"
// @Param        tag_id   path      string      true  "标签ID"
// @Success      200      {object}  Response    "标签详情"
// @Failure      404      {object}  Response    "标签不存在"
// @Router       /knowledge-bases/{id}/tags/{tag_id} [get]
func (h *TagHandler) GetTag(c *gin.Context) {
	tagID := c.Param("tag_id")

	tag, err := h.svc.GetTag(c.Request.Context(), tagID)
	if err != nil {
		errorResponse(c, err)
		return
	}

	success(c, tag)
}

// ListTags 列出标签
// @Summary      列出标签
// @Description  获取知识库下的所有标签
// @Tags         标签管理
// @Accept       json
// @Produce      json
// @Param        id         path      string  true   "知识库ID"
// @Param        page       query     int     false  "页码"
// @Param        page_size  query     int     false  "每页数量"
// @Param        keyword    query     string  false  "关键词搜索"
// @Success      200        {object}  Response  "标签列表"
// @Router       /knowledge-bases/{id}/tags [get]
func (h *TagHandler) ListTags(c *gin.Context) {
	kbID := c.Param("id")

	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")

	resp, err := h.svc.ListTags(c.Request.Context(), kbID, page, pageSize, keyword)
	if err != nil {
		errorResponse(c, err)
		return
	}

	success(c, resp)
}

// GetAllTags 获取知识库的所有标签（不分页）
// @Summary      获取所有标签
// @Description  获取知识库下的所有标签（不分页）
// @Tags         标签管理
// @Accept       json
// @Produce      json
// @Param        id   path      string      true  "知识库ID"
// @Success      200  {object}  Response    "标签列表"
// @Router       /knowledge-bases/{id}/tags/all [get]
func (h *TagHandler) GetAllTags(c *gin.Context) {
	kbID := c.Param("id")

	tags, err := h.svc.GetTagsByKnowledgeBaseID(c.Request.Context(), kbID)
	if err != nil {
		errorResponse(c, err)
		return
	}

	success(c, gin.H{"tags": tags})
}

// UpdateTag 更新标签
// @Summary      更新标签
// @Description  更新标签信息
// @Tags         标签管理
// @Accept       json
// @Produce      json
// @Param        id       path      string              true  "知识库ID"
// @Param        tag_id   path      string              true  "标签ID"
// @Param        request  body      UpdateTagRequest    true  "标签更新信息"
// @Success      200      {object}  Response            "更新后的标签"
// @Failure      400      {object}  Response            "请求参数错误"
// @Failure      404      {object}  Response            "标签不存在"
// @Router       /knowledge-bases/{id}/tags/{tag_id} [put]
func (h *TagHandler) UpdateTag(c *gin.Context) {
	tagID := c.Param("tag_id")

	var req UpdateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: -1, Message: "请求参数不合法: " + err.Error()})
		return
	}

	tag, err := h.svc.UpdateTag(c.Request.Context(), tagID, &svctag.UpdateTagRequest{
		Name:      req.Name,
		Color:     req.Color,
		SortOrder: req.SortOrder,
	})
	if err != nil {
		errorResponse(c, err)
		return
	}

	success(c, tag)
}

// DeleteTag 删除标签
// @Summary      删除标签
// @Description  删除标签
// @Tags         标签管理
// @Accept       json
// @Produce      json
// @Param        id       path      string      true  "知识库ID"
// @Param        tag_id   path      string      true  "标签ID"
// @Success      200      {object}  Response    "删除成功"
// @Failure      404      {object}  Response    "标签不存在"
// @Router       /knowledge-bases/{id}/tags/{tag_id} [delete]
func (h *TagHandler) DeleteTag(c *gin.Context) {
	tagID := c.Param("tag_id")

	if err := h.svc.DeleteTag(c.Request.Context(), tagID); err != nil {
		errorResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, Response{Code: 0, Message: "删除成功"})
}
