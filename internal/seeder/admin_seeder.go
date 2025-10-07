package seeder

import (
	"go-base-project/internal/model"
	"errors"

	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// CreateAdminUser creates default super admin user if it doesn't exist in database.
// This function is idempotent - safe to run multiple times.
// Super admin bypasses all permission checks via backend logic.
func CreateAdminUser(db *gorm.DB, adminPassword string) error {
	adminUsername := "superadm"

	// Cek apakah pengguna admin sudah ada
	var user model.User
	err := db.Where("username = ?", adminUsername).First(&user).Error

	if err == nil {
		// Pengguna sudah ada, tidak perlu melakukan apa-apa.
		log.Info().Msgf("Super admin user '%s' already exists. Skipping creation.", adminUsername)
		return nil
	}

	// Jika error bukan karena 'record not found', kembalikan error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// Pengguna belum ada, buat yang baru
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
