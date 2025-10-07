package service

import (
	"beresin-backend/internal/dto"
	"context"

	"github.com/google/uuid"
)

// RoleServiceInterface mendefinisikan kontrak untuk layanan manajemen role.
type RoleServiceInterface interface {
	CreateRole(ctx context.Context, req dto.CreateRoleRequest, userLevel int) (*dto.RoleResponse, error)
	ListRoles(ctx context.Context, userLevel int) ([]dto.RoleResponse, error)     // Returns roles below user's level
	GetRoleByID(ctx context.Context, roleID uuid.UUID) (*dto.RoleResponse, error) // Get role by ID for level checking
	UpdateRole(ctx context.Context, roleID uuid.UUID, req dto.UpdateRoleRequest, userLevel int) (*dto.RoleResponse, error)
	UpdateRolePermissions(ctx context.Context, roleID uuid.UUID, permissionNames []string) error

	// DISABLED: Role Approval Workflow methods
	// CreateRoleApprovalRequest(ctx context.Context, req dto.CreateRoleApprovalRequest, requestedBy uuid.UUID) (*dto.RoleApprovalResponse, error)
	// ListRoleApprovalRequests(ctx context.Context) ([]dto.RoleApprovalResponse, error)
	// ApproveRejectRoleRequest(ctx context.Context, approvalID uuid.UUID, decision dto.ApprovalDecisionRequest, approverID uuid.UUID) (*dto.RoleApprovalResponse, error)
	GetPredefinedRoleOptions(ctx context.Context, userLevel int) ([]dto.PredefinedRoleOption, error)

	// Permission Management methods
	ListPermissions(ctx context.Context) ([]dto.PermissionResponse, error)
	CreatePermission(ctx context.Context, req dto.CreatePermissionRequest) (*dto.PermissionResponse, error)
	UpdatePermission(ctx context.Context, id uuid.UUID, req dto.UpdatePermissionRequest) (*dto.PermissionResponse, error)
	DeletePermission(ctx context.Context, id uuid.UUID) error

	// Organization-specific role methods
	GetRolesForOrganizationType(ctx context.Context, organizationType string, userLevel int) ([]dto.RoleResponse, error)
	AssignRoleToUserInOrganization(ctx context.Context, req dto.OrganizationRoleAssignmentRequest) (*dto.OrganizationRoleResponse, error)
	GetUserRolesInOrganization(ctx context.Context, userID, organizationID uuid.UUID) ([]dto.OrganizationRoleResponse, error)
}
