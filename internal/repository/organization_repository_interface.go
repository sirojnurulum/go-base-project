package repository

import (
	"go-base-project/internal/dto"
	"go-base-project/internal/model"
	"context"

	"github.com/google/uuid"
)

// OrganizationRepositoryInterface defines the interface for organization data operations
type OrganizationRepositoryInterface interface {
	// Basic CRUD operations
	Create(ctx context.Context, org *model.Organization) (*model.Organization, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Organization, error)
	FindByIDWithDetails(ctx context.Context, id uuid.UUID) (*model.Organization, error)
	FindByCode(ctx context.Context, code string) (*model.Organization, error)
	Update(ctx context.Context, org *model.Organization) (*model.Organization, error)
	Delete(ctx context.Context, id uuid.UUID) error

	// Query operations with enhanced functionality
	FindAll(ctx context.Context, filters map[string]interface{}) ([]model.Organization, error)
	List(ctx context.Context, page, limit int, search string, filters map[string]interface{}) ([]model.Organization, int64, error)
	Count(ctx context.Context, filters map[string]interface{}) (int64, error)
	FindByType(ctx context.Context, orgType string) ([]model.Organization, error)
	FindByParent(ctx context.Context, parentID uuid.UUID) ([]model.Organization, error)
	FindRootOrganizations(ctx context.Context) ([]model.Organization, error)
	FindActiveOrganizations(ctx context.Context) ([]model.Organization, error)

	// Hierarchy management operations
	GetOrganizationHierarchy(ctx context.Context, orgID uuid.UUID) ([]model.Organization, error)
	GetChildrenRecursive(ctx context.Context, parentID uuid.UUID) ([]model.Organization, error)
	GetParentChain(ctx context.Context, orgID uuid.UUID) ([]model.Organization, error)

	// User-Organization operations
	AddUserToOrganization(ctx context.Context, userOrg *model.UserOrganization) error
	RemoveUserFromOrganization(ctx context.Context, userID, orgID uuid.UUID) error
	UpdateUserOrganizationRole(ctx context.Context, userID, orgID uuid.UUID, role string) error
	FindUserOrganizations(ctx context.Context, userID uuid.UUID) ([]model.UserOrganization, error)
	FindOrganizationUsers(ctx context.Context, orgID uuid.UUID) ([]model.UserOrganization, error)
	FindActiveUserOrganizations(ctx context.Context, userID uuid.UUID) ([]model.UserOrganization, error)
	FindActiveOrganizationUsers(ctx context.Context, orgID uuid.UUID) ([]model.UserOrganization, error)

	// Validation and utility operations
	CheckCodeExists(ctx context.Context, code string) (bool, error)
	CheckCodeExistsExcluding(ctx context.Context, code string, excludeID uuid.UUID) (bool, error)
	GetAllExistingCodes(ctx context.Context) ([]string, error)

	// Organization statistics and analytics
	GetOrganizationStats(ctx context.Context, orgID uuid.UUID) (*dto.OrganizationStatsResponse, error)
	GetHierarchyStats(ctx context.Context) (*dto.HierarchyStatsResponse, error)
}
