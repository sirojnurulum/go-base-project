package repository

import (
	"beresin-backend/internal/model"

	"github.com/google/uuid"
)

// OrganizationRepositoryInterface defines the contract for organization repository
type OrganizationRepositoryInterface interface {
	Create(organization *model.Organization) error
	GetByID(id uuid.UUID) (*model.Organization, error)
	GetByCode(code string) (*model.Organization, error)
	GetAll() ([]model.Organization, error)
	Update(organization *model.Organization) error
	Delete(id uuid.UUID) error
	GetMembers(organizationID uuid.UUID) ([]model.UserOrganization, error)
	AddMember(userOrganization *model.UserOrganization) error
	RemoveMember(userID, organizationID uuid.UUID) error
	GetUserOrganizations(userID uuid.UUID) ([]model.UserOrganization, error)
	GetChildren(parentID uuid.UUID) ([]model.Organization, error)
}
