package faq

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/ashwinyue/next-ai/internal/model"
	"github.com/ashwinyue/next-ai/internal/repository"
	"github.com/google/uuid"
)

// EntryService FAQ条目服务（增强版）
type EntryService struct {
	repo *repository.Repositories
}

// NewEntryService 创建FAQ条目服务
func NewEntryService(repo *repository.Repositories) *EntryService {
	return &EntryService{repo: repo}
}

// ========== 请求/响应类型 ==========

// CreateEntryRequest 创建FAQ条目请求
type CreateEntryRequest struct {
	StandardQuestion  string   `json:"standard_question" binding:"required"`
	SimilarQuestions  []string `json:"similar_questions"`
	NegativeQuestions []string `json:"negative_questions"`
	Answers           []string `json:"answers" binding:"required"`
	AnswerStrategy    string   `json:"answer_strategy"`
	Category          string   `json:"category"`
	IsEnabled         *bool    `json:"is_enabled"`
	IsRecommended     *bool    `json:"is_recommended"`
	Priority          int      `json:"priority"`
	Source            string   `json:"source"`
}

// UpdateEntryRequest 更新FAQ条目请求
type UpdateEntryRequest struct {
	StandardQuestion  string   `json:"standard_question"`
	SimilarQuestions  []string `json:"similar_questions"`
	NegativeQuestions []string `json:"negative_questions"`
	Answers           []string `json:"answers"`
	AnswerStrategy    string   `json:"answer_strategy"`
	Category          string   `json:"category"`
	IsEnabled         *bool    `json:"is_enabled"`
	IsRecommended     *bool    `json:"is_recommended"`
	Priority          int      `json:"priority"`
}

// ListEntriesRequest 列出FAQ条目请求
type ListEntriesRequest struct {
	Category  string `json:"category"`
	IsEnabled *bool  `json:"is_enabled"`
	Page      int    `json:"page"`
	Size      int    `json:"size"`
}

// ListEntriesResponse 列出FAQ条目响应
type ListEntriesResponse struct {
	Items []*model.FAQEntry `json:"items"`
	Total int64             `json:"total"`
	Page  int               `json:"page"`
	Size  int               `json:"size"`
}

// BatchUpsertRequest 批量导入请求
type BatchUpsertRequest struct {
	Entries []CreateEntryRequest `json:"entries" binding:"required"`
	Mode    string               `json:"mode" binding:"oneof=append replace"` // append, replace
}

// BatchUpsertResponse 批量导入响应
type BatchUpsertResponse struct {
	TaskID string `json:"task_id"`
}

// ExportEntry 导出的FAQ条目（简化格式）
type ExportEntry struct {
	StandardQuestion  string   `json:"standard_question"`
	SimilarQuestions  []string `json:"similar_questions"`
	Answers           []string `json:"answers"`
	Category          string   `json:"category"`
}

// ========== 核心方法 ==========

// CreateEntry 创建FAQ条目
func (s *EntryService) CreateEntry(ctx context.Context, req *CreateEntryRequest) (*model.FAQEntry, error) {
	// 规范化数据
	similarQuestions := normalizeStrings(req.SimilarQuestions)
	negativeQuestions := normalizeStrings(req.NegativeQuestions)
	answers := normalizeStrings(req.Answers)

	if len(answers) == 0 {
		return nil, fmt.Errorf("answers cannot be empty")
	}

	// 设置默认值
	answerStrategy := req.AnswerStrategy
	if answerStrategy == "" {
		answerStrategy = model.AnswerStrategyAll
	}

	isEnabled := true
	if req.IsEnabled != nil {
		isEnabled = *req.IsEnabled
	}

	isRecommended := false
	if req.IsRecommended != nil {
		isRecommended = *req.IsRecommended
	}

	// 创建FAQ条目
	entry := &model.FAQEntry{
		ID:       uuid.New().String(),
		Priority: req.Priority,
		Category: req.Category,
		Source:   req.Source,
	}

	// 设置问题
	entry.StandardQuestion = strings.TrimSpace(req.StandardQuestion)
	if err := entry.SetSimilarQuestions(similarQuestions); err != nil {
		return nil, fmt.Errorf("failed to set similar questions: %w", err)
	}
	if err := entry.SetNegativeQuestions(negativeQuestions); err != nil {
		return nil, fmt.Errorf("failed to set negative questions: %w", err)
	}
	if err := entry.SetAnswers(answers); err != nil {
		return nil, fmt.Errorf("failed to set answers: %w", err)
	}

	entry.AnswerStrategy = answerStrategy
	entry.IsEnabled = isEnabled
	entry.IsRecommended = isRecommended
	entry.ContentHash = calculateContentHash(entry.StandardQuestion, similarQuestions, negativeQuestions, answers)
	entry.Version = 1
	entry.CreatedAt = time.Now()
	entry.UpdatedAt = time.Now()

	if err := s.repo.FAQ.CreateEntry(entry); err != nil {
		return nil, fmt.Errorf("failed to create FAQ entry: %w", err)
	}

	return entry, nil
}

// GetEntry 获取FAQ条目
func (s *EntryService) GetEntry(ctx context.Context, id string) (*model.FAQEntry, error) {
	entry, err := s.repo.FAQ.GetEntryByID(id)
	if err != nil {
		return nil, fmt.Errorf("FAQ entry not found: %w", err)
	}
	// 增加命中次数
	_ = s.repo.FAQ.IncrementEntryHitCount(id)
	return entry, nil
}

// ListEntries 列出FAQ条目
func (s *EntryService) ListEntries(ctx context.Context, req *ListEntriesRequest) (*ListEntriesResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 || req.Size > 100 {
		req.Size = 20
	}

	offset := (req.Page - 1) * req.Size

	var entries []*model.FAQEntry
	var total int64
	var err error

	if req.IsEnabled != nil {
		entries, total, err = s.repo.FAQ.ListEntriesByStatus(*req.IsEnabled, req.Category, offset, req.Size)
	} else {
		entries, total, err = s.repo.FAQ.ListEntries(req.Category, offset, req.Size)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list FAQ entries: %w", err)
	}

	return &ListEntriesResponse{
		Items: entries,
		Total: total,
		Page:  req.Page,
		Size:  req.Size,
	}, nil
}

// UpdateEntry 更新FAQ条目
func (s *EntryService) UpdateEntry(ctx context.Context, id string, req *UpdateEntryRequest) (*model.FAQEntry, error) {
	entry, err := s.repo.FAQ.GetEntryByID(id)
	if err != nil {
		return nil, fmt.Errorf("FAQ entry not found: %w", err)
	}

	// 更新字段
	if req.StandardQuestion != "" {
		entry.StandardQuestion = strings.TrimSpace(req.StandardQuestion)
	}

	if req.SimilarQuestions != nil {
		if err := entry.SetSimilarQuestions(normalizeStrings(req.SimilarQuestions)); err != nil {
			return nil, fmt.Errorf("failed to set similar questions: %w", err)
		}
	}

	if req.NegativeQuestions != nil {
		if err := entry.SetNegativeQuestions(normalizeStrings(req.NegativeQuestions)); err != nil {
			return nil, fmt.Errorf("failed to set negative questions: %w", err)
		}
	}

	if req.Answers != nil {
		answers := normalizeStrings(req.Answers)
		if len(answers) == 0 {
			return nil, fmt.Errorf("answers cannot be empty")
		}
		if err := entry.SetAnswers(answers); err != nil {
			return nil, fmt.Errorf("failed to set answers: %w", err)
		}
	}

	if req.AnswerStrategy != "" {
		entry.AnswerStrategy = req.AnswerStrategy
	}

	if req.Category != "" {
		entry.Category = req.Category
	}

	if req.IsEnabled != nil {
		entry.IsEnabled = *req.IsEnabled
	}

	if req.IsRecommended != nil {
		entry.IsRecommended = *req.IsRecommended
	}

	if req.Priority > 0 {
		entry.Priority = req.Priority
	}

	// 重新计算哈希并增加版本号
	similarQuestions := entry.GetSimilarQuestions()
	negativeQuestions := entry.GetNegativeQuestions()
	answers := entry.GetAnswers()
	entry.ContentHash = calculateContentHash(entry.StandardQuestion, similarQuestions, negativeQuestions, answers)
	entry.Version++
	entry.UpdatedAt = time.Now()

	if err := s.repo.FAQ.UpdateEntry(entry); err != nil {
		return nil, fmt.Errorf("failed to update FAQ entry: %w", err)
	}

	return entry, nil
}

// DeleteEntry 删除FAQ条目
func (s *EntryService) DeleteEntry(ctx context.Context, id string) error {
	if err := s.repo.FAQ.DeleteEntry(id); err != nil {
		return fmt.Errorf("failed to delete FAQ entry: %w", err)
	}
	return nil
}

// DeleteEntries 批量删除FAQ条目
func (s *EntryService) DeleteEntries(ctx context.Context, ids []string) error {
	if err := s.repo.FAQ.DeleteEntries(ids); err != nil {
		return fmt.Errorf("failed to delete FAQ entries: %w", err)
	}
	return nil
}

// SearchEntries 搜索FAQ条目
func (s *EntryService) SearchEntries(ctx context.Context, keyword string, limit int) ([]*model.FAQEntry, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	entries, err := s.repo.FAQ.SearchEntries(keyword, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search FAQ entries: %w", err)
	}

	return entries, nil
}

// UpdateEntryCategoryBatch 批量更新FAQ条目分类
func (s *EntryService) UpdateEntryCategoryBatch(ctx context.Context, updates map[string]string) error {
	if err := s.repo.FAQ.UpdateEntryCategoryBatch(updates); err != nil {
		return fmt.Errorf("failed to update categories: %w", err)
	}
	return nil
}

// UpdateEntryFieldsBatch 批量更新FAQ条目字段
func (s *EntryService) UpdateEntryFieldsBatch(ctx context.Context, req *model.FAQEntryFieldsBatchUpdate) error {
	if err := s.repo.FAQ.UpdateEntryFieldsBatch(req); err != nil {
		return fmt.Errorf("failed to update fields: %w", err)
	}
	return nil
}

// ExportEntries 导出FAQ条目为CSV格式
func (s *EntryService) ExportEntries(ctx context.Context, category string) ([]byte, error) {
	// 获取所有启用的条目
	entries, _, err := s.repo.FAQ.ListEntriesByStatus(true, category, 0, 10000)
	if err != nil {
		return nil, fmt.Errorf("failed to get FAQ entries: %w", err)
	}

	// 转换为导出格式
	exportData := make([]ExportEntry, 0, len(entries))
	for _, entry := range entries {
		exportData = append(exportData, ExportEntry{
			StandardQuestion: entry.StandardQuestion,
			SimilarQuestions: entry.GetSimilarQuestions(),
			Answers:           entry.GetAnswers(),
			Category:          entry.Category,
		})
	}

	// 转换为JSON
	data, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal export data: %w", err)
	}

	return data, nil
}

// BatchUpsert 批量导入FAQ条目
func (s *EntryService) BatchUpsert(ctx context.Context, req *BatchUpsertRequest) (*BatchUpsertResponse, error) {
	taskID := uuid.New().String()

	// 创建导入进度记录
	progress := &model.FAQImportProgress{
		TaskID:    taskID,
		Status:    model.FAQImportStatusPending,
		Total:     len(req.Entries),
		Processed: 0,
		Progress:  0,
		Message:   "导入任务已创建",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.FAQ.CreateImportProgress(progress); err != nil {
		return nil, fmt.Errorf("failed to create import progress: %w", err)
	}

	// 异步处理导入
	go s.processBatchUpsert(taskID, req)

	return &BatchUpsertResponse{TaskID: taskID}, nil
}

// GetImportProgress 获取导入进度
func (s *EntryService) GetImportProgress(ctx context.Context, taskID string) (*model.FAQImportProgress, error) {
	progress, err := s.repo.FAQ.GetImportProgress(taskID)
	if err != nil {
		return nil, fmt.Errorf("import progress not found: %w", err)
	}
	return progress, nil
}

// processBatchUpsert 处理批量导入
func (s *EntryService) processBatchUpsert(taskID string, req *BatchUpsertRequest) {
	// 更新状态为处理中
	s.updateImportProgress(taskID, model.FAQImportStatusProcessing, 0, 0, "开始处理")

	successCount := 0
	failCount := 0

	for i, entryReq := range req.Entries {
		createReq := &CreateEntryRequest{
			StandardQuestion:  entryReq.StandardQuestion,
			SimilarQuestions:  entryReq.SimilarQuestions,
			NegativeQuestions: entryReq.NegativeQuestions,
			Answers:           entryReq.Answers,
			AnswerStrategy:    entryReq.AnswerStrategy,
			Category:          entryReq.Category,
			IsEnabled:         entryReq.IsEnabled,
			IsRecommended:     entryReq.IsRecommended,
			Priority:          entryReq.Priority,
			Source:            "batch_import",
		}

		_, err := s.CreateEntry(context.Background(), createReq)
		if err != nil {
			failCount++
			s.updateImportProgress(taskID, model.FAQImportStatusProcessing, 0, 0,
				fmt.Sprintf("处理第 %d 条记录失败: %v", i+1, err))
		} else {
			successCount++
		}

		// 更新进度
		progress := int(float64(i+1) / float64(len(req.Entries)) * 100)
		s.updateImportProgress(taskID, model.FAQImportStatusProcessing, progress, i+1,
			fmt.Sprintf("已处理 %d/%d 条记录", i+1, len(req.Entries)))
	}

	// 完成
	status := model.FAQImportStatusCompleted
	message := fmt.Sprintf("导入完成：成功 %d 条，失败 %d 条", successCount, failCount)
	if failCount > 0 {
		status = model.FAQImportStatusFailed
	}
	s.updateImportProgress(taskID, status, 100, len(req.Entries), message)
}

// updateImportProgress 更新导入进度
func (s *EntryService) updateImportProgress(taskID, status string, progressVal, processed int, message string) {
	progress := &model.FAQImportProgress{
		TaskID:    taskID,
		Status:    status,
		Progress:  progressVal,
		Processed: processed,
		Message:   message,
		UpdatedAt: time.Now(),
	}
	_ = s.repo.FAQ.UpdateImportProgress(progress)
}

// ========== 辅助方法 ==========

// normalizeStrings 规范化字符串数组
func normalizeStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	dedup := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, v := range values {
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		dedup = append(dedup, trimmed)
	}
	if len(dedup) == 0 {
		return nil
	}
	return dedup
}

// calculateContentHash 计算FAQ内容哈希
func calculateContentHash(standardQuestion string, similarQuestions, negativeQuestions, answers []string) string {
	// 对数组进行排序
	similarQs := make([]string, len(similarQuestions))
	copy(similarQs, similarQuestions)
	sort.Strings(similarQs)

	negativeQs := make([]string, len(negativeQuestions))
	copy(negativeQs, negativeQuestions)
	sort.Strings(negativeQs)

	ans := make([]string, len(answers))
	copy(ans, answers)
	sort.Strings(ans)

	// 构建哈希字符串
	var builder strings.Builder
	builder.WriteString(strings.TrimSpace(standardQuestion))
	builder.WriteString("|")
	builder.WriteString(strings.Join(similarQs, ","))
	builder.WriteString("|")
	builder.WriteString(strings.Join(negativeQs, ","))
	builder.WriteString("|")
	builder.WriteString(strings.Join(ans, ","))

	// 计算SHA256
	hash := sha256.Sum256([]byte(builder.String()))
	return hex.EncodeToString(hash[:])
}
