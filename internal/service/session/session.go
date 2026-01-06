package session

import (
	"context"
	"sync"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/redis/go-redis/v9"
)

// Manager 会话管理器（简化版）
type Manager struct {
	mu     sync.RWMutex
	memory map[string]*Session
	redis  *redis.Client
}

// Session 会话状态
type Session struct {
	ID        string
	Messages  []*schema.Message
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewManager 创建会话管理器
func NewManager(redisClient *redis.Client) *Manager {
	return &Manager{
		memory: make(map[string]*Session),
		redis:  redisClient,
	}
}

// Get 获取会话
func (m *Manager) Get(ctx context.Context, sessionID string) (*Session, error) {
	m.mu.RLock()
	sess, ok := m.memory[sessionID]
	m.mu.RUnlock()

	if ok {
		return sess, nil
	}

	// 从 Redis 获取（可选）
	// TODO: 实现持久化加载

	// 创建新会话
	sess = &Session{
		ID:        sessionID,
		Messages:  []*schema.Message{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	m.mu.Lock()
	m.memory[sessionID] = sess
	m.mu.Unlock()

	return sess, nil
}

// Append 追加消息
func (m *Manager) Append(ctx context.Context, sessionID string, msg *schema.Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	sess, ok := m.memory[sessionID]
	if !ok {
		var err error
		sess, err = m.Get(ctx, sessionID)
		if err != nil {
			return err
		}
	}

	sess.Messages = append(sess.Messages, msg)
	sess.UpdatedAt = time.Now()

	// TODO: 同步到 Redis
	return nil
}

// GetHistory 获取历史消息
func (m *Manager) GetHistory(ctx context.Context, sessionID string) ([]*schema.Message, error) {
	sess, err := m.Get(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	return sess.Messages, nil
}

// Clear 清空会话
func (m *Manager) Clear(ctx context.Context, sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.memory, sessionID)

	// TODO: 从 Redis 删除
	return nil
}
