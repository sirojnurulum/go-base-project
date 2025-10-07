package model

import (
	"github.com/google/uuid"
)

// RoleOrganizationType represents which organization types a role is applicable to
type RoleOrganizationType struct {
	RoleID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"role_id"`
	OrganizationType string    `gorm:"type:varchar(50);primaryKey;check:organization_type IN ('platform', 'holding', 'company', 'store')" json:"organization_type"`

	// Relationships
	Role Role `gorm:"foreignKey:RoleID" json:"role,omitempty"`
}

// TableName sets the table name for RoleOrganizationType
func (RoleOrganizationType) TableName() string {
	return "role_organization_types"
}
