// Package mcp 提供 MCP 服务管理
// 参考: github.com/cloudwego/eino-ext/components/tool/mcp/officialmcp
package mcp

import (
	"context"
	"fmt"

	"github.com/ashwinyue/next-ai/internal/model"
	"github.com/ashwinyue/next-ai/internal/repository"
)

// Service MCP 服务管理
type Service struct {
	repo *repository.Repositories
}

// NewService 创建 MCP 服务
func NewService(repo *repository.Repositories) *Service {
	return &Service{
		repo: repo,
	}
}

// CreateMCPServiceRequest 创建 MCP 服务请求
type CreateMCPServiceRequest struct {
	Name          string                 `json:"name" binding:"required"`
	Description   string                 `json:"description"`
	TransportType model.MCPTransportType `json:"transport_type" binding:"required"`
	URL           *string                `json:"url,omitempty"`
	Headers       model.MCPHeaders       `json:"headers"`
	AuthConfig    *model.MCPAuthConfig   `json:"auth_config"`
	StdioConfig   *model.MCPStdioConfig  `json:"stdio_config,omitempty"`
	EnvVars       model.MCPEnvVars       `json:"env_vars"`
}

// CreateMCPService 创建 MCP 服务
func (s *Service) CreateMCPService(ctx context.Context, req *CreateMCPServiceRequest) (*model.MCPService, error) {
	svc := &model.MCPService{
		Name:           req.Name,
		Description:    req.Description,
		Enabled:        true,
		TransportType:  req.TransportType,
		URL:             req.URL,
		Headers:        req.Headers,
		AuthConfig:     req.AuthConfig,
		StdioConfig:    req.StdioConfig,
		EnvVars:        req.EnvVars,
		AdvancedConfig: model.GetDefaultAdvancedConfig(),
	}

	if err := s.repo.MCP.Create(svc); err != nil {
		return nil, fmt.Errorf("failed to create MCP service: %w", err)
	}

	return svc, nil
}

// ListMCPServices 列出 MCP 服务
func (s *Service) ListMCPServices(ctx context.Context) ([]*model.MCPService, error) {
	svcs, err := s.repo.MCP.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list MCP services: %w", err)
	}
	return svcs, nil
}

// GetMCPService 获取 MCP 服务详情
func (s *Service) GetMCPService(ctx context.Context, id string) (*model.MCPService, error) {
	svc, err := s.repo.MCP.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("MCP service not found: %w", err)
	}
	return svc, nil
}

// UpdateMCPServiceRequest 更新 MCP 服务请求
type UpdateMCPServiceRequest struct {
	Name          *string                 `json:"name,omitempty"`
	Description   *string                 `json:"description,omitempty"`
	Enabled       *bool                   `json:"enabled,omitempty"`
	TransportType *model.MCPTransportType `json:"transport_type,omitempty"`
	URL           *string                 `json:"url,omitempty"`
	Headers       model.MCPHeaders        `json:"headers,omitempty"`
	AuthConfig    *model.MCPAuthConfig    `json:"auth_config,omitempty"`
	StdioConfig   *model.MCPStdioConfig   `json:"stdio_config,omitempty"`
	EnvVars       model.MCPEnvVars        `json:"env_vars,omitempty"`
}

// UpdateMCPService 更新 MCP 服务
func (s *Service) UpdateMCPService(ctx context.Context, id string, req *UpdateMCPServiceRequest) (*model.MCPService, error) {
	svc, err := s.repo.MCP.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("MCP service not found: %w", err)
	}

	// 更新字段
	if req.Name != nil {
		svc.Name = *req.Name
	}
	if req.Description != nil {
		svc.Description = *req.Description
	}
	if req.Enabled != nil {
		svc.Enabled = *req.Enabled
	}
	if req.TransportType != nil {
		svc.TransportType = *req.TransportType
	}
	if req.URL != nil {
		svc.URL = req.URL
	}
	if len(req.Headers) > 0 {
		svc.Headers = req.Headers
	}
	if req.AuthConfig != nil {
		svc.AuthConfig = req.AuthConfig
	}
	if req.StdioConfig != nil {
		svc.StdioConfig = req.StdioConfig
	}
	if len(req.EnvVars) > 0 {
		svc.EnvVars = req.EnvVars
	}

	if err := s.repo.MCP.Update(svc); err != nil {
		return nil, fmt.Errorf("failed to update MCP service: %w", err)
	}

	return svc, nil
}

// DeleteMCPService 删除 MCP 服务
func (s *Service) DeleteMCPService(ctx context.Context, id string) error {
	if err := s.repo.MCP.Delete(id); err != nil {
		return fmt.Errorf("failed to delete MCP service: %w", err)
	}
	return nil
}

// TestMCPService 测试 MCP 服务连接
// 注意: 完整实现需要使用 github.com/modelcontextprotocol/go-sdk/mcp
func (s *Service) TestMCPService(ctx context.Context, id string) (*model.MCPTestResult, error) {
	svc, err := s.repo.MCP.GetByID(id)
	if err != nil {
		return &model.MCPTestResult{
			Success: false,
			Message: "MCP service not found",
		}, nil
	}

	// 简化版: 检查服务配置是否完整
	// SSE 和 HTTP Streamable 都需要 URL
	if (svc.TransportType == model.MCPTransportSSE || svc.TransportType == model.MCPTransportHTTPStreamable) && svc.URL == nil {
		return &model.MCPTestResult{
			Success: false,
			Message: "SSE/HTTP transport requires URL",
		}, nil
	}

	if svc.TransportType == model.MCPTransportStdio && svc.StdioConfig == nil {
		return &model.MCPTestResult{
			Success: false,
			Message: "Stdio transport requires command",
		}, nil
	}

	// 实际连接测试需要集成 go-sdk 的 mcp.ClientSession
	return &model.MCPTestResult{
		Success: true,
		Message: "Configuration is valid (actual connection test requires go-sdk integration)",
	}, nil
}

// GetMCPServiceTools 获取 MCP 服务提供的工具列表
// 注意: 完整实现需要使用 eino-ext/components/tool/mcp/officialmcp.GetTools
func (s *Service) GetMCPServiceTools(ctx context.Context, id string) ([]*model.MCPTool, error) {
	svc, err := s.repo.MCP.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("MCP service not found: %w", err)
	}

	// 简化版: 返回空工具列表
	// 实际实现需要:
	// 1. 创建 mcp.ClientSession (使用 go-sdk)
	// 2. 调用 cli.ListTools(ctx)
	// 3. 转换为 model.MCPTool 列表
	_ = svc

	return []*model.MCPTool{}, nil
}

// GetMCPServiceResources 获取 MCP 服务提供的资源列表
func (s *Service) GetMCPServiceResources(ctx context.Context, id string) ([]*model.MCPResource, error) {
	svc, err := s.repo.MCP.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("MCP service not found: %w", err)
	}
	_ = svc

	// 实际实现需要使用 cli.ListResources(ctx)
	return []*model.MCPResource{}, nil
}

// ConvertToEinoTools 将 MCP 服务转换为 Eino 工具列表
// 参考: github.com/cloudwego/eino-ext/components/tool/mcp/officialmcp.GetTools
func (s *Service) ConvertToEinoTools(ctx context.Context, serviceID string) ([]interface{}, error) {
	svc, err := s.repo.MCP.GetByID(serviceID)
	if err != nil {
		return nil, fmt.Errorf("MCP service not found: %w", err)
	}

	// 简化版: 返回空工具列表
	// 实际实现:
	// import mcptool "github.com/cloudwego/eino-ext/components/tool/mcp/officialmcp"
	// tools, err := mcptool.GetTools(ctx, &mcptool.Config{
	//     Cli: mcpClient,
	// })
	_ = svc

	return []interface{}{}, nil
}

// EnableMCPService 启用 MCP 服务
func (s *Service) EnableMCPService(ctx context.Context, id string) error {
	if err := s.repo.MCP.UpdateEnabled(id, true); err != nil {
		return fmt.Errorf("failed to enable MCP service: %w", err)
	}
	return nil
}

// DisableMCPService 禁用 MCP 服务
func (s *Service) DisableMCPService(ctx context.Context, id string) error {
	if err := s.repo.MCP.UpdateEnabled(id, false); err != nil {
		return fmt.Errorf("failed to disable MCP service: %w", err)
	}
	return nil
}
