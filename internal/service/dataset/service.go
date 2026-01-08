package dataset

import (
	"context"
	"fmt"

	"github.com/ashwinyue/next-ai/internal/model"
	"github.com/ashwinyue/next-ai/internal/repository"
	"github.com/google/uuid"
)

// Service 数据集服务
type Service struct {
	repo *repository.Repositories
}

// NewService 创建数据集服务
func NewService(repo *repository.Repositories) *Service {
	return &Service{repo: repo}
}

// CreateDatasetRequest 创建数据集请求
type CreateDatasetRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Source      string `json:"source"`
	SourcePath  string `json:"source_path"`
	TenantID    string `json:"tenant_id"`
}

// CreateDataset 创建数据集
func (s *Service) CreateDataset(ctx context.Context, req *CreateDatasetRequest) (*model.Dataset, error) {
	dataset := &model.Dataset{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Source:      req.Source,
		SourcePath:  req.SourcePath,
		TenantID:    req.TenantID,
	}

	if err := s.repo.Dataset.Create(dataset); err != nil {
		return nil, fmt.Errorf("failed to create dataset: %w", err)
	}

	return dataset, nil
}

// GetDataset 获取数据集
func (s *Service) GetDataset(ctx context.Context, id string) (*model.Dataset, error) {
	dataset, err := s.repo.Dataset.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("dataset not found: %w", err)
	}
	return dataset, nil
}

// ListDatasets 列出数据集
func (s *Service) ListDatasets(ctx context.Context, tenantID string, page, size int) ([]*model.Dataset, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 20
	}

	offset := (page - 1) * size

	datasets, err := s.repo.Dataset.List(tenantID, offset, size)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list datasets: %w", err)
	}

	total, err := s.repo.Dataset.Count(tenantID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count datasets: %w", err)
	}

	return datasets, total, nil
}

// UpdateDataset 更新数据集
func (s *Service) UpdateDataset(ctx context.Context, id string, req *CreateDatasetRequest) (*model.Dataset, error) {
	dataset, err := s.repo.Dataset.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("dataset not found: %w", err)
	}

	if req.Name != "" {
		dataset.Name = req.Name
	}
	if req.Description != "" {
		dataset.Description = req.Description
	}

	if err := s.repo.Dataset.Update(dataset); err != nil {
		return nil, fmt.Errorf("failed to update dataset: %w", err)
	}

	return dataset, nil
}

// DeleteDataset 删除数据集
func (s *Service) DeleteDataset(ctx context.Context, id string) error {
	if err := s.repo.Dataset.Delete(id); err != nil {
		return fmt.Errorf("failed to delete dataset: %w", err)
	}
	return nil
}

// ========== QA 对操作 ==========

// QAPairRequest QA 对请求
type QAPairRequest struct {
	DatasetID  string   `json:"dataset_id"`
	Question   string   `json:"question"`
	Answer     string   `json:"answer"`
	PassageIDs []string `json:"passage_ids"`
	Passages   []string `json:"passages"`
}

// CreateQAPair 创建 QA 对
func (s *Service) CreateQAPair(ctx context.Context, req *QAPairRequest) (*model.QAPair, error) {
	// 验证数据集是否存在
	if _, err := s.repo.Dataset.GetByID(req.DatasetID); err != nil {
		return nil, fmt.Errorf("dataset not found: %w", err)
	}

	pair := &model.QAPair{
		ID:         uuid.New().String(),
		DatasetID:  req.DatasetID,
		Question:   req.Question,
		Answer:     req.Answer,
		PassageIDs: req.PassageIDs,
		Passages:   req.Passages,
	}

	if err := s.repo.Dataset.CreateQAPair(pair); err != nil {
		return nil, fmt.Errorf("failed to create QA pair: %w", err)
	}

	// 更新数据集记录数
	s.updateRecordCount(ctx, req.DatasetID)

	return pair, nil
}

// CreateQAPairsBatch 批量创建 QA 对
func (s *Service) CreateQAPairsBatch(ctx context.Context, datasetID string, pairs []*QAPairRequest) (int, error) {
	// 验证数据集是否存在
	if _, err := s.repo.Dataset.GetByID(datasetID); err != nil {
		return 0, fmt.Errorf("dataset not found: %w", err)
	}

	qaPairs := make([]*model.QAPair, len(pairs))
	for i, req := range pairs {
		qaPairs[i] = &model.QAPair{
			ID:         uuid.New().String(),
			DatasetID:  datasetID,
			Question:   req.Question,
			Answer:     req.Answer,
			PassageIDs: req.PassageIDs,
			Passages:   req.Passages,
		}
	}

	if err := s.repo.Dataset.CreateQAPairsBatch(qaPairs); err != nil {
		return 0, fmt.Errorf("failed to create QA pairs: %w", err)
	}

	// 更新数据集记录数
	s.updateRecordCount(ctx, datasetID)

	return len(qaPairs), nil
}

// GetQAPairs 获取数据集的所有 QA 对
func (s *Service) GetQAPairs(ctx context.Context, datasetID string) ([]*model.QAPair, error) {
	pairs, err := s.repo.Dataset.GetQAPairs(datasetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get QA pairs: %w", err)
	}
	return pairs, nil
}

// GetQAPair 获取单个 QA 对
func (s *Service) GetQAPair(ctx context.Context, id string) (*model.QAPair, error) {
	pair, err := s.repo.Dataset.GetQAPairByID(id)
	if err != nil {
		return nil, fmt.Errorf("QA pair not found: %w", err)
	}
	return pair, nil
}

// updateRecordCount 更新数据集记录数
func (s *Service) updateRecordCount(ctx context.Context, datasetID string) {
	pairs, _ := s.repo.Dataset.GetQAPairs(datasetID)
	dataset, _ := s.repo.Dataset.GetByID(datasetID)
	if dataset != nil {
		dataset.RecordCount = len(pairs)
		_ = s.repo.Dataset.Update(dataset)
	}
}

// ImportFromJSON 从 JSON 导入 QA 对
func (s *Service) ImportFromJSON(ctx context.Context, datasetID string, data []JSONQAPair) (int, error) {
	pairs := make([]*QAPairRequest, len(data))
	for i, item := range data {
		pairs[i] = &QAPairRequest{
			DatasetID:  datasetID,
			Question:   item.Question,
			Answer:     item.Answer,
			PassageIDs: item.PassageIDs,
			Passages:   item.Passages,
		}
	}

	return s.CreateQAPairsBatch(ctx, datasetID, pairs)
}

// JSONQAPair JSON 格式的 QA 对
type JSONQAPair struct {
	Question   string   `json:"question"`
	Answer     string   `json:"answer"`
	PassageIDs []string `json:"passage_ids,omitempty"`
	Passages   []string `json:"passages,omitempty"`
}
