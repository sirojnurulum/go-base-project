package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OrganizationType represents the type of organization
type OrganizationType string

const (
	OrganizationTypePlatform OrganizationType = "platform"
	OrganizationTypeCompany  OrganizationType = "company"
	OrganizationTypeStore    OrganizationType = "store"
)

// Organization represents an organization in the system
type Organization struct {
	ID          uuid.UUID          `gorm:"type:uuid;primary_key;default:uuid_generate_v7()" json:"id"`
	Name        string             `gorm:"type:varchar(255);not null" json:"name"`
	Code        string             `gorm:"type:varchar(50);unique;not null" json:"code"`
	Type        OrganizationType   `gorm:"type:varchar(50);not null" json:"type"`
	Description string             `gorm:"type:text" json:"description"`
	ParentID    *uuid.UUID         `gorm:"type:uuid" json:"parent_id"`
	Parent      *Organization      `gorm:"foreignKey:ParentID" json:"parent"`
	Children    []Organization     `gorm:"foreignKey:ParentID" json:"children"`
	Members     []UserOrganization `gorm:"foreignKey:OrganizationID" json:"members"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
	DeletedAt   gorm.DeletedAt     `gorm:"index" json:"-"`
}

// UserOrganization represents the many-to-many relationship between users and organizations
type UserOrganization struct {
	ID             uuid.UUID    `gorm:"type:uuid;primary_key;default:uuid_generate_v7()" json:"id"`
	UserID         uuid.UUID    `gorm:"type:uuid;not null" json:"user_id"`
	OrganizationID uuid.UUID    `gorm:"type:uuid;not null" json:"organization_id"`
	User           User         `gorm:"foreignKey:UserID" json:"user"`
	Organization   Organization `gorm:"foreignKey:OrganizationID" json:"organization"`
	JoinedAt       time.Time    `json:"joined_at"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
}
