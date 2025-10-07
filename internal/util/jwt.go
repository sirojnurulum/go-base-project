package util

import (
	"beresin-backend/internal/constant"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// JWTClaims represents the custom claims for JWT tokens.
type JWTClaims struct {
	UserID         uuid.UUID  `json:"user_id"`
	RoleID         uuid.UUID  `json:"role_id"`
	OrganizationID *uuid.UUID `json:"organization_id,omitempty"`
	jwt.RegisteredClaims
}

// RefreshTokenClaims represents the claims for refresh tokens.
type RefreshTokenClaims struct {
	jwt.RegisteredClaims // Only contains standard claims (sub, exp, etc.)
}

// GenerateAccessToken creates a new access token with user claims.
func GenerateAccessToken(userID, roleID uuid.UUID, organizationID *uuid.UUID, secret string) (string, error) {
	claims := JWTClaims{
		UserID:         userID,
		RoleID:         roleID,
		OrganizationID: organizationID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)), // 15 minutes
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// GenerateRefreshToken creates a new refresh token.
func GenerateRefreshToken(userID uuid.UUID, secret string) (string, error) {
	claims := RefreshTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)), // 7 days
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// VerifyAndGetClaims parses token from header, verifies it, and returns custom claims.
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

// VerifyRefreshToken verifies a refresh token and returns the user ID.
func VerifyRefreshToken(tokenString, secret string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return uuid.Nil, err
	}

	if claims, ok := token.Claims.(*RefreshTokenClaims); ok && token.Valid {
		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			return uuid.Nil, err
		}
		return userID, nil
	}

	return uuid.Nil, errors.New("invalid refresh token")
}
