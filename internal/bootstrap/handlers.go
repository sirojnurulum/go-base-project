package bootstrap

import (
	"beresin-backend/internal/config"
	"beresin-backend/internal/handler"
	"beresin-backend/internal/util"
)

// Handlers menampung semua instance handler untuk aplikasi.
type Handlers struct {
	Auth   *handler.AuthHandler
	Health *handler.HealthHandler
	Role   *handler.RoleHandler
	User   *handler.UserHandler
}

// InitHandlers menginisialisasi semua handler untuk aplikasi.
func InitHandlers(services *Services, cfg config.Config) *Handlers {
	googleOauthConfig := util.SetupGoogleOauth(cfg)

	authHandler := handler.NewAuthHandler(services.Auth, googleOauthConfig, cfg)
	healthHandler := handler.NewHealthHandler()
	roleHandler := handler.NewRoleHandler(services.Role)
	userHandler := handler.NewUserHandler(services.User)

	return &Handlers{
		Auth:   authHandler,
		Health: healthHandler,
		Role:   roleHandler,
		User:   userHandler,
	}
}
