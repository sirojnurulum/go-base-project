package service

import (
	"beresin-backend/internal/dto"
)

// AuthService mendefinisikan kontrak untuk layanan otentikasi.
type AuthService interface {
	Login(username, password string) (*dto.LoginResult, error)
	// Modifikasi: RefreshToken sekarang mengembalikan refresh token baru juga.
	RefreshToken(tokenString string) (newAccessToken string, newRefreshToken string, err error)
	LoginWithGoogle(userInfo dto.GoogleUserInfo) (*dto.LoginResult, error)
	Logout(refreshToken string) error
}
