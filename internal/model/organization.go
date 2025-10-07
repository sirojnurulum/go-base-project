package model

import (
	"time"

	"github.com/google/uuid"
)

// Organization represents an organization in the system (holding, company, or store)
type Organization struct {
	ID                   uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v7()" json:"id"`
	Name                 string     `gorm:"type:varchar(100);not null" json:"name"`
	Code                 string     `gorm:"type:varchar(8);unique;not null" json:"code"`
	OrganizationType     string     `gorm:"type:varchar(20);not null;check:organization_type IN ('holding','company','store')" json:"organization_type"`
	ParentOrganizationID *uuid.UUID `gorm:"type:uuid" json:"parent_organization_id,omitempty"`
	Description          string     `gorm:"type:text" json:"description,omitempty"`
	CreatedBy            uuid.UUID  `gorm:"type:uuid;not null" json:"created_by"`
	IsActive             bool       `gorm:"type:boolean;not null;default:true" json:"is_active"`
	CreatedAt            time.Time  `gorm:"default:now()" json:"created_at"`
	UpdatedAt            time.Time  `gorm:"default:now()" json:"updated_at"`

	// Relationships
	ParentOrganization *Organization  `gorm:"foreignKey:ParentOrganizationID" json:"parent_organization,omitempty"`
	ChildOrganizations []Organization `gorm:"foreignKey:ParentOrganizationID" json:"child_organizations,omitempty"`
	Creator            User           `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	Users              []User         `gorm:"many2many:user_organizations;" json:"users,omitempty"`
}

// TableName sets the table name for Organization
func (Organization) TableName() string {
	return "organizations"
}
