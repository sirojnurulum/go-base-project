package service

import (
	"beresin-backend/internal/cache"
	"beresin-backend/internal/repository"
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type authorizationService struct {
	roleRepo repository.RoleRepositoryInterface
	userRepo repository.UserRepositoryInterface
	redis    *redis.Client
}

// NewAuthorizationService creates a new authorization service instance
func NewAuthorizationService(roleRepo repository.RoleRepositoryInterface, userRepo repository.UserRepositoryInterface, redis *redis.Client) AuthorizationServiceInterface {
	return &authorizationService{
		roleRepo: roleRepo,
		userRepo: userRepo,
		redis:    redis,
	}
}

// getPermissionsForRole fetches the list of permission names for a given role from the database.
func (s *authorizationService) getPermissionsForRole(ctx context.Context, roleID uuid.UUID) ([]string, error) {
	return s.roleRepo.FindPermissionsByRoleID(ctx, roleID)
}

// cachePermissionsForRole stores a role's permissions in Redis.
func (s *authorizationService) cachePermissionsForRole(ctx context.Context, roleID uuid.UUID, permissions []string) {
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
// Super admin roles automatically get all permissions without database lookup.
func (s *authorizationService) GetAndCachePermissionsForRole(ctx context.Context, roleID uuid.UUID) ([]string, error) {
	// First check if this role is a super admin
	isSuperAdmin, err := s.roleRepo.IsRoleSuperAdmin(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to check if role is super admin: %w", err)
	}

	// Super admin gets ALL permissions automatically
	if isSuperAdmin {
		allPermissions := []string{
			"users:create", "users:read", "users:update", "users:delete",
			"roles:assign", "roles:create", "roles:approve", "roles:update", "roles:read",
			"permissions:create", "permissions:read", "permissions:update", "permissions:delete",
			"dashboard:view", "scanned_data:create",
			"shipping:scan", "shipping:manage", "shipping:read", "shipping:export",
			"organizations:create", "organizations:read", "organizations:update", "organizations:delete", "organizations:manage_members",
		}
		return allPermissions, nil
	}

	// For regular roles, use cache and database lookup
	cacheKey := cache.GetRolePermissionsCacheKey(roleID)
	cachedPermissions, err := s.redis.Get(ctx, cacheKey).Result()

	var permissions []string
	if err == nil {
		// Cache HIT
		_ = json.Unmarshal([]byte(cachedPermissions), &permissions)
		return permissions, nil
	}

	// Cache MISS, get from DB
	permissions, err = s.getPermissionsForRole(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions from db: %w", err)
	}
	// Save to cache for subsequent requests
	s.cachePermissionsForRole(ctx, roleID, permissions)
	return permissions, nil
}

// CheckPermission checks if a role has a required permission.
// Super admin roles automatically have all permissions without database lookup.
func (s *authorizationService) CheckPermission(ctx context.Context, roleID uuid.UUID, requiredPermission string) (bool, error) {
	// First check if this role is a super admin
	isSuperAdmin, err := s.roleRepo.IsRoleSuperAdmin(ctx, roleID)
	if err != nil {
		return false, fmt.Errorf("failed to check if role is super admin: %w", err)
	}

	// Super admin has access to everything
	if isSuperAdmin {
		return true, nil
	}

	// For non-super admin roles, check permissions normally
	permissions, err := s.GetAndCachePermissionsForRole(ctx, roleID)
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
func (s *authorizationService) InvalidateRolePermissionsCache(ctx context.Context, roleID uuid.UUID) error {
	cacheKey := cache.GetRolePermissionsCacheKey(roleID)
	log.Info().Str("cacheKey", cacheKey).Msg("Invalidating permissions cache for role")
	return s.redis.Del(ctx, cacheKey).Err()
}

// IsRoleSuperAdmin checks if a role is a super admin role.
func (s *authorizationService) IsRoleSuperAdmin(ctx context.Context, roleID uuid.UUID) (bool, error) {
	return s.roleRepo.IsRoleSuperAdmin(ctx, roleID)
}

// CheckUserOrganizationAccess checks if a user has access to a specific organization.
// This method validates that there's an active user-organization relationship.
func (s *authorizationService) CheckUserOrganizationAccess(ctx context.Context, userID, organizationID uuid.UUID) (bool, error) {
	// Find the user-organization relationship
	userOrg, err := s.userRepo.FindUserOrganization(ctx, userID, organizationID)
	if err != nil {
		// If no relationship found, user doesn't have access
		return false, nil
	}

	// Check if the relationship is active
	if userOrg != nil && userOrg.IsActive {
		return true, nil
	}

	return false, nil
}

// CheckPermissionInOrganization checks if a user has a specific permission within an organization context.
// This ensures complete role and permission isolation between organizations.
func (s *authorizationService) CheckPermissionInOrganization(ctx context.Context, userID, organizationID uuid.UUID, requiredPermission string) (bool, error) {
	// First, check if user has access to the organization
	hasAccess, err := s.CheckUserOrganizationAccess(ctx, userID, organizationID)
	if err != nil {
		return false, fmt.Errorf("failed to check organization access: %w", err)
	}
	if !hasAccess {
		return false, nil
	}

	// Get user's role in this specific organization
	roleID, err := s.GetUserRoleInOrganization(ctx, userID, organizationID)
	if err != nil {
		return false, fmt.Errorf("failed to get user role in organization: %w", err)
	}
	if roleID == nil {
		// User has access to organization but no role assigned
		return false, nil
	}

	// Check if the role has the required permission
	return s.CheckPermission(ctx, *roleID, requiredPermission)
}

// GetUserRoleInOrganization retrieves the user's role ID within a specific organization.
func (s *authorizationService) GetUserRoleInOrganization(ctx context.Context, userID, organizationID uuid.UUID) (*uuid.UUID, error) {
	userOrg, err := s.userRepo.FindUserOrganization(ctx, userID, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user organization relationship: %w", err)
	}

	if userOrg == nil || !userOrg.IsActive {
		return nil, nil
	}

	return userOrg.RoleID, nil
}

// GetUserPermissionsInOrganization retrieves all permissions a user has within a specific organization.
func (s *authorizationService) GetUserPermissionsInOrganization(ctx context.Context, userID, organizationID uuid.UUID) ([]string, error) {
	// Get user's role in this organization
	roleID, err := s.GetUserRoleInOrganization(ctx, userID, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user role in organization: %w", err)
	}
	if roleID == nil {
		return []string{}, nil
	}

	// Get permissions for this role
	return s.GetAndCachePermissionsForRole(ctx, *roleID)
}

// ValidateRoleAccessibleInOrganization checks if a role can be used within a specific organization type.
// This ensures roles are only used in appropriate organization contexts.
func (s *authorizationService) ValidateRoleAccessibleInOrganization(ctx context.Context, roleID, organizationID uuid.UUID, organizationType string) (bool, error) {
	// First check if this role is a super admin (super admin can be used anywhere)
	isSuperAdmin, err := s.roleRepo.IsRoleSuperAdmin(ctx, roleID)
	if err != nil {
		return false, fmt.Errorf("failed to check if role is super admin: %w", err)
	}
	if isSuperAdmin {
		return true, nil
	}

	// Get organization types that this role is applicable to
	organizationTypes, err := s.roleRepo.FindOrganizationTypesByRoleID(ctx, roleID)
	if err != nil {
		return false, fmt.Errorf("failed to get organization types for role: %w", err)
	}

	// Check if the role is applicable to this organization type
	for _, orgType := range organizationTypes {
		if orgType == organizationType {
			return true, nil
		}
	}

	return false, nil
}
