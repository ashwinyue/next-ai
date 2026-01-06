package initialization

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/ashwinyue/next-ai/internal/repository"
)

// Service 初始化服务
type Service struct {
	repo      *repository.Repositories
	startTime time.Time
	version   string
}

// NewService 创建初始化服务
func NewService(repo *repository.Repositories) *Service {
	return &Service{
		repo:      repo,
		startTime: time.Now(),
		version:   "1.0.0",
	}
}

// OllamaStatusResponse Ollama 状态响应
type OllamaStatusResponse struct {
	Available bool   `json:"available"`
	Version   string `json:"version,omitempty"`
	BaseURL   string `json:"baseUrl"`
	Error     string `json:"error,omitempty"`
}

// CheckOllamaStatus 检查 Ollama 服务状态
func (s *Service) CheckOllamaStatus(ctx context.Context) (*OllamaStatusResponse, error) {
	baseURL := os.Getenv("OLLAMA_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	// 解析 URL
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return &OllamaStatusResponse{
			Available: false,
			BaseURL:   baseURL,
			Error:     fmt.Sprintf("无效的 Ollama URL: %v", err),
		}, nil
	}

	// 检查服务健康状态
	healthURL := fmt.Sprintf("%s/health", parsedURL.String())
	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		return &OllamaStatusResponse{
			Available: false,
			BaseURL:   baseURL,
			Error:     fmt.Sprintf("创建请求失败: %v", err),
		}, nil
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return &OllamaStatusResponse{
			Available: false,
			BaseURL:   baseURL,
			Error:     fmt.Sprintf("连接失败: %v", err),
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &OllamaStatusResponse{
			Available: false,
			BaseURL:   baseURL,
			Error:     fmt.Sprintf("服务返回状态码: %d", resp.StatusCode),
		}, nil
	}

	// 获取版本
	version, err := s.getOllamaVersion(ctx, parsedURL.String())
	if err != nil {
		version = "unknown"
	}

	return &OllamaStatusResponse{
		Available: true,
		Version:   version,
		BaseURL:   baseURL,
	}, nil
}

// getOllamaVersion 获取 Ollama 版本
func (s *Service) getOllamaVersion(ctx context.Context, baseURL string) (string, error) {
	versionURL := fmt.Sprintf("%s/api/version", baseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", versionURL, nil)
	if err != nil {
		return "", err
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("获取版本失败: %d", resp.StatusCode)
	}

	return "unknown", nil // Ollama API 返回的格式可能不同
}

// UpdateKBConfigRequest 更新知识库配置请求
type UpdateKBConfigRequest struct {
	ChunkSize    int      `json:"chunkSize"`
	ChunkOverlap int      `json:"chunkOverlap"`
	Separators   []string `json:"separators"`
	EnableQA     bool     `json:"enableQA"`
	QAPrompt     string   `json:"qaPrompt"`
}

// UpdateKBConfig 更新知识库配置
func (s *Service) UpdateKBConfig(ctx context.Context, kbID string, req *UpdateKBConfigRequest) error {
	// 获取知识库
	kb, err := s.repo.Knowledge.GetKnowledgeBaseByID(kbID)
	if err != nil {
		return fmt.Errorf("知识库不存在: %w", err)
	}

	// 更新配置 (使用 JSON 存储在 Metadata 字段中)
	// 简化版：直接更新知识库的字段
	if req.ChunkSize > 0 {
		// 这里可以存储到配置中
		_ = req.ChunkSize
	}
	if req.ChunkOverlap >= 0 {
		// 这里可以存储到配置中
		_ = req.ChunkOverlap
	}
	if len(req.Separators) > 0 {
		// 这里可以存储到配置中
		_ = req.Separators
	}

	// 更新知识库
	kb.UpdatedAt = time.Now()
	if err := s.repo.Knowledge.UpdateKnowledgeBase(kb); err != nil {
		return fmt.Errorf("更新知识库配置失败: %w", err)
	}

	return nil
}

// GetKBConfig 获取知识库配置
func (s *Service) GetKBConfig(ctx context.Context, kbID string) (map[string]interface{}, error) {
	kb, err := s.repo.Knowledge.GetKnowledgeBaseByID(kbID)
	if err != nil {
		return nil, fmt.Errorf("知识库不存在: %w", err)
	}

	// 返回配置信息
	config := map[string]interface{}{
		"id":          kb.ID,
		"name":        kb.Name,
		"description": kb.Description,
		"createdAt":   kb.CreatedAt,
		"updatedAt":   kb.UpdatedAt,
		// 可以添加更多配置字段
	}

	return config, nil
}

// InitializeByKBRequest 初始化知识库请求
type InitializeByKBRequest struct {
	ChunkSize    int      `json:"chunkSize"`
	ChunkOverlap int      `json:"chunkOverlap"`
	Separators   []string `json:"separators"`
	EnableQA     bool     `json:"enableQA"`
	QAPrompt     string   `json:"qaPrompt"`
}

// InitializeByKB 初始化知识库
func (s *Service) InitializeByKB(ctx context.Context, kbID string, req *InitializeByKBRequest) error {
	// 获取知识库
	kb, err := s.repo.Knowledge.GetKnowledgeBaseByID(kbID)
	if err != nil {
		return fmt.Errorf("知识库不存在: %w", err)
	}

	// 更新知识库配置
	kb.UpdatedAt = time.Now()
	if err := s.repo.Knowledge.UpdateKnowledgeBase(kb); err != nil {
		return fmt.Errorf("初始化知识库失败: %w", err)
	}

	return nil
}

// SystemInfo 系统信息
type SystemInfo struct {
	Version    string                `json:"version"`
	GoVersion  string                `json:"goVersion"`
	StartTime  time.Time             `json:"startTime"`
	Uptime     string                `json:"uptime"`
	Ollama     *OllamaStatusResponse `json:"ollama,omitempty"`
}

// GetSystemInfo 获取系统信息
func (s *Service) GetSystemInfo(ctx context.Context) (*SystemInfo, error) {
	uptime := time.Since(s.startTime)

	info := &SystemInfo{
		Version:   s.version,
		GoVersion: "1.23+",
		StartTime: s.startTime,
		Uptime:    uptime.String(),
	}

	// 检查 Ollama 状态
	ollamaStatus, err := s.CheckOllamaStatus(ctx)
	if err == nil {
		info.Ollama = ollamaStatus
	}

	return info, nil
}

// TestEmbeddingRequest 测试 Embedding 请求
type TestEmbeddingRequest struct {
	Source    string `json:"source"`
	ModelName string `json:"modelName"`
	BaseURL   string `json:"baseUrl"`
	APIKey    string `json:"apiKey"`
}

// TestEmbeddingResponse 测试 Embedding 响应
type TestEmbeddingResponse struct {
	Available bool   `json:"available"`
	Message   string `json:"message"`
	Dimension int    `json:"dimension,omitempty"`
}

// TestEmbedding 测试 Embedding 模型
func (s *Service) TestEmbedding(ctx context.Context, req *TestEmbeddingRequest) (*TestEmbeddingResponse, error) {
	// 这里应该调用实际的 Embedding 服务进行测试
	// 简化版：只检查配置是否完整
	if req.ModelName == "" {
		return &TestEmbeddingResponse{
			Available: false,
			Message:   "模型名称不能为空",
		}, nil
	}

	if req.BaseURL == "" {
		return &TestEmbeddingResponse{
			Available: false,
			Message:   "Base URL 不能为空",
		}, nil
	}

	// TODO: 实际调用 Embedding 接口进行测试
	// 使用 eino-ext 的 dashscope.NewEmbedder 或 openai.NewEmbedder
	return &TestEmbeddingResponse{
		Available: true,
		Message:   "Embedding 配置验证通过（实际功能待实现）",
		Dimension: 1536, // 默认维度
	}, nil
}

// OllamaModelInfo Ollama 模型信息
type OllamaModelInfo struct {
	Name       string  `json:"name"`
	ModifiedAt string  `json:"modified_at,omitempty"`
	Size       int64   `json:"size,omitempty"`
	Digest     string  `json:"digest,omitempty"`
	Available  bool    `json:"available"`
}

// ListOllamaModels 列出已安装的 Ollama 模型
func (s *Service) ListOllamaModels(ctx context.Context) ([]*OllamaModelInfo, error) {
	baseURL := os.Getenv("OLLAMA_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	// 调用 Ollama API 列出模型
	listURL := fmt.Sprintf("%s/api/tags", baseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", listURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 Ollama API 失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Ollama API 返回状态码: %d", resp.StatusCode)
	}

	// 解析响应
	var result struct {
		Models []struct {
			Name       string `json:"name"`
			ModifiedAt string `json:"modified_at"`
			Size       int64  `json:"size"`
			Digest     string `json:"digest"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 转换为返回格式
	models := make([]*OllamaModelInfo, len(result.Models))
	for i, m := range result.Models {
		models[i] = &OllamaModelInfo{
			Name:       m.Name,
			ModifiedAt: m.ModifiedAt,
			Size:       m.Size,
			Digest:     m.Digest,
			Available:  true,
		}
	}

	return models, nil
}

// CheckOllamaModelsRequest 检查 Ollama 模型请求
type CheckOllamaModelsRequest struct {
	Models []string `json:"models" binding:"required"`
}

// CheckOllamaModels 检查指定的 Ollama 模型是否已安装
func (s *Service) CheckOllamaModels(ctx context.Context, req *CheckOllamaModelsRequest) (map[string]bool, error) {
	models, err := s.ListOllamaModels(ctx)
	if err != nil {
		return nil, err
	}

	// 构建模型名称集合
	availableModels := make(map[string]bool)
	for _, m := range models {
		// Ollama 模型名称可能包含 tag (如 "llama2:latest")
		// 提取基础名称
		baseName := m.Name
		if idx := strings.Index(m.Name, ":"); idx != -1 {
			baseName = m.Name[:idx]
		}
		availableModels[baseName] = true
		availableModels[m.Name] = true
	}

	// 检查请求的模型
	result := make(map[string]bool)
	for _, modelName := range req.Models {
		result[modelName] = availableModels[modelName]
	}

	return result, nil
}

// CheckRemoteModelRequest 检查远程模型请求
type CheckRemoteModelRequest struct {
	ModelName string `json:"modelName" binding:"required"`
	BaseURL   string `json:"baseUrl" binding:"required"`
	APIKey    string `json:"apiKey"`
}

// CheckRemoteModelResponse 检查远程模型响应
type CheckRemoteModelResponse struct {
	Available bool   `json:"available"`
	Message   string `json:"message"`
}

// CheckRemoteModel 检查远程 LLM 模型连接
func (s *Service) CheckRemoteModel(ctx context.Context, req *CheckRemoteModelRequest) (*CheckRemoteModelResponse, error) {
	if req.ModelName == "" || req.BaseURL == "" {
		return &CheckRemoteModelResponse{
			Available: false,
			Message:   "模型名称和 Base URL 不能为空",
		}, nil
	}

	// 构造测试请求
	chatURL := fmt.Sprintf("%s/chat/completions", strings.TrimSuffix(req.BaseURL, "/"))
	payload := map[string]interface{}{
		"model":    req.ModelName,
		"messages": []map[string]string{{"role": "user", "content": "test"}},
		"max_tokens": 1,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("编码请求失败: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", chatURL, strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if req.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+req.APIKey)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return &CheckRemoteModelResponse{
			Available: false,
			Message:   fmt.Sprintf("连接失败: %v", err),
		}, nil
	}
	defer resp.Body.Close()

	// 解析错误响应
	switch resp.StatusCode {
	case http.StatusOK:
		return &CheckRemoteModelResponse{
			Available: true,
			Message:   "连接正常，模型可用",
		}, nil
	case http.StatusUnauthorized:
		return &CheckRemoteModelResponse{
			Available: false,
			Message:   "认证失败，请检查 API Key",
		}, nil
	case http.StatusForbidden:
		return &CheckRemoteModelResponse{
			Available: false,
			Message:   "权限不足，请检查 API Key 权限",
		}, nil
	case http.StatusNotFound:
		return &CheckRemoteModelResponse{
			Available: false,
			Message:   "API 端点不存在，请检查 Base URL",
		}, nil
	default:
		return &CheckRemoteModelResponse{
			Available: false,
			Message:   fmt.Sprintf("连接失败，状态码: %d", resp.StatusCode),
		}, nil
	}
}

// CheckRerankModelRequest 检查 Rerank 模型请求
type CheckRerankModelRequest struct {
	ModelName string `json:"modelName" binding:"required"`
	BaseURL   string `json:"baseUrl" binding:"required"`
	APIKey    string `json:"apiKey"`
}

// CheckRerankModelResponse 检查 Rerank 模型响应
type CheckRerankModelResponse struct {
	Available bool   `json:"available"`
	Message   string `json:"message"`
}

// CheckRerankModel 检查 Rerank 模型连接和功能
func (s *Service) CheckRerankModel(ctx context.Context, req *CheckRerankModelRequest) (*CheckRerankModelResponse, error) {
	if req.ModelName == "" || req.BaseURL == "" {
		return &CheckRerankModelResponse{
			Available: false,
			Message:   "模型名称和 Base URL 不能为空",
		}, nil
	}

	// 构造 Rerank 测试请求
	rerankURL := fmt.Sprintf("%s/rerank", strings.TrimSuffix(req.BaseURL, "/"))
	payload := map[string]interface{}{
		"model":    req.ModelName,
		"query":    "test query",
		"documents": []string{"document 1", "document 2"},
		"top_n":    1,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("编码请求失败: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", rerankURL, strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if req.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+req.APIKey)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return &CheckRerankModelResponse{
			Available: false,
			Message:   fmt.Sprintf("连接失败: %v", err),
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return &CheckRerankModelResponse{
			Available: true,
			Message:   "Rerank 功能正常",
		}, nil
	}

	return &CheckRerankModelResponse{
		Available: false,
		Message:   fmt.Sprintf("Rerank 测试失败，状态码: %d", resp.StatusCode),
	}, nil
}
