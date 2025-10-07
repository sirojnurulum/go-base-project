package seeder

import (
	"go-base-project/internal/model"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// SeedRolesAndPermissions seeds predefined roles and permissions in the database.
// This is an idempotent operation. Super admin role gets NO permissions (access bypassed via backend logic).
func SeedRolesAndPermissions(db *gorm.DB) error {
	// 1. Define all system permissions
	permissions := []model.Permission{
		{Name: "users:create", Description: "Can create new users"},
		{Name: "users:read", Description: "Can read user data"},
		{Name: "users:update", Description: "Can update user data"},
		{Name: "users:delete", Description: "Can delete users"},
		{Name: "users:assign-organization", Description: "Can assign users to organizations"},
		{Name: "users:remove-organization", Description: "Can remove users from organizations"},
		{Name: "users:bulk-assign-organization", Description: "Can bulk assign users to organizations"},
		{Name: "users:update-organization-role", Description: "Can update user roles in organizations"},
		{Name: "roles:assign", Description: "Can assign roles to users"},
		{Name: "roles:create", Description: "Can create new roles"},
		{Name: "roles:approve", Description: "Can approve role creation requests"},
		{Name: "dashboard:view", Description: "Can view the main dashboard"},
		{Name: "scanned_data:create", Description: "Can create new scanned data entries"},
		// Shipping Management Permissions
		{Name: "shipping:scan", Description: "Can scan packages for shipping"},
		{Name: "shipping:manage", Description: "Can manage shipping operations and cancellations"},
		// Organization Management Permissions
		{Name: "organizations:create", Description: "Can create new organizations"},
		{Name: "organizations:read", Description: "Can read organization data"},
		{Name: "organizations:update", Description: "Can update organization data"},
		{Name: "organizations:delete", Description: "Can delete organizations"},
		{Name: "organizations:manage_members", Description: "Can manage organization members"},
	}

	// Seed all permissions
	for _, p := range permissions {
		if err := db.FirstOrCreate(&p, model.Permission{Name: p.Name}).Error; err != nil {
			log.Error().Err(err).Msgf("Failed to seed permission: %s", p.Name)
			return err
		}
	}
	log.Info().Msg("Permissions seeded successfully.")

	// 2. Define system roles with hierarchy
	rolesToSeed := []struct {
		Name           string
		Description    string
		Level          int
		IsSystemRole   bool
		PredefinedName string
		Permissions    []string
	}{
		{
			Name:           "super_admin",
			Description:    "System Super Administrator",
			Level:          100, // Platform level - bypasses all permission checks
			IsSystemRole:   true,
			PredefinedName: "Nexus",
			Permissions:    []string{}, // EMPTY - Access bypassed via backend logic
		},
	}

	for _, roleData := range rolesToSeed {
		var role model.Role
		if err := db.FirstOrCreate(&role, model.Role{Name: roleData.Name}).Error; err != nil {
			log.Error().Err(err).Msgf("Failed to seed role: %s", roleData.Name)
			return err
		}

		// Update role with hierarchy information
		role.Description = roleData.Description
		role.Level = roleData.Level
		role.IsSystemRole = roleData.IsSystemRole
		role.PredefinedName = roleData.PredefinedName
		role.IsActive = true

		if err := db.Save(&role).Error; err != nil {
			log.Error().Err(err).Msgf("Failed to update role hierarchy: %s", roleData.Name)
			return err
		}

		// Handle role permissions assignment
		if len(roleData.Permissions) > 0 {
			// Assign specified permissions to role
			var permissionsToAssign []model.Permission
			if err := db.Where("name IN ?", roleData.Permissions).Find(&permissionsToAssign).Error; err != nil {
				return err
			}
			if err := db.Model(&role).Association("Permissions").Replace(&permissionsToAssign); err != nil {
				log.Error().Err(err).Msgf("Failed to associate permissions for role: %s", roleData.Name)
				return err
			}
		} else {
			// Clear all permissions (for super_admin and other bypass roles)
			if err := db.Model(&role).Association("Permissions").Clear(); err != nil {
				log.Error().Err(err).Msgf("Failed to clear permissions for role: %s", roleData.Name)
				return err
			}
		}
	}

	log.Info().Msg("Roles and permission associations seeded successfully.")
	return nil
}
