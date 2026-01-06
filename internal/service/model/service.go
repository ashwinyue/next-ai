// Package model 提供模型管理服务
package model

import (
	"context"
	"fmt"

	"github.com/ashwinyue/next-ai/internal/model"
	"github.com/ashwinyue/next-ai/internal/repository"
)

// Service 模型服务
type Service struct {
	repo *repository.ModelRepository
}

// NewService 创建模型服务
func NewService(repo *repository.ModelRepository) *Service {
	return &Service{
		repo: repo,
	}
}

// CreateModel 创建模型
func (s *Service) CreateModel(ctx context.Context, m *model.Model) error {
	// 如果设为默认，清除同类型的其他默认标记
	if m.IsDefault {
		if err := s.repo.ClearDefaultByType(ctx, m.Type, ""); err != nil {
			return fmt.Errorf("clear default: %w", err)
		}
	}
	return s.repo.Create(ctx, m)
}

// GetModelByID 根据 ID 获取模型
func (s *Service) GetModelByID(ctx context.Context, id string) (*model.Model, error) {
	return s.repo.GetByID(ctx, id)
}

// ListModels 列出模型
func (s *Service) ListModels(ctx context.Context, modelType *model.ModelType, source *model.ModelSource) ([]*model.Model, error) {
	return s.repo.List(ctx, modelType, source)
}

// UpdateModel 更新模型
func (s *Service) UpdateModel(ctx context.Context, m *model.Model) error {
	// 如果设为默认，清除同类型的其他默认标记
	if m.IsDefault {
		if err := s.repo.ClearDefaultByType(ctx, m.Type, m.ID); err != nil {
			return fmt.Errorf("clear default: %w", err)
		}
	}
	return s.repo.Update(ctx, m)
}

// DeleteModel 删除模型
func (s *Service) DeleteModel(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// GetDefaultModel 获取指定类型的默认模型
func (s *Service) GetDefaultModel(ctx context.Context, modelType model.ModelType) (*model.Model, error) {
	return s.repo.GetDefaultByType(ctx, modelType)
}

// ListModelProviders 列出支持的模型提供商
func (s *Service) ListModelProviders(ctx context.Context) []ModelProvider {
	return []ModelProvider{
		{
			Name:        "openai",
			DisplayName: "OpenAI",
			Description: "OpenAI 官方 API",
			Types:       []model.ModelType{model.ModelTypeChatModel, model.ModelTypeEmbedding},
		},
		{
			Name:        "aliyun",
			DisplayName: "阿里云 DashScope",
			Description: "阿里云通义千问 API",
			Types:       []model.ModelType{model.ModelTypeChatModel, model.ModelTypeEmbedding, model.ModelTypeRerank},
		},
		{
			Name:        "zhipu",
			DisplayName: "智谱 AI",
			Description: "智谱 AI GLM API",
			Types:       []model.ModelType{model.ModelTypeChatModel, model.ModelTypeEmbedding, model.ModelTypeRerank},
		},
		{
			Name:        "local",
			DisplayName: "本地模型",
			Description: "通过 Ollama 运行的本地模型",
			Types:       []model.ModelType{model.ModelTypeChatModel, model.ModelTypeEmbedding},
		},
		{
			Name:        "jina",
			DisplayName: "Jina AI",
			Description: "Jina AI 向量和重排序服务",
			Types:       []model.ModelType{model.ModelTypeEmbedding, model.ModelTypeRerank},
		},
	}
}

// ModelProvider 模型提供商信息
type ModelProvider struct {
	Name        string          `json:"name"`
	DisplayName string          `json:"display_name"`
	Description string          `json:"description"`
	Types       []model.ModelType `json:"types"`
}
