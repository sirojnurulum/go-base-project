package seeder

import (
	"beresin-backend/internal/model"
	"errors"

	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// CreateAdminUser membuat pengguna admin default jika belum ada di database.
// Fungsi ini idempoten, artinya aman untuk dijalankan berkali-kali.
func CreateAdminUser(db *gorm.DB, adminPassword string) error {
	adminUsername := "admin"

	// Cek apakah pengguna admin sudah ada
	var user model.User
	err := db.Where("username = ?", adminUsername).First(&user).Error

	if err == nil {
		// Pengguna sudah ada, tidak perlu melakukan apa-apa.
		log.Info().Msgf("Admin user '%s' already exists. Skipping creation.", adminUsername)
		return nil
	}

	// Jika error bukan karena 'record not found', kembalikan error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// Pengguna belum ada, buat yang baru
	log.Info().Msgf("Admin user '%s' not found. Creating a new one...", adminUsername)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Dapatkan role admin dari database
	var adminRole model.Role
	if err := db.Where("name = ?", "admin").First(&adminRole).Error; err != nil {
		log.Error().Err(err).Msg("Could not find 'admin' role. Please run the RBAC seeder first.")
		return err
	}

	adminUser := model.User{
		Username: adminUsername,
		Password: string(hashedPassword),
		RoleID:   &adminRole.ID,
	}

	return db.Create(&adminUser).Error
}
