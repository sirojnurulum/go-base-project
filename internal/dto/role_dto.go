package dto

import "github.com/google/uuid"

// CreateRoleRequest defines the structure for creating a new role.
type CreateRoleRequest struct {
	Name        string `json:"name" validate:"required,min=3,max=50"`
	Description string `json:"description" validate:"max=255"`
}

// RoleResponse defines the structure for a role API response.
type RoleResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

// UpdateRolePermissionsRequest defines the structure for updating a role's permissions.
type UpdateRolePermissionsRequest struct {
	// A list of permission names to assign to the role.
	PermissionNames []string `json:"permission_names" validate:"required,gt=0,dive,min=1"`
}
