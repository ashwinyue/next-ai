// Package rewrite provides query rewriting for multi-turn conversations
package rewrite

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// Service 查询重写服务
type Service struct {
	chatModel model.ChatModel
	config    *Config
}

// Config 重写服务配置
type Config struct {
	// SystemPrompt 系统提示词
	SystemPrompt string `json:"system_prompt"`
	// UserPromptTemplate 用户提示词模板
	UserPromptTemplate string `json:"user_prompt_template"`
	// Enabled 是否启用
	Enabled bool `json:"enabled"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		SystemPrompt: `You are a helpful assistant that rewrites user queries for better information retrieval.
Your task is to rewrite the user's latest question to make it more clear and complete,
incorporating relevant context from the conversation history when necessary.

Rules:
1. If the latest question is complete and clear, return it as-is.
2. If the latest question references previous context (pronouns, implied context, etc.), rewrite it to be standalone.
3. Keep the rewritten question concise but complete.
4. Return ONLY the rewritten question, no explanation.`,
		UserPromptTemplate: `Conversation history:
%s

Latest question: %s

Rewritten question:`,
		Enabled: true,
	}
}

// NewService 创建重写服务
func NewService(chatModel model.ChatModel, cfg *Config) *Service {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Service{
		chatModel: chatModel,
		config:    cfg,
	}
}

// RewriteQuery 重写查询
func (s *Service) RewriteQuery(ctx context.Context, query string, history []*schema.Message) (string, error) {
	if !s.config.Enabled {
		return query, nil
	}

	if s.chatModel == nil {
		return query, nil
	}

	// 没有历史或只有一条消息时直接返回
	if len(history) <= 1 {
		return query, nil
	}

	// 构建对话历史
	historyText := s.buildHistoryText(history)

	// 构建用户提示
	userPrompt := fmt.Sprintf(s.config.UserPromptTemplate, historyText, query)

	// 调用 LLM 重写
	messages := []*schema.Message{
		{
			Role:    schema.System,
			Content: s.config.SystemPrompt,
		},
		{
			Role:    schema.User,
			Content: userPrompt,
		},
	}

	resp, err := s.chatModel.Generate(ctx, messages)
	if err != nil {
		return query, nil
	}

	rewritten := strings.TrimSpace(resp.Content)
	if rewritten == "" {
		return query, nil
	}

	return rewritten, nil
}

// RewriteQueryWithMessages 使用消息格式重写查询
func (s *Service) RewriteQueryWithMessages(ctx context.Context, query string, messages []*schema.Message) (string, error) {
	if !s.config.Enabled {
		return query, nil
	}

	if s.chatModel == nil {
		return query, nil
	}

	if len(messages) == 0 {
		return query, nil
	}

	// 过滤出历史（排除当前查询）
	history := messages[:len(messages)-1]
	if len(history) == 0 {
		return query, nil
	}

	return s.RewriteQuery(ctx, query, history)
}

// buildHistoryText 构建历史文本
func (s *Service) buildHistoryText(history []*schema.Message) string {
	var sb strings.Builder

	for i, msg := range history {
		role := "User"
		switch msg.Role {
		case schema.Assistant:
			role = "Assistant"
		case schema.System:
			role = "System"
		}

		content := strings.TrimSpace(msg.Content)
		if content == "" {
			continue
		}

		sb.WriteString(fmt.Sprintf("%s: %s\n", role, content))

		// 限制历史长度
		if i >= 10 {
			break
		}
	}

	return sb.String()
}

// SetEnabled 设置启用状态
func (s *Service) SetEnabled(enabled bool) {
	s.config.Enabled = enabled
}

// IsEnabled 返回是否启用
func (s *Service) IsEnabled() bool {
	return s.config.Enabled
}

// ShouldRewrite 判断是否需要重写
func (s *Service) ShouldRewrite(query string, history []*schema.Message) bool {
	if !s.config.Enabled {
		return false
	}

	// 有历史时总是重写
	if len(history) > 1 {
		return true
	}

	// 检查常见模式
	queryLower := strings.ToLower(query)
	rewritePatterns := []string{
		" it ", " what ", " which ", " who ", " where ", " when ", " why ", " how ",
		"这个", "那个", "它", "他", "她", "它们", "怎么", "什么", "哪个", "哪里",
	}

	for _, pattern := range rewritePatterns {
		if strings.Contains(queryLower, pattern) {
			return true
		}
	}

	return false
}
