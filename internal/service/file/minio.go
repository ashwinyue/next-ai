package file

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOStorage MinIO 对象存储
type MinIOStorage struct {
	client     *minio.Client
	bucketName string
	urlPrefix  string // 用于生成访问URL
}

// MinIOConfig MinIO 配置
type MinIOConfig struct {
	Endpoint   string
	AccessKey  string
	SecretKey  string
	BucketName string
	UseSSL     bool
	URLPrefix  string
}

// NewMinIOStorage 创建 MinIO 存储服务
func NewMinIOStorage(cfg *MinIOConfig) (*MinIOStorage, error) {
	// 初始化 MinIO 客户端
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize MinIO client: %w", err)
	}

	// 检查 bucket 是否存在，不存在则创建
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.BucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket: %w", err)
	}

	if !exists {
		err = client.MakeBucket(ctx, cfg.BucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return &MinIOStorage{
		client:     client,
		bucketName: cfg.BucketName,
		urlPrefix:  strings.TrimSuffix(cfg.URLPrefix, "/"),
	}, nil
}

// Save 保存文件到 MinIO
func (s *MinIOStorage) Save(ctx context.Context, req *SaveRequest) (string, error) {
	// 生成对象名: {tenantID}/{knowledgeID}/{uuid}.{ext}
	ext := filepath.Ext(req.FileName)
	if ext == "" {
		ext = extensionByContentType(req.ContentType)
	}
	fileID := uuid.New().String()
	objectName := fmt.Sprintf("%s/%s/%s%s", req.TenantID, req.KnowledgeID, fileID, ext)

	// 上传文件
	_, err := s.client.PutObject(ctx, s.bucketName, objectName, req.Reader, req.Size, minio.PutObjectOptions{
		ContentType: req.ContentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to MinIO: %w", err)
	}

	return objectName, nil
}

// Get 获取文件内容
func (s *MinIOStorage) Get(ctx context.Context, filePath string) (io.ReadCloser, error) {
	object, err := s.client.GetObject(ctx, s.bucketName, filePath, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get file from MinIO: %w", err)
	}
	return object, nil
}

// Delete 删除文件
func (s *MinIOStorage) Delete(ctx context.Context, filePath string) error {
	err := s.client.RemoveObject(ctx, s.bucketName, filePath, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// GetURL 获取文件的访问URL
func (s *MinIOStorage) GetURL(filePath string) string {
	return fmt.Sprintf("%s/%s/%s", s.urlPrefix, s.bucketName, filePath)
}
