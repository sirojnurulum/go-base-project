package bootstrap

import (
	"beresin-backend/internal/repository"

	"gorm.io/gorm"
)

// Repositories menampung semua instance repository untuk aplikasi.
type Repositories struct {
	User repository.UserRepository
	Role repository.RoleRepository
}

// InitRepositories menginisialisasi semua repository untuk aplikasi.
func InitRepositories(db *gorm.DB) *Repositories {
	userRepository := repository.NewUserRepository(db)
	roleRepository := repository.NewRoleRepository(db)

	return &Repositories{
		User: userRepository,
		Role: roleRepository,
	}
}
