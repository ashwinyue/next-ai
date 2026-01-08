// Package repository 数据访问层
package repository

import (
	"github.com/ashwinyue/next-ai/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EvaluationTaskRepository 评估任务仓库
type EvaluationTaskRepository struct {
	db *gorm.DB
}

// NewEvaluationTaskRepository 创建评估任务仓库
func NewEvaluationTaskRepository(db *gorm.DB) *EvaluationTaskRepository {
	return &EvaluationTaskRepository{db: db}
}

// Create 创建评估任务
func (r *EvaluationTaskRepository) Create(task *model.EvaluationTask) error {
	if task.ID == "" {
		task.ID = uuid.New().String()
	}
	return r.db.Create(task).Error
}

// GetByID 根据 ID 获取评估任务
func (r *EvaluationTaskRepository) GetByID(id string) (*model.EvaluationTask, error) {
	var task model.EvaluationTask
	err := r.db.Where("id = ?", id).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// List 列出评估任务（支持筛选和分页）
func (r *EvaluationTaskRepository) List(knowledgeBaseID string, limit, offset int) ([]*model.EvaluationTask, int64, error) {
	var tasks []*model.EvaluationTask
	var total int64

	query := r.db.Model(&model.EvaluationTask{})
	if knowledgeBaseID != "" {
		query = query.Where("knowledge_base_id = ?", knowledgeBaseID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&tasks).Error
	return tasks, total, err
}

// Update 更新评估任务
func (r *EvaluationTaskRepository) Update(task *model.EvaluationTask) error {
	return r.db.Save(task).Error
}

// Delete 删除评估任务
func (r *EvaluationTaskRepository) Delete(id string) error {
	return r.db.Delete(&model.EvaluationTask{}, "id = ?", id).Error
}

// UpdateProgress 更新任务进度
func (r *EvaluationTaskRepository) UpdateProgress(id string, progress int, status model.EvaluationTaskStatus) error {
	return r.db.Model(&model.EvaluationTask{}).Where("id = ?", id).Updates(map[string]interface{}{
		"progress": progress,
		"status":   status,
	}).Error
}

// UpdateResult 更新任务结果
func (r *EvaluationTaskRepository) UpdateResult(id string, result *model.EvaluationResult) error {
	return r.db.Model(&model.EvaluationTask{}).Where("id = ?", id).Updates(map[string]interface{}{
		"result":   result,
		"status":   model.EvaluationStatusCompleted,
		"progress": 100,
	}).Error
}

// GetByStatus 根据状态获取任务列表
func (r *EvaluationTaskRepository) GetByStatus(status model.EvaluationTaskStatus) ([]*model.EvaluationTask, error) {
	var tasks []*model.EvaluationTask
	err := r.db.Where("status = ?", status).Find(&tasks).Error
	return tasks, err
}
