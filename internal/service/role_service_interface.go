package service

import (
	"beresin-backend/internal/dto"

	"github.com/google/uuid"
)

// RoleService mendefinisikan kontrak untuk layanan manajemen role.
type RoleService interface {
	CreateRole(req dto.CreateRoleRequest) (*dto.RoleResponse, error)
	UpdateRolePermissions(roleID uuid.UUID, permissionNames []string) error
}
