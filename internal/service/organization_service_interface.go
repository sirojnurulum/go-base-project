package service

import (
	"beresin-backend/internal/dto"
	"beresin-backend/internal/model"

	"github.com/google/uuid"
)

// OrganizationServiceInterface defines the contract for organization service
type OrganizationServiceInterface interface {
	CreateOrganization(req *dto.CreateOrganizationRequest) (*model.Organization, error)
	GetOrganizationByID(id uuid.UUID) (*dto.OrganizationDetailResponse, error)
	GetOrganizationByCode(code string) (*model.Organization, error)
	GetAllOrganizations() ([]dto.OrganizationResponse, error)
	UpdateOrganization(id uuid.UUID, req *dto.UpdateOrganizationRequest) (*model.Organization, error)
	DeleteOrganization(id uuid.UUID) error
	JoinOrganization(userID uuid.UUID, req *dto.JoinOrganizationRequest) error
	LeaveOrganization(userID, organizationID uuid.UUID) error
	GetUserOrganizations(userID uuid.UUID) ([]dto.OrganizationResponse, error)
	GetOrganizationMembers(organizationID uuid.UUID) ([]dto.OrganizationMember, error)
}
