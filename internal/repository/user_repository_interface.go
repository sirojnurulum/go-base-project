package repository

import (
	"beresin-backend/internal/model"
	"context"

	"github.com/google/uuid"
)

type UserRepositoryInterface interface {
	Create(ctx context.Context, user *model.User) error
	FindByUsernameWithRole(ctx context.Context, username string) (*model.User, error)
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByGoogleID(ctx context.Context, googleID string) (*model.User, error)
	FindByIDWithRole(ctx context.Context, id uuid.UUID) (*model.User, error)
	FindByIDWithRoleAndOrganizations(ctx context.Context, id uuid.UUID) (*model.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	FindByOrganizationID(ctx context.Context, organizationID uuid.UUID, offset, limit int) ([]model.User, error)
	List(ctx context.Context, offset, limit int, search string) ([]model.User, error)
	Count(ctx context.Context, search string) (int64, error)
	ListWithFilters(ctx context.Context, offset, limit int, search string, maxLevel int, organizationIDs []uuid.UUID) ([]model.User, error)
	CountWithFilters(ctx context.Context, search string, maxLevel int, organizationIDs []uuid.UUID) (int64, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	CountUsersByRoleLevel(ctx context.Context, level int) (int64, error)

	// User-Organization Management
	CreateUserOrganization(ctx context.Context, userOrg *model.UserOrganization) (*model.UserOrganization, error)
	FindUserOrganization(ctx context.Context, userID, organizationID uuid.UUID) (*model.UserOrganization, error)
	FindUserOrganizationWithRole(ctx context.Context, userID, organizationID uuid.UUID) (*model.UserOrganization, error)
	FindUserOrganizations(ctx context.Context, userID uuid.UUID, offset, limit int) ([]model.UserOrganization, error)
	FindOrganizationMembers(ctx context.Context, organizationID uuid.UUID, offset, limit int) ([]model.UserOrganization, error)
	FindOrganizationMembersWithRoles(ctx context.Context, organizationID uuid.UUID, offset, limit int) ([]model.UserOrganization, error)
	CountUserOrganizations(ctx context.Context, userID uuid.UUID) (int64, error)
	CountOrganizationMembers(ctx context.Context, organizationID uuid.UUID) (int64, error)
	UpdateUserOrganization(ctx context.Context, userOrg *model.UserOrganization) (*model.UserOrganization, error)
	UpdateUserOrganizationRole(ctx context.Context, userID, organizationID uuid.UUID, roleID *uuid.UUID) error
	DeleteUserOrganization(ctx context.Context, userID, organizationID uuid.UUID) error
	BulkCreateUserOrganizations(ctx context.Context, userOrgs []model.UserOrganization) ([]model.UserOrganization, []error)

	// User Organization History methods
	CreateUserOrganizationHistory(ctx context.Context, history *model.UserOrganizationHistory) (*model.UserOrganizationHistory, error)
	FindUserOrganizationHistory(ctx context.Context, userID uuid.UUID, offset, limit int) ([]model.UserOrganizationHistory, error)
	FindOrganizationUserHistory(ctx context.Context, organizationID uuid.UUID, offset, limit int) ([]model.UserOrganizationHistory, error)
	CountUserOrganizationHistory(ctx context.Context, userID uuid.UUID) (int64, error)
	CountOrganizationUserHistory(ctx context.Context, organizationID uuid.UUID) (int64, error)
	FindUserOrganizationHistoryByAction(ctx context.Context, userID uuid.UUID, action string, offset, limit int) ([]model.UserOrganizationHistory, error)
}
