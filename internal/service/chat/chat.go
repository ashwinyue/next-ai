package chat

import (
	"context"
	"fmt"
	"strings"

	"github.com/ashwinyue/next-ai/internal/model"
	"github.com/ashwinyue/next-ai/internal/repository"
	ecomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
)

// Service 聊天服务
type Service struct {
	repo      *repository.Repositories
	chatModel ecomodel.ChatModel
}

// NewService 创建聊天服务
func NewService(repo *repository.Repositories, chatModel ecomodel.ChatModel) *Service {
	return &Service{
		repo:      repo,
		chatModel: chatModel,
	}
}

// CreateSessionRequest 创建会话请求
type CreateSessionRequest struct {
	Title   string `json:"title"`
	UserID  string `json:"user_id"`
	AgentID string `json:"agent_id"`
}

// CreateSession 创建会话
func (s *Service) CreateSession(ctx context.Context, req *CreateSessionRequest) (*model.ChatSession, error) {
	session := &model.ChatSession{
		ID:      uuid.New().String(),
		UserID:  req.UserID,
		AgentID: req.AgentID,
		Title:   req.Title,
		Status:  "active",
	}

	if err := s.repo.Chat.CreateSession(session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

// GetSession 获取会话
func (s *Service) GetSession(ctx context.Context, id string) (*model.ChatSession, error) {
	return s.repo.Chat.GetSessionByID(id)
}

// ListSessionsRequest 列出会话请求
type ListSessionsRequest struct {
	UserID string `json:"user_id"`
	Page   int    `json:"page"`
	Size   int    `json:"size"`
}

// ListSessions 列出会话
func (s *Service) ListSessions(ctx context.Context, req *ListSessionsRequest) ([]*model.ChatSession, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 || req.Size > 100 {
		req.Size = 20
	}

	offset := (req.Page - 1) * req.Size

	sessions, err := s.repo.Chat.ListSessions(req.UserID, offset, req.Size)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list sessions: %w", err)
	}

	total := int64(len(sessions))
	return sessions, total, nil
}

// UpdateSession 更新会话
func (s *Service) UpdateSession(ctx context.Context, id string, req *CreateSessionRequest) (*model.ChatSession, error) {
	session, err := s.repo.Chat.GetSessionByID(id)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	if req.Title != "" {
		session.Title = req.Title
	}
	if req.AgentID != "" {
		session.AgentID = req.AgentID
	}

	if err := s.repo.Chat.UpdateSession(session); err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	return session, nil
}

// DeleteSession 删除会话
func (s *Service) DeleteSession(ctx context.Context, id string) error {
	if err := s.repo.Chat.DeleteSession(id); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

// SendMessageRequest 发送消息请求
type SendMessageRequest struct {
	Role    string `json:"role" binding:"required"`
	Content string `json:"content" binding:"required"`
}

// SendMessage 发送消息
func (s *Service) SendMessage(ctx context.Context, sessionID string, req *SendMessageRequest) (*model.ChatMessage, error) {
	_, err := s.repo.Chat.GetSessionByID(sessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	message := &model.ChatMessage{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Role:      req.Role,
		Content:   req.Content,
	}

	if err := s.repo.Chat.CreateMessage(message); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	return message, nil
}

// GetMessages 获取会话消息
func (s *Service) GetMessages(ctx context.Context, sessionID string) ([]*model.ChatMessage, error) {
	return s.repo.Chat.GetMessagesBySessionID(sessionID)
}

// LoadMessagesRequest 加载消息请求
type LoadMessagesRequest struct {
	Limit      int    `json:"limit"`
	BeforeTime string `json:"before_time"` // RFC3339Nano 格式
}

// LoadMessages 加载消息历史（支持分页和时间筛选）
func (s *Service) LoadMessages(ctx context.Context, sessionID string, req *LoadMessagesRequest) ([]*model.ChatMessage, error) {
	// 验证会话是否存在
	_, err := s.repo.Chat.GetSessionByID(sessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	// 设置默认 limit
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	var messages []*model.ChatMessage

	// 如果没有指定 beforeTime，获取最近的 N 条消息
	if req.BeforeTime == "" {
		messages, err = s.repo.Chat.GetRecentMessagesBySession(sessionID, req.Limit)
	} else {
		// 获取指定时间之前的消息
		messages, err = s.repo.Chat.GetMessagesBySessionBeforeTime(sessionID, req.BeforeTime, req.Limit)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load messages: %w", err)
	}

	return messages, nil
}

// GetMessage 获取单条消息
func (s *Service) GetMessage(ctx context.Context, messageID string) (*model.ChatMessage, error) {
	message, err := s.repo.Chat.GetMessageByID(messageID)
	if err != nil {
		return nil, fmt.Errorf("message not found: %w", err)
	}
	return message, nil
}

// DeleteMessage 删除消息
func (s *Service) DeleteMessage(ctx context.Context, sessionID, messageID string) error {
	// 验证消息是否属于该会话
	message, err := s.repo.Chat.GetMessageByID(messageID)
	if err != nil {
		return fmt.Errorf("message not found: %w", err)
	}

	if message.SessionID != sessionID {
		return fmt.Errorf("message does not belong to this session")
	}

	if err := s.repo.Chat.DeleteMessage(messageID); err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	return nil
}

// ========== 会话标题生成 ==========

// GenerateTitleRequest 生成标题请求
type GenerateTitleRequest struct {
	FirstMessage string `json:"first_message"`
}

// GenerateTitle 生成会话标题
// 根据首条用户消息内容，使用 LLM 自动生成简短的会话标题
func (s *Service) GenerateTitle(ctx context.Context, sessionID string, req *GenerateTitleRequest) (string, error) {
	// 检查会话是否存在
	session, err := s.repo.Chat.GetSessionByID(sessionID)
	if err != nil {
		return "", fmt.Errorf("session not found: %w", err)
	}

	// 如果已有标题，直接返回
	if session.Title != "" {
		return session.Title, nil
	}

	// 如果没有 ChatModel，返回默认标题
	if s.chatModel == nil {
		defaultTitle := generateDefaultTitle(req.FirstMessage)
		session.Title = defaultTitle
		if err := s.repo.Chat.UpdateSession(session); err != nil {
			return "", fmt.Errorf("failed to update session title: %w", err)
		}
		return defaultTitle, nil
	}

	// 使用 ChatModel 生成标题
	prompt := buildTitlePrompt(req.FirstMessage)

	messages := []*schema.Message{
		{Role: schema.System, Content: "You are a helpful assistant that generates concise conversation titles."},
		{Role: schema.User, Content: prompt},
	}

	response, err := s.chatModel.Generate(ctx, messages)
	if err != nil {
		// 降级到默认标题
		defaultTitle := generateDefaultTitle(req.FirstMessage)
		session.Title = defaultTitle
		if updateErr := s.repo.Chat.UpdateSession(session); updateErr != nil {
			return "", fmt.Errorf("failed to update session title: %w", updateErr)
		}
		return defaultTitle, nil
	}

	title := strings.TrimSpace(response.Content)
	// 移除可能的前缀
	title = strings.TrimPrefix(title, "Title: ")
	title = strings.TrimPrefix(title, "标题: ")
	title = strings.TrimPrefix(title, "会话标题: ")
	title = strings.TrimSpace(title)

	// 如果标题为空，使用默认标题
	if title == "" {
		title = generateDefaultTitle(req.FirstMessage)
	}

	// 限制标题长度
	if len(title) > 50 {
		title = title[:50] + "..."
	}

	// 更新会话标题
	session.Title = title
	if err := s.repo.Chat.UpdateSession(session); err != nil {
		return "", fmt.Errorf("failed to update session title: %w", err)
	}

	return title, nil
}

// buildTitlePrompt 构建标题生成提示词
func buildTitlePrompt(firstMessage string) string {
	return fmt.Sprintf(`Generate a short, concise title (max 10 words) for this conversation.

First message: %s

Requirements:
- Use the same language as the first message (Chinese or English)
- Keep it brief and descriptive
- Only output the title, no explanations

Title:`, truncateMessage(firstMessage, 200))
}

// generateDefaultTitle 生成默认标题（无 LLM 时的降级方案）
func generateDefaultTitle(message string) string {
	message = strings.TrimSpace(message)
	if len(message) > 20 {
		return message[:20] + "..."
	}
	if message == "" {
		return "新对话"
	}
	return message
}

// truncateMessage 截断消息
func truncateMessage(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// ========== Agent 聊天集成 ==========

// AgentService Agent 服务接口
// 使用 any 类型避免循环依赖
type AgentService interface {
	StreamWithContext(ctx context.Context, agentID string, req interface{}) (<-chan interface{}, error)
}

// ServiceWithAgent 带 Agent 集成的聊天服务
// 注意：这里使用接口避免循环依赖，实际注入时传入 agent.Service
type ServiceWithAgent struct {
	*Service
	agentSvc AgentService
}

// NewServiceWithAgent 创建带 Agent 集成的聊天服务
func NewServiceWithAgent(chatSvc *Service, agentSvc AgentService) *ServiceWithAgent {
	return &ServiceWithAgent{
		Service:  chatSvc,
		agentSvc: agentSvc,
	}
}

// StreamEvent 流式事件
type StreamEvent struct {
	Type     string `json:"type"` // start, message, tool_call, error, end
	Data     string `json:"data"`
	ToolName string `json:"tool_name,omitempty"`
}

// AgentChatRequest Agent 聊天请求
type AgentChatRequest struct {
	SessionID string `json:"session_id"`
	AgentID   string `json:"agent_id"`
	Query     string `json:"query"`
}

// AgentChat 调用 Agent 进行聊天（流式）
// 兼容 WeKnora API: POST /api/v1/agent-chat/:session_id
func (s *ServiceWithAgent) AgentChat(ctx context.Context, req *AgentChatRequest) (<-chan StreamEvent, error) {
	// 验证会话是否存在
	session, err := s.Service.GetSession(ctx, req.SessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	// 使用会话的 Agent ID（如果请求未指定）
	agentID := req.AgentID
	if agentID == "" && session.AgentID != "" {
		agentID = session.AgentID
	}

	// 构建运行时请求
	runReq := map[string]interface{}{
		"query":      req.Query,
		"session_id": req.SessionID,
	}

	// 调用 Agent 流式执行
	rawCh, err := s.agentSvc.StreamWithContext(ctx, agentID, runReq)
	if err != nil {
		return nil, fmt.Errorf("failed to stream agent: %w", err)
	}

	// 转换事件格式
	outCh := make(chan StreamEvent, 10)
	go func() {
		defer close(outCh)
		for evt := range rawCh {
			switch v := evt.(type) {
			case map[string]interface{}:
				evtType, _ := v["type"].(string)
				data, _ := v["data"].(string)
				toolName, _ := v["tool_name"].(string)
				outCh <- StreamEvent{
					Type:     evtType,
					Data:     data,
					ToolName: toolName,
				}
			case StreamEvent:
				outCh <- v
			}
		}
	}()

	return outCh, nil
}
