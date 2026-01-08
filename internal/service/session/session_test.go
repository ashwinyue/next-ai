// Package session 提供 Session 服务单元测试
package session

import (
	"context"
	"testing"
	"time"

	"github.com/cloudwego/eino/schema"
)

// ========== roleToSchema 测试 ==========

func TestRoleToSchema(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		expected schema.RoleType
	}{
		{
			name:     "user role",
			role:     "user",
			expected: schema.User,
		},
		{
			name:     "system role",
			role:     "system",
			expected: schema.System,
		},
		{
			name:     "assistant role",
			role:     "assistant",
			expected: schema.Assistant,
		},
		{
			name:     "empty role",
			role:     "",
			expected: schema.User, // 默认为 User
		},
		{
			name:     "unknown role",
			role:     "unknown",
			expected: schema.User, // 默认为 User
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := roleToSchema(tt.role)
			if result != tt.expected {
				t.Errorf("roleToSchema() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// ========== DefaultConfig 测试 ==========

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	// 验证默认值
	if cfg.MaxHistoryMessages <= 0 {
		t.Errorf("MaxHistoryMessages = %d, want > 0", cfg.MaxHistoryMessages)
	}
	if cfg.HistoryTTL <= 0 {
		t.Errorf("HistoryTTL = %v, want > 0", cfg.HistoryTTL)
	}
	if cfg.CheckpointTTL <= 0 {
		t.Errorf("CheckpointTTL = %v, want > 0", cfg.CheckpointTTL)
	}
}

// ========== stateKey 测试 ==========

func TestStateKey(t *testing.T) {
	manager := &StateManager{}

	tests := []struct {
		name     string
		sessionID string
		expected string
	}{
		{
			name:     "valid session id",
			sessionID: "session-123",
			expected: "session:state:session-123",
		},
		{
			name:     "empty session id",
			sessionID: "",
			expected: "session:state:",
		},
		{
			name:     "session id with special chars",
			sessionID: "session:123:abc",
			expected: "session:state:session:123:abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.stateKey(tt.sessionID)
			if result != tt.expected {
				t.Errorf("stateKey() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// ========== ActiveStream 基本测试 ==========

func TestActiveStream_Done(t *testing.T) {
	stream := &ActiveStream{
		Done: false,
	}

	if stream.IsDone() {
		t.Error("New ActiveStream should not be done")
	}

	stream.MarkDone()

	if !stream.IsDone() {
		t.Error("After MarkDone(), stream should be done")
	}
}

func TestActiveStream_Content(t *testing.T) {
	stream := &ActiveStream{}

	// 空流
	if stream.GetContent() != "" {
		t.Errorf("GetContent() = %q, want empty string", stream.GetContent())
	}

	// 添加内容
	stream.Content.WriteString("Hello")
	if stream.GetContent() != "Hello" {
		t.Errorf("GetContent() = %q, want 'Hello'", stream.GetContent())
	}
}

// ========== Config 测试 ==========

func TestConfig_DefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// 验证默认配置值
	if cfg.MaxHistoryMessages != 100 {
		t.Errorf("MaxHistoryMessages = %d, want 100", cfg.MaxHistoryMessages)
	}
	if cfg.EnableContextCompression {
		t.Error("EnableContextCompression should be false by default")
	}
	if cfg.MaxContextTokens != 4000 {
		t.Errorf("MaxContextTokens = %d, want 4000", cfg.MaxContextTokens)
	}
	if cfg.CompressionThreshold != 20 {
		t.Errorf("CompressionThreshold = %d, want 20", cfg.CompressionThreshold)
	}
}

// ========== Session 结构测试 ==========

func TestSession_Structure(t *testing.T) {
	session := &Session{
		ID:        "test-session",
		Messages:  []*schema.Message{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if session.ID != "test-session" {
		t.Errorf("ID = %q, want 'test-session'", session.ID)
	}
}

// ========== 消息处理测试 ==========

func TestMessageProcessing(t *testing.T) {
	// 创建测试消息
	msgs := []*schema.Message{
		{Role: schema.User, Content: "Hello"},
		{Role: schema.Assistant, Content: "Hi there"},
		{Role: schema.User, Content: "How are you?"},
	}

	// 验证消息数量
	if len(msgs) != 3 {
		t.Errorf("Message count = %d, want 3", len(msgs))
	}

	// 验证消息顺序
	if msgs[0].Role != schema.User {
		t.Error("First message should be from user")
	}
	if msgs[1].Role != schema.Assistant {
		t.Error("Second message should be from assistant")
	}
}

// ========== State 结构测试 ==========

func TestState_Structure(t *testing.T) {
	state := &State{
		SessionID: "session-123",
		AgentID:   "agent-456",
	}

	if state.SessionID != "session-123" {
		t.Errorf("SessionID = %q, want 'session-123'", state.SessionID)
	}
	if state.AgentID != "agent-456" {
		t.Errorf("AgentID = %q, want 'agent-456'", state.AgentID)
	}

	// 测试消息操作
	state.Messages = append(state.Messages, &schema.Message{
		Role:    schema.User,
		Content: "test message",
	})

	if len(state.Messages) != 1 {
		t.Errorf("Messages count = %d, want 1", len(state.Messages))
	}
}

// ========== StateManager 测试 ==========

func TestStateManager_NewStateManager(t *testing.T) {
	// 使用 nil 参数创建
	manager := NewStateManager(nil, nil, nil, nil)

	if manager == nil {
		t.Fatal("NewStateManager() returned nil")
	}

	// 应该使用默认配置
	cfg := manager.GetConfig()
	if cfg == nil {
		t.Error("GetConfig() returned nil")
	}
}

func TestStateManager_GetConfig(t *testing.T) {
	manager := NewStateManager(nil, nil, nil, nil)
	cfg := manager.GetConfig()

	if cfg == nil {
		t.Fatal("GetConfig() returned nil")
	}

	// 验证默认值
	if cfg.MaxHistoryMessages != 100 {
		t.Errorf("MaxHistoryMessages = %d, want 100", cfg.MaxHistoryMessages)
	}
}

func TestStateManager_UpdateConfig(t *testing.T) {
	manager := NewStateManager(nil, nil, nil, nil)

	// 更新配置
	newCfg := &Config{
		MaxHistoryMessages: 200,
		HistoryTTL:         48 * time.Hour,
	}

	manager.UpdateConfig(newCfg)

	updatedCfg := manager.GetConfig()
	if updatedCfg.MaxHistoryMessages != 200 {
		t.Errorf("MaxHistoryMessages = %d, want 200", updatedCfg.MaxHistoryMessages)
	}
	if updatedCfg.HistoryTTL != 48*time.Hour {
		t.Errorf("HistoryTTL = %v, want 48h", updatedCfg.HistoryTTL)
	}
}

// ========== Manager 测试 ==========

func TestNewManager(t *testing.T) {
	// nil redis 客户端
	manager := NewManager(nil)

	if manager == nil {
		t.Fatal("NewManager() returned nil")
	}
	if manager.memory == nil {
		t.Error("manager.memory should be initialized")
	}
	if manager.activeStreams == nil {
		t.Error("manager.activeStreams should be initialized")
	}
}

func TestManager_Get_CreateNew(t *testing.T) {
	manager := NewManager(nil)
	ctx := context.Background()

	session, err := manager.Get(ctx, "new-session")

	if err != nil {
		t.Errorf("Get() unexpected error: %v", err)
	}
	if session == nil {
		t.Fatal("Get() returned nil session")
	}
	if session.ID != "new-session" {
		t.Errorf("ID = %q, want 'new-session'", session.ID)
	}
	if len(session.Messages) != 0 {
		t.Errorf("Messages should be empty, got %d", len(session.Messages))
	}
}

func TestManager_Get_Existing(t *testing.T) {
	manager := NewManager(nil)
	ctx := context.Background()

	// 第一次获取会创建
	session1, _ := manager.Get(ctx, "test-session")
	session1.Messages = append(session1.Messages, &schema.Message{Role: schema.User, Content: "test"})

	// 第二次获取应该返回相同的会话
	session2, err := manager.Get(ctx, "test-session")

	if err != nil {
		t.Errorf("Get() unexpected error: %v", err)
	}
	if len(session2.Messages) != 1 {
		t.Errorf("Messages count = %d, want 1", len(session2.Messages))
	}
}

func TestManager_Append(t *testing.T) {
	manager := NewManager(nil)
	ctx := context.Background()

	// 先创建会话
	manager.Get(ctx, "test-session")

	// 追加消息
	msg := &schema.Message{Role: schema.User, Content: "Hello"}
	err := manager.Append(ctx, "test-session", msg)

	if err != nil {
		t.Errorf("Append() unexpected error: %v", err)
	}

	// 验证消息已追加
	session, _ := manager.Get(ctx, "test-session")
	if len(session.Messages) != 1 {
		t.Errorf("Messages count = %d, want 1", len(session.Messages))
	}
}

func TestManager_GetHistory(t *testing.T) {
	manager := NewManager(nil)
	ctx := context.Background()

	msgs := []*schema.Message{
		{Role: schema.User, Content: "Hello"},
		{Role: schema.Assistant, Content: "Hi"},
	}

	// 追加消息
	for _, msg := range msgs {
		manager.Append(ctx, "test-session", msg)
	}

	// 获取历史
	history, err := manager.GetHistory(ctx, "test-session")

	if err != nil {
		t.Errorf("GetHistory() unexpected error: %v", err)
	}
	if len(history) != 2 {
		t.Errorf("History count = %d, want 2", len(history))
	}
}

func TestManager_Clear(t *testing.T) {
	manager := NewManager(nil)
	ctx := context.Background()

	// 创建会话并添加消息
	manager.Get(ctx, "test-session")
	manager.Append(ctx, "test-session", &schema.Message{Role: schema.User, Content: "test"})

	// 清空会话
	err := manager.Clear(ctx, "test-session")

	if err != nil {
		t.Errorf("Clear() unexpected error: %v", err)
	}

	// 验证会话已从内存中删除
	_, exists := manager.memory["test-session"]
	if exists {
		t.Error("Session should be removed from memory after Clear")
	}
}

// ========== ActiveStream 更多测试 ==========

func TestActiveStream_AppendChunk(t *testing.T) {
	stream := &ActiveStream{}

	stream.AppendChunk("Hello")
	stream.AppendChunk(" World")

	if stream.GetContent() != "Hello World" {
		t.Errorf("GetContent() = %q, want 'Hello World'", stream.GetContent())
	}
}

func TestActiveStream_GetPartialChunks(t *testing.T) {
	stream := &ActiveStream{}

	stream.AppendChunk("chunk1")
	stream.AppendChunk("chunk2")

	chunks := stream.GetPartialChunks()

	if len(chunks) != 2 {
		t.Errorf("GetPartialChunks() returned %d chunks, want 2", len(chunks))
	}
	if chunks[0] != "chunk1" {
		t.Errorf("First chunk = %q, want 'chunk1'", chunks[0])
	}
}

func TestActiveStream_IsDone_Concurrent(t *testing.T) {
	stream := &ActiveStream{Done: false}

	if stream.IsDone() {
		t.Error("New stream should not be done")
	}

	// 标记完成
	stream.MarkDone()

	if !stream.IsDone() {
		t.Error("Stream should be done after MarkDone")
	}
}

func TestActiveStream_Content_Accumulation(t *testing.T) {
	stream := &ActiveStream{}

	// 多次追加
	for i := 0; i < 5; i++ {
		stream.AppendChunk("test")
	}

	expected := "testtesttesttesttest"
	if stream.GetContent() != expected {
		t.Errorf("GetContent() = %q, want %q", stream.GetContent(), expected)
	}
}

// ========== Manager Stream 测试 ==========

func TestManager_RegisterStream(t *testing.T) {
	manager := NewManager(nil)

	_, cancel := context.WithCancel(context.Background())
	stream := manager.RegisterStream("session-1", "msg-1", cancel)

	if stream == nil {
		t.Fatal("RegisterStream() returned nil")
	}
	if stream.SessionID != "session-1" {
		t.Errorf("SessionID = %q, want 'session-1'", stream.SessionID)
	}
	if stream.MessageID != "msg-1" {
		t.Errorf("MessageID = %q, want 'msg-1'", stream.MessageID)
	}
	if stream.IsDone() {
		t.Error("New stream should not be done")
	}
}

func TestManager_GetStream(t *testing.T) {
	manager := NewManager(nil)
	_, cancel := context.WithCancel(context.Background())

	// 注册流
	manager.RegisterStream("session-1", "msg-1", cancel)

	// 获取流
	stream := manager.GetStream("session-1", "msg-1")

	if stream == nil {
		t.Fatal("GetStream() returned nil")
	}

	// 获取不存在的流
	nilStream := manager.GetStream("session-1", "msg-2")
	if nilStream != nil {
		t.Error("GetStream() should return nil for non-existent stream")
	}
}

func TestManager_UnregisterStream(t *testing.T) {
	manager := NewManager(nil)
	_, cancel := context.WithCancel(context.Background())

	// 注册流
	manager.RegisterStream("session-1", "msg-1", cancel)

	// 注销流
	manager.UnregisterStream("session-1", "msg-1")

	// 验证流已删除
	stream := manager.GetStream("session-1", "msg-1")
	if stream != nil {
		t.Error("Stream should be removed after UnregisterStream")
	}
}

func TestManager_StopStream(t *testing.T) {
	manager := NewManager(nil)
	_, cancel := context.WithCancel(context.Background())

	// 注册流
	manager.RegisterStream("session-1", "msg-1", cancel)

	// 停止流
	stopped := manager.StopStream("session-1", "msg-1")

	if !stopped {
		t.Error("StopStream() should return true")
	}

	// 验证流已删除
	stream := manager.GetStream("session-1", "msg-1")
	if stream != nil {
		t.Error("Stream should be removed after StopStream")
	}
}

func TestManager_StopStream_NonExistent(t *testing.T) {
	manager := NewManager(nil)

	// 停止不存在的流
	stopped := manager.StopStream("session-1", "msg-1")

	if stopped {
		t.Error("StopStream() should return false for non-existent stream")
	}
}

// ========== RedisCheckpointStore 测试 ==========

func TestNewRedisCheckpointStore(t *testing.T) {
	// nil redis 客户端
	store := NewRedisCheckpointStore(nil, time.Hour)

	if store == nil {
		t.Fatal("NewRedisCheckpointStore() returned nil")
	}
}
