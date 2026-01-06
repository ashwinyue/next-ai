package tool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ashwinyue/next-ai/internal/model"
	"github.com/ashwinyue/next-ai/internal/repository"
	"github.com/google/uuid"
)

// Service 工具服务
type Service struct {
	repo *repository.Repositories
}

// NewService 创建工具服务
func NewService(repo *repository.Repositories) *Service {
	return &Service{repo: repo}
}

// ToolConfig 工具配置结构
type ToolConfig map[string]interface{}

// RegisterToolRequest 注册工具请求
type RegisterToolRequest struct {
	Name        string     `json:"name" binding:"required"`
	DisplayName string     `json:"display_name"`
	Description string     `json:"description"`
	Type        string     `json:"type" binding:"required"`
	Config      ToolConfig `json:"config"`
}

// RegisterTool 注册工具
func (s *Service) RegisterTool(ctx context.Context, req *RegisterToolRequest) (*model.Tool, error) {
	// 检查名称是否已存在
	if _, err := s.repo.Tool.GetByName(req.Name); err == nil {
		return nil, fmt.Errorf("tool name already exists")
	}

	configJSON, err := json.Marshal(req.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	tool := &model.Tool{
		ID:          uuid.New().String(),
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Type:        req.Type,
		Config:      string(configJSON),
		IsActive:    true,
	}

	if err := s.repo.Tool.Create(tool); err != nil {
		return nil, fmt.Errorf("failed to create tool: %w", err)
	}

	return tool, nil
}

// GetTool 获取工具
func (s *Service) GetTool(ctx context.Context, id string) (*model.Tool, error) {
	return s.repo.Tool.GetByID(id)
}

// ListToolsRequest 列出工具请求
type ListToolsRequest struct {
	Page int `json:"page"`
	Size int `json:"size"`
}

// ListTools 列出工具
func (s *Service) ListTools(ctx context.Context, req *ListToolsRequest) ([]*model.Tool, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 || req.Size > 100 {
		req.Size = 20
	}

	offset := (req.Page - 1) * req.Size
	return s.repo.Tool.List(offset, req.Size)
}

// ListActiveTools 列出活跃工具
func (s *Service) ListActiveTools(ctx context.Context) ([]*model.Tool, error) {
	return s.repo.Tool.ListActive()
}

// UpdateTool 更新工具
func (s *Service) UpdateTool(ctx context.Context, id string, req *RegisterToolRequest) (*model.Tool, error) {
	tool, err := s.repo.Tool.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("tool not found: %w", err)
	}

	tool.Name = req.Name
	tool.DisplayName = req.DisplayName
	tool.Description = req.Description
	tool.Type = req.Type

	configJSON, err := json.Marshal(req.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}
	tool.Config = string(configJSON)

	if err := s.repo.Tool.Update(tool); err != nil {
		return nil, fmt.Errorf("failed to update tool: %w", err)
	}

	return tool, nil
}

// UnregisterTool 注销工具
func (s *Service) UnregisterTool(ctx context.Context, id string) error {
	if err := s.repo.Tool.Delete(id); err != nil {
		return fmt.Errorf("failed to delete tool: %w", err)
	}
	return nil
}
