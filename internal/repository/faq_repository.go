package repository

import (
	"github.com/ashwinyue/next-ai/internal/model"
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

// ========== 简单 FAQ (兼容旧版) ==========

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

// ========== FAQEntry (增强版) ==========

// CreateEntry 创建FAQ条目
func (r *FAQRepository) CreateEntry(entry *model.FAQEntry) error {
	return r.db.Create(entry).Error
}

// GetEntryByID 获取FAQ条目
func (r *FAQRepository) GetEntryByID(id string) (*model.FAQEntry, error) {
	var entry model.FAQEntry
	err := r.db.Where("id = ?", id).First(&entry).Error
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

// ListEntries 列出FAQ条目
func (r *FAQRepository) ListEntries(category string, offset, limit int) ([]*model.FAQEntry, int64, error) {
	var entries []*model.FAQEntry
	var total int64

	query := r.db.Model(&model.FAQEntry{})
	if category != "" {
		query = query.Where("category = ?", category)
	}
	query.Count(&total)

	err := query.Order("priority DESC, hit_count DESC, created_at DESC").
		Offset(offset).Limit(limit).
		Find(&entries).Error
	return entries, total, err
}

// ListEntriesByStatus 按状态列出FAQ条目
func (r *FAQRepository) ListEntriesByStatus(isEnabled bool, category string, offset, limit int) ([]*model.FAQEntry, int64, error) {
	var entries []*model.FAQEntry
	var total int64

	query := r.db.Model(&model.FAQEntry{}).Where("is_enabled = ?", isEnabled)
	if category != "" {
		query = query.Where("category = ?", category)
	}
	query.Count(&total)

	err := query.Order("priority DESC, hit_count DESC, created_at DESC").
		Offset(offset).Limit(limit).
		Find(&entries).Error
	return entries, total, err
}

// UpdateEntry 更新FAQ条目
func (r *FAQRepository) UpdateEntry(entry *model.FAQEntry) error {
	return r.db.Save(entry).Error
}

// DeleteEntry 删除FAQ条目
func (r *FAQRepository) DeleteEntry(id string) error {
	return r.db.Delete(&model.FAQEntry{}, "id = ?", id).Error
}

// DeleteEntries 批量删除FAQ条目
func (r *FAQRepository) DeleteEntries(ids []string) error {
	return r.db.Delete(&model.FAQEntry{}, "id IN ?", ids).Error
}

// SearchEntries 搜索FAQ条目
func (r *FAQRepository) SearchEntries(keyword string, limit int) ([]*model.FAQEntry, error) {
	var entries []*model.FAQEntry
	err := r.db.Where("is_enabled = ? AND (standard_question LIKE ? OR similar_questions LIKE ?)",
		true, "%"+keyword+"%", "%"+keyword+"%").
		Order("priority DESC, hit_count DESC").
		Limit(limit).
		Find(&entries).Error
	return entries, err
}

// UpdateEntryCategoryBatch 批量更新FAQ条目分类
func (r *FAQRepository) UpdateEntryCategoryBatch(updates map[string]string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for id, category := range updates {
			if err := tx.Model(&model.FAQEntry{}).Where("id = ?", id).
				Update("category", category).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// UpdateEntryFieldsBatch 批量更新FAQ条目字段
func (r *FAQRepository) UpdateEntryFieldsBatch(req *model.FAQEntryFieldsBatchUpdate) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 按ID更新
		for id, update := range req.ByID {
			updates := make(map[string]interface{})
			if update.IsEnabled != nil {
				updates["is_enabled"] = *update.IsEnabled
			}
			if update.IsRecommended != nil {
				updates["is_recommended"] = *update.IsRecommended
			}
			if update.Category != nil {
				updates["category"] = *update.Category
			}
			if len(updates) > 0 {
				if err := tx.Model(&model.FAQEntry{}).Where("id = ?", id).Updates(updates).Error; err != nil {
					return err
				}
			}
		}

		// 按分类更新
		for category, update := range req.ByCategory {
			updates := make(map[string]interface{})
			if update.IsEnabled != nil {
				updates["is_enabled"] = *update.IsEnabled
			}
			if update.IsRecommended != nil {
				updates["is_recommended"] = *update.IsRecommended
			}
			if update.Category != nil {
				updates["category"] = *update.Category
			}
			if len(updates) > 0 {
				query := tx.Model(&model.FAQEntry{}).Where("category = ?", category)
				if len(req.ExcludeIDs) > 0 {
					query = query.Where("id NOT IN ?", req.ExcludeIDs)
				}
				if err := query.Updates(updates).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// GetByContentHash 根据内容哈希获取FAQ条目
func (r *FAQRepository) GetByContentHash(hash string) (*model.FAQEntry, error) {
	var entry model.FAQEntry
	err := r.db.Where("content_hash = ?", hash).First(&entry).Error
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

// IncrementEntryHitCount 增加FAQ条目命中次数
func (r *FAQRepository) IncrementEntryHitCount(id string) error {
	return r.db.Model(&model.FAQEntry{}).Where("id = ?", id).
		UpdateColumn("hit_count", gorm.Expr("hit_count + ?", 1)).Error
}

// ========== FAQImportProgress ==========

// CreateImportProgress 创建导入进度
func (r *FAQRepository) CreateImportProgress(progress *model.FAQImportProgress) error {
	return r.db.Create(progress).Error
}

// GetImportProgress 获取导入进度
func (r *FAQRepository) GetImportProgress(taskID string) (*model.FAQImportProgress, error) {
	var progress model.FAQImportProgress
	err := r.db.Where("task_id = ?", taskID).First(&progress).Error
	if err != nil {
		return nil, err
	}
	return &progress, nil
}

// UpdateImportProgress 更新导入进度
func (r *FAQRepository) UpdateImportProgress(progress *model.FAQImportProgress) error {
	return r.db.Save(progress).Error
}

// DeleteImportProgress 删除导入进度
func (r *FAQRepository) DeleteImportProgress(taskID string) error {
	return r.db.Delete(&model.FAQImportProgress{}, "task_id = ?", taskID).Error
}
