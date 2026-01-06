package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/ashwinyue/next-rag/next-ai/internal/config"
	agentmodel "github.com/ashwinyue/next-rag/next-ai/internal/model"
	"github.com/ashwinyue/next-rag/next-ai/internal/repository"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
)

// Service Agent 服务
// 参考 eino-examples，直接使用 eino ADK，不做额外封装
type Service struct {
	repo     *repository.Repositories
	cfg      *config.Config
	allTools []tool.BaseTool
}

// NewService 创建 Agent 服务
func NewService(repo *repository.Repositories, cfg *config.Config, allTools []tool.BaseTool) *Service {
	return &Service{
		repo:     repo,
		cfg:      cfg,
		allTools: allTools,
	}
}

// CreateAgentRequest 创建 Agent 请求
type CreateAgentRequest struct {
	Name         string   `json:"name" binding:"required"`
	DisplayName  string   `json:"display_name"`
	Description  string   `json:"description"`
	SystemPrompt string   `json:"system_prompt"`
	Tools        []string `json:"tools"`
	MaxIter      int      `json:"max_iterations"`
	Temperature  float64  `json:"temperature"`
	MaxTokens    int      `json:"max_tokens"`
	Model        string   `json:"model"`
}

// CreateAgent 创建 Agent
func (s *Service) CreateAgent(ctx context.Context, req *CreateAgentRequest) (*agentmodel.Agent, error) {
	if _, err := s.repo.Agent.GetByName(req.Name); err == nil {
		return nil, fmt.Errorf("agent name already exists")
	}

	// 构建 AgentConfig
	agentConfig := agentmodel.AgentConfig{
		SystemPrompt: req.SystemPrompt,
		Temperature:  req.Temperature,
		MaxTokens:    req.MaxTokens,
		Model:        req.Model,
		Tools:        req.Tools,
	}

	configJSON, err := json.Marshal(agentConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	agent := &agentmodel.Agent{
		ID:          uuid.New().String(),
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Config:      string(configJSON),
		IsActive:    true,
	}

	if err := s.repo.Agent.Create(agent); err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	return agent, nil
}

// GetAgent 获取 Agent
func (s *Service) GetAgent(ctx context.Context, id string) (*agentmodel.Agent, error) {
	return s.repo.Agent.GetByID(id)
}

// ListAgentsRequest 列出 Agent 请求
type ListAgentsRequest struct {
	Page int `json:"page"`
	Size int `json:"size"`
}

// ListAgents 列出 Agent
func (s *Service) ListAgents(ctx context.Context, req *ListAgentsRequest) ([]*agentmodel.Agent, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 || req.Size > 100 {
		req.Size = 20
	}

	offset := (req.Page - 1) * req.Size
	return s.repo.Agent.List(offset, req.Size)
}

// ListActiveAgents 列出活跃 Agent
func (s *Service) ListActiveAgents(ctx context.Context) ([]*agentmodel.Agent, error) {
	return s.repo.Agent.ListActive()
}

// UpdateAgent 更新 Agent
func (s *Service) UpdateAgent(ctx context.Context, id string, req *CreateAgentRequest) (*agentmodel.Agent, error) {
	agentModel, err := s.repo.Agent.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("agent not found: %w", err)
	}

	agentModel.Name = req.Name
	agentModel.DisplayName = req.DisplayName
	agentModel.Description = req.Description

	agentConfig := agentmodel.AgentConfig{
		SystemPrompt: req.SystemPrompt,
		Temperature:  req.Temperature,
		MaxTokens:    req.MaxTokens,
		Model:        req.Model,
		Tools:        req.Tools,
	}

	configJSON, err := json.Marshal(agentConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}
	agentModel.Config = string(configJSON)

	if err := s.repo.Agent.Update(agentModel); err != nil {
		return nil, fmt.Errorf("failed to update agent: %w", err)
	}

	return agentModel, nil
}

// DeleteAgent 删除 Agent
func (s *Service) DeleteAgent(ctx context.Context, id string) error {
	if err := s.repo.Agent.Delete(id); err != nil {
		return fmt.Errorf("failed to delete agent: %w", err)
	}
	return nil
}

// GetAgentConfig 获取 Agent 配置
func (s *Service) GetAgentConfig(ctx context.Context, id string) (*agentmodel.AgentConfig, error) {
	agentModel, err := s.repo.Agent.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("agent not found: %w", err)
	}

	var cfg agentmodel.AgentConfig
	if err := json.Unmarshal([]byte(agentModel.Config), &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// RunRequest 运行 Agent 请求
type RunRequest struct {
	Query     string `json:"query" binding:"required"`
	SessionID string `json:"session_id"`
}

// RunResponse 运行响应
type RunResponse struct {
	Answer string `json:"answer"`
}

// StreamEvent 流式事件
type StreamEvent struct {
	Type     string `json:"type"`             // start, message, tool_call, error, end
	Data     string `json:"data"`
	ToolName string `json:"tool_name,omitempty"`
}

// newToolCallingChatModel 创建支持工具调用的 ChatModel
func (s *Service) newToolCallingChatModel(ctx context.Context) (model.ToolCallingChatModel, error) {
	aiCfg := s.cfg.AI

	var apiKey, baseURL, modelName string

	switch aiCfg.Provider {
	case "openai":
		apiKey = aiCfg.OpenAI.APIKey
		baseURL = aiCfg.OpenAI.BaseURL
		modelName = aiCfg.OpenAI.Model
	case "alibaba", "qwen", "dashscope":
		apiKey = aiCfg.Alibaba.AccessKeySecret
		baseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
		modelName = aiCfg.Alibaba.Model
	case "deepseek":
		apiKey = aiCfg.DeepSeek.APIKey
		baseURL = aiCfg.DeepSeek.BaseURL
		modelName = aiCfg.DeepSeek.Model
	default:
		return nil, fmt.Errorf("unsupported ai provider: %s", aiCfg.Provider)
	}

	if apiKey == "" {
		return nil, fmt.Errorf("api_key is required for provider: %s", aiCfg.Provider)
	}

	if modelName == "" {
		modelName = "gpt-4o-mini"
	}

	temperature := float32(0.7)

	return openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:      apiKey,
		BaseURL:     baseURL,
		Model:       modelName,
		Temperature: &temperature,
	})
}

// createAgent 创建 eino Agent
func (s *Service) createAgent(ctx context.Context, name, description, systemPrompt string, selectedTools []tool.BaseTool, maxIterations int) (*adk.ChatModelAgent, error) {
	chatModel, err := s.newToolCallingChatModel(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat model: %w", err)
	}

	if maxIterations <= 0 {
		maxIterations = 10
	}

	agentCfg := &adk.ChatModelAgentConfig{
		Name:          name,
		Description:   description,
		Instruction:   systemPrompt,
		Model:         chatModel,
		MaxIterations: maxIterations,
	}

	// 添加工具
	if len(selectedTools) > 0 {
		agentCfg.ToolsConfig = adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: selectedTools,
			},
		}
	}

	return adk.NewChatModelAgent(ctx, agentCfg)
}

// Run 运行 Agent（同步）
func (s *Service) Run(ctx context.Context, agentID string, req *RunRequest) (*RunResponse, error) {
	agentModel, err := s.repo.Agent.GetByID(agentID)
	if err != nil {
		return nil, fmt.Errorf("agent not found: %w", err)
	}

	agentConfig, err := s.GetAgentConfig(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("invalid agent config: %w", err)
	}

	// 获取指定工具
	var selectedTools []tool.BaseTool
	if len(agentConfig.Tools) > 0 {
		selectedTools, err = GetToolsByName(ctx, agentConfig.Tools, s.allTools)
		if err != nil {
			return nil, fmt.Errorf("failed to get tools: %w", err)
		}
	} else {
		selectedTools = s.allTools
	}

	// 创建 eino Agent
	einoAgent, err := s.createAgent(ctx, agentModel.Name, agentModel.Description, agentConfig.SystemPrompt, selectedTools, agentConfig.MaxTokens)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	// 加载历史消息
	var history []*schema.Message
	if req.SessionID != "" {
		history = s.loadHistory(ctx, req.SessionID)
	}

	// 构建输入消息
	messages := buildMessages(history, req.Query)

	// 运行 Agent
	iter := einoAgent.Run(ctx, &adk.AgentInput{
		Messages:        messages,
		EnableStreaming: false,
	})

	// 收集结果
	var result string
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}

		if event.Err != nil {
			if event.Err == io.EOF {
				break
			}
			return nil, fmt.Errorf("agent event error: %w", event.Err)
		}

		if event.Output != nil && event.Output.MessageOutput != nil {
			msg, err := event.Output.MessageOutput.GetMessage()
			if err != nil {
				continue
			}
			if msg.Role == schema.Assistant {
				result = msg.Content
			}
		}
	}

	// 保存消息到会话
	if req.SessionID != "" {
		s.saveMessage(ctx, req.SessionID, "user", req.Query)
		s.saveMessage(ctx, req.SessionID, "assistant", result)
	}

	return &RunResponse{Answer: result}, nil
}

// Stream 运行 Agent（流式）
func (s *Service) Stream(ctx context.Context, agentID string, req *RunRequest) (<-chan StreamEvent, error) {
	agentModel, err := s.repo.Agent.GetByID(agentID)
	if err != nil {
		return nil, fmt.Errorf("agent not found: %w", err)
	}

	agentConfig, err := s.GetAgentConfig(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("invalid agent config: %w", err)
	}

	// 获取指定工具
	var selectedTools []tool.BaseTool
	if len(agentConfig.Tools) > 0 {
		selectedTools, err = GetToolsByName(ctx, agentConfig.Tools, s.allTools)
		if err != nil {
			return nil, fmt.Errorf("failed to get tools: %w", err)
		}
	} else {
		selectedTools = s.allTools
	}

	// 创建 eino Agent
	einoAgent, err := s.createAgent(ctx, agentModel.Name, agentModel.Description, agentConfig.SystemPrompt, selectedTools, agentConfig.MaxTokens)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	// 加载历史消息
	var history []*schema.Message
	if req.SessionID != "" {
		history = s.loadHistory(ctx, req.SessionID)
	}

	// 构建输入消息
	messages := buildMessages(history, req.Query)

	// 流式运行 Agent
	iter := einoAgent.Run(ctx, &adk.AgentInput{
		Messages:        messages,
		EnableStreaming: true,
	})

	outCh := make(chan StreamEvent, 10)

	go func() {
		defer close(outCh)

		var fullAnswer string
		for {
			event, ok := iter.Next()
			if !ok {
				outCh <- StreamEvent{Type: "end"}
				break
			}

			if event.Err != nil {
				if event.Err == io.EOF {
					outCh <- StreamEvent{Type: "end"}
					break
				}
				outCh <- StreamEvent{Type: "error", Data: event.Err.Error()}
				continue
			}

			// 处理不同类型的事件
			if event.Output != nil && event.Output.MessageOutput != nil {
				msgVar := event.Output.MessageOutput

				// 流式消息
				if msgVar.IsStreaming && msgVar.MessageStream != nil {
					outCh <- StreamEvent{Type: "start"}

					for {
						chunk, err := msgVar.MessageStream.Recv()
						if err == io.EOF {
							break
						}
						if err != nil {
							outCh <- StreamEvent{Type: "error", Data: err.Error()}
							break
						}

						outCh <- StreamEvent{
							Type: "message",
							Data: chunk.Content,
						}

						// 收集完整答案
						fullAnswer += chunk.Content
					}
				} else if msgVar.Message != nil {
					// 非流式消息
					if msgVar.Role == schema.Assistant {
						outCh <- StreamEvent{
							Type: "message",
							Data: msgVar.Message.Content,
						}
						fullAnswer = msgVar.Message.Content
					} else if msgVar.Role == schema.Tool {
						outCh <- StreamEvent{
							Type:     "tool_call",
							ToolName: msgVar.ToolName,
							Data:     msgVar.Message.Content,
						}
					}
				}
			}

			// 处理 Action
			if event.Action != nil {
				if event.Action.Exit {
					outCh <- StreamEvent{Type: "end"}
					// 结束时保存
					if req.SessionID != "" {
						s.saveMessage(ctx, req.SessionID, "user", req.Query)
						s.saveMessage(ctx, req.SessionID, "assistant", fullAnswer)
					}
					return
				}
				if event.Action.TransferToAgent != nil {
					outCh <- StreamEvent{
						Type:     "transfer",
						ToolName: event.Action.TransferToAgent.DestAgentName,
					}
				}
			}
		}

		// 结束时保存
		if req.SessionID != "" {
			s.saveMessage(ctx, req.SessionID, "user", req.Query)
			s.saveMessage(ctx, req.SessionID, "assistant", fullAnswer)
		}
	}()

	return outCh, nil
}

// loadHistory 从数据库加载历史消息
func (s *Service) loadHistory(ctx context.Context, sessionID string) []*schema.Message {
	messages, err := s.repo.Chat.GetMessagesBySessionID(sessionID)
	if err != nil {
		return nil
	}

	result := make([]*schema.Message, 0, len(messages))
	for _, msg := range messages {
		var role schema.RoleType
		switch msg.Role {
		case "user":
			role = schema.User
		case "assistant":
			role = schema.Assistant
		case "system":
			role = schema.System
		default:
			role = schema.User
		}
		result = append(result, &schema.Message{
			Role:    role,
			Content: msg.Content,
		})
	}
	return result
}

// saveMessage 保存消息到数据库
func (s *Service) saveMessage(ctx context.Context, sessionID, role, content string) {
	_ = s.repo.Chat.CreateMessage(&agentmodel.ChatMessage{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Role:      role,
		Content:   content,
	})
}

// buildMessages 构建消息列表
func buildMessages(history []*schema.Message, query string) []adk.Message {
	result := make([]adk.Message, 0, len(history)+1)
	for _, msg := range history {
		result = append(result, &schema.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}
	result = append(result, &schema.Message{
		Role:    schema.User,
		Content: query,
	})
	return result
}

// RunAgent 运行 Agent（内部方法）
func (s *Service) RunAgent(ctx context.Context, agentID string, query string, history []*schema.Message) (string, error) {
	agentModel, err := s.repo.Agent.GetByID(agentID)
	if err != nil {
		return "", fmt.Errorf("agent not found: %w", err)
	}

	agentConfig, err := s.GetAgentConfig(ctx, agentID)
	if err != nil {
		return "", fmt.Errorf("invalid agent config: %w", err)
	}

	// 获取指定工具
	var selectedTools []tool.BaseTool
	if len(agentConfig.Tools) > 0 {
		selectedTools, err = GetToolsByName(ctx, agentConfig.Tools, s.allTools)
		if err != nil {
			return "", fmt.Errorf("failed to get tools: %w", err)
		}
	} else {
		selectedTools = s.allTools
	}

	// 创建 eino Agent
	einoAgent, err := s.createAgent(ctx, agentModel.Name, agentModel.Description, agentConfig.SystemPrompt, selectedTools, agentConfig.MaxTokens)
	if err != nil {
		return "", fmt.Errorf("failed to create agent: %w", err)
	}

	// 构建输入消息
	messages := buildMessages(history, query)

	// 运行 Agent
	iter := einoAgent.Run(ctx, &adk.AgentInput{
		Messages:        messages,
		EnableStreaming: false,
	})

	// 收集结果
	var result string
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}

		if event.Err != nil {
			if event.Err == io.EOF {
				break
			}
			return "", fmt.Errorf("agent event error: %w", event.Err)
		}

		if event.Output != nil && event.Output.MessageOutput != nil {
			msg, err := event.Output.MessageOutput.GetMessage()
			if err != nil {
				continue
			}
			if msg.Role == schema.Assistant {
				result = msg.Content
			}
		}
	}

	return result, nil
}

// GetToolsByName 根据名称获取工具
func GetToolsByName(ctx context.Context, names []string, allTools []tool.BaseTool) ([]tool.BaseTool, error) {
	if len(names) == 0 {
		return allTools, nil
	}

	toolMap := make(map[string]tool.BaseTool)
	for _, t := range allTools {
		info, err := t.Info(ctx)
		if err != nil {
			continue
		}
		toolMap[info.Name] = t
	}

	result := make([]tool.BaseTool, 0, len(names))
	for _, name := range names {
		t, ok := toolMap[name]
		if !ok {
			return nil, fmt.Errorf("tool not found: %s", name)
		}
		result = append(result, t)
	}

	return result, nil
}

// ListToolNames 列出所有工具名称
func ListToolNames(ctx context.Context, allTools []tool.BaseTool) []string {
	names := make([]string, 0, len(allTools))
	for _, t := range allTools {
		info, err := t.Info(ctx)
		if err != nil {
			continue
		}
		names = append(names, info.Name)
	}
	return names
}
