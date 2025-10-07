package util

import (
	"go-base-project/internal/constant"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// JWTClaims adalah struct untuk custom claims JWT kita.
type JWTClaims struct {
	UserID         uuid.UUID  `json:"user_id"`
	RoleID         uuid.UUID  `json:"role_id"`
	OrganizationID *uuid.UUID `json:"organization_id,omitempty"`
	jwt.RegisteredClaims
}

// VerifyAndGetClaims mem-parsing token dari header, memverifikasinya, dan mengembalikan custom claims.
func VerifyAndGetClaims(c echo.Context, jwtSecret string) (*JWTClaims, error) {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return nil, errors.New(constant.ErrMsgMissingAuthHeader)
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader { // No "Bearer " prefix found
		return nil, errors.New(constant.ErrMsgInvalidOrExpiredToken)
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, errors.New(constant.ErrMsgInvalidOrExpiredToken)
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New(constant.ErrMsgInvalidOrExpiredToken)
}

// GenerateAccessToken membuat access token baru.
func GenerateAccessToken(userID, roleID uuid.UUID, secret string) (string, error) {
	claims := &JWTClaims{
		UserID: userID,
		RoleID: roleID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			Subject:   userID.String(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// GenerateAccessTokenWithOrganization membuat access token baru dengan organization context.
func GenerateAccessTokenWithOrganization(userID, roleID uuid.UUID, organizationID *uuid.UUID, secret string) (string, error) {
	claims := &JWTClaims{
		UserID:         userID,
		RoleID:         roleID,
		OrganizationID: organizationID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			Subject:   userID.String(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// GenerateRefreshToken membuat refresh token baru.
func GenerateRefreshToken(userID uuid.UUID, secret string) (string, error) {
	// Jika test hook diatur, gunakan untuk mengembalikan token yang dapat diprediksi.
	if testRefreshTokenHook != nil {
		return testRefreshTokenHook(), nil
	}

	// Refresh token tidak perlu membawa role_id, hanya user_id (subject).
	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)), // 7 days
		Subject:   userID.String(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// testRefreshTokenHook adalah variabel level paket untuk tes agar dapat mengatur refresh token yang dapat diprediksi.
var testRefreshTokenHook func() string

// SetTestRefreshToken adalah helper untuk tes agar dapat mengatur refresh token yang dapat diprediksi.
// Ini hanya boleh digunakan dalam file tes. Ingatlah untuk meresetnya dengan `defer util.SetTestRefreshToken("")`.
func SetTestRefreshToken(token string) {
	if token == "" {
		testRefreshTokenHook = nil
		return
	}
	testRefreshTokenHook = func() string {
		return token
	}
}
