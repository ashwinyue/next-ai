// Package model 提供 MCP 服务相关的数据模型
package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MCPTransportType MCP 传输类型
type MCPTransportType string

const (
	MCPTransportSSE            MCPTransportType = "sse"             // Server-Sent Events
	MCPTransportHTTPStreamable MCPTransportType = "http-streamable" // HTTP Streamable
	MCPTransportStdio          MCPTransportType = "stdio"           // Standard Input/Output
)

// MCPService MCP 服务配置
type MCPService struct {
	ID             string             `json:"id" gorm:"type:varchar(36);primaryKey"`
	Name           string             `json:"name" gorm:"type:varchar(255);not null"`
	Description    string             `json:"description" gorm:"type:text"`
	Enabled        bool               `json:"enabled" gorm:"default:true;index"`
	TransportType  MCPTransportType   `json:"transport_type" gorm:"type:varchar(50);not null"`
	URL            *string            `json:"url,omitempty" gorm:"type:varchar(512)"` // SSE/HTTP Streamable 需要
	Headers        MCPHeaders         `json:"headers" gorm:"type:json"`
	AuthConfig     *MCPAuthConfig     `json:"auth_config" gorm:"type:json"`
	AdvancedConfig *MCPAdvancedConfig `json:"advanced_config" gorm:"type:json"`
	StdioConfig    *MCPStdioConfig    `json:"stdio_config,omitempty" gorm:"type:json"` // Stdio 传输需要
	EnvVars        MCPEnvVars         `json:"env_vars,omitempty" gorm:"type:json"`     // 环境变量

	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// MCPHeaders HTTP 头映射
type MCPHeaders map[string]string

// MCPAuthConfig MCP 认证配置
type MCPAuthConfig struct {
	APIKey        string            `json:"api_key,omitempty"`
	Token         string            `json:"token,omitempty"`
	CustomHeaders map[string]string `json:"custom_headers,omitempty"`
}

// MCPAdvancedConfig MCP 高级配置
type MCPAdvancedConfig struct {
	Timeout    int `json:"timeout"`     // 超时时间（秒），默认 30
	RetryCount int `json:"retry_count"` // 重试次数，默认 3
	RetryDelay int `json:"retry_delay"` // 重试延迟（秒），默认 1
}

// MCPStdioConfig Stdio 传输配置
type MCPStdioConfig struct {
	Command string   `json:"command"` // 命令: "uvx" 或 "npx"
	Args    []string `json:"args"`    // 命令参数数组
}

// MCPEnvVars 环境变量映射
type MCPEnvVars map[string]string

// MCPTool MCP 服务提供的工具
type MCPTool struct {
	ID          string          `json:"id"`
	ServiceID   string          `json:"service_id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema"` // JSON Schema
	CreatedAt   time.Time       `json:"created_at"`
}

// MCPResource MCP 服务提供的资源
type MCPResource struct {
	ID          string    `json:"id"`
	ServiceID   string    `json:"service_id"`
	URI         string    `json:"uri"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	MimeType    string    `json:"mime_type,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// MCPTestResult MCP 服务测试结果
type MCPTestResult struct {
	Success   bool           `json:"success"`
	Message   string         `json:"message,omitempty"`
	Tools     []*MCPTool     `json:"tools,omitempty"`
	Resources []*MCPResource `json:"resources,omitempty"`
}

// BeforeCreate GORM 钩子
func (m *MCPService) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	return nil
}

// TableName 指定表名
func (MCPService) TableName() string {
	return "mcp_services"
}

// Value 实现 driver.Valuer 接口 for MCPHeaders
func (h MCPHeaders) Value() (driver.Value, error) {
	if h == nil {
		return nil, nil
	}
	return json.Marshal(h)
}

// Scan 实现 sql.Scanner 接口 for MCPHeaders
func (h *MCPHeaders) Scan(value interface{}) error {
	if value == nil {
		*h = nil
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, h)
}

// Value 实现 driver.Valuer 接口 for MCPAuthConfig
func (c *MCPAuthConfig) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c)
}

// Scan 实现 sql.Scanner 接口 for MCPAuthConfig
func (c *MCPAuthConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, c)
}

// Value 实现 driver.Valuer 接口 for MCPAdvancedConfig
func (c *MCPAdvancedConfig) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c)
}

// Scan 实现 sql.Scanner 接口 for MCPAdvancedConfig
func (c *MCPAdvancedConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, c)
}

// Value 实现 driver.Valuer 接口 for MCPStdioConfig
func (c *MCPStdioConfig) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c)
}

// Scan 实现 sql.Scanner 接口 for MCPStdioConfig
func (c *MCPStdioConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, c)
}

// Value 实现 driver.Valuer 接口 for MCPEnvVars
func (e MCPEnvVars) Value() (driver.Value, error) {
	if e == nil {
		return nil, nil
	}
	return json.Marshal(e)
}

// Scan 实现 sql.Scanner 接口 for MCPEnvVars
func (e *MCPEnvVars) Scan(value interface{}) error {
	if value == nil {
		*e = nil
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, e)
}

// GetDefaultAdvancedConfig 返回默认高级配置
func GetDefaultAdvancedConfig() *MCPAdvancedConfig {
	return &MCPAdvancedConfig{
		Timeout:    30,
		RetryCount: 3,
		RetryDelay: 1,
	}
}
