package service

import (
	"beresin-backend/internal/apperror"
	"beresin-backend/internal/dto"
	"beresin-backend/internal/model"
	"beresin-backend/internal/repository"
	"beresin-backend/internal/util"
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// userService implements the UserService interface for user management.
type userService struct {
	userRepo repository.UserRepository
	roleRepo repository.RoleRepository
}

// NewUserService creates a new instance of userService.
func NewUserService(userRepo repository.UserRepository, roleRepo repository.RoleRepository) UserService {
	return &userService{
		userRepo: userRepo,
		roleRepo: roleRepo,
	}
}

// CreateUser creates a new user based on the provided DTO.
func (s *userService) CreateUser(req dto.CreateUserRequest) (*dto.UserResponse, error) {
	// Cek apakah username atau email sudah ada
	_, err := s.userRepo.FindByUsername(req.Username)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.NewConflictError("username already exists")
	}
	_, err = s.userRepo.FindByEmail(req.Email)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.NewConflictError("email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to hash password: %w", err))
	}

	newUser := &model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		RoleID:   &req.RoleID,
	}

	if err := s.userRepo.Create(newUser); err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to create user: %w", err))
	}

	// Fetch the newly created user with their role for the response
	createdUser, err := s.userRepo.FindByIDWithRole(newUser.ID)
	if err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to fetch created user: %w", err))
	}

	return &dto.UserResponse{
		ID:        createdUser.ID,
		Username:  createdUser.Username,
		Email:     createdUser.Email,
		Role:      createdUser.Role.Name,
		AvatarURL: createdUser.AvatarURL,
	}, nil
}

// ListUsers retrieves a paginated list of users.
func (s *userService) ListUsers(page, limit int, search string) (*dto.PagedUserResponse, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	offset := (page - 1) * limit

	users, err := s.userRepo.List(offset, limit, search)
	if err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to list users: %w", err))
	}

	total, err := s.userRepo.Count(search)
	if err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to count users: %w", err))
	}

	userResponses := make([]dto.UserResponse, len(users))
	for i, u := range users {
		userResponses[i] = dto.UserResponse{
			ID:        u.ID,
			Username:  u.Username,
			Email:     u.Email,
			Role:      u.Role.Name,
			AvatarURL: u.AvatarURL,
		}
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return &dto.PagedUserResponse{
		Users:      userResponses,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

// GetUserByID retrieves a single user by their ID.
func (s *userService) GetUserByID(id uuid.UUID) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByIDWithRole(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NewNotFoundError("user")
		}
		return nil, apperror.NewInternalError(fmt.Errorf("failed to find user: %w", err))
	}

	return &dto.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role.Name,
		AvatarURL: user.AvatarURL,
	}, nil
}

// UpdateUser updates a user's data with security validations.
func (s *userService) UpdateUser(id uuid.UUID, req dto.UpdateUserRequest) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByIDWithRole(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NewNotFoundError("user")
		}
		return nil, apperror.NewInternalError(fmt.Errorf("failed to find user for update: %w", err))
	}

	if req.Username != "" {
		user.Username = req.Username
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.RoleID != uuid.Nil {
		user.RoleID = &req.RoleID
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to update user: %w", err))
	}

	// Fetch the updated data again to ensure role data is consistent
	updatedUser, err := s.userRepo.FindByIDWithRole(id)
	if err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to fetch updated user: %w", err))
	}

	return &dto.UserResponse{
		ID:        updatedUser.ID,
		Username:  updatedUser.Username,
		Email:     updatedUser.Email,
		Role:      updatedUser.Role.Name,
		AvatarURL: updatedUser.AvatarURL,
	}, nil
}

// DeleteUser deletes a user.
func (s *userService) DeleteUser(id uuid.UUID) error {
	// First, check if the user exists
	_, err := s.userRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NewNotFoundError("user")
		}
		return apperror.NewInternalError(fmt.Errorf("failed to find user for deletion: %w", err))
	}

	// Call the Delete method from the repository
	if err := s.userRepo.Delete(id); err != nil {
		return apperror.NewInternalError(fmt.Errorf("failed to delete user: %w", err))
	}

	return nil
}

// UpdateUserWithContext updates a user's data with security validations based on current user context.
func (s *userService) UpdateUserWithContext(ctx context.Context, currentUserID, targetUserID uuid.UUID, req dto.UpdateUserRequest) (*dto.UserResponse, error) {
	// Prevent self-role modification
	if currentUserID == targetUserID && req.RoleID != uuid.Nil {
		return nil, apperror.NewForbiddenError(util.GetUserFriendlyError("cannot_change_own_role"))
	}

	// Get current user with role
	currentUser, err := s.userRepo.FindByIDWithRole(currentUserID)
	if err != nil {
		return nil, apperror.NewUnauthorizedError(util.GetUserFriendlyError("authorization_context_not_found"))
	}

	if currentUser.Role == nil {
		return nil, apperror.NewUnauthorizedError(util.GetUserFriendlyError("current_user_no_role"))
	}

	// Get target user with role
	targetUser, err := s.userRepo.FindByIDWithRole(targetUserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NewNotFoundError("user")
		}
		return nil, apperror.NewInternalError(fmt.Errorf("failed to find target user: %w", err))
	}

	// If role change is requested, validate it
	if req.RoleID != uuid.Nil {
		// Get target role
		targetRole, err := s.roleRepo.FindByID(req.RoleID)
		if err != nil {
			return nil, apperror.NewNotFoundError("role")
		}

		// Validate role assignment authorization
		if err := s.validateRoleAssignment(currentUser, targetUser, targetRole); err != nil {
			return nil, err
		}

		targetUser.RoleID = &req.RoleID
	}

	// Update other fields
	if req.Username != "" {
		targetUser.Username = req.Username
	}
	if req.Email != "" {
		targetUser.Email = req.Email
	}

	// Save changes
	if err := s.userRepo.Update(targetUser); err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to update user: %w", err))
	}

	// Fetch updated user with role
	updatedUser, err := s.userRepo.FindByIDWithRole(targetUserID)
	if err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to fetch updated user: %w", err))
	}

	roleName := "No Role"
	if updatedUser.Role != nil {
		roleName = updatedUser.Role.Name
	}

	return &dto.UserResponse{
		ID:        updatedUser.ID,
		Username:  updatedUser.Username,
		Email:     updatedUser.Email,
		Role:      roleName,
		AvatarURL: updatedUser.AvatarURL,
	}, nil
}

// validateRoleAssignment validates if current user can assign target role to target user
func (s *userService) validateRoleAssignment(currentUser, targetUser *model.User, targetRole *model.Role) error {
	// Super admin protection - only super admins can modify super admin accounts
	if targetUser.Role != nil && targetUser.Role.Name == "super_admin" && currentUser.Role.Name != "super_admin" {
		return apperror.NewForbiddenError(util.GetUserFriendlyError("only_superadmin_can_modify"))
	}

	// Super admin assignment - only super admins can assign super admin role
	if targetRole.Name == "super_admin" && currentUser.Role.Name != "super_admin" {
		return apperror.NewForbiddenError(util.GetUserFriendlyError("only_superadmin_can_assign"))
	}

	// Role hierarchy validation - users can only assign roles with lower or equal level
	if currentUser.Role.Level <= targetRole.Level && currentUser.Role.Name != "super_admin" {
		return apperror.NewForbiddenError(util.GetUserFriendlyError("insufficient_authority_level"))
	}

	return nil
}
