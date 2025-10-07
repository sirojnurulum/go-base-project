package repository

import (
	"beresin-backend/internal/model"
	"context"

	"github.com/google/uuid"
)

type RoleRepositoryInterface interface {
	// Basic CRUD operations
	Create(ctx context.Context, role *model.Role) (*model.Role, error)
	FindAll(ctx context.Context) ([]model.Role, error)
	FindAllActive(ctx context.Context) ([]model.Role, error)
	FindSystemRoles(ctx context.Context) ([]model.Role, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Role, error)
	FindByIDWithPermissions(ctx context.Context, id uuid.UUID) (*model.Role, error)
	FindByName(ctx context.Context, name string) (*model.Role, error)
	FindByNameWithPermissions(ctx context.Context, name string) (*model.Role, error)
	List(ctx context.Context, offset, limit int, search string) ([]model.Role, error)
	Count(ctx context.Context, search string) (int64, error)
	Update(ctx context.Context, role *model.Role) (*model.Role, error)
	Delete(ctx context.Context, id uuid.UUID) error

	// Permission-related operations
	FindPermissionsByRoleID(ctx context.Context, roleID uuid.UUID) ([]string, error)
	UpdateRolePermissions(ctx context.Context, roleID uuid.UUID, permissionNames []string) error

	// Permission CRUD operations
	GetAllPermissions(ctx context.Context) ([]model.Permission, error)
	CreatePermission(ctx context.Context, permission *model.Permission) error
	FindPermissionByName(ctx context.Context, name string) (*model.Permission, error)
	FindPermissionByID(ctx context.Context, id uuid.UUID) (*model.Permission, error)
	UpdatePermission(ctx context.Context, permission *model.Permission) error
	DeletePermission(ctx context.Context, id uuid.UUID) error
	CheckPermissionExists(ctx context.Context, name string) (bool, error)

	// Role Approval Repository methods
	CreateRoleApproval(ctx context.Context, approval *model.RoleApproval) (*model.RoleApproval, error)
	FindAllRoleApprovals(ctx context.Context) ([]model.RoleApproval, error)
	FindRoleApprovalsByStatus(ctx context.Context, status string) ([]model.RoleApproval, error)
	FindRoleApprovalsByRequester(ctx context.Context, requesterID uuid.UUID) ([]model.RoleApproval, error)
	FindRoleApprovalByID(ctx context.Context, id uuid.UUID) (*model.RoleApproval, error)
	UpdateRoleApproval(ctx context.Context, approval *model.RoleApproval) (*model.RoleApproval, error)

	// Organization-specific role methods
	FindRolesByOrganizationType(ctx context.Context, organizationType string) ([]model.Role, error)
	CreateRoleOrganizationTypes(ctx context.Context, roleID uuid.UUID, organizationTypes []string) error
	UpdateRoleOrganizationTypes(ctx context.Context, roleID uuid.UUID, organizationTypes []string) error
	DeleteRoleOrganizationTypes(ctx context.Context, roleID uuid.UUID) error
	FindOrganizationTypesByRoleID(ctx context.Context, roleID uuid.UUID) ([]string, error)

	// Super admin check
	IsRoleSuperAdmin(ctx context.Context, roleID uuid.UUID) (bool, error)
}
