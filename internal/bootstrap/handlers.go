package bootstrap

import (
	"go-base-project/internal/config"
	"go-base-project/internal/handler"
	"go-base-project/internal/util"
)

// Handlers menampung semua instance handler untuk aplikasi.
type Handlers struct {
	Auth         *handler.AuthHandler
	Health       *handler.HealthHandler
	Organization *handler.OrganizationHandler
	Role         *handler.RoleHandler
	User         *handler.UserHandler
}

// InitHandlers menginisialisasi semua handler untuk aplikasi.
func InitHandlers(services *Services, cfg config.Config) *Handlers {
	googleOauthConfig := util.SetupGoogleOauth(cfg)

	authHandler := handler.NewAuthHandler(services.Auth, googleOauthConfig, cfg)
	healthHandler := handler.NewHealthHandler()
	organizationHandler := handler.NewOrganizationHandler(services.Organization)
	roleHandler := handler.NewRoleHandler(services.Role)
	userHandler := handler.NewUserHandler(services.User)

	return &Handlers{
		Auth:         authHandler,
		Health:       healthHandler,
		Organization: organizationHandler,
		Role:         roleHandler,
		User:         userHandler,
	}
}
