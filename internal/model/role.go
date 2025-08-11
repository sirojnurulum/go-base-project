package model

import (
	"time"

	"github.com/google/uuid"
)

// Role merepresentasikan peran pengguna dalam sistem.
type Role struct {
	ID          uuid.UUID    `gorm:"type:uuid;primary_key;default:uuid_generate_v7()" json:"id"`
	Name        string       `gorm:"type:varchar(50);unique;not null" json:"name"`
	Description string       `gorm:"type:text" json:"description"`
	Level       int          `gorm:"type:int;not null;default:0" json:"level"`          // Role hierarchy level
	IsSystem    bool         `gorm:"type:bool;not null;default:false" json:"is_system"` // System vs custom role
	IsActive    bool         `gorm:"type:bool;not null;default:true" json:"is_active"`  // Active status
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions"`    // Relasi Many-to-Many
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}
