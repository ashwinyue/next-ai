// Package chat 提供 Chat 服务单元测试
package chat

import (
	"context"
	"errors"
	"testing"

	"github.com/ashwinyue/next-ai/internal/model"
	"github.com/google/uuid"
)

// mockChatRepository Mock Chat Repository
type mockChatRepository struct {
	sessions       map[string]*model.ChatSession
	messages       map[string][]*model.ChatMessage
	createError    error
	getError       error
	updateError    error
	deleteError    error
	createMsgError error
}

func newMockChatRepo() *mockChatRepository {
	return &mockChatRepository{
		sessions: make(map[string]*model.ChatSession),
		messages: make(map[string][]*model.ChatMessage),
	}
}

func (m *mockChatRepository) CreateSession(session *model.ChatSession) error {
	if m.createError != nil {
		return m.createError
	}
	m.sessions[session.ID] = session
	return nil
}

func (m *mockChatRepository) GetSessionByID(id string) (*model.ChatSession, error) {
	if m.getError != nil {
		return nil, m.getError
	}
	if session, ok := m.sessions[id]; ok {
		return session, nil
	}
	return nil, errors.New("session not found")
}

func (m *mockChatRepository) ListSessions(userID string, offset, limit int) ([]*model.ChatSession, error) {
	result := make([]*model.ChatSession, 0)
	for _, session := range m.sessions {
		if session.UserID == userID {
			result = append(result, session)
		}
	}
	return result, nil
}

func (m *mockChatRepository) UpdateSession(session *model.ChatSession) error {
	if m.updateError != nil {
		return m.updateError
	}
	if _, ok := m.sessions[session.ID]; !ok {
		return errors.New("session not found")
	}
	m.sessions[session.ID] = session
	return nil
}

func (m *mockChatRepository) DeleteSession(id string) error {
	if m.deleteError != nil {
		return m.deleteError
	}
	delete(m.sessions, id)
	return nil
}

func (m *mockChatRepository) CreateMessage(message *model.ChatMessage) error {
	if m.createMsgError != nil {
		return m.createMsgError
	}
	if m.messages[message.SessionID] == nil {
		m.messages[message.SessionID] = []*model.ChatMessage{}
	}
	m.messages[message.SessionID] = append(m.messages[message.SessionID], message)
	return nil
}

func (m *mockChatRepository) GetMessagesBySessionID(sessionID string) ([]*model.ChatMessage, error) {
	if messages, ok := m.messages[sessionID]; ok {
		return messages, nil
	}
	return []*model.ChatMessage{}, nil
}

func (m *mockChatRepository) GetMessageByID(id string) (*model.ChatMessage, error) {
	for _, messages := range m.messages {
		for _, msg := range messages {
			if msg.ID == id {
				return msg, nil
			}
		}
	}
	return nil, errors.New("message not found")
}

func (m *mockChatRepository) DeleteMessage(id string) error {
	for sessionID, messages := range m.messages {
		for i, msg := range messages {
			if msg.ID == id {
				m.messages[sessionID] = append(messages[:i], messages[i+1:]...)
				return nil
			}
		}
	}
	return errors.New("message not found")
}

func (m *mockChatRepository) GetRecentMessagesBySession(sessionID string, limit int) ([]*model.ChatMessage, error) {
	if messages, ok := m.messages[sessionID]; ok {
		if len(messages) > limit {
			return messages[len(messages)-limit:], nil
		}
		return messages, nil
	}
	return []*model.ChatMessage{}, nil
}

func (m *mockChatRepository) GetMessagesBySessionBeforeTime(sessionID, beforeTime string, limit int) ([]*model.ChatMessage, error) {
	return m.GetRecentMessagesBySession(sessionID, limit)
}

// ChatRepository Chat Repository 接口（用于测试）
type ChatRepository interface {
	CreateSession(session *model.ChatSession) error
	GetSessionByID(id string) (*model.ChatSession, error)
	ListSessions(userID string, offset, limit int) ([]*model.ChatSession, error)
	UpdateSession(session *model.ChatSession) error
	DeleteSession(id string) error
	CreateMessage(message *model.ChatMessage) error
	GetMessagesBySessionID(sessionID string) ([]*model.ChatMessage, error)
	GetMessageByID(id string) (*model.ChatMessage, error)
	DeleteMessage(id string) error
	GetRecentMessagesBySession(sessionID string, limit int) ([]*model.ChatMessage, error)
	GetMessagesBySessionBeforeTime(sessionID, beforeTime string, limit int) ([]*model.ChatMessage, error)
}

// testService 测试服务包装器
type testService struct {
	repo ChatRepository
}

func newTestService(repo ChatRepository) *testService {
	return &testService{repo: repo}
}

func (s *testService) CreateSession(ctx context.Context, req *CreateSessionRequest) (*model.ChatSession, error) {
	session := &model.ChatSession{
		ID:      uuid.New().String(),
		UserID:  req.UserID,
		AgentID: req.AgentID,
		Title:   req.Title,
		Status:  "active",
	}

	if err := s.repo.CreateSession(session); err != nil {
		return nil, errors.New("failed to create session: " + err.Error())
	}

	return session, nil
}

func (s *testService) GetSession(ctx context.Context, id string) (*model.ChatSession, error) {
	return s.repo.GetSessionByID(id)
}

func (s *testService) ListSessions(ctx context.Context, req *ListSessionsRequest) ([]*model.ChatSession, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 || req.Size > 100 {
		req.Size = 20
	}

	offset := (req.Page - 1) * req.Size

	sessions, err := s.repo.ListSessions(req.UserID, offset, req.Size)
	if err != nil {
		return nil, 0, errors.New("failed to list sessions: " + err.Error())
	}

	total := int64(len(sessions))
	return sessions, total, nil
}

func (s *testService) UpdateSession(ctx context.Context, id string, req *CreateSessionRequest) (*model.ChatSession, error) {
	session, err := s.repo.GetSessionByID(id)
	if err != nil {
		return nil, errors.New("session not found: " + err.Error())
	}

	if req.Title != "" {
		session.Title = req.Title
	}
	if req.AgentID != "" {
		session.AgentID = req.AgentID
	}

	if err := s.repo.UpdateSession(session); err != nil {
		return nil, errors.New("failed to update session: " + err.Error())
	}

	return session, nil
}

func (s *testService) DeleteSession(ctx context.Context, id string) error {
	if err := s.repo.DeleteSession(id); err != nil {
		return errors.New("failed to delete session: " + err.Error())
	}
	return nil
}

func (s *testService) SendMessage(ctx context.Context, sessionID string, req *SendMessageRequest) (*model.ChatMessage, error) {
	_, err := s.repo.GetSessionByID(sessionID)
	if err != nil {
		return nil, errors.New("session not found: " + err.Error())
	}

	message := &model.ChatMessage{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Role:      req.Role,
		Content:   req.Content,
	}

	if err := s.repo.CreateMessage(message); err != nil {
		return nil, errors.New("failed to create message: " + err.Error())
	}

	return message, nil
}

func (s *testService) GetMessages(ctx context.Context, sessionID string) ([]*model.ChatMessage, error) {
	return s.repo.GetMessagesBySessionID(sessionID)
}

func (s *testService) LoadMessages(ctx context.Context, sessionID string, req *LoadMessagesRequest) ([]*model.ChatMessage, error) {
	_, err := s.repo.GetSessionByID(sessionID)
	if err != nil {
		return nil, errors.New("session not found: " + err.Error())
	}

	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	var messages []*model.ChatMessage
	if req.BeforeTime == "" {
		messages, err = s.repo.GetRecentMessagesBySession(sessionID, req.Limit)
	} else {
		messages, err = s.repo.GetMessagesBySessionBeforeTime(sessionID, req.BeforeTime, req.Limit)
	}

	if err != nil {
		return nil, errors.New("failed to load messages: " + err.Error())
	}

	return messages, nil
}

func (s *testService) GetMessage(ctx context.Context, messageID string) (*model.ChatMessage, error) {
	message, err := s.repo.GetMessageByID(messageID)
	if err != nil {
		return nil, errors.New("message not found: " + err.Error())
	}
	return message, nil
}

func (s *testService) DeleteMessage(ctx context.Context, sessionID, messageID string) error {
	message, err := s.repo.GetMessageByID(messageID)
	if err != nil {
		return errors.New("message not found: " + err.Error())
	}

	if message.SessionID != sessionID {
		return errors.New("message does not belong to this session")
	}

	if err := s.repo.DeleteMessage(messageID); err != nil {
		return errors.New("failed to delete message: " + err.Error())
	}

	return nil
}

// ========== 测试用例 ==========

func TestCreateSession(t *testing.T) {
	tests := []struct {
		name        string
		req         *CreateSessionRequest
		setupRepo   func(*mockChatRepository)
		wantErr     bool
		errContains string
	}{
		{
			name: "create session with valid data",
			req: &CreateSessionRequest{
				UserID:  "user-123",
				AgentID: "agent-456",
				Title:   "Test Session",
			},
			setupRepo: func(repo *mockChatRepository) {},
			wantErr:   false,
		},
		{
			name: "create session with minimal data",
			req: &CreateSessionRequest{
				UserID: "user-123",
			},
			setupRepo: func(repo *mockChatRepository) {},
			wantErr:   false,
		},
		{
			name: "create session with repository error",
			req: &CreateSessionRequest{
				UserID: "user-123",
				Title:  "Test",
			},
			setupRepo: func(repo *mockChatRepository) {
				repo.createError = errors.New("database error")
			},
			wantErr:     true,
			errContains: "failed to create session",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockRepo := newMockChatRepo()
			tt.setupRepo(mockRepo)

			svc := newTestService(mockRepo)

			session, err := svc.CreateSession(ctx, tt.req)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CreateSession() expected error, got nil")
				}
				if tt.errContains != "" && err != nil {
					if !contains(err.Error(), tt.errContains) {
						t.Errorf("Error = %v, want contain %q", err, tt.errContains)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("CreateSession() unexpected error: %v", err)
			}
			if session == nil {
				t.Fatal("CreateSession() returned nil session")
			}
			if session.UserID != tt.req.UserID {
				t.Errorf("UserID = %s, want %s", session.UserID, tt.req.UserID)
			}
			if session.Status != "active" {
				t.Errorf("Status = %s, want active", session.Status)
			}
		})
	}
}

func TestGetSession(t *testing.T) {
	ctx := context.Background()
	mockRepo := newMockChatRepo()

	testSession := &model.ChatSession{
		ID:     "session-123",
		UserID: "user-456",
		Title:  "Test Session",
		Status: "active",
	}
	mockRepo.sessions[testSession.ID] = testSession

	svc := newTestService(mockRepo)

	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "get existing session",
			id:      "session-123",
			wantErr: false,
		},
		{
			name:    "get non-existent session",
			id:      "non-existent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session, err := svc.GetSession(ctx, tt.id)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetSession() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("GetSession() unexpected error: %v", err)
			}
			if session == nil {
				t.Fatal("GetSession() returned nil session")
			}
			if session.ID != tt.id {
				t.Errorf("ID = %s, want %s", session.ID, tt.id)
			}
		})
	}
}

func TestListSessions(t *testing.T) {
	tests := []struct {
		name      string
		setupRepo func(*mockChatRepository)
		req       *ListSessionsRequest
		wantCount int
		wantErr   bool
	}{
		{
			name: "list sessions for user",
			setupRepo: func(repo *mockChatRepository) {
				repo.sessions["s1"] = &model.ChatSession{ID: "s1", UserID: "user-123"}
				repo.sessions["s2"] = &model.ChatSession{ID: "s2", UserID: "user-123"}
				repo.sessions["s3"] = &model.ChatSession{ID: "s3", UserID: "other-user"}
			},
			req: &ListSessionsRequest{
				UserID: "user-123",
				Page:   1,
				Size:   10,
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "default page and size",
			setupRepo: func(repo *mockChatRepository) {
				repo.sessions["s1"] = &model.ChatSession{ID: "s1", UserID: "user-123"}
			},
			req: &ListSessionsRequest{
				UserID: "user-123",
				Page:   0,
				Size:   0,
			},
			wantCount: 1,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockRepo := newMockChatRepo()
			tt.setupRepo(mockRepo)

			svc := newTestService(mockRepo)

			sessions, total, err := svc.ListSessions(ctx, tt.req)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ListSessions() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ListSessions() unexpected error: %v", err)
			}
			if len(sessions) != tt.wantCount {
				t.Errorf("Count = %d, want %d", len(sessions), tt.wantCount)
			}
			if total != int64(tt.wantCount) {
				t.Errorf("Total = %d, want %d", total, tt.wantCount)
			}
		})
	}
}

func TestUpdateSession(t *testing.T) {
	ctx := context.Background()
	mockRepo := newMockChatRepo()

	testSession := &model.ChatSession{
		ID:      "session-123",
		UserID:  "user-456",
		Title:   "Original Title",
		AgentID: "agent-1",
		Status:  "active",
	}
	mockRepo.sessions[testSession.ID] = testSession

	svc := newTestService(mockRepo)

	tests := []struct {
		name        string
		id          string
		req         *CreateSessionRequest
		wantErr     bool
		errContains string
	}{
		{
			name: "update title",
			id:   "session-123",
			req: &CreateSessionRequest{
				Title: "Updated Title",
			},
			wantErr: false,
		},
		{
			name: "update agent",
			id:   "session-123",
			req: &CreateSessionRequest{
				AgentID: "new-agent",
			},
			wantErr: false,
		},
		{
			name: "update non-existent session",
			id:   "non-existent",
			req: &CreateSessionRequest{
				Title: "Title",
			},
			wantErr:     true,
			errContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session, err := svc.UpdateSession(ctx, tt.id, tt.req)

			if tt.wantErr {
				if err == nil {
					t.Errorf("UpdateSession() expected error, got nil")
				}
				if tt.errContains != "" && err != nil {
					if !contains(err.Error(), tt.errContains) {
						t.Errorf("Error = %v, want contain %q", err, tt.errContains)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("UpdateSession() unexpected error: %v", err)
			}
			if session == nil {
				t.Fatal("UpdateSession() returned nil session")
			}

			if tt.req.Title != "" && session.Title != tt.req.Title {
				t.Errorf("Title = %s, want %s", session.Title, tt.req.Title)
			}
			if tt.req.AgentID != "" && session.AgentID != tt.req.AgentID {
				t.Errorf("AgentID = %s, want %s", session.AgentID, tt.req.AgentID)
			}
		})
	}
}

func TestDeleteSession(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		setupRepo   func(*mockChatRepository)
		wantErr     bool
		errContains string
	}{
		{
			name: "delete existing session",
			id:   "session-123",
			setupRepo: func(repo *mockChatRepository) {
				repo.sessions["session-123"] = &model.ChatSession{ID: "session-123"}
			},
			wantErr: false,
		},
		{
			name: "delete with repository error",
			id:   "session-123",
			setupRepo: func(repo *mockChatRepository) {
				repo.sessions["session-123"] = &model.ChatSession{ID: "session-123"}
				repo.deleteError = errors.New("database error")
			},
			wantErr:     true,
			errContains: "failed to delete",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockRepo := newMockChatRepo()
			tt.setupRepo(mockRepo)

			svc := newTestService(mockRepo)

			err := svc.DeleteSession(ctx, tt.id)

			if tt.wantErr {
				if err == nil {
					t.Errorf("DeleteSession() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("DeleteSession() unexpected error: %v", err)
			}
		})
	}
}

func TestSendMessage(t *testing.T) {
	ctx := context.Background()
	mockRepo := newMockChatRepo()

	testSession := &model.ChatSession{
		ID:     "session-123",
		UserID: "user-456",
	}
	mockRepo.sessions[testSession.ID] = testSession

	svc := newTestService(mockRepo)

	tests := []struct {
		name        string
		sessionID   string
		req         *SendMessageRequest
		wantErr     bool
		errContains string
	}{
		{
			name:      "send user message",
			sessionID: "session-123",
			req: &SendMessageRequest{
				Role:    "user",
				Content: "Hello, world!",
			},
			wantErr: false,
		},
		{
			name:      "send assistant message",
			sessionID: "session-123",
			req: &SendMessageRequest{
				Role:    "assistant",
				Content: "Hi there!",
			},
			wantErr: false,
		},
		{
			name:      "send to non-existent session",
			sessionID: "non-existent",
			req: &SendMessageRequest{
				Role:    "user",
				Content: "Hello",
			},
			wantErr:     true,
			errContains: "session not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message, err := svc.SendMessage(ctx, tt.sessionID, tt.req)

			if tt.wantErr {
				if err == nil {
					t.Errorf("SendMessage() expected error, got nil")
				}
				if tt.errContains != "" && err != nil {
					if !contains(err.Error(), tt.errContains) {
						t.Errorf("Error = %v, want contain %q", err, tt.errContains)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("SendMessage() unexpected error: %v", err)
			}
			if message == nil {
				t.Fatal("SendMessage() returned nil message")
			}
			if message.Role != tt.req.Role {
				t.Errorf("Role = %s, want %s", message.Role, tt.req.Role)
			}
			if message.Content != tt.req.Content {
				t.Errorf("Content = %s, want %s", message.Content, tt.req.Content)
			}
		})
	}
}

func TestGetMessages(t *testing.T) {
	ctx := context.Background()
	mockRepo := newMockChatRepo()

	sessionID := "session-123"
	mockRepo.sessions[sessionID] = &model.ChatSession{ID: sessionID}
	mockRepo.messages[sessionID] = []*model.ChatMessage{
		{ID: "msg1", Role: "user", Content: "Hello"},
		{ID: "msg2", Role: "assistant", Content: "Hi"},
	}

	svc := newTestService(mockRepo)

	tests := []struct {
		name      string
		sessionID string
		wantCount int
		wantErr   bool
	}{
		{
			name:      "get messages for session",
			sessionID: "session-123",
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "get messages for empty session",
			sessionID: "empty-session",
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messages, err := svc.GetMessages(ctx, tt.sessionID)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetMessages() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("GetMessages() unexpected error: %v", err)
			}
			if len(messages) != tt.wantCount {
				t.Errorf("Count = %d, want %d", len(messages), tt.wantCount)
			}
		})
	}
}

func TestLoadMessages(t *testing.T) {
	ctx := context.Background()
	mockRepo := newMockChatRepo()

	sessionID := "session-123"
	mockRepo.sessions[sessionID] = &model.ChatSession{ID: sessionID}
	mockRepo.messages[sessionID] = []*model.ChatMessage{
		{ID: "msg1", Role: "user", Content: "Message 1"},
		{ID: "msg2", Role: "assistant", Content: "Response 1"},
		{ID: "msg3", Role: "user", Content: "Message 2"},
	}

	svc := newTestService(mockRepo)

	tests := []struct {
		name        string
		sessionID   string
		req         *LoadMessagesRequest
		wantCount   int
		wantErr     bool
		errContains string
	}{
		{
			name:      "load messages with default limit",
			sessionID: "session-123",
			req:       &LoadMessagesRequest{},
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:      "load messages with custom limit",
			sessionID: "session-123",
			req:       &LoadMessagesRequest{Limit: 2},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "load messages for non-existent session",
			sessionID: "non-existent",
			req:       &LoadMessagesRequest{},
			wantCount: 0,
			wantErr:   true,
			errContains: "session not found",
		},
		{
			name:      "enforce max limit",
			sessionID: "session-123",
			req:       &LoadMessagesRequest{Limit: 200},
			wantCount: 3,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messages, err := svc.LoadMessages(ctx, tt.sessionID, tt.req)

			if tt.wantErr {
				if err == nil {
					t.Errorf("LoadMessages() expected error, got nil")
				}
				if tt.errContains != "" && err != nil {
					if !contains(err.Error(), tt.errContains) {
						t.Errorf("Error = %v, want contain %q", err, tt.errContains)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("LoadMessages() unexpected error: %v", err)
			}
			if len(messages) != tt.wantCount {
				t.Errorf("Count = %d, want %d", len(messages), tt.wantCount)
			}
		})
	}
}

func TestDeleteMessage(t *testing.T) {
	ctx := context.Background()
	mockRepo := newMockChatRepo()

	sessionID := "session-123"
	messageID := "msg-1"
	mockRepo.sessions[sessionID] = &model.ChatSession{ID: sessionID}
	mockRepo.messages[sessionID] = []*model.ChatMessage{
		{ID: messageID, SessionID: sessionID, Role: "user", Content: "Hello"},
		{ID: "msg-2", SessionID: "other-session", Role: "user", Content: "Other"},
	}

	svc := newTestService(mockRepo)

	tests := []struct {
		name        string
		sessionID   string
		messageID   string
		wantErr     bool
		errContains string
	}{
		{
			name:      "delete message from session",
			sessionID: "session-123",
			messageID: "msg-1",
			wantErr:   false,
		},
		{
			name:      "message does not belong to session",
			sessionID: "session-123",
			messageID: "msg-2",
			wantErr:   true,
			errContains: "does not belong to this session",
		},
		{
			name:      "message not found",
			sessionID: "session-123",
			messageID: "non-existent",
			wantErr:   true,
			errContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.DeleteMessage(ctx, tt.sessionID, tt.messageID)

			if tt.wantErr {
				if err == nil {
					t.Errorf("DeleteMessage() expected error, got nil")
				}
				if tt.errContains != "" && err != nil {
					if !contains(err.Error(), tt.errContains) {
						t.Errorf("Error = %v, want contain %q", err, tt.errContains)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("DeleteMessage() unexpected error: %v", err)
			}
		})
	}
}

func TestGenerateDefaultTitle(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected string
	}{
		{
			name:     "short message",
			message:  "Hello",
			expected: "Hello",
		},
		{
			name:     "long message should be truncated",
			message:  "This is a very long message that exceeds twenty characters",
			expected: "This is a very long ...", // 20 chars + "..."
		},
		{
			name:     "empty message",
			message:  "",
			expected: "新对话",
		},
		{
			name:     "whitespace only",
			message:  "   ",
			expected: "新对话",
		},
		{
			name:     "exactly 20 characters",
			message:  "12345678901234567890",
			expected: "12345678901234567890",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateDefaultTitle(tt.message)
			if result != tt.expected {
				t.Errorf("generateDefaultTitle() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestTruncateMessage(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		maxLen   int
		expected string
	}{
		{
			name:     "short string",
			s:        "Hello",
			maxLen:   10,
			expected: "Hello",
		},
		{
			name:     "exactly max length",
			s:        "12345",
			maxLen:   5,
			expected: "12345",
		},
		{
			name:     "exceeds max length",
			s:        "1234567890",
			maxLen:   5,
			expected: "12345...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateMessage(tt.s, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncateMessage() = %s, want %s", result, tt.expected)
			}
		})
	}
}

// 辅助函数
func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
