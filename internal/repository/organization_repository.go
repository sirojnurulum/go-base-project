package repository

import (
	"beresin-backend/internal/model"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OrganizationRepository implements OrganizationRepositoryInterface
type OrganizationRepository struct {
	db *gorm.DB
}

// NewOrganizationRepository creates a new instance of OrganizationRepository
func NewOrganizationRepository(db *gorm.DB) OrganizationRepositoryInterface {
	return &OrganizationRepository{db: db}
}

// Create creates a new organization
func (r *OrganizationRepository) Create(organization *model.Organization) error {
	return r.db.Create(organization).Error
}

// GetByID retrieves an organization by ID
func (r *OrganizationRepository) GetByID(id uuid.UUID) (*model.Organization, error) {
	var organization model.Organization
	err := r.db.Preload("Parent").Preload("Children").Preload("Members.User").First(&organization, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &organization, nil
}

// GetByCode retrieves an organization by code
func (r *OrganizationRepository) GetByCode(code string) (*model.Organization, error) {
	var organization model.Organization
	err := r.db.Preload("Parent").Preload("Children").First(&organization, "code = ?", code).Error
	if err != nil {
		return nil, err
	}
	return &organization, nil
}

// GetAll retrieves all organizations
func (r *OrganizationRepository) GetAll() ([]model.Organization, error) {
	var organizations []model.Organization
	err := r.db.Preload("Parent").Find(&organizations).Error
	return organizations, err
}

// Update updates an existing organization
func (r *OrganizationRepository) Update(organization *model.Organization) error {
	return r.db.Save(organization).Error
}

// Delete soft deletes an organization
func (r *OrganizationRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Organization{}, "id = ?", id).Error
}

// GetMembers retrieves all members of an organization
func (r *OrganizationRepository) GetMembers(organizationID uuid.UUID) ([]model.UserOrganization, error) {
	var members []model.UserOrganization
	err := r.db.Preload("User").Preload("User.Role").Where("organization_id = ?", organizationID).Find(&members).Error
	return members, err
}

// AddMember adds a user to an organization
func (r *OrganizationRepository) AddMember(userOrganization *model.UserOrganization) error {
	userOrganization.JoinedAt = time.Now()
	return r.db.Create(userOrganization).Error
}

// RemoveMember removes a user from an organization
func (r *OrganizationRepository) RemoveMember(userID, organizationID uuid.UUID) error {
	return r.db.Where("user_id = ? AND organization_id = ?", userID, organizationID).Delete(&model.UserOrganization{}).Error
}

// GetUserOrganizations retrieves all organizations a user belongs to
func (r *OrganizationRepository) GetUserOrganizations(userID uuid.UUID) ([]model.UserOrganization, error) {
	var userOrgs []model.UserOrganization
	err := r.db.Preload("Organization").Where("user_id = ?", userID).Find(&userOrgs).Error
	return userOrgs, err
}

// GetChildren retrieves all child organizations of a parent
func (r *OrganizationRepository) GetChildren(parentID uuid.UUID) ([]model.Organization, error) {
	var children []model.Organization
	err := r.db.Where("parent_id = ?", parentID).Find(&children).Error
	return children, err
}
