package model

import "time"

// FAQ 常见问题
type FAQ struct {
	ID          string    `gorm:"primaryKey;size:36"`
	Question    string    `gorm:"type:text;uniqueIndex"`
	Answer      string    `gorm:"type:text"`
	Category    string    `gorm:"size:100;index"`
	Priority    int       `gorm:"default:0"`
	HitCount    int       `gorm:"default:0"`
	IsActive    bool      `gorm:"index;default:true"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

func (FAQ) TableName() string {
	return "faqs"
}
