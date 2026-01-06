package model

import "time"

// Tool 工具定义
type Tool struct {
	ID          string    `gorm:"primaryKey;size:36"`
	Name        string    `gorm:"size:100;uniqueIndex"`
	DisplayName string    `gorm:"size:255"`
	Description string    `gorm:"type:text"`
	Type        string    `gorm:"size:20;index"` // builtin, custom
	Config      string    `gorm:"type:jsonb"`
	IsActive    bool      `gorm:"index;default:true"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

func (Tool) TableName() string {
	return "tools"
}
