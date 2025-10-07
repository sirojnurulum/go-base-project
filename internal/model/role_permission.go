package model

import "github.com/google/uuid"

// RolePermission represents the many-to-many relationship between roles and permissions.
type RolePermission struct {
	RoleID       uuid.UUID  `gorm:"type:uuid;primaryKey" json:"role_id"`
	PermissionID uuid.UUID  `gorm:"type:uuid;primaryKey" json:"permission_id"`
	Role         Role       `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	Permission   Permission `gorm:"foreignKey:PermissionID" json:"permission,omitempty"`
}

// TableName sets the table name for RolePermission
func (RolePermission) TableName() string {
	return "role_permissions"
}
