package seeder

import (
	"errors"
	"go-base-project/internal/model"

	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// CreateAdminUser creates default super admin user if it doesn't exist in database.
// This function is idempotent - safe to run multiple times.
// Super admin bypasses all permission checks via backend logic.
func CreateAdminUser(db *gorm.DB, adminUsername, adminPassword string) error {
	// Check if admin user already exists
	var user model.User
	err := db.Where("username = ?", adminUsername).First(&user).Error

	if err == nil {
		// User already exists, no need to do anything.
		log.Info().Msgf("Super admin user '%s' already exists. Skipping creation.", adminUsername)
		return nil
	}

	// If error is not because of 'record not found', return error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// User doesn't exist yet, create a new one
	log.Info().Msgf("Super admin user '%s' not found. Creating a new one...", adminUsername)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Get super_admin role from database
	var superAdminRole model.Role
	if err := db.Where("name = ?", "super_admin").First(&superAdminRole).Error; err != nil {
		log.Error().Err(err).Msg("Could not find 'super_admin' role. Please run the RBAC seeder first.")
		return err
	}

	adminUser := model.User{
		Username: adminUsername,
		Password: string(hashedPassword),
		RoleID:   &superAdminRole.ID,
	}

	return db.Create(&adminUser).Error
}
