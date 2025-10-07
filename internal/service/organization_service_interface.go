package service

import (
	"go-base-project/internal/dto"
	"context"

	"github.com/google/uuid"
)

// OrganizationService defines the interface for organization business logic
type OrganizationServiceInterface interface {
	// Organization CRUD
	CreateOrganization(ctx context.Context, req dto.CreateOrganizationRequest, createdBy uuid.UUID) (*dto.OrganizationResponse, error)
	GetOrganizationByID(ctx context.Context, id uuid.UUID) (*dto.OrganizationResponse, error)
	GetOrganizationByCode(ctx context.Context, code string) (*dto.OrganizationResponse, error)
	UpdateOrganization(ctx context.Context, id uuid.UUID, req dto.UpdateOrganizationRequest, updatedBy uuid.UUID) (*dto.OrganizationResponse, error)
	DeleteOrganization(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error

	// Organization queries
	ListOrganizations(ctx context.Context, req dto.ListOrganizationsRequest) (*dto.PaginatedOrganizationsResponse, error)
	GetOrganizationsByType(ctx context.Context, orgType string) ([]dto.OrganizationResponse, error)
	GetChildOrganizations(ctx context.Context, parentID uuid.UUID) ([]dto.OrganizationResponse, error)

	// User-Organization management
	JoinOrganization(ctx context.Context, userID uuid.UUID, req dto.JoinOrganizationRequest) (*dto.UserOrganizationResponse, error)
	LeaveOrganization(ctx context.Context, userID, orgID uuid.UUID) error
	GetUserOrganizations(ctx context.Context, userID uuid.UUID) ([]dto.UserOrganizationResponse, error)
	GetOrganizationMembers(ctx context.Context, orgID uuid.UUID) ([]dto.UserOrganizationResponse, error)

	// Organization code utilities
	GenerateUniqueCode(ctx context.Context, organizationName string) (string, error)
	ValidateOrganizationAccess(ctx context.Context, userID, orgID uuid.UUID) (bool, error)

	// Complete organization structure creation (Platform admin feature)
	CreateCompleteOrganizationStructure(ctx context.Context, req dto.CreateCompleteStructureRequest, createdBy uuid.UUID) (*dto.CompleteOrganizationStructureResponse, error)

	// Organization statistics based on user level
	GetOrganizationStatistics(ctx context.Context, userID uuid.UUID) (*dto.OrganizationStatisticsResponse, error)

	// Hierarchical access control
	GetAccessibleOrganizationIDs(ctx context.Context, currentUserID uuid.UUID, currentUserLevel int) ([]uuid.UUID, error)
}
