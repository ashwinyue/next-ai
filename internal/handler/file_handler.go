package handler

import (
	"io"
	"strconv"

	filesvc "github.com/ashwinyue/next-ai/internal/service/file"
	"github.com/gin-gonic/gin"
)

// FileHandler 文件处理器
type FileHandler struct {
	fileSvc *filesvc.Service
}

// NewFileHandler 创建文件处理器
func NewFileHandler(fileSvc *filesvc.Service) *FileHandler {
	return &FileHandler{
		fileSvc: fileSvc,
	}
}

// UploadFile 上传文件
// @Summary      上传文件
// @Description  上传文件到存储服务
// @Tags         文件管理
// @Accept       multipart/form-data
// @Produce      json
// @Param        tenant_id formData string true "租户ID"
// @Param        file      formData file   true "文件"
// @Success      200        {object}  Response  "上传成功"
// @Failure      500        {object}  Response  "服务器错误"
// @Router       /files/upload [post]
func (h *FileHandler) UploadFile(c *gin.Context) {
	tenantID := c.PostForm("tenant_id")

	if tenantID == "" {
		BadRequest(c, "tenant_id is required")
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		BadRequest(c, "$1"+err.Error())
		return
	}

	// 打开文件
	f, err := fileHeader.Open()
	if err != nil {
		Error(c, err)
		return
	}
	defer f.Close()

	// 保存文件
	storedFile, err := h.fileSvc.SaveFile(c.Request.Context(), &filesvc.SaveFileRequest{
		FileName:    fileHeader.Filename,
		ContentType: fileHeader.Header.Get("Content-Type"),
		Size:        fileHeader.Size,
		Reader:      f,
		TenantID:    tenantID,
	})
	if err != nil {
		Error(c, err)
		return
	}

	Created(c, storedFile)
}

// GetFile 获取文件
// @Summary      获取文件
// @Description  获取文件内容
// @Tags         文件管理
// @Accept       json
// @Produce      octet-stream
// @Param        id   path      string  true "文件ID"
// @Success      200  {file}    file    "文件内容"
// @Failure      404  {object}  Response  "文件不存在"
// @Router       /files/{id} [get]
func (h *FileHandler) GetFile(c *gin.Context) {
	id := c.Param("id")

	storedFile, reader, err := h.fileSvc.GetFile(c.Request.Context(), id)
	if err != nil {
		Error(c, err)
		return
	}
	defer reader.Close()

	// 设置响应头
	c.Header("Content-Type", storedFile.ContentType)
	c.Header("Content-Disposition", "attachment; filename="+storedFile.FileName)
	c.Header("Content-Length", strconv.FormatInt(storedFile.FileSize, 10))

	// 流式传输文件
	_, err = io.Copy(c.Writer, reader)
	if err != nil {
		Error(c, err)
		return
	}
}

// GetFileURL 获取文件访问URL
// @Summary      获取文件URL
// @Description  获取文件的访问URL
// @Tags         文件管理
// @Accept       json
// @Produce      json
// @Param        id   path      string  true "文件ID"
// @Success      200  {object}  Response  "文件URL"
// @Failure      404  {object}  Response  "文件不存在"
// @Router       /files/{id}/url [get]
func (h *FileHandler) GetFileURL(c *gin.Context) {
	id := c.Param("id")

	url, err := h.fileSvc.GetFileURL(id)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{"url": url})
}

// DeleteFile 删除文件
// @Summary      删除文件
// @Description  删除文件
// @Tags         文件管理
// @Accept       json
// @Produce      json
// @Param        id   path      string  true "文件ID"
// @Success      200  {object}  Response  "删除成功"
// @Failure      500  {object}  Response  "服务器错误"
// @Router       /files/{id} [delete]
func (h *FileHandler) DeleteFile(c *gin.Context) {
	id := c.Param("id")

	if err := h.fileSvc.DeleteFile(c.Request.Context(), id); err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{"message": "File deleted successfully"})
}
