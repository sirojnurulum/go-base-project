package bootstrap

import (
	"beresin-backend/internal/config"
	"beresin-backend/internal/service"

	"github.com/go-redis/redis/v8"
)

// Services menampung semua instance service untuk aplikasi.
type Services struct {
	Auth          service.AuthServiceInterface
	Role          service.RoleServiceInterface
	User          service.UserServiceInterface
	Authorization service.AuthorizationServiceInterface
	Organization  service.OrganizationServiceInterface
}

// InitServices menginisialisasi semua service untuk aplikasi.
func InitServices(repos *Repositories, redisClient *redis.Client, cfg config.Config) *Services {
	authorizationService := service.NewAuthorizationService(repos.Role, repos.User, redisClient)
	authService := service.NewAuthService(repos.User, repos.Role, authorizationService, redisClient, cfg.JWTSecret)
	roleService := service.NewRoleService(repos.Role, authorizationService)
	userService := service.NewUserService(repos.User, repos.Role, nil, redisClient) // nil for organizationService for now
	organizationService := service.NewOrganizationService(repos.Organization, repos.User)

	return &Services{
		Auth:          authService,
		Role:          roleService,
		User:          userService,
		Authorization: authorizationService,
		Organization:  organizationService,
	}
}
