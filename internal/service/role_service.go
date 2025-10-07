package service

import (
	"beresin-backend/internal/apperror"
	"beresin-backend/internal/constant"
	"beresin-backend/internal/dto"
	"beresin-backend/internal/model"
	"beresin-backend/internal/repository"
	"beresin-backend/internal/util"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// roleService implements the RoleService interface for role management.
type roleService struct {
	roleRepo             repository.RoleRepositoryInterface
	authorizationService AuthorizationServiceInterface
}

// NewRoleService creates a new instance of roleService.
func NewRoleService(roleRepo repository.RoleRepositoryInterface, authorizationService AuthorizationServiceInterface) RoleServiceInterface {
	return &roleService{
		roleRepo:             roleRepo,
		authorizationService: authorizationService,
	}
}
func (s *roleService) CreateRole(ctx context.Context, req dto.CreateRoleRequest, userLevel int) (*dto.RoleResponse, error) {
	// Hierarchical validation: User can only create roles below their own level
	if req.Level >= userLevel {
		return nil, apperror.NewValidationError(fmt.Sprintf("Cannot create role with level %d. Role level must be below your current level (%d)", req.Level, userLevel))
	}

	// Business validation: Level 100 is reserved for platform admin only
	if req.Level >= constant.RoleLevelPlatformAdmin {
		return nil, apperror.NewValidationError("Level 100 is reserved for platform admin and cannot be created")
	}

	// Validate role level constraints - minimum store staff level
	minLevel := constant.RoleLevelStoreStaff        // Level 10
	maxLevel := constant.RoleLevelPlatformAdmin - 1 // Level 99
	if req.Level < minLevel || req.Level >= constant.RoleLevelPlatformAdmin {
		return nil, apperror.NewValidationError(fmt.Sprintf("Role level must be between %d and %d", minLevel, maxLevel))
	}

	// Use repository to check if role name already exists (more efficient)
	existingRole, err := s.roleRepo.FindByName(ctx, req.Name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Error().Err(err).Str("role_name", req.Name).Msg("Failed to check existing role")
		return nil, apperror.NewInternalError(fmt.Errorf("failed to validate role name: %w", err))
	}
	if existingRole != nil {
		return nil, apperror.NewConflictError(fmt.Sprintf("Role with name '%s' already exists", req.Name))
	}

	// Create role model with proper field values
	newRole := &model.Role{
		Name:           req.Name,
		Description:    req.Description,
		Level:          req.Level,
		PredefinedName: req.PredefinedName,
		IsSystemRole:   false, // User-created roles are not system roles
		IsActive:       true,  // New roles are active by default
	}

	// Use repository create method
	createdRole, err := s.roleRepo.Create(ctx, newRole)
	if err != nil {
		log.Error().Err(err).Interface("role", newRole).Msg("Failed to create role")
		return nil, apperror.NewInternalError(fmt.Errorf("failed to create role: %w", err))
	}

	// Determine organization types to assign
	organizationTypes := req.OrganizationTypes
	if len(organizationTypes) == 0 {
		// Default: assign to all organization types if none specified
		organizationTypes = []string{"platform", "holding", "company", "store"}
	}

	// Create organization type mappings
	err = s.roleRepo.CreateRoleOrganizationTypes(ctx, createdRole.ID, organizationTypes)
	if err != nil {
		log.Error().Err(err).Str("role_id", createdRole.ID.String()).Msg("Failed to create role organization types")
		// Note: Role is already created, so we continue but log the error
	}

	// Get the created role with organization types for response
	roleResponse := util.MapRoleToResponse(createdRole)
	roleResponse.OrganizationTypes = organizationTypes

	return roleResponse, nil
}

// ListRoles retrieves all roles below the user's level (hierarchical access control).
// Users can only see roles that are below their own level for security.
func (s *roleService) ListRoles(ctx context.Context, userLevel int) ([]dto.RoleResponse, error) {
	// Get all roles from repository
	roles, err := s.roleRepo.FindAll(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch roles from repository")
		return nil, apperror.NewInternalError(fmt.Errorf("failed to fetch roles: %w", err))
	}

	var filteredRoles []model.Role

	// Apply hierarchical filtering: users can only see roles with level below their own level
	// ALWAYS exclude super_admin role from the list for security and UI consistency
	if userLevel >= constant.RoleLevelSuperAdmin {
		// Super Admin (level 100) sees all roles except super_admin role itself
		for _, role := range roles {
			// Exclude super_admin role from being displayed in Role Management
			if role.Name != "super_admin" && role.PredefinedName != "super_admin" {
				filteredRoles = append(filteredRoles, role)
			}
		}
	} else {
		// All other users (including Platform Admins level 99) only see roles below their level
		// and exclude super_admin role
		for _, role := range roles {
			if role.Level < userLevel && role.Name != "super_admin" && role.PredefinedName != "super_admin" {
				filteredRoles = append(filteredRoles, role)
			}
		}
	}

	// Convert to response format
	response := make([]dto.RoleResponse, len(filteredRoles))
	for i, role := range filteredRoles {
		roleResponse := util.MapRoleToResponse(&role)

		// Map permissions if available
		if len(role.Permissions) > 0 {
			permissions := make([]string, len(role.Permissions))
			for j, permission := range role.Permissions {
				permissions[j] = permission.Name
			}
			roleResponse.Permissions = permissions
		}

		response[i] = *roleResponse
	}

	log.Info().
		Int("user_level", userLevel).
		Int("total_roles", len(roles)).
		Int("filtered_roles", len(response)).
		Msg("Successfully retrieved roles with hierarchical filtering")
	return response, nil
}

// GetRoleByID retrieves a single role by ID.
func (s *roleService) GetRoleByID(ctx context.Context, roleID uuid.UUID) (*dto.RoleResponse, error) {
	role, err := s.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NewNotFoundError("role")
		}
		log.Error().Err(err).Str("role_id", roleID.String()).Msg("Failed to fetch role by ID from repository")
		return nil, apperror.NewInternalError(fmt.Errorf("failed to fetch role: %w", err))
	}

	roleResponse := util.MapRoleToResponse(role)
	log.Info().Str("role_id", roleID.String()).Int("level", role.Level).Msg("Successfully retrieved role by ID")
	return roleResponse, nil
}

// UpdateRole updates an existing role with hierarchical validation.
func (s *roleService) UpdateRole(ctx context.Context, roleID uuid.UUID, req dto.UpdateRoleRequest, userLevel int) (*dto.RoleResponse, error) {
	// Get existing role first
	existingRole, err := s.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NewNotFoundError("role")
		}
		log.Error().Err(err).Str("role_id", roleID.String()).Msg("Failed to fetch role for update")
		return nil, apperror.NewInternalError(fmt.Errorf("failed to fetch role: %w", err))
	}

	// Hierarchical validation: User can only update roles below their own level
	if existingRole.Level >= userLevel {
		return nil, apperror.NewValidationError(fmt.Sprintf("Cannot update role with level %d. You can only update roles below your level (%d)", existingRole.Level, userLevel))
	}

	// Hierarchical validation: User can only set role level below their own level
	if req.Level >= userLevel {
		return nil, apperror.NewValidationError(fmt.Sprintf("Cannot set role level to %d. Role level must be below your current level (%d)", req.Level, userLevel))
	}

	// Business validation: Level 100 is reserved for platform admin only
	if req.Level >= constant.RoleLevelPlatformAdmin {
		return nil, apperror.NewValidationError("Level 100 is reserved for platform admin and cannot be set")
	}

	// Validate role level constraints - minimum store staff level
	minLevel := constant.RoleLevelStoreStaff        // Level 10
	maxLevel := constant.RoleLevelPlatformAdmin - 1 // Level 99
	if req.Level < minLevel || req.Level >= constant.RoleLevelPlatformAdmin {
		return nil, apperror.NewValidationError(fmt.Sprintf("Role level must be between %d and %d", minLevel, maxLevel))
	}

	// Check if new name conflicts with existing roles (excluding current role)
	if req.Name != existingRole.Name {
		existingByName, err := s.roleRepo.FindByName(ctx, req.Name)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Str("role_name", req.Name).Msg("Failed to check existing role name")
			return nil, apperror.NewInternalError(fmt.Errorf("failed to validate role name: %w", err))
		}
		if existingByName != nil && existingByName.ID != roleID {
			return nil, apperror.NewConflictError(fmt.Sprintf("Role with name '%s' already exists", req.Name))
		}
	}

	// Update role fields
	existingRole.Name = req.Name
	existingRole.Description = req.Description
	existingRole.Level = req.Level
	existingRole.PredefinedName = req.PredefinedName

	// Update role using repository
	updatedRole, err := s.roleRepo.Update(ctx, existingRole)
	if err != nil {
		log.Error().Err(err).Interface("role", existingRole).Msg("Failed to update role")
		return nil, apperror.NewInternalError(fmt.Errorf("failed to update role: %w", err))
	}

	// Update organization type mappings if provided
	if len(req.OrganizationTypes) > 0 {
		err = s.roleRepo.UpdateRoleOrganizationTypes(ctx, roleID, req.OrganizationTypes)
		if err != nil {
			log.Error().Err(err).Str("role_id", roleID.String()).Msg("Failed to update role organization types")
			// Continue but log error
		}
	}

	log.Info().
		Str("role_id", roleID.String()).
		Str("role_name", updatedRole.Name).
		Int("old_level", existingRole.Level).
		Int("new_level", req.Level).
		Int("user_level", userLevel).
		Msg("Successfully updated role")

	// Return structured response with organization types
	roleResponse := util.MapRoleToResponse(updatedRole)
	if len(req.OrganizationTypes) > 0 {
		roleResponse.OrganizationTypes = req.OrganizationTypes
	}

	return roleResponse, nil
}

// UpdateRolePermissions updates the permissions associated with a role.
func (s *roleService) UpdateRolePermissions(ctx context.Context, roleID uuid.UUID, permissionNames []string) error {
	// Call a single repository method that handles the transaction.
	if err := s.roleRepo.UpdateRolePermissions(ctx, roleID, permissionNames); err != nil {
		// Translate repository error to AppError
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NewNotFoundError("role")
		}
		// For other database errors, return as an internal error
		return apperror.NewInternalError(fmt.Errorf("failed to update role permissions in db: %w", err))
	}

	if err := s.authorizationService.InvalidateRolePermissionsCache(ctx, roleID); err != nil {
		log.Error().Err(err).Msgf("CRITICAL: DB updated but failed to invalidate cache for role %s", roleID)
	}

	return nil
}

// DISABLED: Role approval functionality
/*
// CreateRoleApprovalRequest creates a new role approval request.
func (s *roleService) CreateRoleApprovalRequest(ctx context.Context, req dto.CreateRoleApprovalRequest, requestedBy uuid.UUID) (*dto.RoleApprovalResponse, error) {
	// Additional validation: Level 100 is reserved for superadmin only
	if req.RequestedLevel >= 100 {
		return nil, apperror.NewValidationError("Level 100 is reserved for superadmin and cannot be requested")
	}

	newApproval := &model.RoleApproval{
		RequestedRoleName: req.RequestedRoleName,
		RequestedLevel:    req.RequestedLevel,
		Description:       req.Description,
		RequestedBy:       requestedBy,
		Status:            "pending",
	}

	createdApproval, err := s.roleRepo.CreateRoleApproval(ctx, newApproval)
	if err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to create role approval request: %w", err))
	}

	// Fetch the approval with user information
	approvalWithUsers, err := s.roleRepo.FindRoleApprovalByID(ctx, createdApproval.ID)
	if err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to fetch created approval: %w", err))
	}

	return s.mapRoleApprovalToResponse(approvalWithUsers), nil
}
*/

/*
// ListRoleApprovalRequests retrieves all role approval requests.
func (s *roleService) ListRoleApprovalRequests(ctx context.Context) ([]dto.RoleApprovalResponse, error) {
	approvals, err := s.roleRepo.FindAllRoleApprovals(ctx)
	if err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to fetch role approval requests: %w", err))
	}

	response := make([]dto.RoleApprovalResponse, len(approvals))
	for i, approval := range approvals {
		response[i] = *s.mapRoleApprovalToResponse(&approval)
	}

	return response, nil
}
*/

/*
// ApproveRejectRoleRequest approves or rejects a role approval request.
func (s *roleService) ApproveRejectRoleRequest(ctx context.Context, approvalID uuid.UUID, decision dto.ApprovalDecisionRequest, approverID uuid.UUID) (*dto.RoleApprovalResponse, error) {
	approval, err := s.roleRepo.FindRoleApprovalByID(ctx, approvalID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NewNotFoundError("role approval request")
		}
		return nil, apperror.NewInternalError(fmt.Errorf("failed to fetch approval request: %w", err))
	}

	if approval.Status != "pending" {
		return nil, apperror.NewConflictError("approval request has already been processed")
	}

	approval.Status = decision.Status
	approval.Reason = decision.Reason
	approval.ApproverID = &approverID

	updatedApproval, err := s.roleRepo.UpdateRoleApproval(ctx, approval)
	if err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to update approval request: %w", err))
	}

	// If approved, create the actual role
	if decision.Status == "approved" {
		roleRequest := dto.CreateRoleRequest{
			Name:           approval.RequestedRoleName,
			Description:    approval.Description,
			Level:          approval.RequestedLevel,
			PredefinedName: s.getPredefinedNameByLevel(approval.RequestedLevel),
		}

		// For approved role creation, use system-level access (level 100) to bypass hierarchical validation
		// since the approval process already validates the appropriateness of the role level
		_, err := s.CreateRole(ctx, roleRequest, constant.RoleLevelPlatformAdmin)
		if err != nil {
			// Log the error but don't fail the approval process
			log.Error().Err(err).Msgf("Failed to create role after approval for request %s", approvalID)
		}
	}

	return s.mapRoleApprovalToResponse(updatedApproval), nil
}
*/

// GetPredefinedRoleOptions returns the available predefined role options based on user's level (hierarchical access control).
func (s *roleService) GetPredefinedRoleOptions(ctx context.Context, userLevel int) ([]dto.PredefinedRoleOption, error) {
	// Define all predefined role options
	allOptions := []dto.PredefinedRoleOption{
		{
			Name:        "platform_admin",
			Level:       99,
			Description: "Platform Administrator - Highest manageable access level for system-wide operations",
		},
		// {
		// 	Name:        "platform_staff",
		// 	Level:       76,
		// 	Description: "Platform Staff - Advanced platform operations and cross-organization management",
		// },
		{
			Name:        "holding_admin",
			Level:       75, // Level 51-75: Holding level (keeping 75 as highest holding level)
			Description: "Holding Administrator - Manage holding companies and their subsidiaries",
		},
		// {
		// 	Name:        "holding_staff",
		// 	Level:       51,
		// 	Description: "Holding Staff - Support holding-level operations and reporting",
		// },
		{
			Name:        "company_admin",
			Level:       50, // Level 26-50: Company level (keeping 50 as highest company level)
			Description: "Company Administrator - Manage company operations and store networks",
		},
		// {
		// 	Name:        "company_staff",
		// 	Level:       26,
		// 	Description: "Company Staff - Company-level operational access and store coordination",
		// },
		{
			Name:        "store_admin",
			Level:       25,
			Description: "Store Administrator - Manage individual store operations and staff",
		},
		// {
		// 	Name:        "store_staff",
		// 	Level:       1,
		// 	Description: "Store Staff - Basic store operational access and customer service",
		// },
	}

	// Filter options: only return roles with level below user's level (hierarchical access control)
	// Special case: userLevel 0 means new user without role - show basic role options
	var filteredOptions []dto.PredefinedRoleOption
	for _, option := range allOptions {
		if userLevel == 0 {
			// For new users, only show store and company level roles (not platform/holding)
			if option.Level <= 50 {
				filteredOptions = append(filteredOptions, option)
			}
		} else if option.Level < userLevel {
			filteredOptions = append(filteredOptions, option)
		}
	}

	log.Info().
		Int("user_level", userLevel).
		Int("total_options", len(allOptions)).
		Int("filtered_options", len(filteredOptions)).
		Bool("is_new_user", userLevel == 0).
		Msg("Successfully retrieved predefined role options with hierarchical filtering")

	return filteredOptions, nil
}

/*
// Helper function to map RoleApproval model to response DTO
func (s *roleService) mapRoleApprovalToResponse(approval *model.RoleApproval) *dto.RoleApprovalResponse {
	response := &dto.RoleApprovalResponse{
		ID:                approval.ID,
		RequestedRoleName: approval.RequestedRoleName,
		RequestedLevel:    approval.RequestedLevel,
		Description:       approval.Description,
		RequestedBy:       approval.RequestedBy,
		Status:            approval.Status,
		Reason:            approval.Reason,
		CreatedAt:         approval.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:         approval.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if approval.RequestedByUser.Username != "" {
		response.RequestedByUser = approval.RequestedByUser.Username
	}

	if approval.ApproverID != nil {
		response.ApproverID = approval.ApproverID
		if approval.Approver != nil && approval.Approver.Username != "" {
			approverUsername := approval.Approver.Username
			response.Approver = &approverUsername
		}
	}

	return response
}
*/

// Helper function to get predefined name by level
func (s *roleService) getPredefinedNameByLevel(level int) string {
	switch level {
	case 1000:
		return "Nexus"
	case 500:
		return "Oreo"
	case 50:
		return "Jellybean"
	default:
		return "Custom"
	}
}

// Permission Management Service Methods

// ListPermissions returns all permissions in the system.
func (s *roleService) ListPermissions(ctx context.Context) ([]dto.PermissionResponse, error) {
	permissions, err := s.roleRepo.GetAllPermissions(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve permissions from database")
		return nil, apperror.NewInternalError(err)
	}

	var response []dto.PermissionResponse
	for _, permission := range permissions {
		response = append(response, dto.PermissionResponse{
			ID:          permission.ID,
			Name:        permission.Name,
			Description: permission.Description,
			CreatedAt:   permission.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   permission.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return response, nil
}

// CreatePermission creates a new permission in the system.
func (s *roleService) CreatePermission(ctx context.Context, req dto.CreatePermissionRequest) (*dto.PermissionResponse, error) {
	// Input validation
	if req.Name == "" {
		return nil, apperror.NewValidationError("Permission name is required")
	}

	// Use repository method to check if permission already exists (efficient)
	exists, err := s.roleRepo.CheckPermissionExists(ctx, req.Name)
	if err != nil {
		log.Error().Err(err).Str("permission_name", req.Name).Msg("Failed to check permission existence")
		return nil, apperror.NewInternalError(fmt.Errorf("failed to validate permission: %w", err))
	}
	if exists {
		return nil, apperror.NewConflictError(fmt.Sprintf("Permission with name '%s' already exists", req.Name))
	}

	// Create permission model
	permission := &model.Permission{
		Name:        req.Name,
		Description: req.Description,
	}

	// Use repository method to create permission
	if err := s.roleRepo.CreatePermission(ctx, permission); err != nil {
		log.Error().Err(err).Interface("permission", permission).Msg("Failed to create permission")
		return nil, apperror.NewInternalError(fmt.Errorf("failed to create permission: %w", err))
	}

	log.Info().Str("permission_name", permission.Name).Str("permission_id", permission.ID.String()).Msg("Permission created successfully")

	// Return structured response using helper method
	return util.MapPermissionToResponse(permission), nil
}

// UpdatePermission updates an existing permission.
func (s *roleService) UpdatePermission(ctx context.Context, id uuid.UUID, req dto.UpdatePermissionRequest) (*dto.PermissionResponse, error) {
	// Find existing permission
	permission, err := s.roleRepo.FindPermissionByID(ctx, id)
	if err != nil {
		log.Error().Err(err).Str("permission_id", id.String()).Msg("Failed to find permission")
		return nil, apperror.NewInternalError(err)
	}
	if permission == nil {
		return nil, apperror.NewNotFoundError("Permission not found")
	}

	// Check if new name conflicts with existing permissions (excluding current permission)
	if req.Name != permission.Name {
		existingPermission, err := s.roleRepo.FindPermissionByName(ctx, req.Name)
		if err != nil {
			log.Error().Err(err).Str("permission_name", req.Name).Msg("Failed to check permission name uniqueness")
			return nil, apperror.NewInternalError(err)
		}
		if existingPermission != nil && existingPermission.ID != id {
			return nil, apperror.NewConflictError(fmt.Sprintf("permission with name '%s' already exists", req.Name))
		}
	}

	// Update permission fields
	permission.Name = req.Name
	permission.Description = req.Description

	if err := s.roleRepo.UpdatePermission(ctx, permission); err != nil {
		log.Error().Err(err).Interface("permission", permission).Msg("Failed to update permission")
		return nil, apperror.NewInternalError(err)
	}

	return &dto.PermissionResponse{
		ID:          permission.ID,
		Name:        permission.Name,
		Description: permission.Description,
		CreatedAt:   permission.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   permission.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

// DeletePermission deletes a permission from the system.
func (s *roleService) DeletePermission(ctx context.Context, id uuid.UUID) error {
	// Check if permission exists
	permission, err := s.roleRepo.FindPermissionByID(ctx, id)
	if err != nil {
		log.Error().Err(err).Str("permission_id", id.String()).Msg("Failed to find permission")
		return apperror.NewInternalError(err)
	}
	if permission == nil {
		return apperror.NewNotFoundError("Permission not found")
	}

	// Attempt to delete (repository will check if permission is in use)
	if err := s.roleRepo.DeletePermission(ctx, id); err != nil {
		if err.Error() == "cannot delete permission: it is currently assigned to one or more roles" {
			return apperror.NewConflictError("Cannot delete permission: it is currently assigned to one or more roles")
		}
		log.Error().Err(err).Str("permission_id", id.String()).Msg("Failed to delete permission")
		return apperror.NewInternalError(err)
	}

	log.Info().Str("permission_id", id.String()).Str("permission_name", permission.Name).Msg("Permission deleted successfully")
	return nil
}

// Organization-specific role methods implementation

// GetRolesForOrganizationType returns roles that are applicable to a specific organization type.
func (s *roleService) GetRolesForOrganizationType(ctx context.Context, organizationType string, userLevel int) ([]dto.RoleResponse, error) {
	// Validate organization type
	validTypes := []string{"platform", "holding", "company", "store"}
	isValid := false
	for _, validType := range validTypes {
		if organizationType == validType {
			isValid = true
			break
		}
	}
	if !isValid {
		return nil, apperror.NewValidationError("Invalid organization type. Must be one of: platform, holding, company, store")
	}

	// Get roles for the organization type
	roles, err := s.roleRepo.FindRolesByOrganizationType(ctx, organizationType)
	if err != nil {
		log.Error().Err(err).Str("organization_type", organizationType).Msg("Failed to get roles for organization type")
		return nil, apperror.NewInternalError(err)
	}

	// Filter roles based on user level (users can only see/assign roles below their level)
	var filteredRoles []model.Role
	for _, role := range roles {
		if role.Level < userLevel {
			filteredRoles = append(filteredRoles, role)
		}
	}

	// Convert to DTOs
	roleResponses := make([]dto.RoleResponse, len(filteredRoles))
	for i, role := range filteredRoles {
		permissionNames := make([]string, len(role.Permissions))
		for j, permission := range role.Permissions {
			permissionNames[j] = permission.Name
		}

		// Get organization types for this role
		organizationTypes := make([]string, len(role.OrganizationTypes))
		for j, orgType := range role.OrganizationTypes {
			organizationTypes[j] = orgType.OrganizationType
		}

		roleResponses[i] = dto.RoleResponse{
			ID:                role.ID,
			Name:              role.Name,
			Description:       role.Description,
			Level:             role.Level,
			IsSystemRole:      role.IsSystemRole,
			PredefinedName:    role.PredefinedName,
			IsActive:          role.IsActive,
			OrganizationTypes: organizationTypes,
			Permissions:       permissionNames,
		}
	}

	return roleResponses, nil
}

// AssignRoleToUserInOrganization assigns a role to a user within a specific organization context.
func (s *roleService) AssignRoleToUserInOrganization(ctx context.Context, req dto.OrganizationRoleAssignmentRequest) (*dto.OrganizationRoleResponse, error) {
	// This functionality is handled by the user service's AssignUserToOrganization method
	// which already includes role assignment. This method can be implemented for more
	// specific role management if needed.

	// For now, return a placeholder implementation
	return nil, apperror.NewValidationError("Use UserService.AssignUserToOrganization for role assignment")
}

// GetUserRolesInOrganization retrieves all roles a user has in a specific organization.
func (s *roleService) GetUserRolesInOrganization(ctx context.Context, userID, organizationID uuid.UUID) ([]dto.OrganizationRoleResponse, error) {
	// This functionality is handled by the user service's GetUserOrganizations method
	// which already includes role information. This method can be implemented for more
	// specific role queries if needed.

	// For now, return a placeholder implementation
	return nil, apperror.NewValidationError("Use UserService.GetUserOrganizations for user role information")
}
