package faq

import (
	"context"
	"fmt"

	"github.com/ashwinyue/next-ai/internal/model"
	"github.com/ashwinyue/next-ai/internal/repository"
	"github.com/google/uuid"
)

// Service FAQ服务
type Service struct {
	repo *repository.Repositories
}

// NewService 创建FAQ服务
func NewService(repo *repository.Repositories) *Service {
	return &Service{repo: repo}
}

// CreateFAQRequest 创建FAQ请求
type CreateFAQRequest struct {
	Question string `json:"question" binding:"required"`
	Answer   string `json:"answer" binding:"required"`
	Category string `json:"category"`
	Priority int    `json:"priority"`
}

// CreateFAQ 创建FAQ
func (s *Service) CreateFAQ(ctx context.Context, req *CreateFAQRequest) (*model.FAQ, error) {
	faq := &model.FAQ{
		ID:       uuid.New().String(),
		Question: req.Question,
		Answer:   req.Answer,
		Category: req.Category,
		Priority: req.Priority,
		IsActive: true,
	}

	if err := s.repo.FAQ.Create(faq); err != nil {
		return nil, fmt.Errorf("failed to create FAQ: %w", err)
	}

	return faq, nil
}

// GetFAQ 获取FAQ
func (s *Service) GetFAQ(ctx context.Context, id string) (*model.FAQ, error) {
	faq, err := s.repo.FAQ.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("FAQ not found: %w", err)
	}
	// 增加命中次数
	_ = s.repo.FAQ.IncrementHitCount(id)
	return faq, nil
}

// ListFAQsRequest 列出FAQ请求
type ListFAQsRequest struct {
	Category string `json:"category"`
	Page     int    `json:"page"`
	Size     int    `json:"size"`
}

// ListFAQs 列出FAQ
func (s *Service) ListFAQs(ctx context.Context, req *ListFAQsRequest) ([]*model.FAQ, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 || req.Size > 100 {
		req.Size = 20
	}

	offset := (req.Page - 1) * req.Size
	return s.repo.FAQ.List(req.Category, offset, req.Size)
}

// ListActiveFAQs 列出活跃FAQ
func (s *Service) ListActiveFAQs(ctx context.Context, category string) ([]*model.FAQ, error) {
	return s.repo.FAQ.ListActive(category)
}

// UpdateFAQ 更新FAQ
func (s *Service) UpdateFAQ(ctx context.Context, id string, req *CreateFAQRequest) (*model.FAQ, error) {
	faq, err := s.repo.FAQ.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("FAQ not found: %w", err)
	}

	faq.Question = req.Question
	faq.Answer = req.Answer
	faq.Category = req.Category
	faq.Priority = req.Priority

	if err := s.repo.FAQ.Update(faq); err != nil {
		return nil, fmt.Errorf("failed to update FAQ: %w", err)
	}

	return faq, nil
}

// DeleteFAQ 删除FAQ
func (s *Service) DeleteFAQ(ctx context.Context, id string) error {
	if err := s.repo.FAQ.Delete(id); err != nil {
		return fmt.Errorf("failed to delete FAQ: %w", err)
	}
	return nil
}

// SearchFAQs 搜索FAQ
func (s *Service) SearchFAQs(ctx context.Context, keyword string, limit int) ([]*model.FAQ, error) {
	if limit <= 0 || limit > 20 {
		limit = 10
	}
	return s.repo.FAQ.Search(keyword, limit)
}
