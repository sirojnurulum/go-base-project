package bootstrap

import (
	"beresin-backend/internal/config"
	"beresin-backend/internal/handler"
	"beresin-backend/internal/util"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// Handlers menampung semua instance handler untuk aplikasi.
type Handlers struct {
	Auth         *handler.AuthHandler
	Health       *handler.HealthHandler
	Role         *handler.RoleHandler
	User         *handler.UserHandler
	Organization *handler.OrganizationHandler
}

// InitHandlers menginisialisasi semua handler untuk aplikasi.
func InitHandlers(services *Services, cfg config.Config, db *gorm.DB, redis *redis.Client) *Handlers {
	googleOauthConfig := util.SetupGoogleOauth(cfg)

	authHandler := handler.NewAuthHandler(services.Auth, googleOauthConfig, cfg)
	healthHandler := handler.NewHealthHandler(db, redis)
	roleHandler := handler.NewRoleHandler(services.Role)
	userHandler := handler.NewUserHandler(services.User)
	organizationHandler := handler.NewOrganizationHandler(services.Organization)

	return &Handlers{
		Auth:         authHandler,
		Health:       healthHandler,
		Role:         roleHandler,
		User:         userHandler,
		Organization: organizationHandler,
	}
}
