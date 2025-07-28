package service

import (
	"beresin-backend/internal/apperror"
	"beresin-backend/internal/dto"
	"beresin-backend/internal/model"
	"beresin-backend/internal/repository"
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
}

// NewUserService creates a new instance of userService.
func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
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

// UpdateUser updates a user's data.
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
