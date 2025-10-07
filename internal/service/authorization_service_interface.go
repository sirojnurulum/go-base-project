package service

import (
	"context"

	"github.com/google/uuid"
)

// AuthorizationService defines the contract for authorization and permission-related logic.
type AuthorizationServiceInterface interface {
	CheckPermission(ctx context.Context, roleID uuid.UUID, requiredPermission string) (bool, error)
	InvalidateRolePermissionsCache(ctx context.Context, roleID uuid.UUID) error
	GetAndCachePermissionsForRole(ctx context.Context, roleID uuid.UUID) ([]string, error)
	CheckUserOrganizationAccess(ctx context.Context, userID, organizationID uuid.UUID) (bool, error)
	IsRoleSuperAdmin(ctx context.Context, roleID uuid.UUID) (bool, error)

	// Multi-tenant role isolation methods
	CheckPermissionInOrganization(ctx context.Context, userID, organizationID uuid.UUID, requiredPermission string) (bool, error)
	GetUserRoleInOrganization(ctx context.Context, userID, organizationID uuid.UUID) (*uuid.UUID, error)
	GetUserPermissionsInOrganization(ctx context.Context, userID, organizationID uuid.UUID) ([]string, error)
	ValidateRoleAccessibleInOrganization(ctx context.Context, roleID, organizationID uuid.UUID, organizationType string) (bool, error)
}
