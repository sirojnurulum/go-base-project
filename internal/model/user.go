package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user account in the system.
type User struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:uuid_generate_v7()" json:"id"`
	Username  string         `gorm:"type:varchar(50);unique;not null" json:"username"`
	Email     string         `gorm:"type:varchar(255);unique" json:"email"`
	Password  string         `gorm:"type:varchar(255)" json:"-"` // Jangan ekspos password di JSON
	RoleID    *uuid.UUID     `gorm:"type:uuid" json:"role_id"`   // Foreign key, pointer agar bisa NULL
	Role      *Role          `gorm:"foreignKey:RoleID" json:"role"`
	GoogleID  string         `gorm:"type:varchar(255);unique" json:"google_id"`
	AvatarURL string         `gorm:"type:text" json:"avatar_url"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
