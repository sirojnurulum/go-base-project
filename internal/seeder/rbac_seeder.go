package seeder

import (
	"beresin-backend/internal/model"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// SeedRolesAndPermissions membuat peran dan izin yang telah ditentukan sebelumnya dalam database.
// Ini adalah operasi idempoten.
func SeedRolesAndPermissions(db *gorm.DB) error {
	// 1. Definisikan semua izin yang ada di sistem
	permissions := []model.Permission{
		{Name: "users:create", Description: "Can create new users"},
		{Name: "users:read", Description: "Can read user data"},
		{Name: "users:update", Description: "Can update user data"},
		{Name: "users:delete", Description: "Can delete users"},
		{Name: "roles:assign", Description: "Can assign roles to users"},
		{Name: "dashboard:view", Description: "Can view the main dashboard"},
		{Name: "roles:create", Description: "Can create new roles"},
		{Name: "scanned_data:create", Description: "Can create new scanned data entries"},
	}

	var allPermissionNames []string
	for _, p := range permissions {
		allPermissionNames = append(allPermissionNames, p.Name)
		if err := db.FirstOrCreate(&p, model.Permission{Name: p.Name}).Error; err != nil {
			log.Error().Err(err).Msgf("Failed to seed permission: %s", p.Name)
			return err
		}
	}
	log.Info().Msg("Permissions seeded successfully.")

	// 2. Definisikan peran dan petakan izinnya
	rolesToSeed := map[string][]string{
		"admin": allPermissionNames, // Admin secara dinamis mendapatkan semua izin yang terdaftar
		"user": { // Pengguna dasar hanya dapat melihat dashboard
			"dashboard:view",
		},
	}

	for roleName, permissionNames := range rolesToSeed {
		var role model.Role
		if err := db.FirstOrCreate(&role, model.Role{Name: roleName}).Error; err != nil {
			log.Error().Err(err).Msgf("Failed to seed role: %s", roleName)
			return err
		}

		var permissionsToAssign []model.Permission
		if err := db.Where("name IN ?", permissionNames).Find(&permissionsToAssign).Error; err != nil {
			return err
		}

		if err := db.Model(&role).Association("Permissions").Replace(&permissionsToAssign); err != nil {
			log.Error().Err(err).Msgf("Failed to associate permissions for role: %s", roleName)
			return err
		}
	}

	log.Info().Msg("Roles and permission associations seeded successfully.")
	return nil
}
