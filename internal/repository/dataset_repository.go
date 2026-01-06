package repository

import (
	"github.com/ashwinyue/next-ai/internal/model"
	"gorm.io/gorm"
)

// DatasetRepository 数据集仓库
type DatasetRepository struct {
	db *gorm.DB
}

// NewDatasetRepository 创建数据集仓库
func NewDatasetRepository(db *gorm.DB) *DatasetRepository {
	return &DatasetRepository{db: db}
}

// Create 创建数据集
func (r *DatasetRepository) Create(dataset *model.Dataset) error {
	return r.db.Create(dataset).Error
}

// GetByID 根据ID获取数据集
func (r *DatasetRepository) GetByID(id string) (*model.Dataset, error) {
	var dataset model.Dataset
	err := r.db.Where("id = ?", id).First(&dataset).Error
	if err != nil {
		return nil, err
	}
	return &dataset, nil
}

// List 列出数据集
func (r *DatasetRepository) List(tenantID string, offset, limit int) ([]*model.Dataset, error) {
	var datasets []*model.Dataset
	query := r.db.Order("created_at DESC").Offset(offset).Limit(limit)
	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}
	err := query.Find(&datasets).Error
	return datasets, err
}

// Update 更新数据集
func (r *DatasetRepository) Update(dataset *model.Dataset) error {
	return r.db.Save(dataset).Error
}

// Delete 删除数据集
func (r *DatasetRepository) Delete(id string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 删除关联的 QA 对
		if err := tx.Delete(&model.QAPair{}, "dataset_id = ?", id).Error; err != nil {
			return err
		}
		// 删除数据集
		return tx.Delete(&model.Dataset{}, "id = ?", id).Error
	})
}

// ========== QA 对操作 ==========

// CreateQAPair 创建 QA 对
func (r *DatasetRepository) CreateQAPair(pair *model.QAPair) error {
	return r.db.Create(pair).Error
}

// CreateQAPairsBatch 批量创建 QA 对
func (r *DatasetRepository) CreateQAPairsBatch(pairs []*model.QAPair) error {
	if len(pairs) == 0 {
		return nil
	}
	return r.db.Create(&pairs).Error
}

// GetQAPairs 获取数据集的所有 QA 对
func (r *DatasetRepository) GetQAPairs(datasetID string) ([]*model.QAPair, error) {
	var pairs []*model.QAPair
	err := r.db.Where("dataset_id = ?", datasetID).Find(&pairs).Error
	return pairs, err
}

// GetQAPairByID 获取单个 QA 对
func (r *DatasetRepository) GetQAPairByID(id string) (*model.QAPair, error) {
	var pair model.QAPair
	err := r.db.Where("id = ?", id).First(&pair).Error
	if err != nil {
		return nil, err
	}
	return &pair, nil
}

// DeleteQAPairs 删除数据集的所有 QA 对
func (r *DatasetRepository) DeleteQAPairs(datasetID string) error {
	return r.db.Delete(&model.QAPair{}, "dataset_id = ?", datasetID).Error
}
