package dto

import "github.com/google/uuid"

// CreateRoleRequest defines the structure for creating a new role.
type CreateRoleRequest struct {
	Name              string   `json:"name" validate:"required,min=3,max=50"`
	Description       string   `json:"description" validate:"max=255"`
	Level             int      `json:"level" validate:"required,min=0,max=99"`                                            // Level 100 reserved for superadmin
	PredefinedName    string   `json:"predefined_name" validate:"required,min=3,max=50"`                                  // NEW: Android-style name
	OrganizationTypes []string `json:"organization_types" validate:"omitempty,dive,oneof=platform holding company store"` // Organization context
}

// UpdateRoleRequest defines the structure for updating an existing role.
type UpdateRoleRequest struct {
	Name              string   `json:"name" validate:"required,min=3,max=50"`
	Description       string   `json:"description" validate:"max=255"`
	Level             int      `json:"level" validate:"required,min=0,max=99"`                                            // Level 100 reserved for superadmin
	PredefinedName    string   `json:"predefined_name" validate:"required,min=3,max=50"`                                  // Android-style name
	OrganizationTypes []string `json:"organization_types" validate:"omitempty,dive,oneof=platform holding company store"` // Organization context
}

// RoleResponse defines the structure for a role API response.
type RoleResponse struct {
	ID                uuid.UUID `json:"id"`
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	Level             int       `json:"level"`                        // NEW: Hierarchy level
	IsSystemRole      bool      `json:"is_system_role"`               // NEW: System role protection
	PredefinedName    string    `json:"predefined_name"`              // NEW: Android-style name
	IsActive          bool      `json:"is_active"`                    // NEW: Role activation status
	OrganizationTypes []string  `json:"organization_types,omitempty"` // Organization context
	Permissions       []string  `json:"permissions,omitempty"`
}

// UpdateRolePermissionsRequest defines the structure for updating a role's permissions.
type UpdateRolePermissionsRequest struct {
	// A list of permission names to assign to the role.
	PermissionNames []string `json:"permission_names" validate:"dive,min=1"`
}

// CreateRoleApprovalRequest defines the structure for requesting role creation approval.
type CreateRoleApprovalRequest struct {
	RequestedRoleName string `json:"requested_role_name" validate:"required,min=3,max=50"`
	RequestedLevel    int    `json:"requested_level" validate:"required,min=0,max=99"` // Level 100 reserved for superadmin
	Description       string `json:"description" validate:"max=255"`
}

// RoleApprovalResponse defines the structure for role approval API response.
type RoleApprovalResponse struct {
	ID                uuid.UUID  `json:"id"`
	RequestedRoleName string     `json:"requested_role_name"`
	RequestedLevel    int        `json:"requested_level"`
	Description       string     `json:"description"`
	RequestedBy       uuid.UUID  `json:"requested_by"`
	RequestedByUser   string     `json:"requested_by_user"` // Username of requester
	ApproverID        *uuid.UUID `json:"approver_id"`
	Approver          *string    `json:"approver,omitempty"` // Username of approver
	Status            string     `json:"status"`
	Reason            string     `json:"reason"`
	CreatedAt         string     `json:"created_at"`
	UpdatedAt         string     `json:"updated_at"`
}

// ApprovalDecisionRequest defines the structure for approving/rejecting role requests.
type ApprovalDecisionRequest struct {
	Status string `json:"status" validate:"required,oneof=approved rejected"`
	Reason string `json:"reason" validate:"max=255"`
}

// PredefinedRoleOption defines available predefined role names and levels.
type PredefinedRoleOption struct {
	Name        string `json:"name"`
	Level       int    `json:"level"`
	Description string `json:"description"`
}

// Permission Management DTOs

// CreatePermissionRequest defines the structure for creating a new permission.
type CreatePermissionRequest struct {
	Name        string `json:"name" validate:"required,min=3,max=100"`
	Description string `json:"description" validate:"max=255"`
}

// PermissionResponse defines the structure for a permission API response.
type PermissionResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   string    `json:"created_at"`
	UpdatedAt   string    `json:"updated_at"`
}

// UpdatePermissionRequest defines the structure for updating a permission.
type UpdatePermissionRequest struct {
	Name        string `json:"name" validate:"required,min=3,max=100"`
	Description string `json:"description" validate:"max=255"`
}

// Organization-specific Role DTOs

// GetRolesForOrganizationRequest defines the structure for getting roles suitable for an organization type.
type GetRolesForOrganizationRequest struct {
	OrganizationType string `query:"organization_type" validate:"required,oneof=platform holding company store"`
}

// OrganizationRoleAssignmentRequest defines the structure for assigning role with organization context.
type OrganizationRoleAssignmentRequest struct {
	UserID         uuid.UUID `json:"user_id" validate:"required"`
	OrganizationID uuid.UUID `json:"organization_id" validate:"required"`
	RoleID         uuid.UUID `json:"role_id" validate:"required"`
}

// OrganizationRoleResponse defines the structure for organization-specific role response.
type OrganizationRoleResponse struct {
	RoleID           uuid.UUID `json:"role_id"`
	RoleName         string    `json:"role_name"`
	RoleLevel        int       `json:"role_level"`
	OrganizationID   uuid.UUID `json:"organization_id"`
	OrganizationName string    `json:"organization_name"`
	OrganizationType string    `json:"organization_type"`
	IsActive         bool      `json:"is_active"`
	AssignedAt       string    `json:"assigned_at"`
}
