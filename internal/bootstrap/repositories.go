package bootstrap

import (
	"go-base-project/internal/repository"

	"gorm.io/gorm"
)

// Repositories menampung semua instance repository untuk aplikasi.
type Repositories struct {
	Organization repository.OrganizationRepositoryInterface
	User         repository.UserRepositoryInterface
	Role         repository.RoleRepositoryInterface
}

// InitRepositories menginisialisasi semua repository untuk aplikasi.
func InitRepositories(db *gorm.DB) *Repositories {
	organizationRepository := repository.NewOrganizationRepository(db)
	userRepository := repository.NewUserRepository(db)
	roleRepository := repository.NewRoleRepository(db)

	return &Repositories{
		Organization: organizationRepository,
		User:         userRepository,
		Role:         roleRepository,
	}
}
