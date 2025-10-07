package bootstrap

import (
	"go-base-project/internal/config"
	"go-base-project/internal/service"

	"github.com/go-redis/redis/v8"
)

// Services menampung semua instance service untuk aplikasi.
type Services struct {
	Auth          service.AuthServiceInterface
	Organization  service.OrganizationServiceInterface
	Role          service.RoleServiceInterface
	User          service.UserServiceInterface
	Authorization service.AuthorizationServiceInterface
}

// InitServices menginisialisasi semua service untuk aplikasi.
func InitServices(repos *Repositories, redisClient *redis.Client, cfg config.Config) *Services {
	authorizationService := service.NewAuthorizationService(repos.Role, repos.User, redisClient)
	authService := service.NewAuthService(repos.User, repos.Role, authorizationService, redisClient, cfg.JWTSecret)
	organizationService := service.NewOrganizationService(repos.Organization, repos.User)
	roleService := service.NewRoleService(repos.Role, authorizationService)
	userService := service.NewUserService(repos.User, repos.Role, organizationService, redisClient)

	return &Services{
		Auth:          authService,
		Organization:  organizationService,
		Role:          roleService,
		User:          userService,
		Authorization: authorizationService,
	}
}
