package service

import (
	"go-base-project/internal/apperror"
	"go-base-project/internal/constant"
	"go-base-project/internal/dto"
	"go-base-project/internal/model"
	"go-base-project/internal/repository"
	"go-base-project/internal/util"
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type organizationService struct {
	orgRepo  repository.OrganizationRepositoryInterface
	userRepo repository.UserRepositoryInterface
}

// NewOrganizationService creates a new organization service instance
func NewOrganizationService(orgRepo repository.OrganizationRepositoryInterface, userRepo repository.UserRepositoryInterface) OrganizationServiceInterface {
	return &organizationService{
		orgRepo:  orgRepo,
		userRepo: userRepo,
	}
}

// CreateOrganization creates a new organization with enhanced validation and error handling
func (s *organizationService) CreateOrganization(ctx context.Context, req dto.CreateOrganizationRequest, createdBy uuid.UUID) (*dto.OrganizationResponse, error) {
	// Validate organization type
	if err := s.validateOrganizationType(req.OrganizationType); err != nil {
		return nil, err
	}

	// Generate unique code using business logic in service layer
	code, err := s.GenerateUniqueCode(ctx, req.Name)
	if err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to generate organization code: %w", err))
	}

	// Validate parent organization if specified
	if req.ParentOrganizationID != nil {
		if err := s.validateParentOrganization(ctx, *req.ParentOrganizationID); err != nil {
			return nil, err
		}
	}

	// Create organization using repository
	org := &model.Organization{
		Name:                 req.Name,
		Code:                 code,
		OrganizationType:     req.OrganizationType,
		ParentOrganizationID: req.ParentOrganizationID,
		Description:          req.Description,
		CreatedBy:            createdBy,
		IsActive:             true,
	}

	createdOrg, err := s.orgRepo.Create(ctx, org)
	if err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to create organization: %w", err))
	}

	// Auto-join creator to organization as Organization Creator using repository method
	// TODO: Replace with proper role ID lookup for "Organization Creator" role
	userOrg := &model.UserOrganization{
		UserID:         createdBy,
		OrganizationID: createdOrg.ID,
		RoleID:         nil, // TODO: Set to Organization Creator role ID
		IsActive:       true,
	}

	if err := s.orgRepo.AddUserToOrganization(ctx, userOrg); err != nil {
		// Log the error but don't fail the organization creation
		// TODO: Replace with proper structured logging
		fmt.Printf("Warning: Failed to auto-join creator to organization %s: %v\n", createdOrg.ID, err)
	}

	return util.MapOrganizationToResponse(createdOrg), nil
}

// GetOrganizationByID retrieves organization by ID with enhanced error handling
func (s *organizationService) GetOrganizationByID(ctx context.Context, id uuid.UUID) (*dto.OrganizationResponse, error) {
	org, err := s.orgRepo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.NewNotFoundError("organization")
		}
		return nil, apperror.NewInternalError(fmt.Errorf("failed to find organization: %w", err))
	}

	return util.MapOrganizationToResponse(org), nil
}

// GetOrganizationByCode retrieves organization by code
func (s *organizationService) GetOrganizationByCode(ctx context.Context, code string) (*dto.OrganizationResponse, error) {
	org, err := s.orgRepo.FindByCode(ctx, code)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.NewNotFoundError("organization")
		}
		return nil, apperror.NewInternalError(fmt.Errorf("failed to find organization: %w", err))
	}

	return util.MapOrganizationToResponse(org), nil
}

// UpdateOrganization updates an existing organization
func (s *organizationService) UpdateOrganization(ctx context.Context, id uuid.UUID, req dto.UpdateOrganizationRequest, updatedBy uuid.UUID) (*dto.OrganizationResponse, error) {
	org, err := s.orgRepo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.NewNotFoundError("organization")
		}
		return nil, apperror.NewInternalError(fmt.Errorf("failed to find organization: %w", err))
	}

	// Update fields if provided
	if req.Name != "" {
		org.Name = req.Name
	}
	if req.Description != "" {
		org.Description = req.Description
	}
	if req.IsActive != nil {
		org.IsActive = *req.IsActive
	}

	updatedOrg, err := s.orgRepo.Update(ctx, org)
	if err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to update organization: %w", err))
	}

	return util.MapOrganizationToResponse(updatedOrg), nil
}

// DeleteOrganization deletes an organization
func (s *organizationService) DeleteOrganization(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error {
	// Check if organization exists
	_, err := s.orgRepo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return apperror.NewNotFoundError("organization")
		}
		return apperror.NewInternalError(fmt.Errorf("failed to find organization: %w", err))
	}

	// Check if organization has child organizations
	children, err := s.orgRepo.FindByParent(ctx, id)
	if err != nil {
		return apperror.NewInternalError(fmt.Errorf("failed to check child organizations: %w", err))
	}

	if len(children) > 0 {
		return apperror.NewAppError(http.StatusBadRequest, "Cannot delete organization with child organizations", nil)
	}

	if err := s.orgRepo.Delete(ctx, id); err != nil {
		return apperror.NewInternalError(fmt.Errorf("failed to delete organization: %w", err))
	}

	return nil
}

// ListOrganizations retrieves organizations with filters using repository's enhanced List method
func (s *organizationService) ListOrganizations(ctx context.Context, req dto.ListOrganizationsRequest) (*dto.PaginatedOrganizationsResponse, error) {
	// Get current user from context for hierarchical filtering
	var accessibleOrgIDs []uuid.UUID

	if currentUserID, ok := ctx.Value("current_user_id").(uuid.UUID); ok {
		// Get current user with role information
		currentUser, err := s.userRepo.FindByIDWithRole(ctx, currentUserID)
		if err != nil {
			return nil, apperror.NewInternalError(fmt.Errorf("failed to find current user: %w", err))
		}

		// Determine current user's level
		var currentUserLevel int
		if currentUser.Role != nil {
			currentUserLevel = currentUser.Role.Level
		}

		// Super Admin (level 100) and Platform Admin (level >= 76) can see ALL organizations without filtering
		if currentUserLevel >= 76 {
			// No filtering for system maintainers - they can see all organizations
		} else {
			// For non-platform users, apply hierarchical filtering
			accessibleOrgIDs, err = s.GetAccessibleOrganizationIDs(ctx, currentUserID, currentUserLevel)
			if err != nil {
				return nil, apperror.NewInternalError(fmt.Errorf("failed to get accessible organizations: %w", err))
			}
		}
	}

	// Apply hierarchical filtering to base filters
	filters := s.buildOrganizationFilters(req)

	// Add organization ID filtering only for non-platform users
	if len(accessibleOrgIDs) > 0 {
		filters["accessible_org_ids"] = accessibleOrgIDs
	}

	// Use repository List method with pagination and hierarchical filtering
	orgs, total, err := s.orgRepo.List(ctx, req.Page, req.Limit, req.Search, filters)
	if err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to list organizations: %w", err))
	}

	response := make([]dto.OrganizationResponse, len(orgs))
	for i, org := range orgs {
		response[i] = *util.MapOrganizationToResponse(&org)
	}

	// Calculate pagination info
	totalPages := int((total + int64(req.Limit) - 1) / int64(req.Limit))

	return &dto.PaginatedOrganizationsResponse{
		Organizations: response,
		Pagination: dto.PaginationInfo{
			Page:       req.Page,
			Limit:      req.Limit,
			Total:      int(total),
			TotalPages: totalPages,
		},
	}, nil
} // GetOrganizationsByType retrieves organizations by type
func (s *organizationService) GetOrganizationsByType(ctx context.Context, orgType string) ([]dto.OrganizationResponse, error) {
	orgs, err := s.orgRepo.FindByType(ctx, orgType)
	if err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to find organizations by type: %w", err))
	}

	response := make([]dto.OrganizationResponse, len(orgs))
	for i, org := range orgs {
		response[i] = *util.MapOrganizationToResponse(&org)
	}

	return response, nil
}

// GetChildOrganizations retrieves child organizations
func (s *organizationService) GetChildOrganizations(ctx context.Context, parentID uuid.UUID) ([]dto.OrganizationResponse, error) {
	orgs, err := s.orgRepo.FindByParent(ctx, parentID)
	if err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to find child organizations: %w", err))
	}

	response := make([]dto.OrganizationResponse, len(orgs))
	for i, org := range orgs {
		response[i] = *util.MapOrganizationToResponse(&org)
	}

	return response, nil
}

// JoinOrganization allows user to join an organization with enhanced validation
func (s *organizationService) JoinOrganization(ctx context.Context, userID uuid.UUID, req dto.JoinOrganizationRequest) (*dto.UserOrganizationResponse, error) {
	// Validate organization exists and is active
	org, err := s.orgRepo.FindByCode(ctx, req.OrganizationCode)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.NewNotFoundError("organization with code")
		}
		return nil, apperror.NewInternalError(fmt.Errorf("failed to find organization: %w", err))
	}

	if !org.IsActive {
		return nil, apperror.NewAppError(http.StatusBadRequest, "Organization is not active", nil)
	}

	// Check if user already in organization using repository method
	if exists, err := s.checkUserOrganizationMembership(ctx, userID, org.ID); err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to check user membership: %w", err))
	} else if exists {
		return nil, apperror.NewConflictError("User already member of this organization")
	}

	// Add user to organization with Member role
	// TODO: Replace with proper role ID lookup for "Member" role
	userOrg := &model.UserOrganization{
		UserID:         userID,
		OrganizationID: org.ID,
		RoleID:         nil, // TODO: Set to Member role ID
		IsActive:       true,
	}

	if err := s.orgRepo.AddUserToOrganization(ctx, userOrg); err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to join organization: %w", err))
	}

	return s.mapToUserOrganizationResponse(userOrg, org), nil
}

// LeaveOrganization allows user to leave an organization
func (s *organizationService) LeaveOrganization(ctx context.Context, userID, orgID uuid.UUID) error {
	if err := s.orgRepo.RemoveUserFromOrganization(ctx, userID, orgID); err != nil {
		return apperror.NewInternalError(fmt.Errorf("failed to leave organization: %w", err))
	}
	return nil
}

// GetUserOrganizations retrieves all organizations for a user
func (s *organizationService) GetUserOrganizations(ctx context.Context, userID uuid.UUID) ([]dto.UserOrganizationResponse, error) {
	userOrgs, err := s.orgRepo.FindUserOrganizations(ctx, userID)
	if err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to find user organizations: %w", err))
	}

	response := make([]dto.UserOrganizationResponse, len(userOrgs))
	for i, userOrg := range userOrgs {
		response[i] = *s.mapToUserOrganizationResponse(&userOrg, &userOrg.Organization)
	}

	return response, nil
}

// GetOrganizationMembers retrieves all members of an organization
func (s *organizationService) GetOrganizationMembers(ctx context.Context, orgID uuid.UUID) ([]dto.UserOrganizationResponse, error) {
	userOrgs, err := s.orgRepo.FindOrganizationUsers(ctx, orgID)
	if err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to find organization members: %w", err))
	}

	response := make([]dto.UserOrganizationResponse, len(userOrgs))
	for i, userOrg := range userOrgs {
		org, _ := s.orgRepo.FindByID(ctx, userOrg.OrganizationID)
		response[i] = *s.mapToUserOrganizationResponse(&userOrg, org)
	}

	return response, nil
}

// checkUserOrganizationMembership checks if user is already a member of the organization
func (s *organizationService) checkUserOrganizationMembership(ctx context.Context, userID, orgID uuid.UUID) (bool, error) {
	userOrgs, err := s.orgRepo.FindUserOrganizations(ctx, userID)
	if err != nil {
		return false, err
	}

	for _, userOrg := range userOrgs {
		if userOrg.OrganizationID == orgID && userOrg.IsActive {
			return true, nil
		}
	}

	return false, nil
}

// buildOrganizationFilters constructs filters map from request parameters
func (s *organizationService) buildOrganizationFilters(req dto.ListOrganizationsRequest) map[string]interface{} {
	filters := make(map[string]interface{})

	if req.OrganizationType != "" {
		filters["organization_type"] = req.OrganizationType
	}
	if req.ParentOrganizationID != nil {
		filters["parent_organization_id"] = *req.ParentOrganizationID
	}
	if req.IsActive != nil {
		filters["is_active"] = *req.IsActive
	}

	return filters
}

// validateOrganizationType validates if the organization type is valid
func (s *organizationService) validateOrganizationType(orgType string) error {
	validTypes := []string{
		constant.OrganizationTypeHolding,
		constant.OrganizationTypeCompany,
		constant.OrganizationTypeStore,
	}

	for _, validType := range validTypes {
		if orgType == validType {
			return nil
		}
	}

	return apperror.NewAppError(http.StatusBadRequest, "Invalid organization type", nil)
}

// validateParentOrganization validates if parent organization exists and hierarchy is valid
func (s *organizationService) validateParentOrganization(ctx context.Context, parentID uuid.UUID) error {
	// Check if parent organization exists
	_, err := s.orgRepo.FindByID(ctx, parentID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return apperror.NewNotFoundError("parent organization")
		}
		return apperror.NewInternalError(fmt.Errorf("failed to validate parent organization: %w", err))
	}
	return nil
}

// buildOrganizationFilters constructs filters map from request parameters

// GenerateUniqueCode generates a unique organization code using business logic in service layer
func (s *organizationService) GenerateUniqueCode(ctx context.Context, organizationName string) (string, error) {
	// Get all existing codes from repository (data access only)
	existingCodes, err := s.orgRepo.GetAllExistingCodes(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to fetch existing codes: %w", err)
	}

	// Use utility function for business logic (code generation)
	uniqueCode := util.GenerateUniqueOrganizationCode(organizationName, existingCodes)

	return uniqueCode, nil
}

// ValidateOrganizationAccess checks if user has access to organization using repository methods
func (s *organizationService) ValidateOrganizationAccess(ctx context.Context, userID, orgID uuid.UUID) (bool, error) {
	return s.checkUserOrganizationMembership(ctx, userID, orgID)
}

// mapToUserOrganizationResponse converts UserOrganization model to response DTO using utility function
func (s *organizationService) mapToUserOrganizationResponse(userOrg *model.UserOrganization, org *model.Organization) *dto.UserOrganizationResponse {
	orgResponse := util.MapOrganizationToResponse(org)

	return &dto.UserOrganizationResponse{
		UserID:         userOrg.UserID,
		OrganizationID: userOrg.OrganizationID,
		Organization:   *orgResponse,
		RoleID:         userOrg.RoleID,
		JoinedAt:       userOrg.JoinedAt,
		IsActive:       userOrg.IsActive,
	}
}

// CreateCompleteOrganizationStructure creates the complete organization structure (Holding->Company->Store)
// This is used by Platform Admins to onboard new users with complete business structure
func (s *organizationService) CreateCompleteOrganizationStructure(ctx context.Context, req dto.CreateCompleteStructureRequest, createdBy uuid.UUID) (*dto.CompleteOrganizationStructureResponse, error) {
	// Validate that the creator is Platform level (90-100)
	creator, err := s.userRepo.FindByIDWithRole(ctx, createdBy)
	if err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to find creator user: %w", err))
	}

	// Check if creator is platform level
	isPlatformUser := creator.Role != nil && (creator.Role.Level >= constant.RoleLevelPlatformManager)
	if !isPlatformUser {
		return nil, apperror.NewForbiddenError("Only Platform level users can create complete organization structures")
	}

	// Validate that target user exists
	targetUser, err := s.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		return nil, apperror.NewNotFoundError("target user not found")
	}

	// Step 1: Create Holding Company (root level)
	holdingCode, err := s.GenerateUniqueCode(ctx, req.HoldingName)
	if err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to generate holding code: %w", err))
	}

	holding := &model.Organization{
		Name:             req.HoldingName,
		Code:             holdingCode,
		OrganizationType: constant.OrganizationTypeHolding,
		Description:      req.HoldingDescription,
		CreatedBy:        createdBy,
		IsActive:         true,
	}

	createdHolding, err := s.orgRepo.Create(ctx, holding)
	if err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to create holding organization: %w", err))
	}

	// Step 2: Create Company under Holding
	companyCode, err := s.GenerateUniqueCode(ctx, req.CompanyName)
	if err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to generate company code: %w", err))
	}

	company := &model.Organization{
		Name:                 req.CompanyName,
		Code:                 companyCode,
		OrganizationType:     constant.OrganizationTypeCompany,
		ParentOrganizationID: &createdHolding.ID,
		Description:          req.CompanyDescription,
		CreatedBy:            createdBy,
		IsActive:             true,
	}

	createdCompany, err := s.orgRepo.Create(ctx, company)
	if err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to create company organization: %w", err))
	}

	// Step 3: Create Store under Company
	storeCode, err := s.GenerateUniqueCode(ctx, req.StoreName)
	if err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to generate store code: %w", err))
	}

	store := &model.Organization{
		Name:                 req.StoreName,
		Code:                 storeCode,
		OrganizationType:     constant.OrganizationTypeStore,
		ParentOrganizationID: &createdCompany.ID,
		Description:          req.StoreDescription,
		CreatedBy:            createdBy,
		IsActive:             true,
	}

	createdStore, err := s.orgRepo.Create(ctx, store)
	if err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to create store organization: %w", err))
	}

	// Step 4: Assign user to all organizations (typically they'll be holding owner)
	organizations := []*model.Organization{createdHolding, createdCompany, createdStore}
	var userMemberships []dto.UserOrganizationResponse

	for _, org := range organizations {
		userOrg := &model.UserOrganization{
			UserID:         req.UserID,
			OrganizationID: org.ID,
			IsActive:       true,
		}

		if err := s.orgRepo.AddUserToOrganization(ctx, userOrg); err != nil {
			return nil, apperror.NewInternalError(fmt.Errorf("failed to add user to organization %s: %w", org.Name, err))
		}

		// Add to response
		membership := dto.UserOrganizationResponse{
			UserID:         req.UserID,
			OrganizationID: org.ID,
			Organization:   *util.MapOrganizationToResponse(org),
			JoinedAt:       userOrg.JoinedAt,
			IsActive:       true,
		}
		userMemberships = append(userMemberships, membership)
	}

	// Build response
	response := &dto.CompleteOrganizationStructureResponse{
		Message:        fmt.Sprintf("Complete organization structure created successfully for user %s", targetUser.Username),
		HoldingID:      createdHolding.ID,
		CompanyID:      createdCompany.ID,
		StoreID:        createdStore.ID,
		UserAssignedTo: req.UserID,
		UserRole:       req.UserRole,
		Organizations: []dto.OrganizationResponse{
			*util.MapOrganizationToResponse(createdHolding),
			*util.MapOrganizationToResponse(createdCompany),
			*util.MapOrganizationToResponse(createdStore),
		},
		UserMemberships: userMemberships,
	}

	return response, nil
}

// GetOrganizationStatistics returns organization statistics based on user's role level
func (s *organizationService) GetOrganizationStatistics(ctx context.Context, userID uuid.UUID) (*dto.OrganizationStatisticsResponse, error) {
	// Get user with role information
	user, err := s.userRepo.FindByIDWithRole(ctx, userID)
	if err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to get user information: %w", err))
	}
	if user == nil {
		return nil, apperror.NewValidationError("User not found")
	}

	// Get organization counts based on user's role level
	var platformCount, holdingCount, companyCount, storeCount int64

	// Determine what statistics to show based on user's role level
	if user.Role != nil {
		// Platform users (highest level) can see all statistics
		if user.Role.Level >= constant.RoleLevelPlatformManager {
			// For now, count all organizations by type (later we can add platform filtering)
			if platformCount, err = s.countOrganizationsByType(ctx, constant.OrganizationTypePlatform); err != nil {
				return nil, apperror.NewInternalError(fmt.Errorf("failed to count platform organizations: %w", err))
			}
			if holdingCount, err = s.countOrganizationsByType(ctx, constant.OrganizationTypeHolding); err != nil {
				return nil, apperror.NewInternalError(fmt.Errorf("failed to count holding organizations: %w", err))
			}
			if companyCount, err = s.countOrganizationsByType(ctx, constant.OrganizationTypeCompany); err != nil {
				return nil, apperror.NewInternalError(fmt.Errorf("failed to count company organizations: %w", err))
			}
			if storeCount, err = s.countOrganizationsByType(ctx, constant.OrganizationTypeStore); err != nil {
				return nil, apperror.NewInternalError(fmt.Errorf("failed to count store organizations: %w", err))
			}
		} else if user.Role.Level >= constant.RoleLevelHoldingOwner {
			// Holding users can see holding, company, and store counts
			// TODO: Filter by user's organization access
			if holdingCount, err = s.countOrganizationsByType(ctx, constant.OrganizationTypeHolding); err != nil {
				return nil, apperror.NewInternalError(fmt.Errorf("failed to count holding organizations: %w", err))
			}
			if companyCount, err = s.countOrganizationsByType(ctx, constant.OrganizationTypeCompany); err != nil {
				return nil, apperror.NewInternalError(fmt.Errorf("failed to count company organizations: %w", err))
			}
			if storeCount, err = s.countOrganizationsByType(ctx, constant.OrganizationTypeStore); err != nil {
				return nil, apperror.NewInternalError(fmt.Errorf("failed to count store organizations: %w", err))
			}
		} else if user.Role.Level >= constant.RoleLevelCompanyManager {
			// Company users can only see company and store counts
			// TODO: Filter by user's organization access
			if companyCount, err = s.countOrganizationsByType(ctx, constant.OrganizationTypeCompany); err != nil {
				return nil, apperror.NewInternalError(fmt.Errorf("failed to count company organizations: %w", err))
			}
			if storeCount, err = s.countOrganizationsByType(ctx, constant.OrganizationTypeStore); err != nil {
				return nil, apperror.NewInternalError(fmt.Errorf("failed to count store organizations: %w", err))
			}
		} else {
			// Store level users can only see store counts
			// TODO: Filter by user's organization access
			if storeCount, err = s.countOrganizationsByType(ctx, constant.OrganizationTypeStore); err != nil {
				return nil, apperror.NewInternalError(fmt.Errorf("failed to count store organizations: %w", err))
			}
		}
	}

	return &dto.OrganizationStatisticsResponse{
		PlatformCount: platformCount,
		HoldingCount:  holdingCount,
		CompanyCount:  companyCount,
		StoreCount:    storeCount,
	}, nil
}

// countOrganizationsByType is a helper method to count organizations by type
func (s *organizationService) countOrganizationsByType(ctx context.Context, orgType string) (int64, error) {
	filters := map[string]interface{}{
		"organization_type": orgType,
	}
	return s.orgRepo.Count(ctx, filters)
}

// GetAccessibleOrganizationIDs returns all organization IDs that the current user can access
// based on their role level and organization hierarchy
func (s *organizationService) GetAccessibleOrganizationIDs(ctx context.Context, currentUserID uuid.UUID, currentUserLevel int) ([]uuid.UUID, error) {
	// Platform users (level >= 76) can access all organizations
	if currentUserLevel >= 76 {
		return s.getAllOrganizationIDs(ctx)
	}

	// Get user's direct organization assignments
	userOrgs, err := s.userRepo.FindUserOrganizations(ctx, currentUserID, 0, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to get user organizations: %w", err)
	}

	if len(userOrgs) == 0 {
		return []uuid.UUID{}, nil
	}

	var accessibleOrgIDs []uuid.UUID
	processedOrgIDs := make(map[uuid.UUID]bool)

	// For each organization the user belongs to, get all organizations they can access
	for _, userOrg := range userOrgs {
		if processedOrgIDs[userOrg.OrganizationID] {
			continue
		}
		processedOrgIDs[userOrg.OrganizationID] = true

		// Get the organization with its details
		org, err := s.orgRepo.FindByIDWithDetails(ctx, userOrg.OrganizationID)
		if err != nil {
			continue // Skip if organization not found
		}

		// Add current organization
		accessibleOrgIDs = append(accessibleOrgIDs, org.ID)

		// Based on organization type and user level, determine accessible child organizations
		childOrgIDs, err := s.getAccessibleChildOrganizations(ctx, org, currentUserLevel)
		if err != nil {
			continue // Skip if error getting children
		}

		for _, childID := range childOrgIDs {
			if !processedOrgIDs[childID] {
				accessibleOrgIDs = append(accessibleOrgIDs, childID)
				processedOrgIDs[childID] = true
			}
		}
	}

	return accessibleOrgIDs, nil
}

// getAccessibleChildOrganizations recursively gets all child organizations that the user can access
func (s *organizationService) getAccessibleChildOrganizations(ctx context.Context, org *model.Organization, userLevel int) ([]uuid.UUID, error) {
	var childIDs []uuid.UUID

	// Get direct children of current organization
	children, err := s.orgRepo.FindByParent(ctx, org.ID)
	if err != nil {
		return childIDs, err
	}

	for _, child := range children {
		// Determine if user can access this child based on organization hierarchy rules
		if s.canAccessChildOrganization(org.OrganizationType, child.OrganizationType, userLevel) {
			childIDs = append(childIDs, child.ID)

			// Recursively get children of this child
			grandChildIDs, err := s.getAccessibleChildOrganizations(ctx, &child, userLevel)
			if err != nil {
				continue // Skip if error, but continue with other children
			}
			childIDs = append(childIDs, grandChildIDs...)
		}
	}

	return childIDs, nil
}

// canAccessChildOrganization determines if a user can access a child organization
// based on organization hierarchy and user level
func (s *organizationService) canAccessChildOrganization(parentType, childType string, userLevel int) bool {
	// Organization hierarchy rules:
	// - Holding (level 51-75) can access Company and Store
	// - Company (level 26-50) can access Store only
	// - Store (level 1-25) cannot access other organizations

	switch parentType {
	case constant.OrganizationTypeHolding:
		// Holding users can access Company and Store children
		return childType == constant.OrganizationTypeCompany || childType == constant.OrganizationTypeStore
	case constant.OrganizationTypeCompany:
		// Company users can access Store children only
		return childType == constant.OrganizationTypeStore
	case constant.OrganizationTypeStore:
		// Store users cannot access child organizations (stores don't have children typically)
		return false
	default:
		return false
	}
}

// getAllOrganizationIDs returns all organization IDs for platform users
func (s *organizationService) getAllOrganizationIDs(ctx context.Context) ([]uuid.UUID, error) {
	orgs, err := s.orgRepo.FindAll(ctx, map[string]interface{}{})
	if err != nil {
		return nil, fmt.Errorf("failed to get all organizations: %w", err)
	}

	var orgIDs []uuid.UUID
	for _, org := range orgs {
		orgIDs = append(orgIDs, org.ID)
	}

	return orgIDs, nil
}
