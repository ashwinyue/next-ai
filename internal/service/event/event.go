// Package event provides event handling for agent execution
package event

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// EventType 事件类型
type EventType string

const (
	// EventStart 开始事件
	EventStart EventType = "start"
	// EventEnd 结束事件
	EventEnd EventType = "end"
	// EventError 错误事件
	EventError EventType = "error"
	// EventTool 工具调用事件
	EventTool EventType = "tool"
	// EventMessage 消息事件
	EventMessage EventType = "message"
)

// Event Agent 事件
type Event struct {
	ID        string                 `json:"id"`
	SessionID string                 `json:"session_id"`
	AgentID   string                 `json:"agent_id"`
	EventType EventType              `json:"event_type"`
	Component string                 `json:"component,omitempty"`
	Name      string                 `json:"name,omitempty"`
	Data      string                 `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Store 事件存储接口
type Store interface {
	SaveEvent(ctx context.Context, evt *Event) error
	GetEvents(ctx context.Context, sessionID string) ([]*Event, error)
	GetEventsByType(ctx context.Context, sessionID string, eventType EventType) ([]*Event, error)
	ClearEvents(ctx context.Context, sessionID string) error
}

// Handler 事件处理器接口
type Handler interface {
	Handle(ctx context.Context, evt *Event) error
}

// EventHandlerFunc 函数类型的事件处理器
type EventHandlerFunc func(ctx context.Context, evt *Event) error

// Handle 实现 Handler 接口
func (f EventHandlerFunc) Handle(ctx context.Context, evt *Event) error {
	return f(ctx, evt)
}

// Service 事件服务
type Service struct {
	sessionID string
	agentID   string
	store     Store
}

// NewService 创建事件服务
func NewService(sessionID, agentID string, store Store) (*Service, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("sessionID cannot be empty")
	}
	if agentID == "" {
		return nil, fmt.Errorf("agentID cannot be empty")
	}

	return &Service{
		sessionID: sessionID,
		agentID:   agentID,
		store:     store,
	}, nil
}

// Handle 处理事件
func (s *Service) Handle(ctx context.Context, evt *Event) error {
	if s.store != nil {
		return s.store.SaveEvent(ctx, evt)
	}
	return nil
}

// OnStart 开始时触发
func (s *Service) OnStart(ctx context.Context) error {
	evt := &Event{
		ID:        s.generateEventID(),
		SessionID: s.sessionID,
		AgentID:   s.agentID,
		EventType: EventStart,
		Data:      "session started",
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	return s.Handle(ctx, evt)
}

// OnEnd 结束时触发
func (s *Service) OnEnd(ctx context.Context, result string) error {
	evt := &Event{
		ID:        s.generateEventID(),
		SessionID: s.sessionID,
		AgentID:   s.agentID,
		EventType: EventEnd,
		Data:      fmt.Sprintf("result: %s", result),
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	return s.Handle(ctx, evt)
}

// OnError 错误时触发
func (s *Service) OnError(ctx context.Context, err error) error {
	evt := &Event{
		ID:        s.generateEventID(),
		SessionID: s.sessionID,
		AgentID:   s.agentID,
		EventType: EventError,
		Data:      fmt.Sprintf("error: %v", err),
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	return s.Handle(ctx, evt)
}

// OnToolCall 工具调用时触发
func (s *Service) OnToolCall(ctx context.Context, toolName string, input interface{}) error {
	evt := &Event{
		ID:        s.generateEventID(),
		SessionID: s.sessionID,
		AgentID:   s.agentID,
		EventType: EventTool,
		Component: "tool",
		Name:      toolName,
		Data:      fmt.Sprintf("tool_call: %s with input: %v", toolName, input),
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	return s.Handle(ctx, evt)
}

// generateEventID 生成事件 ID
func (s *Service) generateEventID() string {
	return "evt_" + uuid.New().String()
}

// ========== EventBus ==========

// EventBus 事件总线
type EventBus struct {
	store       Store
	subscribers map[string][]Handler
	mu          sync.RWMutex
}

// NewEventBus 创建事件总线
func NewEventBus(store Store) *EventBus {
	return &EventBus{
		store:       store,
		subscribers: make(map[string][]Handler),
	}
}

// SaveEvent 保存事件
func (b *EventBus) SaveEvent(ctx context.Context, evt *Event) error {
	if b.store != nil {
		return b.store.SaveEvent(ctx, evt)
	}
	return nil
}

// GetEvents 获取会话的所有事件
func (b *EventBus) GetEvents(ctx context.Context, sessionID string) ([]*Event, error) {
	if b.store != nil {
		return b.store.GetEvents(ctx, sessionID)
	}
	return []*Event{}, nil
}

// GetEventsByType 获取特定类型的事件
func (b *EventBus) GetEventsByType(ctx context.Context, sessionID string, eventType EventType) ([]*Event, error) {
	if b.store != nil {
		return b.store.GetEventsByType(ctx, sessionID, eventType)
	}
	return []*Event{}, nil
}

// ClearEvents 清空会话事件
func (b *EventBus) ClearEvents(ctx context.Context, sessionID string) error {
	if b.store != nil {
		return b.store.ClearEvents(ctx, sessionID)
	}
	return nil
}

// Subscribe 订阅事件
func (b *EventBus) Subscribe(ctx context.Context, handler Handler) error {
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.subscribers["*"] = append(b.subscribers["*"], handler)

	return nil
}

// Unsubscribe 取消订阅
func (b *EventBus) Unsubscribe(ctx context.Context, handler Handler) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if handlers, ok := b.subscribers["*"]; ok {
		for i, h := range handlers {
			if h == handler {
				b.subscribers["*"] = append(handlers[:i], handlers[i+1:]...)
				break
			}
		}
	}

	return nil
}

// Publish 发布事件
func (b *EventBus) Publish(ctx context.Context, evt *Event) error {
	// 1. 保存事件
	if err := b.SaveEvent(ctx, evt); err != nil {
		return err
	}

	// 2. 通知订阅者
	b.mu.RLock()
	defer b.mu.RUnlock()

	// 通知所有订阅者
	if handlers, ok := b.subscribers["*"]; ok {
		for _, handler := range handlers {
			// 异步处理
			go func(h Handler) {
				_ = h.Handle(ctx, evt)
			}(handler)
		}
	}

	return nil
}

// ========== SessionManager ==========

// SessionManager 会话事件管理器
type SessionManager struct {
	bus    *EventBus
	agents map[string]string // sessionID -> agentID
	mu     sync.RWMutex
}

// NewSessionManager 创建会话事件管理器
func NewSessionManager(bus *EventBus) *SessionManager {
	return &SessionManager{
		bus:    bus,
		agents: make(map[string]string),
	}
}

// RegisterSession 注册会话
func (m *SessionManager) RegisterSession(ctx context.Context, sessionID, agentID string) (*Service, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.agents[sessionID] = agentID

	svc, err := NewService(sessionID, agentID, nil)
	if err != nil {
		return nil, err
	}

	// 订阅事件总线
	if err := m.bus.Subscribe(ctx, svc); err != nil {
		return nil, err
	}

	return svc, nil
}

// UnregisterSession 注销会话
func (m *SessionManager) UnregisterSession(ctx context.Context, sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.agents, sessionID)

	// 清理事件
	return m.bus.ClearEvents(ctx, sessionID)
}

// GetSessionEvents 获取会话事件
func (m *SessionManager) GetSessionEvents(ctx context.Context, sessionID string) ([]*Event, error) {
	return m.bus.GetEvents(ctx, sessionID)
}

// Publish 发布会话事件
func (m *SessionManager) Publish(ctx context.Context, evt *Event) error {
	return m.bus.Publish(ctx, evt)
}
