package repository

import (
	"github.com/ashwinyue/next-rag/next-ai/internal/model"
	"gorm.io/gorm"
)

// FAQRepository FAQ数据访问
type FAQRepository struct {
	db *gorm.DB
}

// NewFAQRepository 创建FAQ仓库
func NewFAQRepository(db *gorm.DB) *FAQRepository {
	return &FAQRepository{db: db}
}

// Create 创建FAQ
func (r *FAQRepository) Create(faq *model.FAQ) error {
	return r.db.Create(faq).Error
}

// GetByID 获取FAQ
func (r *FAQRepository) GetByID(id string) (*model.FAQ, error) {
	var faq model.FAQ
	err := r.db.Where("id = ?", id).First(&faq).Error
	if err != nil {
		return nil, err
	}
	return &faq, nil
}

// List 列出FAQ
func (r *FAQRepository) List(category string, offset, limit int) ([]*model.FAQ, error) {
	var faqs []*model.FAQ
	query := r.db.Order("priority DESC, hit_count DESC").Offset(offset).Limit(limit)
	if category != "" {
		query = query.Where("category = ?", category)
	}
	err := query.Find(&faqs).Error
	return faqs, err
}

// ListActive 列出活跃FAQ
func (r *FAQRepository) ListActive(category string) ([]*model.FAQ, error) {
	var faqs []*model.FAQ
	query := r.db.Where("is_active = ?", true).Order("priority DESC, hit_count DESC")
	if category != "" {
		query = query.Where("category = ?", category)
	}
	err := query.Find(&faqs).Error
	return faqs, err
}

// Update 更新FAQ
func (r *FAQRepository) Update(faq *model.FAQ) error {
	return r.db.Save(faq).Error
}

// Delete 删除FAQ
func (r *FAQRepository) Delete(id string) error {
	return r.db.Delete(&model.FAQ{}, "id = ?", id).Error
}

// IncrementHitCount 增加命中次数
func (r *FAQRepository) IncrementHitCount(id string) error {
	return r.db.Model(&model.FAQ{}).Where("id = ?", id).
		UpdateColumn("hit_count", gorm.Expr("hit_count + ?", 1)).Error
}

// Search 搜索FAQ
func (r *FAQRepository) Search(keyword string, limit int) ([]*model.FAQ, error) {
	var faqs []*model.FAQ
	err := r.db.Where("is_active = ? AND question LIKE ?", true, "%"+keyword+"%").
		Order("priority DESC, hit_count DESC").
		Limit(limit).
		Find(&faqs).Error
	return faqs, err
}
