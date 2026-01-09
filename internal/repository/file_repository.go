package repository

import (
	"github.com/ashwinyue/next-ai/internal/model"
	"gorm.io/gorm"
)

// FileRepository 文件仓库
type FileRepository struct {
	db *gorm.DB
}

// NewFileRepository 创建文件仓库
func NewFileRepository(db *gorm.DB) *FileRepository {
	return &FileRepository{db: db}
}

// Create 创建文件记录
func (r *FileRepository) Create(file *model.StoredFile) error {
	return r.db.Create(file).Error
}

// GetByID 根据ID获取文件
func (r *FileRepository) GetByID(id string) (*model.StoredFile, error) {
	var file model.StoredFile
	err := r.db.Where("id = ?", id).First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

// Delete 删除文件记录
func (r *FileRepository) Delete(id string) error {
	return r.db.Delete(&model.StoredFile{}, "id = ?", id).Error
}
