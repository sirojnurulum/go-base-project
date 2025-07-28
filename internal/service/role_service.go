package service

import (
	"beresin-backend/internal/apperror"
	"beresin-backend/internal/dto"
	"beresin-backend/internal/model"
	"beresin-backend/internal/repository"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// roleService implements the RoleService interface for role management.
type roleService struct {
	roleRepo             repository.RoleRepository
	authorizationService AuthorizationService // To call cache invalidation
}

// NewRoleService creates a new instance of roleService.
func NewRoleService(roleRepo repository.RoleRepository, authorizationService AuthorizationService) RoleService {
	return &roleService{
		roleRepo:             roleRepo,
		authorizationService: authorizationService,
	}
}
func (s *roleService) CreateRole(req dto.CreateRoleRequest) (*dto.RoleResponse, error) {
	// Check if a role with the same name already exists
	_, err := s.roleRepo.FindRoleByName(req.Name)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		// If the error is not 'not found', it means the role already exists or another error occurred
		return nil, apperror.NewConflictError(fmt.Sprintf("role with name '%s' already exists", req.Name))
	}

	newRole := &model.Role{
		Name:        req.Name,
		Description: req.Description,
	}

	createdRole, err := s.roleRepo.Create(newRole)
	if err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to create role: %w", err))
	}

	return &dto.RoleResponse{
		ID:          createdRole.ID,
		Name:        createdRole.Name,
		Description: createdRole.Description,
	}, nil
}

// UpdateRolePermissions updates the permissions associated with a role.
func (s *roleService) UpdateRolePermissions(roleID uuid.UUID, permissionNames []string) error {
	// Call a single repository method that handles the transaction.
	if err := s.roleRepo.UpdateRolePermissions(roleID, permissionNames); err != nil {
		// Translate repository error to AppError
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NewNotFoundError("role")
		}
		// For other database errors, return as an internal error
		return apperror.NewInternalError(fmt.Errorf("failed to update role permissions in db: %w", err))
	}

	if err := s.authorizationService.InvalidateRolePermissionsCache(roleID); err != nil {
		log.Error().Err(err).Msgf("CRITICAL: DB updated but failed to invalidate cache for role %s", roleID)
	}

	return nil
}
