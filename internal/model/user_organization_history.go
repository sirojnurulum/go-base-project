package model

import (
	"time"

	"github.com/google/uuid"
)

// UserOrganizationHistory tracks changes to user-organization assignments
type UserOrganizationHistory struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v7()" json:"id"`
	UserID         uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	OrganizationID uuid.UUID `gorm:"type:uuid;not null" json:"organization_id"`
	Action         string    `gorm:"type:varchar(50);not null;check:action IN ('assigned', 'removed', 'role_updated', 'status_changed')" json:"action"`
	PreviousRole   string    `gorm:"type:varchar(50)" json:"previous_role,omitempty"`
	NewRole        string    `gorm:"type:varchar(50)" json:"new_role,omitempty"`
	PreviousStatus *bool     `gorm:"type:boolean" json:"previous_status,omitempty"`
	NewStatus      *bool     `gorm:"type:boolean" json:"new_status,omitempty"`
	ActionBy       uuid.UUID `gorm:"type:uuid;not null" json:"action_by"`
	ActionAt       time.Time `gorm:"default:now()" json:"action_at"`
	Reason         string    `gorm:"type:text" json:"reason,omitempty"`

	// Relationships
	User         User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Organization Organization `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	Actor        User         `gorm:"foreignKey:ActionBy" json:"actor,omitempty"`
}

// TableName sets the table name for UserOrganizationHistory
func (UserOrganizationHistory) TableName() string {
	return "user_organization_history"
}
