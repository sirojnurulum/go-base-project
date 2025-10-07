package service

import (
	"go-base-project/internal/dto"
	"context"

	"github.com/google/uuid"
)

// AuthService mendefinisikan kontrak untuk layanan otentikasi.
type AuthServiceInterface interface {
	Login(ctx context.Context, username, password string) (*dto.LoginResult, error)
	// Modifikasi: RefreshToken sekarang mengembalikan refresh token baru juga.
	RefreshToken(ctx context.Context, tokenString string) (newAccessToken string, newRefreshToken string, err error)
	LoginWithGoogle(ctx context.Context, userInfo dto.GoogleUserInfo) (*dto.LoginResult, error)
	Logout(ctx context.Context, refreshToken string) error
	GetUserWithPermissions(ctx context.Context, userID string) (*dto.LoginResult, error)
	SwitchOrganizationContext(ctx context.Context, userID, roleID uuid.UUID, organizationID string) (*dto.SwitchOrganizationResult, error)
}
