package service

import (
	"beresin-backend/internal/dto"
	"context"

	"github.com/google/uuid"
)

// UserService mendefinisikan kontrak untuk layanan manajemen pengguna.
type UserServiceInterface interface {
	CreateUser(ctx context.Context, req dto.CreateUserRequest) (*dto.UserResponse, error)
	ListUsers(ctx context.Context, page, limit int, search string) (*dto.PagedUserResponse, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*dto.UserResponse, error)
	UpdateUser(ctx context.Context, id uuid.UUID, req dto.UpdateUserRequest) (*dto.UserResponse, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error

	// User-Organization Management
	AssignUserToOrganization(ctx context.Context, req dto.AssignUserToOrganizationRequest) (*dto.UserOrganizationResponse, error)
	RemoveUserFromOrganization(ctx context.Context, userID, organizationID uuid.UUID) error
	UpdateUserOrganizationRole(ctx context.Context, userID, organizationID uuid.UUID, req dto.UpdateUserOrganizationRequest) (*dto.UserOrganizationResponse, error)
	GetUserOrganizations(ctx context.Context, userID uuid.UUID, page, limit int) (*dto.PagedUserOrganizationResponse, error)
	GetOrganizationMembers(ctx context.Context, organizationID uuid.UUID, page, limit int) (*dto.PagedUserOrganizationResponse, error)
	BulkAssignUsersToOrganization(ctx context.Context, req dto.BulkAssignUsersToOrganizationRequest) (*dto.BulkAssignResponse, error)
	// GetUserOrganizationHistory retrieves the organization assignment history for a user.
	GetUserOrganizationHistory(ctx context.Context, userID uuid.UUID, page, limit int) (*dto.PagedUserOrganizationHistoryResponse, error)

	// GetOrganizationUserHistory retrieves the user assignment history for an organization.
	GetOrganizationUserHistory(ctx context.Context, organizationID uuid.UUID, page, limit int) (*dto.PagedUserOrganizationHistoryResponse, error)

	// LogUserOrganizationAction manually logs a user organization action (for manual tracking).
	LogUserOrganizationAction(ctx context.Context, req dto.LogUserOrganizationActionRequest, actionBy uuid.UUID) (*dto.UserOrganizationHistoryResponse, error)
}
