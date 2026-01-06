package repository

import (
	"github.com/ashwinyue/next-ai/internal/model"
	"gorm.io/gorm"
)

// ChatRepository 聊天数据访问
type ChatRepository struct {
	db *gorm.DB
}

// NewChatRepository 创建聊天仓库
func NewChatRepository(db *gorm.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

// CreateSession 创建会话
func (r *ChatRepository) CreateSession(session *model.ChatSession) error {
	return r.db.Create(session).Error
}

// GetSessionByID 获取会话
func (r *ChatRepository) GetSessionByID(id string) (*model.ChatSession, error) {
	var session model.ChatSession
	err := r.db.Preload("Messages").Where("id = ?", id).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// ListSessions 列出会话
func (r *ChatRepository) ListSessions(userID string, offset, limit int) ([]*model.ChatSession, error) {
	var sessions []*model.ChatSession
	query := r.db.Order("created_at DESC").Offset(offset).Limit(limit)
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	err := query.Find(&sessions).Error
	return sessions, err
}

// UpdateSession 更新会话
func (r *ChatRepository) UpdateSession(session *model.ChatSession) error {
	return r.db.Save(session).Error
}

// DeleteSession 删除会话
func (r *ChatRepository) DeleteSession(id string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&model.ChatMessage{}, "session_id = ?", id).Error; err != nil {
			return err
		}
		return tx.Delete(&model.ChatSession{}, "id = ?", id).Error
	})
}

// CreateMessage 创建消息
func (r *ChatRepository) CreateMessage(msg *model.ChatMessage) error {
	return r.db.Create(msg).Error
}

// GetMessagesBySessionID 获取会话消息
func (r *ChatRepository) GetMessagesBySessionID(sessionID string) ([]*model.ChatMessage, error) {
	var messages []*model.ChatMessage
	err := r.db.Where("session_id = ?", sessionID).Order("created_at ASC").Find(&messages).Error
	return messages, err
}

// GetRecentMessagesBySession 获取会话最近的 N 条消息
func (r *ChatRepository) GetRecentMessagesBySession(sessionID string, limit int) ([]*model.ChatMessage, error) {
	var messages []*model.ChatMessage
	err := r.db.Where("session_id = ?", sessionID).
		Order("created_at DESC").
		Limit(limit).
		Find(&messages).Error
	return messages, err
}

// GetMessagesBySessionBeforeTime 获取会话指定时间之前的消息
func (r *ChatRepository) GetMessagesBySessionBeforeTime(sessionID string, beforeTime string, limit int) ([]*model.ChatMessage, error) {
	var messages []*model.ChatMessage
	query := r.db.Where("session_id = ?", sessionID)
	if beforeTime != "" {
		query = query.Where("created_at < ?", beforeTime)
	}
	err := query.Order("created_at DESC").Limit(limit).Find(&messages).Error
	return messages, err
}

// GetMessageByID 获取单条消息
func (r *ChatRepository) GetMessageByID(messageID string) (*model.ChatMessage, error) {
	var message model.ChatMessage
	err := r.db.Where("id = ?", messageID).First(&message).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

// DeleteMessage 删除消息
func (r *ChatRepository) DeleteMessage(messageID string) error {
	return r.db.Delete(&model.ChatMessage{}, "id = ?", messageID).Error
}
