package service

import (
	"beresin-backend/internal/repository"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

// AuthorizationService defines the contract for authorization and permission-related logic.
type AuthorizationService interface {
	CheckPermission(roleID uuid.UUID, requiredPermission string) (bool, error)
	InvalidateRolePermissionsCache(roleID uuid.UUID) error
	GetAndCachePermissionsForRole(roleID uuid.UUID) ([]string, error)
}

func NewAuthorizationService(roleRepo repository.RoleRepository, redis *redis.Client) AuthorizationService {
	return &authorizationService{
		roleRepo: roleRepo,
		redis:    redis,
	}
}
