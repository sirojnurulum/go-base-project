package repository

import (
	"beresin-backend/internal/model"

	"github.com/google/uuid"
)

type RoleRepository interface {
	Create(role *model.Role) (*model.Role, error)
	FindRoleByName(name string) (*model.Role, error)
	FindPermissionsByRoleID(roleID uuid.UUID) ([]string, error)
	UpdateRolePermissions(roleID uuid.UUID, permissionNames []string) error
}
