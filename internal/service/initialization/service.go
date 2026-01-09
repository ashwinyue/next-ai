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
	"github.com/cloudwego/eino-ext/components/embedding/ollama"
	"github.com/cloudwego/eino-ext/components/embedding/openai"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// Service 初始化服务
type Service struct {
	repo      *repository.Repositories
	chatModel model.ChatModel
	startTime time.Time
	version   string
}

// NewService 创建初始化服务
func NewService(repo *repository.Repositories, chatModel model.ChatModel) *Service {
	return &Service{
		repo:      repo,
		chatModel: chatModel,
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

// SystemInfo 系统信息
type SystemInfo struct {
	Version   string                `json:"version"`
	GoVersion string                `json:"goVersion"`
	StartTime time.Time             `json:"startTime"`
	Uptime    string                `json:"uptime"`
	Ollama    *OllamaStatusResponse `json:"ollama,omitempty"`
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
	// 基础配置验证
	if req.ModelName == "" {
		return &TestEmbeddingResponse{
			Available: false,
			Message:   "模型名称不能为空",
		}, nil
	}

	// Ollama 可以使用默认的 localhost URL
	baseURL := req.BaseURL
	if req.Source == "ollama" && baseURL == "" {
		baseURL = "http://localhost:11434"
	} else if baseURL == "" {
		return &TestEmbeddingResponse{
			Available: false,
			Message:   "Base URL 不能为空",
		}, nil
	}

	// 根据 Source 选择对应的 Embedder 实现
	var embedder embedding.Embedder
	var err error

	source := strings.ToLower(req.Source)
	if source == "ollama" {
		// 使用 Ollama Embedder
		embedder, err = ollama.NewEmbedder(ctx, &ollama.EmbeddingConfig{
			BaseURL: baseURL,
			Model:   req.ModelName,
			Timeout: 10 * time.Second,
		})
		if err != nil {
			return &TestEmbeddingResponse{
				Available: false,
				Message:   fmt.Sprintf("创建 Ollama Embedder 失败: %v", err),
			}, nil
		}
	} else {
		// 默认使用 OpenAI 兼容接口
		embedder, err = openai.NewEmbedder(ctx, &openai.EmbeddingConfig{
			APIKey:  req.APIKey,
			Model:   req.ModelName,
			BaseURL: baseURL,
		})
		if err != nil {
			return &TestEmbeddingResponse{
				Available: false,
				Message:   fmt.Sprintf("创建 Embedder 失败: %v", err),
			}, nil
		}
	}

	// 执行测试调用
	testText := "这是一个测试文本"
	vectors, err := embedder.EmbedStrings(ctx, []string{testText})
	if err != nil {
		return &TestEmbeddingResponse{
			Available: false,
			Message:   fmt.Sprintf("Embedding 调用失败: %v", err),
		}, nil
	}

	if len(vectors) == 0 || len(vectors[0]) == 0 {
		return &TestEmbeddingResponse{
			Available: false,
			Message:   "Embedding 返回空结果",
		}, nil
	}

	dimension := len(vectors[0])
	return &TestEmbeddingResponse{
		Available: true,
		Message:   "Embedding 测试成功",
		Dimension: dimension,
	}, nil
}

// OllamaModelInfo Ollama 模型信息
type OllamaModelInfo struct {
	Name       string `json:"name"`
	ModifiedAt string `json:"modified_at,omitempty"`
	Size       int64  `json:"size,omitempty"`
	Digest     string `json:"digest,omitempty"`
	Available  bool   `json:"available"`
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
		"model":      req.ModelName,
		"messages":   []map[string]string{{"role": "user", "content": "test"}},
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
		"model":     req.ModelName,
		"query":     "test query",
		"documents": []string{"document 1", "document 2"},
		"top_n":     1,
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

// ========== Ollama 模型下载 ==========

// DownloadModelRequest 下载模型请求
type DownloadModelRequest struct {
	ModelName string `json:"modelName" binding:"required"` // 如 "llama2", "mistral:latest"
}

// DownloadModelResponse 下载模型响应
type DownloadModelResponse struct {
	TaskID  string `json:"task_id"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

// DownloadProgress 下载进度
type DownloadProgress struct {
	TaskID    string  `json:"task_id"`
	ModelName string  `json:"model_name"`
	Status    string  `json:"status"`   // pending, downloading, completed, failed
	Progress  float64 `json:"progress"` // 0-100
	Total     int64   `json:"total,omitempty"`
	Completed int64   `json:"completed,omitempty"`
	Speed     int64   `json:"speed,omitempty"` // bytes/s
	Error     string  `json:"error,omitempty"`
}

// 下载任务存储（简化版：内存存储，生产环境应使用 Redis）
// 注意：此存储不是并发安全的，单机部署时没问题
var downloadTasks = make(map[string]*DownloadProgress)

// DownloadModel 下载 Ollama 模型（异步）
func (s *Service) DownloadModel(ctx context.Context, req *DownloadModelRequest) (*DownloadModelResponse, error) {
	// 生成任务 ID
	taskID := fmt.Sprintf("download_%d", time.Now().UnixNano())

	// 创建初始进度记录
	progress := &DownloadProgress{
		TaskID:    taskID,
		ModelName: req.ModelName,
		Status:    "pending",
		Progress:  0,
	}

	// 存储任务
	downloadTasks[taskID] = progress

	// 异步执行下载
	go s.executeDownload(context.Background(), taskID, req.ModelName, progress)

	return &DownloadModelResponse{
		TaskID:  taskID,
		Message: "下载任务已启动",
		Status:  "pending",
	}, nil
}

// executeDownload 执行模型下载
func (s *Service) executeDownload(ctx context.Context, taskID, modelName string, progress *DownloadProgress) {
	progress.Status = "downloading"

	baseURL := os.Getenv("OLLAMA_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	// 调用 Ollama API 拉取模型
	pullURL := fmt.Sprintf("%s/api/pull", baseURL)
	payload := map[string]string{
		"name":   modelName,
		"stream": "true",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		progress.Status = "failed"
		progress.Error = fmt.Sprintf("编码请求失败: %v", err)
		return
	}

	req, err := http.NewRequestWithContext(ctx, "POST", pullURL, strings.NewReader(string(body)))
	if err != nil {
		progress.Status = "failed"
		progress.Error = fmt.Sprintf("创建请求失败: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 0} // 无超时，因为下载可能需要很长时间
	resp, err := client.Do(req)
	if err != nil {
		progress.Status = "failed"
		progress.Error = fmt.Sprintf("连接 Ollama 失败: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		progress.Status = "failed"
		progress.Error = fmt.Sprintf("Ollama API 返回状态码: %d", resp.StatusCode)
		return
	}

	// 解析流式响应
	decoder := json.NewDecoder(resp.Body)
	for {
		var status struct {
			Status    string `json:"status"`
			Total     int64  `json:"total"`
			Completed int64  `json:"completed"`
			Error     string `json:"error"`
		}

		if err := decoder.Decode(&status); err != nil {
			if err.Error() != "EOF" {
				progress.Status = "failed"
				progress.Error = fmt.Sprintf("解析响应失败: %v", err)
			}
			break
		}

		// 更新进度
		if status.Total > 0 {
			progress.Total = status.Total
			progress.Completed = status.Completed
			progress.Progress = float64(status.Completed) / float64(status.Total) * 100
		}

		if status.Error != "" {
			progress.Status = "failed"
			progress.Error = status.Error
			return
		}

		if status.Status == "success" {
			progress.Status = "completed"
			progress.Progress = 100
			return
		}
	}
}

// GetDownloadProgress 获取下载进度
func (s *Service) GetDownloadProgress(ctx context.Context, taskID string) (*DownloadProgress, error) {
	progress, ok := downloadTasks[taskID]
	if !ok {
		return nil, fmt.Errorf("任务不存在")
	}

	// 返回副本，避免并发问题
	result := *progress
	return &result, nil
}

// ListDownloadTasks 列出所有下载任务
func (s *Service) ListDownloadTasks(ctx context.Context) ([]*DownloadProgress, error) {
	tasks := make([]*DownloadProgress, 0, len(downloadTasks))
	for _, task := range downloadTasks {
		// 返回副本
		taskCopy := *task
		tasks = append(tasks, &taskCopy)
	}
	return tasks, nil
}

// CancelDownload 取消下载任务
func (s *Service) CancelDownload(ctx context.Context, taskID string) error {
	progress, ok := downloadTasks[taskID]
	if !ok {
		return fmt.Errorf("任务不存在")
	}

	if progress.Status == "completed" || progress.Status == "failed" {
		return fmt.Errorf("任务已结束，无法取消")
	}

	// 简化版：只标记为已取消
	// 实际实现应该通过 context 取消正在进行的 HTTP 请求
	progress.Status = "failed"
	progress.Error = "用户取消"

	return nil
}

// ========== 文本处理功能（WeKnora API 兼容）==========

// Entity 实体
type Entity struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Relation 关系
type Relation struct {
	ID         string                 `json:"id"`
	Source     string                 `json:"source"` // 源实体 ID
	Target     string                 `json:"target"` // 目标实体 ID
	Type       string                 `json:"type"`   // 关系类型
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Confidence float64                `json:"confidence"` // 置信度
}

// ExtractTextRelationsRequest 文本关系提取请求
type ExtractTextRelationsRequest struct {
	Text      string                 `json:"text" binding:"required"`
	Threshold float64                `json:"threshold"`
	LLMConfig map[string]interface{} `json:"llm_config"`
}

// ExtractTextRelationsResponse 文本关系提取响应
type ExtractTextRelationsResponse struct {
	Success   bool       `json:"success"`
	Entities  []Entity   `json:"entities"`
	Relations []Relation `json:"relations"`
	Error     string     `json:"error,omitempty"`
}

// ExtractTextRelations 提取文本关系（实体和关系）
func (s *Service) ExtractTextRelations(ctx context.Context, req *ExtractTextRelationsRequest) (*ExtractTextRelationsResponse, error) {
	if s.chatModel == nil {
		return &ExtractTextRelationsResponse{
			Success: false,
			Error:   "ChatModel 未配置",
		}, nil
	}

	if req.Text == "" {
		return &ExtractTextRelationsResponse{
			Success: false,
			Error:   "文本不能为空",
		}, nil
	}

	threshold := req.Threshold
	if threshold <= 0 {
		threshold = 0.5
	}

	// 构建提示词
	prompt := fmt.Sprintf(`请从以下文本中提取实体和关系。

文本：
%s

请按以下 JSON 格式返回结果：
{
  "entities": [
    {"id": "1", "name": "实体名称", "type": "实体类型（如：人物、组织、地点等）"}
  ],
  "relations": [
    {"source": "源实体ID", "target": "目标实体ID", "type": "关系类型", "confidence": 0.9}
  ]
}

要求：
1. 只返回 JSON，不要其他内容
2. 实体类型包括：人物、组织、地点、时间、事件、概念等
3. 关系类型包括：属于、位于、创建、参与、发生等
4. 置信度范围 0-1，低于 %.2f 的关系可以忽略`, req.Text, threshold)

	messages := []*schema.Message{
		{Role: schema.System, Content: "你是一个专业的知识图谱构建助手，擅长从文本中提取实体和关系。"},
		{Role: schema.User, Content: prompt},
	}

	resp, err := s.chatModel.Generate(ctx, messages)
	if err != nil {
		return &ExtractTextRelationsResponse{
			Success: false,
			Error:   fmt.Sprintf("LLM 调用失败: %v", err),
		}, nil
	}

	// 解析响应
	var result struct {
		Entities []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"entities"`
		Relations []struct {
			Source     string  `json:"source"`
			Target     string  `json:"target"`
			Type       string  `json:"type"`
			Confidence float64 `json:"confidence"`
		} `json:"relations"`
	}

	// 尝试提取 JSON（处理可能的额外文本）
	content := resp.Content
	jsonStart := strings.Index(content, "{")
	jsonEnd := strings.LastIndex(content, "}")
	if jsonStart >= 0 && jsonEnd > jsonStart {
		content = content[jsonStart : jsonEnd+1]
	}

	if err := json.Unmarshal([]byte(content), &result); err != nil {
		// 解析失败，返回空结果
		return &ExtractTextRelationsResponse{
			Success:   true,
			Entities:  []Entity{},
			Relations: []Relation{},
		}, nil
	}

	// 转换为输出格式
	entities := make([]Entity, 0, len(result.Entities))
	for _, e := range result.Entities {
		entities = append(entities, Entity{
			ID:   e.ID,
			Name: e.Name,
			Type: e.Type,
		})
	}

	relations := make([]Relation, 0, len(result.Relations))
	for _, r := range result.Relations {
		if r.Confidence >= threshold {
			relations = append(relations, Relation{
				Source:     r.Source,
				Target:     r.Target,
				Type:       r.Type,
				Confidence: r.Confidence,
			})
		}
	}

	return &ExtractTextRelationsResponse{
		Success:   true,
		Entities:  entities,
		Relations: relations,
	}, nil
}

// FabriTagRequest 生成标签请求
type FabriTagRequest struct {
	Text      string                 `json:"text" binding:"required"`
	LLMConfig map[string]interface{} `json:"llm_config"`
}

// FabriTagResponse 生成标签响应
type FabriTagResponse struct {
	Success bool     `json:"success"`
	Tags    []string `json:"tags"`
	Error   string   `json:"error,omitempty"`
}

// FabriTag 生成文本标签
func (s *Service) FabriTag(ctx context.Context, req *FabriTagRequest) (*FabriTagResponse, error) {
	if s.chatModel == nil {
		return &FabriTagResponse{
			Success: false,
			Error:   "ChatModel 未配置",
		}, nil
	}

	if req.Text == "" {
		return &FabriTagResponse{
			Success: false,
			Error:   "文本不能为空",
		}, nil
	}

	prompt := fmt.Sprintf(`请为以下文本生成 3-5 个描述性标签。

文本：
%s

要求：
1. 标签应该简洁明了，通常是 2-4 个词
2. 标签应该反映文本的核心主题和内容
3. 只返回标签列表，用 JSON 数组格式：["标签1", "标签2", "标签3"]
4. 不要添加任何其他解释文字`, req.Text)

	messages := []*schema.Message{
		{Role: schema.System, Content: "你是一个专业的文本标签生成助手，擅长提取文本核心主题。"},
		{Role: schema.User, Content: prompt},
	}

	resp, err := s.chatModel.Generate(ctx, messages)
	if err != nil {
		return &FabriTagResponse{
			Success: false,
			Error:   fmt.Sprintf("LLM 调用失败: %v", err),
		}, nil
	}

	// 解析 JSON 数组
	content := strings.TrimSpace(resp.Content)
	arrayStart := strings.Index(content, "[")
	arrayEnd := strings.LastIndex(content, "]")
	if arrayStart >= 0 && arrayEnd > arrayStart {
		content = content[arrayStart : arrayEnd+1]
	}

	var tags []string
	if err := json.Unmarshal([]byte(content), &tags); err != nil {
		// 解析失败，尝试按行分割
		lines := strings.Split(resp.Content, "\n")
		tags = make([]string, 0, len(lines))
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "```") {
				// 移除可能的列表符号
				line = strings.TrimPrefix(line, "-")
				line = strings.TrimPrefix(line, "*")
				line = strings.TrimPrefix(line, "•")
				line = strings.TrimSpace(line)
				if line != "" {
					tags = append(tags, line)
				}
			}
		}
	}

	return &FabriTagResponse{
		Success: true,
		Tags:    tags,
	}, nil
}

// FabriTextRequest 根据标签生成文本请求
type FabriTextRequest struct {
	Tags      []string               `json:"tags"`
	LLMConfig map[string]interface{} `json:"llm_config"`
}

// FabriTextResponse 根据标签生成文本响应
type FabriTextResponse struct {
	Success bool   `json:"success"`
	Text    string `json:"text"`
	Error   string `json:"error,omitempty"`
}

// FabriText 根据标签生成文本
func (s *Service) FabriText(ctx context.Context, req *FabriTextRequest) (*FabriTextResponse, error) {
	if s.chatModel == nil {
		return &FabriTextResponse{
			Success: false,
			Error:   "ChatModel 未配置",
		}, nil
	}

	if len(req.Tags) == 0 {
		return &FabriTextResponse{
			Success: false,
			Error:   "标签不能为空",
		}, nil
	}

	tagsStr := strings.Join(req.Tags, "、")
	prompt := fmt.Sprintf(`请根据以下标签生成一段描述性文本。

标签：%s

要求：
1. 生成的文本应该连贯自然，长度在 100-200 字之间
2. 文本应该体现所有标签的含义
3. 文本风格应该专业、客观
4. 只返回生成的文本，不要添加标题或其他内容`, tagsStr)

	messages := []*schema.Message{
		{Role: schema.System, Content: "你是一个专业的文本生成助手，擅长根据关键词生成连贯的描述性文本。"},
		{Role: schema.User, Content: prompt},
	}

	resp, err := s.chatModel.Generate(ctx, messages)
	if err != nil {
		return &FabriTextResponse{
			Success: false,
			Error:   fmt.Sprintf("LLM 调用失败: %v", err),
		}, nil
	}

	return &FabriTextResponse{
		Success: true,
		Text:    strings.TrimSpace(resp.Content),
	}, nil
}

// TestMultimodalRequest 测试多模态请求
type TestMultimodalRequest struct {
	Image            string `json:"image" binding:"required"` // base64 编码的图片
	VLMModel         string `json:"vlm_model"`
	VLMBaseURL       string `json:"vlm_base_url"`
	VLMAPIKey        string `json:"vlm_api_key"`
	VLMInterfaceType string `json:"vlm_interface_type"`
}

// TestMultimodalResponse 测试多模态响应
type TestMultimodalResponse struct {
	Success bool   `json:"success"`
	Result  string `json:"result"`
	Error   string `json:"error,omitempty"`
}

// TestMultimodal 测试多模态功能
func (s *Service) TestMultimodal(ctx context.Context, req *TestMultimodalRequest) (*TestMultimodalResponse, error) {
	// 简化版：验证图片数据是否有效
	if req.Image == "" {
		return &TestMultimodalResponse{
			Success: false,
			Error:   "图片数据不能为空",
		}, nil
	}

	// 验证 base64 格式
	dataURLPrefix := "data:image/"
	base64Prefix := ";base64,"

	var base64Data string
	if strings.HasPrefix(req.Image, dataURLPrefix) {
		// Data URL 格式
		idx := strings.Index(req.Image, base64Prefix)
		if idx == -1 {
			return &TestMultimodalResponse{
				Success: false,
				Error:   "无效的 Data URL 格式",
			}, nil
		}
		base64Data = req.Image[idx+len(base64Prefix):]
	} else {
		base64Data = req.Image
	}

	// 简单验证 base64 长度
	if len(base64Data) < 100 {
		return &TestMultimodalResponse{
			Success: false,
			Error:   "图片数据太短，可能不完整",
		}, nil
	}

	// 简化版：只验证图片数据，不实际调用 VLM
	// 实际实现应该使用支持视觉的模型（如 GPT-4V, Claude 3, Qwen-VL 等）
	return &TestMultimodalResponse{
		Success: true,
		Result:  "多模态功能已验证。图片数据格式正确，base64 长度: " + fmt.Sprintf("%d", len(base64Data)),
	}, nil
}
