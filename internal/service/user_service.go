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
	"math"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// userService implements the UserService interface for user management.
type userService struct {
	userRepo   repository.UserRepositoryInterface
	roleRepo   repository.RoleRepositoryInterface
	orgService OrganizationServiceInterface
	redis      *redis.Client
}

// NewUserService creates a new instance of userService.
func NewUserService(userRepo repository.UserRepositoryInterface, roleRepo repository.RoleRepositoryInterface, orgService OrganizationServiceInterface, redisClient *redis.Client) UserServiceInterface {
	return &userService{
		userRepo:   userRepo,
		roleRepo:   roleRepo,
		orgService: orgService,
		redis:      redisClient,
	}
}

// CreateUser creates a new user with enhanced validation and error handling
func (s *userService) CreateUser(ctx context.Context, req dto.CreateUserRequest) (*dto.UserResponse, error) {
	// Validate username uniqueness using existing repository method
	_, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err == nil {
		return nil, apperror.NewConflictError("username already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to check username availability: %w", err))
	}

	// Validate email uniqueness using existing repository method
	_, err = s.userRepo.FindByEmail(ctx, req.Email)
	if err == nil {
		return nil, apperror.NewConflictError("email already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to check email availability: %w", err))
	}

	// Validate role exists
	if err := s.validateRoleExists(ctx, req.RoleID); err != nil {
		return nil, err
	}

	// Set default auth provider if not specified
	authProvider := req.AuthProvider
	if authProvider == "" {
		authProvider = "local"
	}

	// Validate auth provider and associated data
	if authProvider == "google" {
		if req.GoogleID == nil || *req.GoogleID == "" {
			return nil, apperror.NewValidationError("google_id is required for Google authentication")
		}
		// Check if GoogleID already exists
		existingUser, err := s.userRepo.FindByGoogleID(ctx, *req.GoogleID)
		if err == nil && existingUser != nil {
			return nil, apperror.NewConflictError("google_id already exists")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NewInternalError(fmt.Errorf("failed to check google_id availability: %w", err))
		}
	} else if authProvider == "local" {
		if req.Password == "" {
			return nil, apperror.NewValidationError("password is required for local authentication")
		}
		// Ensure GoogleID is nil for local users
		req.GoogleID = nil
	}

	// Hash password if provided (required for local auth, optional for OAuth)
	var hashedPassword string
	if req.Password != "" {
		var err error
		hashedPassword, err = s.hashPassword(req.Password)
		if err != nil {
			return nil, apperror.NewInternalError(fmt.Errorf("failed to hash password: %w", err))
		}
	}

	// Create user with validated data
	newUser := &model.User{
		Username:     req.Username,
		Email:        req.Email,
		Password:     hashedPassword,
		RoleID:       &req.RoleID,
		AuthProvider: authProvider,
		GoogleID:     req.GoogleID,
	}

	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to create user: %w", err))
	}

	// Fetch the newly created user with role details using repository method
	createdUser, err := s.userRepo.FindByIDWithRole(ctx, newUser.ID)
	if err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to fetch created user: %w", err))
	}

	return util.MapUserToResponse(createdUser), nil
}

// ListUsers retrieves a paginated list of users with level and organization filtering
func (s *userService) ListUsers(ctx context.Context, page, limit int, search string) (*dto.PagedUserResponse, error) {
	// Apply default pagination values and validate using utility function
	page, limit, offset := util.ValidateAndSetPaginationParams(page, limit)

	// Get current user making the request
	currentUserID, ok := ctx.Value("current_user_id").(uuid.UUID)
	if !ok {
		return nil, apperror.NewAppError(http.StatusUnauthorized, constant.ErrMsgAuthorizationContextMissing, nil)
	}

	// Get current user with role information
	currentUser, err := s.userRepo.FindByIDWithRole(ctx, currentUserID)
	if err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to find current user: %w", err))
	}

	// Determine current user's level
	var currentUserLevel int
	if currentUser.Role != nil {
		currentUserLevel = currentUser.Role.Level
	} else {
		// If no role, set very low level
		currentUserLevel = 0
	}

	// Super Admin (level 100) can see ALL users except themselves
	// Platform Admin (level >= 76) can see users with level < their level, except themselves
	if currentUserLevel >= 76 {
		// For Platform Admin: set max level they can see
		var maxLevelToSee int
		if currentUserLevel >= 100 {
			// Super Admin can see all levels except themselves
			maxLevelToSee = 999 // Use high number to include all levels
		} else {
			// Platform Admin can only see users with level < their level
			maxLevelToSee = currentUserLevel - 1
		}

		// Use filtered method that excludes current user and respects level hierarchy
		users, err := s.userRepo.ListWithFilters(ctx, offset, limit, search, maxLevelToSee, nil)
		if err != nil {
			return nil, apperror.NewInternalError(fmt.Errorf("failed to list users: %w", err))
		}

		// Filter out current user from results
		filteredUsers := make([]model.User, 0, len(users))
		for _, user := range users {
			if user.ID != currentUserID {
				filteredUsers = append(filteredUsers, user)
			}
		}

		// Get total count with same filtering (excluding current user and level restriction)
		total, err := s.userRepo.CountWithFilters(ctx, search, maxLevelToSee, nil)
		if err != nil {
			return nil, apperror.NewInternalError(fmt.Errorf("failed to count users: %w", err))
		}

		// Adjust total count to exclude current user
		if total > 0 {
			total = total - 1
		}

		// Convert users to response DTOs using utility function
		userResponses := make([]dto.UserResponse, len(filteredUsers))
		for i, user := range filteredUsers {
			userResponses[i] = *util.MapUserToResponse(&user)
		}

		// Calculate total pages
		totalPages := int(math.Ceil(float64(total) / float64(limit)))

		return &dto.PagedUserResponse{
			Users:      userResponses,
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		}, nil
	}

	// For non-platform users, apply hierarchical filtering
	// Get organizations accessible to current user using hierarchical filtering
	accessibleOrgIDs, err := s.orgService.GetAccessibleOrganizationIDs(ctx, currentUserID, currentUserLevel)
	if err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to get accessible organizations: %w", err))
	}

	// Fetch users with level and organization filtering
	users, err := s.userRepo.ListWithFilters(ctx, offset, limit, search, currentUserLevel, accessibleOrgIDs)
	if err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to list users: %w", err))
	}

	// Filter out current user from results for non-platform users too
	filteredUsers := make([]model.User, 0, len(users))
	for _, user := range users {
		if user.ID != currentUserID {
			filteredUsers = append(filteredUsers, user)
		}
	}

	// Get total count with the same filters
	total, err := s.userRepo.CountWithFilters(ctx, search, currentUserLevel, accessibleOrgIDs)
	if err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to count users: %w", err))
	}

	// Adjust total count to exclude current user
	if total > 0 {
		total = total - 1
	}

	// Convert users to response DTOs using utility function
	userResponses := make([]dto.UserResponse, len(filteredUsers))
	for i, user := range filteredUsers {
		userResponses[i] = *util.MapUserToResponse(&user)
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return &dto.PagedUserResponse{
		Users:      userResponses,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

// GetUserByID retrieves a single user by their ID with enhanced error handling
func (s *userService) GetUserByID(ctx context.Context, id uuid.UUID) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByIDWithRole(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NewNotFoundError("user")
		}
		return nil, apperror.NewInternalError(fmt.Errorf("failed to find user: %w", err))
	}

	return util.MapUserToResponse(user), nil
}

// UpdateUser updates a user's data with enhanced authorization and validation
func (s *userService) UpdateUser(ctx context.Context, id uuid.UUID, req dto.UpdateUserRequest) (*dto.UserResponse, error) {
	// Fetch the user to be updated
	user, err := s.userRepo.FindByIDWithRole(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NewNotFoundError("user")
		}
		return nil, apperror.NewInternalError(fmt.Errorf("failed to find user for update: %w", err))
	}

	// Get current user making the request
	currentUserID, ok := ctx.Value("current_user_id").(uuid.UUID)
	if !ok {
		return nil, apperror.NewAppError(http.StatusUnauthorized, constant.ErrMsgAuthorizationContextMissing, nil)
	}

	// Validate role change authorization if role is being updated
	if user.RoleID != nil && req.RoleID == nil {
		// Role is being explicitly unassigned (set to null)
		// Validate that current user can remove the existing role
		if err := s.validateRoleChangeAuthorization(ctx, currentUserID, id, *user.RoleID, user); err != nil {
			return nil, err
		}

		// When removing role, also remove all organization assignments
		// since user cannot have organizations without a role
		if err := s.removeAllUserOrganizations(ctx, id); err != nil {
			return nil, apperror.NewInternalError(fmt.Errorf("failed to remove user organizations: %w", err))
		}

		// Invalidate all user sessions since permissions have changed
		if err := s.invalidateUserSessions(ctx, id); err != nil {
			// Log error but don't fail the operation
			fmt.Printf("Warning: failed to invalidate user sessions: %v\n", err)
		}

		user.RoleID = nil
	} else if req.RoleID != nil {
		// Role is being assigned or changed
		if user.RoleID == nil || *user.RoleID != *req.RoleID {
			if err := s.validateRoleChangeAuthorization(ctx, currentUserID, id, *req.RoleID, user); err != nil {
				return nil, err
			}

			// If user is getting a role for the first time or role changed,
			// invalidate sessions to refresh permissions
			if user.RoleID == nil || *user.RoleID != *req.RoleID {
				if err := s.invalidateUserSessions(ctx, id); err != nil {
					// Log error but don't fail the operation
					fmt.Printf("Warning: failed to invalidate user sessions: %v\n", err)
				}
			}
		}
		user.RoleID = req.RoleID
	}
	// If req.RoleID is nil and user.RoleID is nil, no change needed

	// Update basic user information
	s.updateBasicUserInfo(user, req)

	// Clear the Role association when RoleID is nil to prevent GORM conflicts
	if user.RoleID == nil {
		user.Role = nil
	} // Save the updated user
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to update user: %w", err))
	}

	// Fetch the updated user with fresh role data
	updatedUser, err := s.userRepo.FindByIDWithRole(ctx, id)
	if err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to fetch updated user: %w", err))
	}

	return util.MapUserToResponse(updatedUser), nil
}

// DeleteUser deletes a user.
func (s *userService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	// First, check if the user exists
	_, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NewNotFoundError("user")
		}
		return apperror.NewInternalError(fmt.Errorf("failed to find user for deletion: %w", err))
	}

	// Call the Delete method from the repository
	if err := s.userRepo.Delete(ctx, id); err != nil {
		return apperror.NewInternalError(fmt.Errorf("failed to delete user: %w", err))
	}

	return nil
}

// User-Organization Management Implementation

// AssignUserToOrganization assigns a user to an organization with a specific role.
func (s *userService) AssignUserToOrganization(ctx context.Context, req dto.AssignUserToOrganizationRequest) (*dto.UserOrganizationResponse, error) {
	// Validate user exists
	_, err := s.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NewNotFoundError("user")
		}
		return nil, apperror.NewInternalError(fmt.Errorf("failed to find user: %w", err))
	}

	// Check if user is already assigned to this organization
	existingAssignment, err := s.userRepo.FindUserOrganization(ctx, req.UserID, req.OrganizationID)
	if err == nil && existingAssignment != nil {
		return nil, apperror.NewConflictError("User is already assigned to this organization")
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to check existing assignment: %w", err))
	}

	// Create user-organization assignment
	userOrg := &model.UserOrganization{
		UserID:         req.UserID,
		OrganizationID: req.OrganizationID,
		RoleID:         req.RoleID,
		IsActive:       req.IsActive,
		JoinedAt:       time.Now(),
	}

	_, err = s.userRepo.CreateUserOrganization(ctx, userOrg)
	if err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to assign user to organization: %w", err))
	}

	// Get full user-organization data with relationships
	fullUserOrg, err := s.userRepo.FindUserOrganization(ctx, req.UserID, req.OrganizationID)
	if err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to fetch created assignment: %w", err))
	}

	return s.mapUserOrganizationToResponse(fullUserOrg), nil
}

// RemoveUserFromOrganization removes a user from an organization.
func (s *userService) RemoveUserFromOrganization(ctx context.Context, userID, organizationID uuid.UUID) error {
	// Check if assignment exists
	_, err := s.userRepo.FindUserOrganization(ctx, userID, organizationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NewNotFoundError("user organization assignment")
		}
		return apperror.NewInternalError(fmt.Errorf("failed to find assignment: %w", err))
	}

	// Remove assignment
	if err := s.userRepo.DeleteUserOrganization(ctx, userID, organizationID); err != nil {
		return apperror.NewInternalError(fmt.Errorf("failed to remove user from organization: %w", err))
	}

	return nil
}

// UpdateUserOrganizationRole updates a user's role in an organization.
func (s *userService) UpdateUserOrganizationRole(ctx context.Context, userID, organizationID uuid.UUID, req dto.UpdateUserOrganizationRequest) (*dto.UserOrganizationResponse, error) {
	// Find existing assignment
	userOrg, err := s.userRepo.FindUserOrganization(ctx, userID, organizationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NewNotFoundError("user organization assignment")
		}
		return nil, apperror.NewInternalError(fmt.Errorf("failed to find assignment: %w", err))
	}

	// Update assignment
	userOrg.RoleID = req.RoleID
	userOrg.IsActive = req.IsActive

	updatedUserOrg, err := s.userRepo.UpdateUserOrganization(ctx, userOrg)
	if err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to update user organization role: %w", err))
	}

	return s.mapUserOrganizationToResponse(updatedUserOrg), nil
}

// GetUserOrganizations retrieves all organizations that a user belongs to.
func (s *userService) GetUserOrganizations(ctx context.Context, userID uuid.UUID, page, limit int) (*dto.PagedUserOrganizationResponse, error) {
	// Apply default pagination values
	page, limit, offset := util.ValidateAndSetPaginationParams(page, limit)

	// Get user organizations
	userOrgs, err := s.userRepo.FindUserOrganizations(ctx, userID, offset, limit)
	if err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to get user organizations: %w", err))
	}

	// Get total count
	total, err := s.userRepo.CountUserOrganizations(ctx, userID)
	if err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to count user organizations: %w", err))
	}

	// Convert to response DTOs
	userOrgResponses := make([]dto.UserOrganizationResponse, len(userOrgs))
	for i, userOrg := range userOrgs {
		userOrgResponses[i] = *s.mapUserOrganizationToResponse(&userOrg)
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return &dto.PagedUserOrganizationResponse{
		UserOrganizations: userOrgResponses,
		Page:              page,
		Limit:             limit,
		Total:             total,
		TotalPages:        totalPages,
	}, nil
}

// GetOrganizationMembers retrieves all members of an organization.
func (s *userService) GetOrganizationMembers(ctx context.Context, organizationID uuid.UUID, page, limit int) (*dto.PagedUserOrganizationResponse, error) {
	// Apply default pagination values
	page, limit, offset := util.ValidateAndSetPaginationParams(page, limit)

	// Get organization members
	userOrgs, err := s.userRepo.FindOrganizationMembers(ctx, organizationID, offset, limit)
	if err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to get organization members: %w", err))
	}

	// Get total count
	total, err := s.userRepo.CountOrganizationMembers(ctx, organizationID)
	if err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to count organization members: %w", err))
	}

	// Convert to response DTOs
	userOrgResponses := make([]dto.UserOrganizationResponse, len(userOrgs))
	for i, userOrg := range userOrgs {
		userOrgResponses[i] = *s.mapUserOrganizationToResponse(&userOrg)
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return &dto.PagedUserOrganizationResponse{
		UserOrganizations: userOrgResponses,
		Page:              page,
		Limit:             limit,
		Total:             total,
		TotalPages:        totalPages,
	}, nil
}

// BulkAssignUsersToOrganization assigns multiple users to an organization.
func (s *userService) BulkAssignUsersToOrganization(ctx context.Context, req dto.BulkAssignUsersToOrganizationRequest) (*dto.BulkAssignResponse, error) {
	var userOrgs []model.UserOrganization
	var errors []dto.BulkAssignError

	// Prepare user-organization assignments
	for _, userID := range req.UserIDs {
		// Check if user exists
		_, err := s.userRepo.FindByID(ctx, userID)
		if err != nil {
			errors = append(errors, dto.BulkAssignError{
				UserID: userID,
				Error:  "User not found",
			})
			continue
		}

		// Check if already assigned
		_, err = s.userRepo.FindUserOrganization(ctx, userID, req.OrganizationID)
		if err == nil {
			errors = append(errors, dto.BulkAssignError{
				UserID: userID,
				Error:  "User already assigned to organization",
			})
			continue
		}

		userOrgs = append(userOrgs, model.UserOrganization{
			UserID:         userID,
			OrganizationID: req.OrganizationID,
			RoleID:         req.RoleID,
			IsActive:       true,
			JoinedAt:       time.Now(),
		})
	}

	// Bulk create assignments
	created, createErrors := s.userRepo.BulkCreateUserOrganizations(ctx, userOrgs)

	// Convert create errors to BulkAssignError
	for i, err := range createErrors {
		if i < len(userOrgs) {
			errors = append(errors, dto.BulkAssignError{
				UserID: userOrgs[i].UserID,
				Error:  err.Error(),
			})
		}
	}

	// Convert created assignments to responses
	assignments := make([]dto.UserOrganizationResponse, len(created))
	for i, userOrg := range created {
		assignments[i] = *s.mapUserOrganizationToResponse(&userOrg)
	}

	successCount := len(created)
	failureCount := len(errors)

	return &dto.BulkAssignResponse{
		SuccessCount: successCount,
		FailureCount: failureCount,
		Assignments:  assignments,
		Errors:       errors,
		Message:      fmt.Sprintf("%d users assigned successfully, %d failed", successCount, failureCount),
	}, nil
}

// GetUserOrganizationHistory retrieves the organization assignment history for a user.
func (s *userService) GetUserOrganizationHistory(ctx context.Context, userID uuid.UUID, page, limit int) (*dto.PagedUserOrganizationHistoryResponse, error) {
	// Apply default pagination values
	page, limit, offset := util.ValidateAndSetPaginationParams(page, limit)

	// Get user organization history
	history, err := s.userRepo.FindUserOrganizationHistory(ctx, userID, offset, limit)
	if err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to get user organization history: %w", err))
	}

	// Get total count
	total, err := s.userRepo.CountUserOrganizationHistory(ctx, userID)
	if err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to count user organization history: %w", err))
	}

	// Convert to response DTOs
	historyResponses := make([]dto.UserOrganizationHistoryResponse, len(history))
	for i, h := range history {
		historyResponses[i] = *s.mapUserOrganizationHistoryToResponse(&h)
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return &dto.PagedUserOrganizationHistoryResponse{
		History:    historyResponses,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

// GetOrganizationUserHistory retrieves the user assignment history for an organization.
func (s *userService) GetOrganizationUserHistory(ctx context.Context, organizationID uuid.UUID, page, limit int) (*dto.PagedUserOrganizationHistoryResponse, error) {
	// Apply default pagination values
	page, limit, offset := util.ValidateAndSetPaginationParams(page, limit)

	// Get organization user history
	history, err := s.userRepo.FindOrganizationUserHistory(ctx, organizationID, offset, limit)
	if err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to get organization user history: %w", err))
	}

	// Get total count
	total, err := s.userRepo.CountOrganizationUserHistory(ctx, organizationID)
	if err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to count organization user history: %w", err))
	}

	// Convert to response DTOs
	historyResponses := make([]dto.UserOrganizationHistoryResponse, len(history))
	for i, h := range history {
		historyResponses[i] = *s.mapUserOrganizationHistoryToResponse(&h)
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return &dto.PagedUserOrganizationHistoryResponse{
		History:    historyResponses,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

// LogUserOrganizationAction manually logs a user organization action.
func (s *userService) LogUserOrganizationAction(ctx context.Context, req dto.LogUserOrganizationActionRequest, actionBy uuid.UUID) (*dto.UserOrganizationHistoryResponse, error) {
	// Validate user exists
	_, err := s.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NewNotFoundError("user")
		}
		return nil, apperror.NewInternalError(fmt.Errorf("failed to find user: %w", err))
	}

	// Create history record
	history := &model.UserOrganizationHistory{
		UserID:         req.UserID,
		OrganizationID: req.OrganizationID,
		Action:         req.Action,
		PreviousRole:   req.PreviousRole,
		NewRole:        req.NewRole,
		PreviousStatus: req.PreviousStatus,
		NewStatus:      req.NewStatus,
		ActionBy:       actionBy,
		ActionAt:       time.Now(),
		Reason:         req.Reason,
	}

	createdHistory, err := s.userRepo.CreateUserOrganizationHistory(ctx, history)
	if err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to log user organization action: %w", err))
	}

	// Get full history record with relationships
	historyRecords, err := s.userRepo.FindUserOrganizationHistory(ctx, req.UserID, 0, 1)
	if err != nil || len(historyRecords) == 0 {
		// Return basic response if we can't fetch full record
		return s.mapUserOrganizationHistoryToResponse(createdHistory), nil
	}

	return s.mapUserOrganizationHistoryToResponse(&historyRecords[0]), nil
}

// Helper method to map UserOrganization model to response DTO
func (s *userService) mapUserOrganizationToResponse(userOrg *model.UserOrganization) *dto.UserOrganizationResponse {
	response := &dto.UserOrganizationResponse{
		UserID:         userOrg.UserID,
		OrganizationID: userOrg.OrganizationID,
		RoleID:         userOrg.RoleID,
		IsActive:       userOrg.IsActive,
		JoinedAt:       userOrg.JoinedAt,
	}

	// Include organization data if loaded
	if userOrg.Organization.ID != uuid.Nil {
		response.Organization = dto.OrganizationResponse{
			ID:                   userOrg.Organization.ID,
			Name:                 userOrg.Organization.Name,
			Code:                 userOrg.Organization.Code,
			OrganizationType:     userOrg.Organization.OrganizationType,
			ParentOrganizationID: userOrg.Organization.ParentOrganizationID,
			Description:          userOrg.Organization.Description,
			CreatedBy:            userOrg.Organization.CreatedBy,
			IsActive:             userOrg.Organization.IsActive,
			CreatedAt:            userOrg.Organization.CreatedAt,
			UpdatedAt:            userOrg.Organization.UpdatedAt,
		}
	}

	// Include role data if loaded
	if userOrg.Role != nil && userOrg.Role.ID != uuid.Nil {
		response.Role = &dto.RoleResponse{
			ID:             userOrg.Role.ID,
			Name:           userOrg.Role.Name,
			Description:    userOrg.Role.Description,
			Level:          userOrg.Role.Level,
			IsSystemRole:   userOrg.Role.IsSystemRole,
			PredefinedName: userOrg.Role.PredefinedName,
			IsActive:       userOrg.Role.IsActive,
		}
	}

	return response
}

// Helper method to map UserOrganizationHistory model to response DTO
func (s *userService) mapUserOrganizationHistoryToResponse(history *model.UserOrganizationHistory) *dto.UserOrganizationHistoryResponse {
	response := &dto.UserOrganizationHistoryResponse{
		ID:             history.ID,
		UserID:         history.UserID,
		OrganizationID: history.OrganizationID,
		Action:         history.Action,
		PreviousRole:   history.PreviousRole,
		NewRole:        history.NewRole,
		ActionBy:       history.ActionBy,
		ActionAt:       history.ActionAt.Format(time.RFC3339),
		Reason:         history.Reason,
	}

	// Handle boolean pointers for status
	if history.PreviousStatus != nil {
		prevStatus := *history.PreviousStatus
		if prevStatus {
			response.PreviousRole = "active"
		} else {
			response.PreviousRole = "inactive"
		}
	}

	if history.NewStatus != nil {
		newStatus := *history.NewStatus
		if newStatus {
			response.NewRole = "active"
		} else {
			response.NewRole = "inactive"
		}
	}

	return response
}

// Helper methods for user service
// validateRoleChangeAuthorization validates if the current user can change roles
func (s *userService) validateRoleChangeAuthorization(ctx context.Context, currentUserID, targetUserID, newRoleID uuid.UUID, targetUser *model.User) error {
	// Prevent self-role changes
	if currentUserID == targetUserID {
		return apperror.NewAppError(http.StatusForbidden, constant.ErrMsgCannotChangeOwnRole, nil)
	}

	// Get current user's role to check authorization
	currentUser, err := s.userRepo.FindByIDWithRole(ctx, currentUserID)
	if err != nil {
		return apperror.NewInternalError(fmt.Errorf("failed to find current user: %w", err))
	}

	if currentUser.Role == nil || currentUser.RoleID == nil {
		return apperror.NewAppError(http.StatusForbidden, constant.ErrMsgCurrentUserHasNoRole, nil)
	}

	// Get the new role information
	newRole, err := s.roleRepo.FindByID(ctx, newRoleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NewNotFoundError("role")
		}
		return apperror.NewInternalError(fmt.Errorf("failed to find new role: %w", err))
	}

	// Superadmin protection: Only other superadmins can modify superadmin roles
	if targetUser.Role != nil && targetUser.Role.Name == "super_admin" && currentUser.Role.Name != "super_admin" {
		return apperror.NewAppError(http.StatusForbidden, constant.ErrMsgOnlySuperAdminCanModify, nil)
	}

	// Level-based authorization: Users cannot assign roles with equal or higher level than their own
	if newRole.Level >= currentUser.Role.Level {
		return apperror.NewAppError(http.StatusForbidden, constant.ErrMsgInsufficientAuthorityLevel, nil)
	}

	// Prevent assigning superadmin role unless current user is superadmin
	if newRole.Name == "super_admin" && currentUser.Role.Name != "super_admin" {
		return apperror.NewAppError(http.StatusForbidden, constant.ErrMsgOnlySuperAdminCanAssign, nil)
	}

	// Level 100 (superadmin) uniqueness validation - only one superadmin allowed
	if newRole.Level >= 100 {
		existingSuperadminCount, err := s.userRepo.CountUsersByRoleLevel(context.Background(), 100)
		if err != nil {
			return apperror.NewInternalError(fmt.Errorf("failed to check existing superadmin: %w", err))
		}

		// Allow if this is updating the existing superadmin, but not creating a new one
		if existingSuperadminCount > 0 && (targetUser.Role == nil || targetUser.Role.Level < 100) {
			return apperror.NewValidationError("Only one superadmin (level 100) is allowed in the system")
		}
	}

	return nil
}

// updateBasicUserInfo updates basic user information fields
func (s *userService) updateBasicUserInfo(user *model.User, req dto.UpdateUserRequest) {
	if req.Username != "" {
		user.Username = req.Username
	}
	if req.Email != "" {
		user.Email = req.Email
	}
}

// validateRoleExists validates if the specified role exists
func (s *userService) validateRoleExists(ctx context.Context, roleID uuid.UUID) error {
	_, err := s.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NewNotFoundError("role")
		}
		return apperror.NewInternalError(fmt.Errorf("failed to validate role: %w", err))
	}
	return nil
}

// hashPassword hashes a password using bcrypt
func (s *userService) hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// removeAllUserOrganizations removes all organization assignments for a user
func (s *userService) removeAllUserOrganizations(ctx context.Context, userID uuid.UUID) error {
	// Get all user organization assignments
	userOrgs, err := s.userRepo.FindUserOrganizations(ctx, userID, 0, 1000) // Use high limit to get all
	if err != nil {
		return fmt.Errorf("failed to fetch user organizations: %w", err)
	}

	// Remove each organization assignment using the organization service
	for _, userOrg := range userOrgs {
		if err := s.RemoveUserFromOrganization(ctx, userID, userOrg.OrganizationID); err != nil {
			return fmt.Errorf("failed to remove user from organization %s: %w", userOrg.OrganizationID.String(), err)
		}
	}

	return nil
}

// invalidateUserSessions invalidates all active sessions for a user
func (s *userService) invalidateUserSessions(ctx context.Context, userID uuid.UUID) error {
	// Pattern to find all refresh tokens for this user
	// Refresh tokens are stored with user ID as value
	pattern := "*"

	// Scan all keys to find refresh tokens for this user
	iter := s.redis.Scan(ctx, 0, pattern, 0).Iterator()
	var keysToDelete []string

	for iter.Next(ctx) {
		key := iter.Val()
		// Get the value to check if it's this user's token
		val, err := s.redis.Get(ctx, key).Result()
		if err != nil {
			continue // Skip if error reading value
		}

		// If the value matches our user ID, this is a refresh token for this user
		if val == userID.String() {
			keysToDelete = append(keysToDelete, key)
		}
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan redis keys: %w", err)
	}

	// Delete all found refresh tokens
	if len(keysToDelete) > 0 {
		if err := s.redis.Del(ctx, keysToDelete...).Err(); err != nil {
			return fmt.Errorf("failed to delete refresh tokens: %w", err)
		}
	}

	return nil
}
