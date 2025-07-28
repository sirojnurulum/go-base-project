package service

import (
	"beresin-backend/internal/cache"
	"beresin-backend/internal/repository"
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type authorizationService struct {
	roleRepo repository.RoleRepository
	redis    *redis.Client
}

// getPermissionsForRole fetches the list of permission names for a given role from the database.
func (s *authorizationService) getPermissionsForRole(roleID uuid.UUID) ([]string, error) {
	return s.roleRepo.FindPermissionsByRoleID(roleID)
}

// cachePermissionsForRole stores a role's permissions in Redis.
func (s *authorizationService) cachePermissionsForRole(roleID uuid.UUID, permissions []string) {
	cacheKey := cache.GetRolePermissionsCacheKey(roleID)

	permissionsJSON, err := json.Marshal(permissions)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to marshal permissions for role %s for caching", roleID.String())
		return
	}

	if err := s.redis.Set(ctx, cacheKey, permissionsJSON, cache.PermissionsCacheDuration).Err(); err != nil {
		log.Error().Err(err).Msgf("Failed to cache permissions for role %s", roleID.String())
	} else {
		log.Debug().Msgf("Permissions for role %s have been cached.", roleID.String())
	}
}

// GetAndCachePermissionsForRole retrieves permissions for a role, using cache first, and populates cache on miss.
func (s *authorizationService) GetAndCachePermissionsForRole(roleID uuid.UUID) ([]string, error) {
	cacheKey := cache.GetRolePermissionsCacheKey(roleID)
	cachedPermissions, err := s.redis.Get(ctx, cacheKey).Result()

	var permissions []string
	if err == nil {
		// Cache HIT
		_ = json.Unmarshal([]byte(cachedPermissions), &permissions)
		return permissions, nil
	}

	// Cache MISS, get from DB
	permissions, err = s.getPermissionsForRole(roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions from db: %w", err)
	}
	// Save to cache for subsequent requests
	s.cachePermissionsForRole(roleID, permissions)
	return permissions, nil
}

// CheckPermission checks if a role has a required permission.
func (s *authorizationService) CheckPermission(roleID uuid.UUID, requiredPermission string) (bool, error) {
	permissions, err := s.GetAndCachePermissionsForRole(roleID)
	if err != nil {
		return false, err
	}

	for _, p := range permissions {
		if p == requiredPermission {
			return true, nil
		}
	}
	return false, nil
}

// InvalidateRolePermissionsCache removes the permissions cache for a role.
func (s *authorizationService) InvalidateRolePermissionsCache(roleID uuid.UUID) error {
	cacheKey := cache.GetRolePermissionsCacheKey(roleID)
	log.Info().Str("cacheKey", cacheKey).Msg("Invalidating permissions cache for role")
	return s.redis.Del(ctx, cacheKey).Err()
}
