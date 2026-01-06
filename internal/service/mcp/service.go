// Package mcp 提供 MCP 服务管理
// 使用官方 go-sdk: github.com/modelcontextprotocol/go-sdk
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

// CreateMCPServieRequest 创建 MCP 服务请求
type CreateMCPServiceRequest struct {
	Name          string                  `json:"name" binding:"required"`
	Description   string                  `json:"description"`
	TransportType model.MCPTransportType  `json:"transport_type" binding:"required"`
	URL           *string                 `json:"url,omitempty"`
	Headers       model.MCPHeaders         `json:"headers"`
	AuthConfig    *model.MCPAuthConfig     `json:"auth_config"`
	StdioConfig   *model.MCPStdioConfig    `json:"stdio_config,omitempty"`
	EnvVars       model.MCPEnvVars         `json:"env_vars"`
}

// CreateMCPService 创建 MCP 服务
func (s *Service) CreateMCPService(ctx context.Context, req *CreateMCPServiceRequest) (*model.MCPService, error) {
	svc := &model.MCPService{
		Name:          req.Name,
		Description:   req.Description,
		Enabled:       true,
		TransportType: req.TransportType,
		URL:           req.URL,
		Headers:       req.Headers,
		AuthConfig:    req.AuthConfig,
		StdioConfig:   req.StdioConfig,
		EnvVars:       req.EnvVars,
		AdvancedConfig: model.GetDefaultAdvancedConfig(),
	}

	// TODO: 保存到数据库
	// 当前简化版本直接返回
	return svc, nil
}

// ListMCPServices 列出 MCP 服务
func (s *Service) ListMCPServices(ctx context.Context) ([]*model.MCPService, error) {
	// TODO: 从数据库查询
	return []*model.MCPService{}, nil
}

// GetMCPService 获取 MCP 服务详情
func (s *Service) GetMCPService(ctx context.Context, id string) (*model.MCPService, error) {
	// TODO: 从数据库查询
	return nil, fmt.Errorf("MCP service not found: %s", id)
}

// UpdateMCPServiceRequest 更新 MCP 服务请求
type UpdateMCPServiceRequest struct {
	Name          *string                 `json:"name,omitempty"`
	Description   *string                 `json:"description,omitempty"`
	Enabled       *bool                   `json:"enabled,omitempty"`
	TransportType *model.MCPTransportType `json:"transport_type,omitempty"`
	URL           *string                 `json:"url,omitempty"`
	Headers       model.MCPHeaders         `json:"headers,omitempty"`
	AuthConfig    *model.MCPAuthConfig     `json:"auth_config,omitempty"`
	StdioConfig   *model.MCPStdioConfig    `json:"stdio_config,omitempty"`
	EnvVars       model.MCPEnvVars         `json:"env_vars,omitempty"`
}

// UpdateMCPService 更新 MCP 服务
func (s *Service) UpdateMCPService(ctx context.Context, id string, req *UpdateMCPServiceRequest) (*model.MCPService, error) {
	// TODO: 从数据库获取并更新
	return nil, fmt.Errorf("not implemented")
}

// DeleteMCPService 删除 MCP 服务
func (s *Service) DeleteMCPService(ctx context.Context, id string) error {
	// TODO: 从数据库删除
	return fmt.Errorf("not implemented")
}

// TestMCPService 测试 MCP 服务连接
func (s *Service) TestMCPService(ctx context.Context, id string) (*model.MCPTestResult, error) {
	// TODO: 使用 eino-ext 的 MCP 组件测试连接
	// 1. 获取服务配置
	// 2. 创建 MCP 客户端 (使用 mcp-go 或 go-sdk)
	// 3. 调用 ListTools / ListResources
	// 4. 返回结果
	return &model.MCPTestResult{
		Success: false,
		Message: "not implemented",
	}, nil
}

// GetMCPServiceTools 获取 MCP 服务提供的工具列表
func (s *Service) GetMCPServiceTools(ctx context.Context, id string) ([]*model.MCPTool, error) {
	// TODO: 使用 eino-ext/mcp.GetTools() 获取工具列表
	// 参考: github.com/cloudwego/eino-ext/components/tool/mcp
	return []*model.MCPTool{}, nil
}

// GetMCPServiceResources 获取 MCP 服务提供的资源列表
func (s *Service) GetMCPServiceResources(ctx context.Context, id string) ([]*model.MCPResource, error) {
	// TODO: 获取资源列表
	return []*model.MCPResource{}, nil
}

// ConvertToEinoTools 将 MCP 服务转换为 Eino 工具列表
// 使用 eino-ext 的 MCP 组件
func (s *Service) ConvertToEinoTools(ctx context.Context, serviceID string) (interface{}, error) {
	// TODO: 使用 eino-ext/components/tool/mcp.GetTools() 转换
	// 示例:
	// import mcptool "github.com/cloudwego/eino-ext/components/tool/mcp"
	// tools, err := mcptool.GetTools(ctx, &mcptool.Config{
	//     Cli: mcpClient,
	// })
	return nil, fmt.Errorf("not implemented")
}
