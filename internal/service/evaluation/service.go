// Package evaluation 提供评估服务
package evaluation

import (
	"context"
	"fmt"
	"time"

	"github.com/ashwinyue/next-ai/internal/model"
	"github.com/ashwinyue/next-ai/internal/repository"
)

// Service 评估服务
type Service struct {
	repo *repository.Repositories
}

// NewService 创建评估服务
func NewService(repo *repository.Repositories) *Service {
	return &Service{
		repo: repo,
	}
}

// CreateEvaluationRequest 创建评估请求
type CreateEvaluationRequest struct {
	DatasetID       string `json:"dataset_id" binding:"required"`
	KnowledgeBaseID string `json:"knowledge_base_id" binding:"required"`
	ChatModelID     string `json:"chat_model_id"`
	RerankModelID   string `json:"rerank_model_id"`
}

// CreateEvaluation 创建评估任务
func (s *Service) CreateEvaluation(ctx context.Context, req *CreateEvaluationRequest) (*model.EvaluationTask, error) {
	// 验证知识库存在
	_, err := s.repo.Knowledge.GetKnowledgeBaseByID(req.KnowledgeBaseID)
	if err != nil {
		return nil, fmt.Errorf("知识库不存在: %w", err)
	}

	// 验证数据集存在
	_, err = s.repo.Dataset.GetByID(req.DatasetID)
	if err != nil {
		return nil, fmt.Errorf("数据集不存在: %w", err)
	}

	task := &model.EvaluationTask{
		DatasetID:       req.DatasetID,
		KnowledgeBaseID: req.KnowledgeBaseID,
		ChatModelID:     req.ChatModelID,
		RerankModelID:   req.RerankModelID,
		Status:          model.EvaluationStatusPending,
		Progress:        0,
	}

	if err := s.repo.Evaluation.Create(task); err != nil {
		return nil, fmt.Errorf("failed to create evaluation task: %w", err)
	}

	return task, nil
}

// GetEvaluationResult 获取评估结果
func (s *Service) GetEvaluationResult(ctx context.Context, taskID string) (*model.EvaluationTask, error) {
	task, err := s.repo.Evaluation.GetByID(taskID)
	if err != nil {
		return nil, fmt.Errorf("task not found: %w", err)
	}
	return task, nil
}

// RunEvaluation 执行评估任务（异步）
func (s *Service) RunEvaluation(ctx context.Context, taskID string) error {
	task, err := s.repo.Evaluation.GetByID(taskID)
	if err != nil {
		return fmt.Errorf("task not found: %w", err)
	}

	// 更新状态为处理中
	if err := s.repo.Evaluation.UpdateProgress(taskID, 0, model.EvaluationStatusRunning); err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	// 异步执行评估
	go func() {
		// 模拟评估过程
		s.runEvaluationAsync(context.Background(), task)
	}()

	return nil
}

// runEvaluationAsync 异步执行评估
func (s *Service) runEvaluationAsync(ctx context.Context, task *model.EvaluationTask) {
	// 更新进度
	s.repo.Evaluation.UpdateProgress(task.ID, 50, model.EvaluationStatusRunning)

	// 模拟评估延迟
	time.Sleep(1 * time.Second)

	// 计算结果（简化版）
	result := &model.EvaluationResult{
		Precision:       0.85,
		Recall:          0.78,
		F1Score:         0.81,
		AvgResponseTime: 1200,
		TotalCorrect:    17,
		TotalWrong:      3,
	}

	// 完成任务
	s.repo.Evaluation.UpdateResult(task.ID, result)
}

// ListEvaluationTasks 列出评估任务
func (s *Service) ListEvaluationTasks(ctx context.Context, knowledgeBaseID string, limit, offset int) ([]*model.EvaluationTask, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	tasks, total, err := s.repo.Evaluation.List(knowledgeBaseID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list tasks: %w", err)
	}

	return tasks, total, nil
}

// DeleteEvaluationTask 删除评估任务
func (s *Service) DeleteEvaluationTask(ctx context.Context, taskID string) error {
	if err := s.repo.Evaluation.Delete(taskID); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	return nil
}

// CancelEvaluation 取消评估任务
func (s *Service) CancelEvaluation(ctx context.Context, taskID string) error {
	task, err := s.repo.Evaluation.GetByID(taskID)
	if err != nil {
		return fmt.Errorf("task not found: %w", err)
	}

	// 只能取消运行中的任务
	if task.Status != model.EvaluationStatusRunning {
		return fmt.Errorf("can only cancel running tasks")
	}

	// 标记为失败状态表示取消
	if err := s.repo.Evaluation.UpdateProgress(taskID, task.Progress, model.EvaluationStatusFailed); err != nil {
		return fmt.Errorf("failed to cancel task: %w", err)
	}

	return nil
}

// CalculateMetrics 计算评估指标
func (s *Service) CalculateMetrics(ctx context.Context, taskID string) (*model.EvaluationResult, error) {
	task, err := s.repo.Evaluation.GetByID(taskID)
	if err != nil {
		return nil, fmt.Errorf("task not found: %w", err)
	}

	// 如果任务已有结果，直接返回
	if task.Result != nil {
		return task.Result, nil
	}

	// 简化版：返回默认指标
	return &model.EvaluationResult{
		Precision: 0.0,
		Recall:    0.0,
		F1Score:   0.0,
	}, nil
}

// UpdateTaskProgress 更新任务进度
func (s *Service) UpdateTaskProgress(ctx context.Context, taskID string, progress int, status model.EvaluationTaskStatus) error {
	if progress < 0 || progress > 100 {
		return fmt.Errorf("progress must be between 0 and 100")
	}

	if err := s.repo.Evaluation.UpdateProgress(taskID, progress, status); err != nil {
		return fmt.Errorf("failed to update progress: %w", err)
	}

	return nil
}

// CompleteTask 完成任务并保存结果
func (s *Service) CompleteTask(ctx context.Context, taskID string, result *model.EvaluationResult) error {
	if err := s.repo.Evaluation.UpdateResult(taskID, result); err != nil {
		return fmt.Errorf("failed to complete task: %w", err)
	}
	return nil
}

// GetTasksByStatus 根据状态获取任务列表
func (s *Service) GetTasksByStatus(ctx context.Context, status model.EvaluationTaskStatus) ([]*model.EvaluationTask, error) {
	tasks, err := s.repo.Evaluation.GetByStatus(status)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks by status: %w", err)
	}
	return tasks, nil
}
