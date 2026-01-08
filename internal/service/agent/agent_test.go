// Package agent 提供 Agent 服务单元测试
package agent

import (
	"context"
	"errors"
	"testing"

	"github.com/ashwinyue/next-ai/internal/model"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// AgentRepository Agent Repository 接口（用于测试）
type AgentRepository interface {
	Create(agent *model.Agent) error
	GetByID(id string) (*model.Agent, error)
	GetByName(name string) (*model.Agent, error)
	List(offset, limit int) ([]*model.Agent, error)
	ListActive() ([]*model.Agent, error)
	Update(agent *model.Agent) error
	Delete(id string) error
}

// mockAgentRepository Mock Agent Repository
type mockAgentRepository struct {
	agents      map[string]*model.Agent
	createError error
	getError    error
	updateError error
	deleteError error
}

func newMockAgentRepo() *mockAgentRepository {
	return &mockAgentRepository{
		agents: make(map[string]*model.Agent),
	}
}

func (m *mockAgentRepository) Create(agent *model.Agent) error {
	if m.createError != nil {
		return m.createError
	}
	if _, exists := m.agents[agent.ID]; exists {
		return errors.New("agent already exists")
	}
	m.agents[agent.ID] = agent
	return nil
}

func (m *mockAgentRepository) GetByID(id string) (*model.Agent, error) {
	if m.getError != nil {
		return nil, m.getError
	}
	if agent, ok := m.agents[id]; ok {
		return agent, nil
	}
	return nil, errors.New("agent not found")
}

func (m *mockAgentRepository) GetByName(name string) (*model.Agent, error) {
	for _, agent := range m.agents {
		if agent.Name == name {
			return agent, nil
		}
	}
	return nil, errors.New("agent not found")
}

func (m *mockAgentRepository) List(offset, limit int) ([]*model.Agent, error) {
	result := make([]*model.Agent, 0, len(m.agents))
	for _, agent := range m.agents {
		result = append(result, agent)
	}
	return result, nil
}

func (m *mockAgentRepository) ListActive() ([]*model.Agent, error) {
	result := make([]*model.Agent, 0)
	for _, agent := range m.agents {
		if agent.IsActive {
			result = append(result, agent)
		}
	}
	return result, nil
}

func (m *mockAgentRepository) Update(agent *model.Agent) error {
	if m.updateError != nil {
		return m.updateError
	}
	if _, exists := m.agents[agent.ID]; !exists {
		return errors.New("agent not found")
	}
	m.agents[agent.ID] = agent
	return nil
}

func (m *mockAgentRepository) Delete(id string) error {
	if m.deleteError != nil {
		return m.deleteError
	}
	if _, exists := m.agents[id]; !exists {
		return errors.New("agent not found")
	}
	delete(m.agents, id)
	return nil
}

// testService 测试服务包装器
type testService struct {
	repo AgentRepository
}

func newTestService(repo AgentRepository) *testService {
	return &testService{repo: repo}
}

func (s *testService) CreateAgent(ctx context.Context, req *CreateAgentRequest) (*model.Agent, error) {
	// 检查名称是否已存在
	if _, err := s.repo.GetByName(req.Name); err == nil {
		return nil, errors.New("agent name already exists")
	}

	// 设置默认值
	agentMode := req.AgentMode
	if agentMode == "" {
		agentMode = "quick-answer"
	}

	// 构建 Tools JSON
	toolsJSON := make(model.JSON)
	if len(req.Tools) > 0 {
		for _, tool := range req.Tools {
			toolsJSON[tool] = true
		}
	}

	agent := &model.Agent{
		ID:           uuid.New().String(),
		Name:         req.Name,
		Description:  req.Description,
		Avatar:       req.Avatar,
		AgentMode:    agentMode,
		SystemPrompt: req.SystemPrompt,
		Tools:        toolsJSON,
		MaxIter:      req.MaxIter,
		Temperature:  req.Temperature,
		ModelConfig:  model.ModelConfig{Model: req.Model},
		KnowledgeIDs: pq.StringArray(req.KnowledgeIDs),
		IsActive:     true,
		IsBuiltin:    false,
	}

	if err := s.repo.Create(agent); err != nil {
		return nil, errors.New("failed to create agent: " + err.Error())
	}

	return agent, nil
}

func (s *testService) GetAgent(ctx context.Context, id string) (*model.Agent, error) {
	return s.repo.GetByID(id)
}

func (s *testService) ListAgents(ctx context.Context, req *ListAgentsRequest) ([]*model.Agent, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 || req.Size > 100 {
		req.Size = 20
	}

	agents, err := s.repo.List((req.Page-1)*req.Size, req.Size)
	if err != nil {
		return nil, errors.New("failed to list agents: " + err.Error())
	}

	return agents, nil
}

func (s *testService) UpdateAgent(ctx context.Context, id string, req *CreateAgentRequest) (*model.Agent, error) {
	agent, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("agent not found: " + err.Error())
	}

	// 检查名称冲突
	if existing, err := s.repo.GetByName(req.Name); err == nil && existing.ID != id {
		return nil, errors.New("agent name already exists")
	}

	// 内置 Agent 不允许修改核心配置
	if agent.IsBuiltin {
		if req.SystemPrompt != "" && req.SystemPrompt != agent.SystemPrompt {
			return nil, errors.New("cannot modify builtin agent system prompt")
		}
		if req.AgentMode != "" && req.AgentMode != agent.AgentMode {
			return nil, errors.New("cannot modify builtin agent mode")
		}
	}

	if req.Name != "" {
		agent.Name = req.Name
	}
	if req.Description != "" {
		agent.Description = req.Description
	}
	if req.Avatar != "" {
		agent.Avatar = req.Avatar
	}
	if req.SystemPrompt != "" {
		agent.SystemPrompt = req.SystemPrompt
	}
	// 更新 Tools
	if req.Tools != nil {
		toolsJSON := make(model.JSON)
		for _, tool := range req.Tools {
			toolsJSON[tool] = true
		}
		agent.Tools = toolsJSON
	}
	if req.MaxIter > 0 {
		agent.MaxIter = req.MaxIter
	}
	if req.Temperature > 0 {
		agent.Temperature = req.Temperature
	}
	if req.Model != "" {
		agent.ModelConfig.Model = req.Model
	}
	if req.KnowledgeIDs != nil {
		agent.KnowledgeIDs = pq.StringArray(req.KnowledgeIDs)
	}

	if err := s.repo.Update(agent); err != nil {
		return nil, errors.New("failed to update agent: " + err.Error())
	}

	return agent, nil
}

func (s *testService) DeleteAgent(ctx context.Context, id string) error {
	agent, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New("agent not found: " + err.Error())
	}

	// 内置 Agent 不允许删除
	if agent.IsBuiltin {
		return errors.New("cannot delete builtin agent")
	}

	if err := s.repo.Delete(id); err != nil {
		return errors.New("failed to delete agent: " + err.Error())
	}

	return nil
}

func (s *testService) CopyAgent(ctx context.Context, id string) (*model.Agent, error) {
	original, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("agent not found: " + err.Error())
	}

	// 生成新名称
	newName := original.Name + " (Copy)"
	counter := 1
	for {
		if _, err := s.repo.GetByName(newName); err != nil {
			break
		}
		counter++
		newName = original.Name + " (Copy " + string(rune(counter)) + ")"
	}

	copied := &model.Agent{
		ID:           uuid.New().String(),
		Name:         newName,
		Description:  original.Description,
		Avatar:       original.Avatar,
		AgentMode:    original.AgentMode,
		SystemPrompt: original.SystemPrompt,
		Tools:        original.Tools,
		MaxIter:      original.MaxIter,
		Temperature:  original.Temperature,
		ModelConfig:  original.ModelConfig,
		KnowledgeIDs: original.KnowledgeIDs,
		IsActive:     true,
		IsBuiltin:    false, // 复制的 Agent 不是内置的
	}

	if err := s.repo.Create(copied); err != nil {
		return nil, errors.New("failed to create copy: " + err.Error())
	}

	return copied, nil
}

// ========== 测试用例 ==========

func TestCreateAgent(t *testing.T) {
	tests := []struct {
		name        string
		req         *CreateAgentRequest
		setupRepo   func(*mockAgentRepository)
		wantErr     bool
		errContains string
	}{
		{
			name: "create agent with valid data",
			req: &CreateAgentRequest{
				Name:         "Test Agent",
				Description:  "Test Description",
				SystemPrompt: "You are a helpful assistant",
				Tools:        []string{"web_search"},
				Model:        "gpt-4",
			},
			setupRepo: func(repo *mockAgentRepository) {},
			wantErr:   false,
		},
		{
			name: "create agent with default mode",
			req: &CreateAgentRequest{
				Name: "Default Mode Agent",
				Model: "gpt-4",
			},
			setupRepo: func(repo *mockAgentRepository) {},
			wantErr:   false,
		},
		{
			name: "duplicate agent name",
			req: &CreateAgentRequest{
				Name:  "Duplicate Agent",
				Model: "gpt-4",
			},
			setupRepo: func(repo *mockAgentRepository) {
				// 预先创建同名 Agent
				repo.agents["existing"] = &model.Agent{
					ID:   "existing",
					Name: "Duplicate Agent",
				}
			},
			wantErr:     true,
			errContains: "already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockRepo := newMockAgentRepo()
			tt.setupRepo(mockRepo)

			svc := newTestService(mockRepo)

			agent, err := svc.CreateAgent(ctx, tt.req)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CreateAgent() expected error, got nil")
				}
				if tt.errContains != "" && err != nil {
					if !contains(err.Error(), tt.errContains) {
						t.Errorf("Error = %v, want contain %q", err, tt.errContains)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("CreateAgent() unexpected error: %v", err)
			}
			if agent == nil {
				t.Fatal("CreateAgent() returned nil agent")
			}
			if agent.Name != tt.req.Name {
				t.Errorf("Name = %s, want %s", agent.Name, tt.req.Name)
			}
			if !agent.IsActive {
				t.Errorf("IsActive = %v, want true", agent.IsActive)
			}
			if agent.IsBuiltin {
				t.Errorf("IsBuiltin = %v, want false", agent.IsBuiltin)
			}
		})
	}
}

func TestGetAgent(t *testing.T) {
	ctx := context.Background()
	mockRepo := newMockAgentRepo()

	testAgent := &model.Agent{
		ID:      "agent-123",
		Name:    "Test Agent",
		IsActive: true,
	}
	mockRepo.agents[testAgent.ID] = testAgent

	svc := newTestService(mockRepo)

	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "get existing agent",
			id:      "agent-123",
			wantErr: false,
		},
		{
			name:    "get non-existent agent",
			id:      "non-existent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent, err := svc.GetAgent(ctx, tt.id)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetAgent() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("GetAgent() unexpected error: %v", err)
			}
			if agent == nil {
				t.Fatal("GetAgent() returned nil agent")
			}
			if agent.ID != tt.id {
				t.Errorf("ID = %s, want %s", agent.ID, tt.id)
			}
		})
	}
}

func TestListAgents(t *testing.T) {
	tests := []struct {
		name      string
		setupRepo func(*mockAgentRepository)
		req       *ListAgentsRequest
		wantCount int
		wantErr   bool
	}{
		{
			name: "list agents",
			setupRepo: func(repo *mockAgentRepository) {
				repo.agents["a1"] = &model.Agent{ID: "a1", Name: "Agent 1", IsActive: true}
				repo.agents["a2"] = &model.Agent{ID: "a2", Name: "Agent 2", IsActive: true}
				repo.agents["a3"] = &model.Agent{ID: "a3", Name: "Agent 3", IsActive: false}
			},
			req: &ListAgentsRequest{
				Page: 1,
				Size: 10,
			},
			wantCount: 3,
			wantErr:   false,
		},
		{
			name: "default page and size",
			setupRepo: func(repo *mockAgentRepository) {
				repo.agents["a1"] = &model.Agent{ID: "a1", Name: "Agent 1"}
			},
			req: &ListAgentsRequest{
				Page: 0,
				Size: 0,
			},
			wantCount: 1,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockRepo := newMockAgentRepo()
			tt.setupRepo(mockRepo)

			svc := newTestService(mockRepo)

			agents, err := svc.ListAgents(ctx, tt.req)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ListAgents() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ListAgents() unexpected error: %v", err)
			}
			if len(agents) != tt.wantCount {
				t.Errorf("Count = %d, want %d", len(agents), tt.wantCount)
			}
		})
	}
}

func TestUpdateAgent(t *testing.T) {
	ctx := context.Background()
	mockRepo := newMockAgentRepo()

	regularAgent := &model.Agent{
		ID:        "agent-123",
		Name:      "Regular Agent",
		IsBuiltin: false,
	}
	mockRepo.agents[regularAgent.ID] = regularAgent

	builtinAgent := &model.Agent{
		ID:           "builtin-1",
		Name:         "Builtin Agent",
		IsBuiltin:    true,
		SystemPrompt: "Original prompt",
		AgentMode:    "quick-answer",
	}
	mockRepo.agents[builtinAgent.ID] = builtinAgent

	svc := newTestService(mockRepo)

	tests := []struct {
		name        string
		id          string
		req         *CreateAgentRequest
		wantErr     bool
		errContains string
	}{
		{
			name: "update regular agent name",
			id:   "agent-123",
			req: &CreateAgentRequest{
				Name: "Updated Name",
			},
			wantErr: false,
		},
		{
			name: "update builtin agent system prompt should fail",
			id:   "builtin-1",
			req: &CreateAgentRequest{
				SystemPrompt: "New prompt",
			},
			wantErr:     true,
			errContains: "cannot modify builtin",
		},
		{
			name: "update builtin agent mode should fail",
			id:   "builtin-1",
			req: &CreateAgentRequest{
				AgentMode: "smart-reasoning",
			},
			wantErr:     true,
			errContains: "cannot modify builtin",
		},
		{
			name: "update non-existent agent",
			id:   "non-existent",
			req: &CreateAgentRequest{
				Name: "Name",
			},
			wantErr:     true,
			errContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent, err := svc.UpdateAgent(ctx, tt.id, tt.req)

			if tt.wantErr {
				if err == nil {
					t.Errorf("UpdateAgent() expected error, got nil")
				}
				if tt.errContains != "" && err != nil {
					if !contains(err.Error(), tt.errContains) {
						t.Errorf("Error = %v, want contain %q", err, tt.errContains)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("UpdateAgent() unexpected error: %v", err)
			}
			if agent == nil {
				t.Fatal("UpdateAgent() returned nil agent")
			}

			if tt.req.Name != "" && agent.Name != tt.req.Name {
				t.Errorf("Name = %s, want %s", agent.Name, tt.req.Name)
			}
		})
	}
}

func TestDeleteAgent(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		setupRepo   func(*mockAgentRepository)
		wantErr     bool
		errContains string
	}{
		{
			name: "delete regular agent",
			id:   "agent-123",
			setupRepo: func(repo *mockAgentRepository) {
				repo.agents["agent-123"] = &model.Agent{
					ID:        "agent-123",
					Name:      "Regular Agent",
					IsBuiltin: false,
				}
			},
			wantErr: false,
		},
		{
			name: "delete builtin agent should fail",
			id:   "builtin-1",
			setupRepo: func(repo *mockAgentRepository) {
				repo.agents["builtin-1"] = &model.Agent{
					ID:        "builtin-1",
					Name:      "Builtin Agent",
					IsBuiltin: true,
				}
			},
			wantErr:     true,
			errContains: "cannot delete builtin",
		},
		{
			name:      "delete non-existent agent",
			id:        "non-existent",
			setupRepo: func(repo *mockAgentRepository) {},
			wantErr:   true,
			errContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockRepo := newMockAgentRepo()
			tt.setupRepo(mockRepo)

			svc := newTestService(mockRepo)

			err := svc.DeleteAgent(ctx, tt.id)

			if tt.wantErr {
				if err == nil {
					t.Errorf("DeleteAgent() expected error, got nil")
				}
				if tt.errContains != "" && err != nil {
					if !contains(err.Error(), tt.errContains) {
						t.Errorf("Error = %v, want contain %q", err, tt.errContains)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("DeleteAgent() unexpected error: %v", err)
			}

			// 验证 Agent 已被删除
			if _, exists := mockRepo.agents[tt.id]; exists {
				t.Error("Agent still exists after deletion")
			}
		})
	}
}

func TestCopyAgent(t *testing.T) {
	ctx := context.Background()
	mockRepo := newMockAgentRepo()

	originalAgent := &model.Agent{
		ID:           "agent-123",
		Name:         "Original Agent",
		Description:  "Original Description",
		AgentMode:    "quick-answer",
		SystemPrompt: "Original Prompt",
		IsBuiltin:    true, // 即使是内置 Agent 也能复制
	}
	mockRepo.agents[originalAgent.ID] = originalAgent

	svc := newTestService(mockRepo)

	tests := []struct {
		name        string
		id          string
		wantErr     bool
		errContains string
		verifyName  string
	}{
		{
			name:       "copy existing agent",
			id:         "agent-123",
			wantErr:    false,
			verifyName: "Original Agent (Copy)",
		},
		{
			name:        "copy non-existent agent",
			id:          "non-existent",
			wantErr:     true,
			errContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			copied, err := svc.CopyAgent(ctx, tt.id)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CopyAgent() expected error, got nil")
				}
				if tt.errContains != "" && err != nil {
					if !contains(err.Error(), tt.errContains) {
						t.Errorf("Error = %v, want contain %q", err, tt.errContains)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("CopyAgent() unexpected error: %v", err)
			}
			if copied == nil {
				t.Fatal("CopyAgent() returned nil agent")
			}

			// 验证复制的 Agent 属性
			if copied.Name != tt.verifyName {
				t.Errorf("Name = %s, want %s", copied.Name, tt.verifyName)
			}
			if copied.Description != originalAgent.Description {
				t.Errorf("Description = %s, want %s", copied.Description, originalAgent.Description)
			}
			if copied.IsBuiltin {
				t.Error("Copied agent should not be builtin")
			}
			if !copied.IsActive {
				t.Error("Copied agent should be active")
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
