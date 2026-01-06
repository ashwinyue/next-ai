package agent

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/ashwinyue/next-ai/internal/config"
	agentmodel "github.com/ashwinyue/next-ai/internal/model"
	"github.com/ashwinyue/next-ai/internal/repository"
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
	Description  string   `json:"description"`
	Avatar       string   `json:"avatar,omitempty"`
	AgentMode    string   `json:"agent_mode,omitempty"`     // quick-answer 或 smart-reasoning
	SystemPrompt string   `json:"system_prompt"`
	Tools        []string `json:"tools"`
	MaxIter      int      `json:"max_iterations"`
	Temperature  float64  `json:"temperature,omitempty"`
	Model        string   `json:"model"`
	KnowledgeIDs []string `json:"knowledge_ids,omitempty"`
}

// CreateAgent 创建 Agent
func (s *Service) CreateAgent(ctx context.Context, req *CreateAgentRequest) (*agentmodel.Agent, error) {
	if _, err := s.repo.Agent.GetByName(req.Name); err == nil {
		return nil, fmt.Errorf("agent name already exists")
	}

	// 默认模式为 quick-answer
	agentMode := req.AgentMode
	if agentMode == "" {
		agentMode = agentmodel.AgentModeQuickAnswer
	}

	// 验证模式
	if agentMode != agentmodel.AgentModeQuickAnswer && agentMode != agentmodel.AgentModeSmartReasoning {
		return nil, fmt.Errorf("invalid agent_mode: %s, must be 'quick-answer' or 'smart-reasoning'", agentMode)
	}

	// 构建 Tools JSON
	toolsJSON := make(agentmodel.JSON)
	if len(req.Tools) > 0 {
		for _, tool := range req.Tools {
			toolsJSON[tool] = true
		}
	}

	// 构建 ModelConfig
	modelConfig := agentmodel.ModelConfig{
		Provider: s.cfg.AI.Provider,
		Model:    req.Model,
	}
	if modelConfig.Model == "" {
		modelConfig.Model = s.cfg.AI.OpenAI.Model
	}

	agent := &agentmodel.Agent{
		ID:           uuid.New().String(),
		Name:         req.Name,
		Description:  req.Description,
		Avatar:       req.Avatar,
		IsBuiltin:    false,
		AgentMode:    agentMode,
		SystemPrompt: req.SystemPrompt,
		ModelConfig:  modelConfig,
		Tools:        toolsJSON,
		MaxIter:      req.MaxIter,
		Temperature:  req.Temperature,
		KnowledgeIDs: req.KnowledgeIDs,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
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

	// 内置 Agent 不允许修改核心配置
	if agentModel.IsBuiltin {
		return nil, fmt.Errorf("builtin agent cannot be updated")
	}

	agentModel.Name = req.Name
	agentModel.Description = req.Description
	agentModel.Avatar = req.Avatar
	agentModel.SystemPrompt = req.SystemPrompt
	agentModel.MaxIter = req.MaxIter
	agentModel.Temperature = req.Temperature
	agentModel.UpdatedAt = time.Now()

	// 更新 AgentMode
	if req.AgentMode != "" {
		if req.AgentMode != agentmodel.AgentModeQuickAnswer && req.AgentMode != agentmodel.AgentModeSmartReasoning {
			return nil, fmt.Errorf("invalid agent_mode: %s", req.AgentMode)
		}
		agentModel.AgentMode = req.AgentMode
	}

	// 更新 Tools
	toolsJSON := make(agentmodel.JSON)
	if len(req.Tools) > 0 {
		for _, tool := range req.Tools {
			toolsJSON[tool] = true
		}
	}
	agentModel.Tools = toolsJSON

	// 更新 KnowledgeIDs
	agentModel.KnowledgeIDs = req.KnowledgeIDs

	// 更新 ModelConfig
	if req.Model != "" {
		agentModel.ModelConfig.Model = req.Model
	}

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

// CopyAgent 复制 Agent
func (s *Service) CopyAgent(ctx context.Context, id string) (*agentmodel.Agent, error) {
	sourceAgent, err := s.repo.Agent.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("agent not found: %w", err)
	}

	// 复制配置，生成新 ID
	newAgent := &agentmodel.Agent{
		ID:           uuid.New().String(),
		Name:         sourceAgent.Name + " (副本)",
		Description:  sourceAgent.Description,
		Avatar:       sourceAgent.Avatar,
		IsBuiltin:    false, // 复制的 Agent 不是内置的
		AgentMode:    sourceAgent.AgentMode,
		SystemPrompt: sourceAgent.SystemPrompt,
		ModelConfig:  sourceAgent.ModelConfig,
		Tools:        sourceAgent.Tools,
		MaxIter:      sourceAgent.MaxIter,
		Temperature:  sourceAgent.Temperature,
		KnowledgeIDs: sourceAgent.KnowledgeIDs,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.repo.Agent.Create(newAgent); err != nil {
		return nil, fmt.Errorf("failed to create copied agent: %w", err)
	}

	return newAgent, nil
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
func (s *Service) newToolCallingChatModel(ctx context.Context, modelConfig agentmodel.ModelConfig) (model.ToolCallingChatModel, error) {
	var apiKey, baseURL, modelName string

	// 从 modelConfig 获取配置
	if modelConfig.APIKey != "" {
		apiKey = modelConfig.APIKey
	}
	if modelConfig.BaseURL != "" {
		baseURL = modelConfig.BaseURL
	}
	if modelConfig.Model != "" {
		modelName = modelConfig.Model
	}

	// 如果没有提供，使用全局配置
	if apiKey == "" || modelName == "" {
		aiCfg := s.cfg.AI
		switch aiCfg.Provider {
		case "openai":
			if apiKey == "" {
				apiKey = aiCfg.OpenAI.APIKey
			}
			if baseURL == "" {
				baseURL = aiCfg.OpenAI.BaseURL
			}
			if modelName == "" {
				modelName = aiCfg.OpenAI.Model
			}
		case "alibaba", "qwen", "dashscope":
			if apiKey == "" {
				apiKey = aiCfg.Alibaba.AccessKeySecret
			}
			if baseURL == "" {
				baseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
			}
			if modelName == "" {
				modelName = aiCfg.Alibaba.Model
			}
		case "deepseek":
			if apiKey == "" {
				apiKey = aiCfg.DeepSeek.APIKey
			}
			if baseURL == "" {
				baseURL = aiCfg.DeepSeek.BaseURL
			}
			if modelName == "" {
				modelName = aiCfg.DeepSeek.Model
			}
		default:
			if apiKey == "" {
				apiKey = aiCfg.OpenAI.APIKey
			}
			if baseURL == "" {
				baseURL = aiCfg.OpenAI.BaseURL
			}
			if modelName == "" {
				modelName = aiCfg.OpenAI.Model
			}
		}
	}

	if apiKey == "" {
		return nil, fmt.Errorf("api_key is required")
	}

	if modelName == "" {
		modelName = "gpt-4o-mini"
	}

	temperature := float32(0.7)
	if temp, ok := modelConfig.Parameters["temperature"].(float64); ok {
		temperature = float32(temp)
	}

	return openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:      apiKey,
		BaseURL:     baseURL,
		Model:       modelName,
		Temperature: &temperature,
	})
}

// createAgent 创建 eino Agent
func (s *Service) createAgent(ctx context.Context, agentModel *agentmodel.Agent, selectedTools []tool.BaseTool) (*adk.ChatModelAgent, error) {
	// 根据 AgentMode 选择不同的实现
	switch agentModel.AgentMode {
	case agentmodel.AgentModeSmartReasoning:
		// React Agent 模式
		return s.createReactAgent(ctx, agentModel, selectedTools)
	default:
		// Quick-answer 模式（默认）
		return s.createChatModelAgent(ctx, agentModel, selectedTools)
	}
}

// createChatModelAgent 创建 ChatModel Agent（quick-answer 模式）
func (s *Service) createChatModelAgent(ctx context.Context, agentModel *agentmodel.Agent, selectedTools []tool.BaseTool) (*adk.ChatModelAgent, error) {
	chatModel, err := s.newToolCallingChatModel(ctx, agentModel.ModelConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat model: %w", err)
	}

	maxIter := agentModel.MaxIter
	if maxIter <= 0 {
		maxIter = 10
	}

	agentCfg := &adk.ChatModelAgentConfig{
		Name:          agentModel.Name,
		Description:   agentModel.Description,
		Instruction:   agentModel.SystemPrompt,
		Model:         chatModel,
		MaxIterations: maxIter,
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

// createReactAgent 创建 React Agent（smart-reasoning 模式）
// 参考 eino-examples/flow/agent/react/react.go
func (s *Service) createReactAgent(ctx context.Context, agentModel *agentmodel.Agent, selectedTools []tool.BaseTool) (*adk.ChatModelAgent, error) {
	chatModel, err := s.newToolCallingChatModel(ctx, agentModel.ModelConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat model: %w", err)
	}

	maxIter := agentModel.MaxIter
	if maxIter <= 0 {
		maxIter = 10
	}

	// 构建 system prompt
	systemPrompt := agentModel.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = "你是一个有用的助手，可以使用工具来帮助用户。"
	}

	// 使用 adk.NewChatModelAgent，它在底层支持 ReAct 模式
	// React Agent 本质上是一个支持工具调用的 ChatModel Agent
	agentCfg := &adk.ChatModelAgentConfig{
		Name:          agentModel.Name,
		Description:   agentModel.Description,
		Instruction:   systemPrompt,
		Model:         chatModel,
		MaxIterations: maxIter,
	}

	// 添加工具
	if len(selectedTools) > 0 {
		agentCfg.ToolsConfig = adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: selectedTools,
			},
		}
	}

	// 注意：adk.NewChatModelAgent 已经支持 ReAct 模式的工具调用循环
	// 如果需要更底层的 React Agent 控制，可以使用 react.NewAgent
	return adk.NewChatModelAgent(ctx, agentCfg)
}

// getToolNames 从 Agent.Tools 获取工具名称列表
func getToolNames(tools agentmodel.JSON) []string {
	var names []string
	for k := range tools {
		names = append(names, k)
	}
	return names
}

// Run 运行 Agent（同步）
func (s *Service) Run(ctx context.Context, agentID string, req *RunRequest) (*RunResponse, error) {
	agentModel, err := s.repo.Agent.GetByID(agentID)
	if err != nil {
		return nil, fmt.Errorf("agent not found: %w", err)
	}

	// 获取指定工具
	toolNames := getToolNames(agentModel.Tools)
	selectedTools, err := GetToolsByName(ctx, toolNames, s.allTools)
	if err != nil {
		// 如果获取工具失败，使用所有工具
		selectedTools = s.allTools
	}

	// 创建 eino Agent
	einoAgent, err := s.createAgent(ctx, agentModel, selectedTools)
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

	// 获取指定工具
	toolNames := getToolNames(agentModel.Tools)
	selectedTools, err := GetToolsByName(ctx, toolNames, s.allTools)
	if err != nil {
		selectedTools = s.allTools
	}

	// 创建 eino Agent
	einoAgent, err := s.createAgent(ctx, agentModel, selectedTools)
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

	// 获取指定工具
	toolNames := getToolNames(agentModel.Tools)
	selectedTools, err := GetToolsByName(ctx, toolNames, s.allTools)
	if err != nil {
		selectedTools = s.allTools
	}

	// 创建 eino Agent
	einoAgent, err := s.createAgent(ctx, agentModel, selectedTools)
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

// ========== 内置 Agent 初始化 ==========

// builtinAgentConfig 内置 Agent 配置模板
type builtinAgentConfig struct {
	ID           string
	Name         string
	Description  string
	Avatar       string
	AgentMode    string
	SystemPrompt string
	ToolNames    []string
	MaxIter      int
	Temperature  float64
}

// getBuiltinAgents 获取所有内置 Agent 配置
func getBuiltinAgents() []builtinAgentConfig {
	return []builtinAgentConfig{
		{
			ID:          agentmodel.BuiltinQuickAnswerID,
			Name:        "快速问答",
			Description: "基于知识库的快速问答助手，适合直接检索回答问题",
			Avatar:      "⚡",
			AgentMode:   agentmodel.AgentModeQuickAnswer,
			SystemPrompt: "你是一个专业的知识库问答助手。请根据检索到的知识库内容回答用户问题。如果知识库中没有相关信息，请诚实告知用户。",
			ToolNames:   []string{"knowledge_search", "list_chunks"},
			MaxIter:     5,
			Temperature: 0.3,
		},
		{
			ID:          agentmodel.BuiltinSmartReasoningID,
			Name:        "智能推理",
			Description: "具备多步推理能力的助手，可以使用多种工具分析问题",
			Avatar:      "🧠",
			AgentMode:   agentmodel.AgentModeSmartReasoning,
			SystemPrompt: "你是一个具备强大推理能力的助手。面对复杂问题时，你可以：\n1. 使用网络搜索获取最新信息\n2. 检索知识库获取专业内容\n3. 使用思考工具进行逻辑分析\n\n请按步骤推理，给出准确的答案。",
			ToolNames:   []string{"web_search", "knowledge_search", "list_chunks", "todo_write"},
			MaxIter:     15,
			Temperature: 0.7,
		},
		{
			ID:          "builtin-deep-researcher",
			Name:        "深度研究",
			Description: "擅长深入研究复杂主题的助手",
			Avatar:      "🔬",
			AgentMode:   agentmodel.AgentModeSmartReasoning,
			SystemPrompt: "你是一个专业的研究助手。面对研究主题时，请：\n1. 先使用 todo_write 创建研究计划\n2. 使用网络搜索获取多来源信息\n3. 使用 grep_chunks 在文档中查找细节\n4. 综合分析得出结论",
			ToolNames:   []string{"web_search", "knowledge_search", "grep_chunks", "list_chunks", "todo_write", "thinking"},
			MaxIter:     20,
			Temperature: 0.5,
		},
		{
			ID:          "builtin-data-analyst",
			Name:        "数据分析",
			Description: "专业的数据分析助手，可以查询和分析数据",
			Avatar:      "📊",
			AgentMode:   agentmodel.AgentModeQuickAnswer,
			SystemPrompt: "你是一个数据分析助手。你可以使用数据库查询工具来获取和分析数据。请根据用户需求提供清晰的数据分析结果。",
			ToolNames:   []string{"database_query", "data_analysis", "data_schema"},
			MaxIter:     10,
			Temperature: 0.3,
		},
		{
			ID:          "builtin-knowledge-graph-expert",
			Name:        "知识图谱专家",
			Description: "专注于知识关系和图谱分析的助手",
			Avatar:      "🕸️",
			AgentMode:   agentmodel.AgentModeSmartReasoning,
			SystemPrompt: "你是一个知识图谱分析专家。请帮助用户理解实体之间的关系，分析知识图谱中的连接。",
			ToolNames:   []string{"knowledge_search", "grep_chunks", "list_chunks"},
			MaxIter:     10,
			Temperature: 0.5,
		},
		{
			ID:          "builtin-document-assistant",
			Name:        "文档助手",
			Description: "专业的文档分析和处理助手",
			Avatar:      "📄",
			AgentMode:   agentmodel.AgentModeQuickAnswer,
			SystemPrompt: "你是一个文档助手。你可以帮助用户搜索、分析文档内容，提取关键信息，解答文档相关问题。",
			ToolNames:   []string{"knowledge_search", "list_chunks", "grep_chunks", "get_document_info"},
			MaxIter:     8,
			Temperature: 0.4,
		},
	}
}

// InitBuiltinAgents 初始化内置 Agent
// 如果内置 Agent 不存在，则创建它们；如果存在但配置不同，则更新它们
func (s *Service) InitBuiltinAgents(ctx context.Context) error {
	configs := getBuiltinAgents()

	for _, cfg := range configs {
		// 检查是否已存在
		existingAgent, err := s.repo.Agent.GetByID(cfg.ID)
		if err != nil {
			// 不存在，创建新的
			newAgent := &agentmodel.Agent{
				ID:           cfg.ID,
				Name:         cfg.Name,
				Description:  cfg.Description,
				Avatar:       cfg.Avatar,
				IsBuiltin:    true,
				AgentMode:    cfg.AgentMode,
				SystemPrompt: cfg.SystemPrompt,
				ModelConfig: agentmodel.ModelConfig{
					Provider: s.cfg.AI.Provider,
					Model:    s.cfg.AI.OpenAI.Model,
				},
				Tools:       make(agentmodel.JSON),
				MaxIter:     cfg.MaxIter,
				Temperature: cfg.Temperature,
				IsActive:    true,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}

			// 设置工具
			for _, toolName := range cfg.ToolNames {
				newAgent.Tools[toolName] = true
			}

			if err := s.repo.Agent.Create(newAgent); err != nil {
				return fmt.Errorf("failed to create builtin agent %s: %w", cfg.Name, err)
			}
		} else {
			// 已存在，检查是否需要更新（仅允许更新非核心字段）
			updated := false
			if existingAgent.Avatar != cfg.Avatar {
				existingAgent.Avatar = cfg.Avatar
				updated = true
			}
			if existingAgent.Description != cfg.Description {
				existingAgent.Description = cfg.Description
				updated = true
			}
			if existingAgent.AgentMode != cfg.AgentMode {
				existingAgent.AgentMode = cfg.AgentMode
				updated = true
			}
			if existingAgent.SystemPrompt != cfg.SystemPrompt {
				existingAgent.SystemPrompt = cfg.SystemPrompt
				updated = true
			}
			if existingAgent.MaxIter != cfg.MaxIter {
				existingAgent.MaxIter = cfg.MaxIter
				updated = true
			}
			if existingAgent.Temperature != cfg.Temperature {
				existingAgent.Temperature = cfg.Temperature
				updated = true
			}

			// 确保是内置标识
			if !existingAgent.IsBuiltin {
				existingAgent.IsBuiltin = true
				updated = true
			}

			if updated {
				existingAgent.UpdatedAt = time.Now()
				if err := s.repo.Agent.Update(existingAgent); err != nil {
					return fmt.Errorf("failed to update builtin agent %s: %w", cfg.Name, err)
				}
			}
		}
	}

	return nil
}

// ListBuiltinAgents 列出内置 Agent
func (s *Service) ListBuiltinAgents(ctx context.Context) ([]*agentmodel.Agent, error) {
	allAgents, err := s.repo.Agent.ListActive()
	if err != nil {
		return nil, err
	}

	var builtinAgents []*agentmodel.Agent
	for _, agent := range allAgents {
		if agent.IsBuiltin {
			builtinAgents = append(builtinAgents, agent)
		}
	}

	return builtinAgents, nil
}
