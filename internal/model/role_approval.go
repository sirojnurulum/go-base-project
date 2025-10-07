package model

import (
	"time"

	"github.com/google/uuid"
)

// RoleApproval represents role approval requests for role creation workflow
type RoleApproval struct {
	ID                uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v7()" json:"id"`
	RequestedRoleName string     `gorm:"type:varchar(50);not null" json:"requested_role_name"`
	RequestedLevel    int        `gorm:"type:integer;not null" json:"requested_level"`
	Description       string     `gorm:"type:text" json:"description"`
	RequestedBy       uuid.UUID  `gorm:"type:uuid;not null" json:"requested_by"`
	RequestedByUser   User       `gorm:"foreignKey:RequestedBy" json:"requested_by_user"`
	ApproverID        *uuid.UUID `gorm:"type:uuid" json:"approver_id"`
	Approver          *User      `gorm:"foreignKey:ApproverID" json:"approver,omitempty"`
	Status            string     `gorm:"type:varchar(20);not null;default:'pending'" json:"status"` // pending, approved, rejected
	Reason            string     `gorm:"type:text" json:"reason"`

	// Organization context fields
	OrganizationID            *uuid.UUID    `gorm:"type:uuid" json:"organization_id,omitempty"`
	Organization              *Organization `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	OrganizationCode          string        `gorm:"type:varchar(8)" json:"organization_code,omitempty"`
	IsNewOrganization         bool          `gorm:"type:boolean;default:false" json:"is_new_organization"`
	RequestedOrganizationName string        `gorm:"type:varchar(100)" json:"requested_organization_name,omitempty"`
	RequestedOrganizationType string        `gorm:"type:varchar(20);check:requested_organization_type IN ('platform','holding','company','store')" json:"requested_organization_type,omitempty"`

	CreatedAt time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:now()" json:"updated_at"`
}

// TableName sets the table name for RoleApproval
func (RoleApproval) TableName() string {
	return "role_approvals"
}
