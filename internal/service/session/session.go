package session

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/redis/go-redis/v9"
)

const (
	// 会话在 Redis 中的过期时间（24小时）
	sessionTTL = 24 * time.Hour
	// Redis key 前缀
	sessionKeyPrefix = "session:"
)

// Manager 会话管理器（简化版）
type Manager struct {
	mu            sync.RWMutex
	memory        map[string]*Session
	activeStreams map[string]*ActiveStream // 活跃流控制
	redis         *redis.Client
}

// Session 会话状态
type Session struct {
	ID        string
	Messages  []*schema.Message
	CreatedAt time.Time
	UpdatedAt time.Time
}

// sessionData 会话数据（用于 Redis 存储）
type sessionData struct {
	ID        string        `json:"id"`
	Messages  []messageData `json:"messages"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

// messageData 消息数据（用于 Redis 存储）
type messageData struct {
	Role    string                 `json:"role"`
	Content string                 `json:"content"`
	Extra   map[string]interface{} `json:"extra,omitempty"`
}

// roleToSchema 将字符串角色转换为 schema.RoleType
func roleToSchema(role string) schema.RoleType {
	switch role {
	case "system":
		return schema.System
	case "assistant":
		return schema.Assistant
	case "user":
		return schema.User
	default:
		return schema.User
	}
}

// ActiveStream 活跃流
type ActiveStream struct {
	SessionID    string
	MessageID    string
	CancelFunc   context.CancelFunc
	Content      strings.Builder
	Done         bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
	partialCache []string // 部分响应缓存
	mu           sync.Mutex
}

// NewManager 创建会话管理器
func NewManager(redisClient *redis.Client) *Manager {
	return &Manager{
		memory:        make(map[string]*Session),
		activeStreams: make(map[string]*ActiveStream),
		redis:         redisClient,
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

	// 从 Redis 加载
	if m.redis != nil {
		if sess := m.loadFromRedis(ctx, sessionID); sess != nil {
			m.mu.Lock()
			m.memory[sessionID] = sess
			m.mu.Unlock()
			return sess, nil
		}
	}

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
		// 需要在锁内重新获取，避免并发问题
		m.mu.Unlock()
		var err error
		sess, err = m.Get(ctx, sessionID)
		m.mu.Lock()
		if err != nil {
			return err
		}
	}

	sess.Messages = append(sess.Messages, msg)
	sess.UpdatedAt = time.Now()

	// 同步到 Redis
	if m.redis != nil {
		if err := m.saveToRedis(ctx, sess); err != nil {
			// 记录错误但不影响主流程
			fmt.Printf("Warning: failed to save session to redis: %v\n", err)
		}
	}

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

	// 从 Redis 删除
	if m.redis != nil {
		key := sessionKeyPrefix + sessionID
		if err := m.redis.Del(ctx, key).Err(); err != nil {
			fmt.Printf("Warning: failed to delete session from redis: %v\n", err)
		}
	}

	return nil
}

// loadFromRedis 从 Redis 加载会话
func (m *Manager) loadFromRedis(ctx context.Context, sessionID string) *Session {
	key := sessionKeyPrefix + sessionID
	data, err := m.redis.Get(ctx, key).Result()
	if err != nil {
		return nil
	}

	var sd sessionData
	if err := json.Unmarshal([]byte(data), &sd); err != nil {
		return nil
	}

	// 转换消息
	messages := make([]*schema.Message, len(sd.Messages))
	for i, md := range sd.Messages {
		messages[i] = &schema.Message{
			Role:    roleToSchema(md.Role),
			Content: md.Content,
			Extra:   md.Extra,
		}
	}

	return &Session{
		ID:        sd.ID,
		Messages:  messages,
		CreatedAt: sd.CreatedAt,
		UpdatedAt: sd.UpdatedAt,
	}
}

// saveToRedis 保存会话到 Redis
func (m *Manager) saveToRedis(ctx context.Context, sess *Session) error {
	key := sessionKeyPrefix + sess.ID

	// 转换消息
	messages := make([]messageData, len(sess.Messages))
	for i, msg := range sess.Messages {
		messages[i] = messageData{
			Role:    string(msg.Role),
			Content: msg.Content,
			Extra:   msg.Extra,
		}
	}

	sd := sessionData{
		ID:        sess.ID,
		Messages:  messages,
		CreatedAt: sess.CreatedAt,
		UpdatedAt: sess.UpdatedAt,
	}

	data, err := json.Marshal(sd)
	if err != nil {
		return err
	}

	return m.redis.Set(ctx, key, data, sessionTTL).Err()
}

// ========== 流控制功能 ==========

// RegisterStream 注册活跃流
func (m *Manager) RegisterStream(sessionID, messageID string, cancelFunc context.CancelFunc) *ActiveStream {
	m.mu.Lock()
	defer m.mu.Unlock()

	stream := &ActiveStream{
		SessionID:    sessionID,
		MessageID:    messageID,
		CancelFunc:   cancelFunc,
		Content:      strings.Builder{},
		Done:         false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		partialCache: make([]string, 0),
	}

	key := sessionID + ":" + messageID
	m.activeStreams[key] = stream
	return stream
}

// UnregisterStream 注销流
func (m *Manager) UnregisterStream(sessionID, messageID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := sessionID + ":" + messageID
	if stream, ok := m.activeStreams[key]; ok {
		stream.Done = true
		delete(m.activeStreams, key)
	}
}

// GetStream 获取活跃流
func (m *Manager) GetStream(sessionID, messageID string) *ActiveStream {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := sessionID + ":" + messageID
	return m.activeStreams[key]
}

// StopStream 停止流
func (m *Manager) StopStream(sessionID, messageID string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := sessionID + ":" + messageID
	stream, ok := m.activeStreams[key]
	if !ok {
		return false
	}

	// 调用取消函数
	if stream.CancelFunc != nil {
		stream.CancelFunc()
	}

	stream.Done = true
	delete(m.activeStreams, key)
	return true
}

// AppendStreamContent 追加流内容
func (s *ActiveStream) AppendChunk(chunk string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Content.WriteString(chunk)
	s.partialCache = append(s.partialCache, chunk)
	s.UpdatedAt = time.Now()
}

// GetStreamContent 获取流内容
func (s *ActiveStream) GetContent() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Content.String()
}

// GetPartialChunks 获取部分缓存块
func (s *ActiveStream) GetPartialChunks() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string{}, s.partialCache...)
}

// IsDone 检查流是否完成
func (s *ActiveStream) IsDone() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Done
}

// MarkDone 标记流完成
func (s *ActiveStream) MarkDone() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Done = true
	s.UpdatedAt = time.Now()
}
