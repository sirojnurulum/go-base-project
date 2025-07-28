package util

import (
	"beresin-backend/internal/config"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// SetupGoogleOauth membuat konfigurasi OAuth2 untuk Google.
func SetupGoogleOauth(cfg config.Config) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  "http://localhost:8080/api/auth/google/callback", // Sesuaikan untuk produksi!
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email",
		},
		Endpoint: google.Endpoint,
	}
}
