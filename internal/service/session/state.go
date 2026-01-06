// Package session 提供会话状态管理
// 参考 next-ai/docs/eino-integration-guide.md
// 直接使用 eino compose.CheckPointStore，避免冗余封装
package session

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/ashwinyue/next-ai/internal/repository"
	"github.com/redis/go-redis/v9"
)

// State 会话状态
type State struct {
	SessionID  string            `json:"session_id"`
	AgentID    string            `json:"agent_id,omitempty"`
	Messages   []*schema.Message `json:"messages"`
	Metadata   map[string]any    `json:"metadata,omitempty"`
	LastActive time.Time         `json:"last_active"`
	Version    int64             `json:"version"`
}

// Config 会话配置
type Config struct {
	MaxHistoryMessages       int           // 最大保留历史消息数
	HistoryTTL                time.Duration // 历史 TTL
	CheckpointTTL             time.Duration // Checkpoint TTL
	EnableContextCompression bool          // 是否启用上下文压缩
	MaxContextTokens         int           // 最大上下文 Token 数
	CompressionThreshold     int           // 触发压缩的阈值
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		MaxHistoryMessages:       100,
		HistoryTTL:                24 * time.Hour,
		CheckpointTTL:             24 * time.Hour,
		EnableContextCompression: false,
		MaxContextTokens:         4000,
		CompressionThreshold:     20,
	}
}

// StateManager 会话状态管理器
// 使用 eino compose.CheckPointStore 持久化状态
type StateManager struct {
	checkpointStore compose.CheckPointStore
	messageRepo     *repository.Repositories
	redisClient     *redis.Client
	config          *Config
	mu              sync.RWMutex
}

// NewStateManager 创建状态管理器
func NewStateManager(
	checkpointStore compose.CheckPointStore,
	repos *repository.Repositories,
	redisClient *redis.Client,
	config *Config,
) *StateManager {
	if config == nil {
		config = DefaultConfig()
	}

	return &StateManager{
		checkpointStore: checkpointStore,
		messageRepo:     repos,
		redisClient:     redisClient,
		config:          config,
	}
}

// SaveState 保存状态到 Checkpoint
func (m *StateManager) SaveState(ctx context.Context, state *State) error {
	if state == nil {
		return fmt.Errorf("state cannot be nil")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	state.LastActive = time.Now()
	state.Version = time.Now().UnixNano() / int64(time.Millisecond)

	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	key := m.stateKey(state.SessionID)
	if err := m.checkpointStore.Set(ctx, key, data); err != nil {
		return fmt.Errorf("failed to save checkpoint: %w", err)
	}

	return nil
}

// LoadState 从 Checkpoint 加载状态
func (m *StateManager) LoadState(ctx context.Context, sessionID string) (*State, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := m.stateKey(sessionID)
	data, found, err := m.checkpointStore.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to load checkpoint: %w", err)
	}

	if !found {
		// 从数据库构建状态
		return m.buildStateFromDB(ctx, sessionID)
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state: %w", err)
	}

	return &state, nil
}

// GetHistory 获取会话历史
func (m *StateManager) GetHistory(ctx context.Context, sessionID string) ([]*schema.Message, error) {
	state, err := m.LoadState(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	return state.Messages, nil
}

// AppendMessage 追加消息
func (m *StateManager) AppendMessage(ctx context.Context, sessionID string, message *schema.Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, err := m.LoadState(ctx, sessionID)
	if err != nil {
		return err
	}

	state.Messages = append(state.Messages, message)

	// 应用历史消息限制
	if m.config.MaxHistoryMessages > 0 && len(state.Messages) > m.config.MaxHistoryMessages {
		state.Messages = state.Messages[len(state.Messages)-m.config.MaxHistoryMessages:]
	}

	return m.SaveState(ctx, state)
}

// AppendMessages 批量追加消息
func (m *StateManager) AppendMessages(ctx context.Context, sessionID string, messages []*schema.Message) error {
	if len(messages) == 0 {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	state, err := m.LoadState(ctx, sessionID)
	if err != nil {
		return err
	}

	state.Messages = append(state.Messages, messages...)

	if m.config.MaxHistoryMessages > 0 && len(state.Messages) > m.config.MaxHistoryMessages {
		state.Messages = state.Messages[len(state.Messages)-m.config.MaxHistoryMessages:]
	}

	return m.SaveState(ctx, state)
}

// ClearHistory 清空历史
func (m *StateManager) ClearHistory(ctx context.Context, sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, err := m.LoadState(ctx, sessionID)
	if err != nil {
		return err
	}

	state.Messages = []*schema.Message{}
	return m.SaveState(ctx, state)
}

// DeleteState 删除状态
func (m *StateManager) DeleteState(ctx context.Context, sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.stateKey(sessionID)
	_ = m.checkpointStore.Set(ctx, key, nil) // 清空即为删除
	return nil
}

// GetConfig 获取配置
func (m *StateManager) GetConfig() *Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// UpdateConfig 更新配置
func (m *StateManager) UpdateConfig(config *Config) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config = config
}

// buildStateFromDB 从数据库构建状态（回退方案）
func (m *StateManager) buildStateFromDB(ctx context.Context, sessionID string) (*State, error) {
	// 从 repository 获取消息
	messages, err := m.messageRepo.Chat.GetMessagesBySessionID(sessionID)
	if err != nil {
		// 返回空状态而不是错误
		return &State{
			SessionID:  sessionID,
			Messages:   []*schema.Message{},
			Metadata:   make(map[string]any),
			LastActive: time.Now(),
			Version:    time.Now().UnixNano() / int64(time.Millisecond),
		}, nil
	}

	// 转换为 schema.Message
	einoMessages := make([]*schema.Message, 0, len(messages))
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
		einoMessages = append(einoMessages, &schema.Message{
			Role:    role,
			Content: msg.Content,
		})
	}

	return &State{
		SessionID:  sessionID,
		Messages:   einoMessages,
		Metadata:   make(map[string]any),
		LastActive: time.Now(),
		Version:    time.Now().UnixNano() / int64(time.Millisecond),
	}, nil
}

// stateKey 生成状态存储 key
func (m *StateManager) stateKey(sessionID string) string {
	return fmt.Sprintf("session:state:%s", sessionID)
}

// RedisCheckpointStore Redis CheckpointStore 实现
type RedisCheckpointStore struct {
	client *redis.Client
	ttl    time.Duration
}

// NewRedisCheckpointStore 创建 Redis CheckpointStore
func NewRedisCheckpointStore(client *redis.Client, ttl time.Duration) compose.CheckPointStore {
	return &RedisCheckpointStore{
		client: client,
		ttl:    ttl,
	}
}

// Get 实现 CheckpointStore.Get
func (s *RedisCheckpointStore) Get(ctx context.Context, key string) ([]byte, bool, error) {
	val, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, false, nil
		}
		return nil, false, err
	}

	if val == "" {
		return nil, false, nil
	}

	return []byte(val), true, nil
}

// Set 实现 CheckpointStore.Set
func (s *RedisCheckpointStore) Set(ctx context.Context, key string, value []byte) error {
	if value == nil {
		// 删除
		return s.client.Del(ctx, key).Err()
	}
	return s.client.Set(ctx, key, value, s.ttl).Err()
}
