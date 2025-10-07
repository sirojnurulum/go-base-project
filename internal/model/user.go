package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user account in the system.
type User struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:uuid_generate_v7()" json:"id"`
	Username     string         `gorm:"type:varchar(50);unique;not null" json:"username"`
	Email        string         `gorm:"type:varchar(255);unique" json:"email"`
	Password     string         `gorm:"type:varchar(255)" json:"-"`         // Don't expose password in JSON
	RoleID       *uuid.UUID     `gorm:"type:uuid" json:"role_id"`           // Foreign key for RBAC system
	GoogleID     *string        `gorm:"type:varchar(255)" json:"google_id"` // Changed to pointer for proper NULL handling
	AvatarURL    string         `gorm:"type:text" json:"avatar_url"`
	AuthProvider string         `gorm:"type:varchar(20);default:'local'" json:"auth_provider"` // Track authentication method
	IsActive     bool           `gorm:"default:true" json:"is_active"`
	Name         string         `gorm:"type:varchar(255)" json:"name"`
	CreatedAt    time.Time      `gorm:"default:now()" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"default:now()" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Role          *Role          `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	Organizations []Organization `gorm:"many2many:user_organizations;" json:"organizations,omitempty"`
}

// TableName sets the table name for User
func (User) TableName() string {
	return "users"
}

// BeforeCreate GORM hook to ensure proper data handling before insert
func (u *User) BeforeCreate(tx *gorm.DB) error {
	// Set default auth provider if empty
	if u.AuthProvider == "" {
		u.AuthProvider = "local"
	}

	// For local users, ensure GoogleID is nil, not empty string
	if u.AuthProvider == "local" && u.GoogleID != nil && *u.GoogleID == "" {
		u.GoogleID = nil
	}

	return nil
}

// BeforeUpdate GORM hook to ensure proper data handling before update
func (u *User) BeforeUpdate(tx *gorm.DB) error {
	// For local users, ensure GoogleID is nil, not empty string
	if u.AuthProvider == "local" && u.GoogleID != nil && *u.GoogleID == "" {
		u.GoogleID = nil
	}

	return nil
}
