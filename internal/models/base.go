package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ErrRecordNotFound is exported for use in test mocks and service layer.
var ErrRecordNotFound = gorm.ErrRecordNotFound

// BaseModel provides common fields for all entities
type BaseModel struct {
	ID        string         `gorm:"type:text;primaryKey" json:"id"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate generates UUID if not set
func (b *BaseModel) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	return nil
}
