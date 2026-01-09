package file

import (
	"context"
	"io"
)

// Storage 文件存储接口
type Storage interface {
	// Save 保存文件，返回文件路径
	Save(ctx context.Context, req *SaveRequest) (string, error)
	// Get 获取文件内容
	Get(ctx context.Context, filePath string) (io.ReadCloser, error)
	// Delete 删除文件
	Delete(ctx context.Context, filePath string) error
	// GetURL 获取文件的访问URL（如果是对象存储）
	GetURL(filePath string) string
}

// SaveRequest 保存文件请求
type SaveRequest struct {
	FileName    string
	ContentType string
	Size        int64
	Reader      io.Reader
	TenantID    string
}

// StorageType 存储类型
type StorageType string

const (
	StorageTypeLocal StorageType = "local"
	StorageTypeMinIO StorageType = "minio"
	StorageTypeCOS   StorageType = "cos"
)
