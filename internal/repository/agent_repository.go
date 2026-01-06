package repository

import (
	"github.com/ashwinyue/next-ai/internal/model"
	"gorm.io/gorm"
)

// AgentRepository Agent数据访问
type AgentRepository struct {
	db *gorm.DB
}

// NewAgentRepository 创建Agent仓库
func NewAgentRepository(db *gorm.DB) *AgentRepository {
	return &AgentRepository{db: db}
}

// Create 创建Agent
func (r *AgentRepository) Create(agent *model.Agent) error {
	return r.db.Create(agent).Error
}

// GetByID 获取Agent
func (r *AgentRepository) GetByID(id string) (*model.Agent, error) {
	var agent model.Agent
	err := r.db.Where("id = ?", id).First(&agent).Error
	if err != nil {
		return nil, err
	}
	return &agent, nil
}

// GetByName 获取Agent
func (r *AgentRepository) GetByName(name string) (*model.Agent, error) {
	var agent model.Agent
	err := r.db.Where("name = ?", name).First(&agent).Error
	if err != nil {
		return nil, err
	}
	return &agent, nil
}

// List 列出Agent
func (r *AgentRepository) List(offset, limit int) ([]*model.Agent, error) {
	var agents []*model.Agent
	err := r.db.Order("created_at DESC").Offset(offset).Limit(limit).Find(&agents).Error
	return agents, err
}

// ListActive 列出活跃Agent
func (r *AgentRepository) ListActive() ([]*model.Agent, error) {
	var agents []*model.Agent
	err := r.db.Where("is_active = ?", true).Order("created_at DESC").Find(&agents).Error
	return agents, err
}

// Update 更新Agent
func (r *AgentRepository) Update(agent *model.Agent) error {
	return r.db.Save(agent).Error
}

// Delete 删除Agent
func (r *AgentRepository) Delete(id string) error {
	return r.db.Delete(&model.Agent{}, "id = ?", id).Error
}
