package file

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// LocalStorage 本地文件存储
type LocalStorage struct {
	basePath  string // 基础路径
	urlPrefix string // URL前缀，用于生成访问URL
}

// NewLocalStorage 创建本地存储服务
func NewLocalStorage(basePath, urlPrefix string) (*LocalStorage, error) {
	// 确保基础路径存在
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &LocalStorage{
		basePath:  basePath,
		urlPrefix: strings.TrimSuffix(urlPrefix, "/"),
	}, nil
}

// Save 保存文件到本地
func (s *LocalStorage) Save(ctx context.Context, req *SaveRequest) (string, error) {
	// 生成文件路径: {basePath}/{tenantID}/{uuid}.{ext}
	ext := filepath.Ext(req.FileName)
	if ext == "" {
		// 根据内容类型推断扩展名
		ext = extensionByContentType(req.ContentType)
	}
	fileID := uuid.New().String()
	relativePath := fmt.Sprintf("%s/%s%s", req.TenantID, fileID, ext)
	fullPath := filepath.Join(s.basePath, relativePath)

	// 创建目录
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// 创建文件
	file, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// 写入内容
	if _, err := io.Copy(file, req.Reader); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return relativePath, nil
}

// Get 获取文件内容
func (s *LocalStorage) Get(ctx context.Context, filePath string) (io.ReadCloser, error) {
	fullPath := filepath.Join(s.basePath, filePath)
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	return file, nil
}

// Delete 删除文件
func (s *LocalStorage) Delete(ctx context.Context, filePath string) error {
	fullPath := filepath.Join(s.basePath, filePath)
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// GetURL 获取文件的访问URL
func (s *LocalStorage) GetURL(filePath string) string {
	return fmt.Sprintf("%s/%s", s.urlPrefix, filePath)
}

// ListBuckets 列出存储桶（本地存储返回空列表）
func (s *LocalStorage) ListBuckets(ctx context.Context) ([]map[string]interface{}, error) {
	// 本地存储没有 bucket 概念，返回空列表
	return []map[string]interface{}{}, nil
}

// extensionByContentType 根据内容类型返回扩展名
func extensionByContentType(contentType string) string {
	switch contentType {
	case "application/pdf":
		return ".pdf"
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		return ".docx"
	case "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
		return ".xlsx"
	case "application/vnd.openxmlformats-officedocument.presentationml.presentation":
		return ".pptx"
	case "text/plain", "text/csv":
		return ".txt"
	case "text/markdown":
		return ".md"
	case "application/json":
		return ".json"
	case "text/html":
		return ".html"
	default:
		return ".bin"
	}
}
