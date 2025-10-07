package model

import (
	"time"

	"github.com/google/uuid"
)

// UserOrganization represents the many-to-many relationship between users and organizations
type UserOrganization struct {
	UserID         uuid.UUID  `gorm:"type:uuid;primaryKey" json:"user_id"`
	OrganizationID uuid.UUID  `gorm:"type:uuid;primaryKey" json:"organization_id"`
	RoleID         *uuid.UUID `gorm:"type:uuid" json:"role_id,omitempty"`
	JoinedAt       time.Time  `gorm:"default:now()" json:"joined_at"`
	IsActive       bool       `gorm:"type:boolean;not null;default:true" json:"is_active"`

	// Relationships
	User         User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Organization Organization `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	Role         *Role        `gorm:"foreignKey:RoleID" json:"role,omitempty"`
}

// TableName sets the table name for UserOrganization
func (UserOrganization) TableName() string {
	return "user_organizations"
}
