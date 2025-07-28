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
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions"` // Relasi Many-to-Many
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}
