package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Role represents a user role in the system with hierarchy.
type Role struct {
	ID                uuid.UUID              `gorm:"type:uuid;primary_key;default:uuid_generate_v7()" json:"id"`
	Name              string                 `gorm:"type:varchar(50);unique;not null" json:"name"`
	Description       string                 `gorm:"type:text" json:"description"`
	Level             int                    `gorm:"type:integer;not null;default:0;check:level >= 0 AND level <= 100" json:"level"`
	IsSystemRole      bool                   `gorm:"type:boolean;not null;default:false" json:"is_system_role"`
	PredefinedName    string                 `gorm:"type:varchar(50)" json:"predefined_name"`
	IsActive          bool                   `gorm:"type:boolean;not null;default:true" json:"is_active"`
	Permissions       []Permission           `gorm:"many2many:role_permissions;" json:"permissions"`
	OrganizationTypes []RoleOrganizationType `gorm:"foreignKey:RoleID" json:"organization_types,omitempty"`
	CreatedAt         time.Time              `gorm:"default:now()" json:"created_at"`
	UpdatedAt         time.Time              `gorm:"default:now()" json:"updated_at"`
	DeletedAt         gorm.DeletedAt         `gorm:"index" json:"-"`
}

// TableName sets the table name for Role
func (Role) TableName() string {
	return "roles"
}
