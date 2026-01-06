package file

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/ashwinyue/next-ai/internal/model"
	"github.com/ashwinyue/next-ai/internal/repository"
	"github.com/google/uuid"
)

// Service 文件服务
type Service struct {
	repo        *repository.Repositories
	storage     Storage
	storageType StorageType
}

// NewService 创建文件服务
func NewService(repo *repository.Repositories, storage Storage, storageType StorageType) *Service {
	return &Service{
		repo:        repo,
		storage:     storage,
		storageType: storageType,
	}
}

// NewServiceFromConfig 从配置创建文件服务
func NewServiceFromConfig(repo *repository.Repositories, storageType StorageType, cfg map[string]string) (*Service, error) {
	var storage Storage
	var err error

	switch storageType {
	case StorageTypeLocal:
		basePath := cfg["base_path"]
		if basePath == "" {
			basePath = "./data/files"
		}
		urlPrefix := cfg["url_prefix"]
		if urlPrefix == "" {
			urlPrefix = "/files"
		}
		storage, err = NewLocalStorage(basePath, urlPrefix)

	case StorageTypeMinIO:
		if cfg["endpoint"] == "" || cfg["access_key"] == "" || cfg["secret_key"] == "" || cfg["bucket"] == "" {
			return nil, fmt.Errorf("missing required MinIO config")
		}
		useSSL := cfg["use_ssl"] == "true"
		urlPrefix := cfg["url_prefix"]
		if urlPrefix == "" {
			urlPrefix = fmt.Sprintf("%s/%s", cfg["endpoint"], cfg["bucket"])
		}
		storage, err = NewMinIOStorage(&MinIOConfig{
			Endpoint:   cfg["endpoint"],
			AccessKey:  cfg["access_key"],
			SecretKey:  cfg["secret_key"],
			BucketName: cfg["bucket"],
			UseSSL:     useSSL,
			URLPrefix:  urlPrefix,
		})

	default:
		return nil, fmt.Errorf("unsupported storage type: %s", storageType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}

	return NewService(repo, storage, storageType), nil
}

// SaveFile 保存文件
func (s *Service) SaveFile(ctx context.Context, req *SaveFileRequest) (*model.StoredFile, error) {
	// 使用存储服务保存文件
	filePath, err := s.storage.Save(ctx, &SaveRequest{
		FileName:    req.FileName,
		ContentType: req.ContentType,
		Size:        req.Size,
		Reader:      req.Reader,
		TenantID:    req.TenantID,
		KnowledgeID: req.KnowledgeID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	// 创建文件记录
	storedFile := &model.StoredFile{
		ID:          uuid.New().String(),
		TenantID:    req.TenantID,
		KnowledgeID: req.KnowledgeID,
		FileName:    req.FileName,
		FileSize:    req.Size,
		ContentType: req.ContentType,
		StorageType: string(s.storageType),
		FilePath:    filePath,
	}

	// 保存到数据库
	if err := s.repo.DB.Create(storedFile).Error; err != nil {
		// 如果数据库保存失败，删除已保存的文件
		_ = s.storage.Delete(ctx, filePath)
		return nil, fmt.Errorf("failed to save file record: %w", err)
	}

	return storedFile, nil
}

// GetFile 获取文件
func (s *Service) GetFile(ctx context.Context, id string) (*model.StoredFile, io.ReadCloser, error) {
	var storedFile model.StoredFile
	if err := s.repo.DB.Where("id = ?", id).First(&storedFile).Error; err != nil {
		return nil, nil, fmt.Errorf("file not found: %w", err)
	}

	reader, err := s.storage.Get(ctx, storedFile.FilePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get file content: %w", err)
	}

	return &storedFile, reader, nil
}

// DeleteFile 删除文件
func (s *Service) DeleteFile(ctx context.Context, id string) error {
	var storedFile model.StoredFile
	if err := s.repo.DB.Where("id = ?", id).First(&storedFile).Error; err != nil {
		return fmt.Errorf("file not found: %w", err)
	}

	// 从存储中删除
	if err := s.storage.Delete(ctx, storedFile.FilePath); err != nil {
		return fmt.Errorf("failed to delete file from storage: %w", err)
	}

	// 从数据库删除
	if err := s.repo.DB.Delete(&storedFile).Error; err != nil {
		return fmt.Errorf("failed to delete file record: %w", err)
	}

	return nil
}

// GetFileURL 获取文件访问URL
func (s *Service) GetFileURL(id string) (string, error) {
	var storedFile model.StoredFile
	if err := s.repo.DB.Where("id = ?", id).First(&storedFile).Error; err != nil {
		return "", fmt.Errorf("file not found: %w", err)
	}

	return s.storage.GetURL(storedFile.FilePath), nil
}

// SaveFileRequest 保存文件请求
type SaveFileRequest struct {
	FileName    string
	ContentType string
	Size        int64
	Reader      io.Reader
	TenantID    string
	KnowledgeID string
}

// SaveFileFromPath 从文件路径保存文件
func (s *Service) SaveFileFromPath(ctx context.Context, filePath, tenantID, knowledgeID string) (*model.StoredFile, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	return s.SaveFile(ctx, &SaveFileRequest{
		FileName:    info.Name(),
		Size:        info.Size(),
		Reader:      file,
		TenantID:    tenantID,
		KnowledgeID: knowledgeID,
	})
}

// ListByKnowledgeID 列出知识库的所有文件
func (s *Service) ListByKnowledgeID(ctx context.Context, knowledgeID string) ([]*model.StoredFile, error) {
	return s.repo.File.GetByKnowledgeID(knowledgeID)
}

// DeleteByKnowledgeID 删除知识库的所有文件
func (s *Service) DeleteByKnowledgeID(ctx context.Context, knowledgeID string) error {
	files, err := s.repo.File.GetByKnowledgeID(knowledgeID)
	if err != nil {
		return err
	}

	for _, f := range files {
		if err := s.storage.Delete(ctx, f.FilePath); err != nil {
			return err
		}
	}

	return s.repo.File.DeleteByKnowledgeID(knowledgeID)
}
