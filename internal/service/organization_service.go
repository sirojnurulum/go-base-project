package service

import (
	"beresin-backend/internal/dto"
	"beresin-backend/internal/model"
	"beresin-backend/internal/repository"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OrganizationService implements OrganizationServiceInterface
type OrganizationService struct {
	orgRepo  repository.OrganizationRepositoryInterface
	userRepo repository.UserRepository
}

// NewOrganizationService creates a new instance of OrganizationService
func NewOrganizationService(
	orgRepo repository.OrganizationRepositoryInterface,
	userRepo repository.UserRepository,
) OrganizationServiceInterface {
	return &OrganizationService{
		orgRepo:  orgRepo,
		userRepo: userRepo,
	}
}

// CreateOrganization creates a new organization
func (s *OrganizationService) CreateOrganization(req *dto.CreateOrganizationRequest) (*model.Organization, error) {
	// Check if code already exists
	existingOrg, err := s.orgRepo.GetByCode(req.Code)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existingOrg != nil {
		return nil, errors.New("organization code already exists")
	}

	// Validate parent organization if specified
	if req.ParentID != nil {
		parentOrg, err := s.orgRepo.GetByID(*req.ParentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("parent organization not found")
			}
			return nil, err
		}

		// Validate organization hierarchy
		if !s.validateOrganizationHierarchy(parentOrg.Type, req.Type) {
			return nil, errors.New("invalid organization hierarchy")
		}
	}

	organization := &model.Organization{
		Name:        req.Name,
		Code:        req.Code,
		Type:        req.Type,
		Description: req.Description,
		ParentID:    req.ParentID,
	}

	err = s.orgRepo.Create(organization)
	if err != nil {
		return nil, err
	}

	return organization, nil
}

// validateOrganizationHierarchy validates the organization hierarchy
func (s *OrganizationService) validateOrganizationHierarchy(parentType, childType model.OrganizationType) bool {
	switch parentType {
	case model.OrganizationTypePlatform:
		return childType == model.OrganizationTypeCompany
	case model.OrganizationTypeCompany:
		return childType == model.OrganizationTypeStore
	case model.OrganizationTypeStore:
		return false // Store cannot have children
	}
	return false
}

// GetOrganizationByID retrieves an organization by ID with detailed information
func (s *OrganizationService) GetOrganizationByID(id uuid.UUID) (*dto.OrganizationDetailResponse, error) {
	org, err := s.orgRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	members, err := s.orgRepo.GetMembers(id)
	if err != nil {
		return nil, err
	}

	response := &dto.OrganizationDetailResponse{
		OrganizationResponse: dto.OrganizationResponse{
			ID:          org.ID,
			Name:        org.Name,
			Code:        org.Code,
			Type:        org.Type,
			Description: org.Description,
			ParentID:    org.ParentID,
			MemberCount: len(members),
		},
	}

	// Add parent information
	if org.Parent != nil {
		response.Parent = &dto.OrganizationResponse{
			ID:   org.Parent.ID,
			Name: org.Parent.Name,
			Code: org.Parent.Code,
			Type: org.Parent.Type,
		}
	}

	// Add children information
	for _, child := range org.Children {
		response.Children = append(response.Children, dto.OrganizationResponse{
			ID:   child.ID,
			Name: child.Name,
			Code: child.Code,
			Type: child.Type,
		})
	}

	// Add members information
	for _, member := range members {
		roleName := "No Role"
		if member.User.Role != nil {
			roleName = member.User.Role.Name
		}

		response.Members = append(response.Members, dto.OrganizationMember{
			UserID:   member.UserID,
			Username: member.User.Username,
			Email:    member.User.Email,
			Role:     roleName,
			JoinedAt: member.JoinedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return response, nil
}

// GetOrganizationByCode retrieves an organization by code
func (s *OrganizationService) GetOrganizationByCode(code string) (*model.Organization, error) {
	return s.orgRepo.GetByCode(code)
}

// GetAllOrganizations retrieves all organizations
func (s *OrganizationService) GetAllOrganizations() ([]dto.OrganizationResponse, error) {
	organizations, err := s.orgRepo.GetAll()
	if err != nil {
		return nil, err
	}

	var responses []dto.OrganizationResponse
	for _, org := range organizations {
		// Count members
		members, _ := s.orgRepo.GetMembers(org.ID)
		memberCount := len(members)

		responses = append(responses, dto.OrganizationResponse{
			ID:          org.ID,
			Name:        org.Name,
			Code:        org.Code,
			Type:        org.Type,
			Description: org.Description,
			ParentID:    org.ParentID,
			MemberCount: memberCount,
		})
	}

	return responses, nil
}

// UpdateOrganization updates an existing organization
func (s *OrganizationService) UpdateOrganization(id uuid.UUID, req *dto.UpdateOrganizationRequest) (*model.Organization, error) {
	org, err := s.orgRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	org.Name = req.Name
	org.Description = req.Description

	err = s.orgRepo.Update(org)
	if err != nil {
		return nil, err
	}

	return org, nil
}

// DeleteOrganization deletes an organization
func (s *OrganizationService) DeleteOrganization(id uuid.UUID) error {
	// Check if organization has children
	children, err := s.orgRepo.GetChildren(id)
	if err != nil {
		return err
	}
	if len(children) > 0 {
		return errors.New("cannot delete organization with children")
	}

	// Check if organization has members
	members, err := s.orgRepo.GetMembers(id)
	if err != nil {
		return err
	}
	if len(members) > 0 {
		return errors.New("cannot delete organization with members")
	}

	return s.orgRepo.Delete(id)
}

// JoinOrganization adds a user to an organization
func (s *OrganizationService) JoinOrganization(userID uuid.UUID, req *dto.JoinOrganizationRequest) error {
	// Find organization by code
	org, err := s.orgRepo.GetByCode(req.Code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("organization not found")
		}
		return err
	}

	// Check if user exists
	_, err = s.userRepo.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	// Check if user is already a member
	userOrgs, err := s.orgRepo.GetUserOrganizations(userID)
	if err != nil {
		return err
	}

	for _, userOrg := range userOrgs {
		if userOrg.OrganizationID == org.ID {
			return errors.New("user is already a member of this organization")
		}
	}

	// Add user to organization
	userOrganization := &model.UserOrganization{
		UserID:         userID,
		OrganizationID: org.ID,
	}

	return s.orgRepo.AddMember(userOrganization)
}

// LeaveOrganization removes a user from an organization
func (s *OrganizationService) LeaveOrganization(userID, organizationID uuid.UUID) error {
	// Check if user is a member
	userOrgs, err := s.orgRepo.GetUserOrganizations(userID)
	if err != nil {
		return err
	}

	isMember := false
	for _, userOrg := range userOrgs {
		if userOrg.OrganizationID == organizationID {
			isMember = true
			break
		}
	}

	if !isMember {
		return errors.New("user is not a member of this organization")
	}

	return s.orgRepo.RemoveMember(userID, organizationID)
}

// GetUserOrganizations retrieves all organizations a user belongs to
func (s *OrganizationService) GetUserOrganizations(userID uuid.UUID) ([]dto.OrganizationResponse, error) {
	userOrgs, err := s.orgRepo.GetUserOrganizations(userID)
	if err != nil {
		return nil, err
	}

	var responses []dto.OrganizationResponse
	for _, userOrg := range userOrgs {
		// Count members
		members, _ := s.orgRepo.GetMembers(userOrg.OrganizationID)
		memberCount := len(members)

		responses = append(responses, dto.OrganizationResponse{
			ID:          userOrg.Organization.ID,
			Name:        userOrg.Organization.Name,
			Code:        userOrg.Organization.Code,
			Type:        userOrg.Organization.Type,
			Description: userOrg.Organization.Description,
			ParentID:    userOrg.Organization.ParentID,
			MemberCount: memberCount,
		})
	}

	return responses, nil
}

// GetOrganizationMembers retrieves all members of an organization
func (s *OrganizationService) GetOrganizationMembers(organizationID uuid.UUID) ([]dto.OrganizationMember, error) {
	members, err := s.orgRepo.GetMembers(organizationID)
	if err != nil {
		return nil, err
	}

	var responses []dto.OrganizationMember
	for _, member := range members {
		roleName := "No Role"
		if member.User.Role != nil {
			roleName = member.User.Role.Name
		}

		responses = append(responses, dto.OrganizationMember{
			UserID:   member.UserID,
			Username: member.User.Username,
			Email:    member.User.Email,
			Role:     roleName,
			JoinedAt: member.JoinedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return responses, nil
}
