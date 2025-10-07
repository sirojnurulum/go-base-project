package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Permission represents an individual permission for an action.
type Permission struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:uuid_generate_v7()" json:"id"`
	Name        string         `gorm:"type:varchar(100);unique;not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	CreatedAt   time.Time      `gorm:"default:now()" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"default:now()" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName sets the table name for Permission
func (Permission) TableName() string {
	return "permissions"
}
