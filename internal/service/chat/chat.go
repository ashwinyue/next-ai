package chat

import (
	"context"
	"fmt"

	"github.com/ashwinyue/next-rag/next-ai/internal/model"
	"github.com/ashwinyue/next-rag/next-ai/internal/repository"
	"github.com/google/uuid"
)

// Service 聊天服务
type Service struct {
	repo *repository.Repositories
}

// NewService 创建聊天服务
func NewService(repo *repository.Repositories) *Service {
	return &Service{repo: repo}
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
