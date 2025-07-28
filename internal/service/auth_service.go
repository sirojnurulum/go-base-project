package service

import (
	"beresin-backend/internal/apperror"
	"beresin-backend/internal/constant"
	"beresin-backend/internal/dto"
	"beresin-backend/internal/model"
	"beresin-backend/internal/repository"
	"beresin-backend/internal/util"
	"beresin-backend/pkg/generator"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// authService implements the AuthService interface for authentication-related logic.
type authService struct {
	userRepo             repository.UserRepository
	roleRepo             repository.RoleRepository
	authorizationService AuthorizationService
	redis                *redis.Client
	jwtSecret            string
}

var ctx = context.Background()

// NewAuthService creates a new instance of authService.
func NewAuthService(userRepo repository.UserRepository, roleRepo repository.RoleRepository, authorizationService AuthorizationService, redis *redis.Client, jwtSecret string) AuthService {
	return &authService{
		userRepo:             userRepo,
		roleRepo:             roleRepo,
		authorizationService: authorizationService,
		redis:                redis,
		jwtSecret:            jwtSecret,
	}
}

// Logout invalidates a refresh token by deleting it from Redis.
func (s *authService) Logout(refreshToken string) error {
	// The key in Redis is the refresh token itself.
	err := s.redis.Del(ctx, refreshToken).Err()
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete refresh token from Redis")
		// We don't return an error to the user, as the main goal is to clear
		// the cookie, which will happen regardless. Logging is crucial.
	}
	return nil // Always succeed from the user's perspective.
}

// Login validates credentials, generates tokens, and stores the refresh token in Redis.
func (s *authService) Login(username, password string) (*dto.LoginResult, error) {
	// 1. Find user by username, preload role
	user, err := s.userRepo.FindByUsernameWithRole(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NewUnauthorizedError(constant.ErrMsgInvalidCredentials)
		}
		return nil, apperror.NewInternalError(fmt.Errorf("failed to find user: %w", err))
	}

	if user == nil || user.Role == nil {
		return nil, apperror.NewUnauthorizedError(constant.ErrMsgInvalidCredentials)
	}
	// 2. Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, apperror.NewUnauthorizedError(constant.ErrMsgInvalidCredentials)
	}

	// 3. Create token and login result DTO
	return s.createLoginResultForUser(user)
}

// RefreshToken validates a refresh token, then issues a new access token and a new refresh token (rotation).
func (s *authService) RefreshToken(tokenString string) (string, string, error) {
	// 1. Parse dan validasi refresh token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, apperror.NewUnauthorizedError("unexpected signing method")
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return "", "", apperror.NewUnauthorizedError("invalid or expired refresh token")
	}

	// 2. Cek apakah refresh token ada di Redis (belum dicabut/dirotasi)
	userIDStrFromRedis, err := s.redis.Get(ctx, tokenString).Result()
	if err == redis.Nil {
		return "", "", apperror.NewUnauthorizedError("refresh token not found or already used") // Token sudah dirotasi atau tidak valid
	} else if err != nil {
		return "", "", apperror.NewInternalError(fmt.Errorf("failed to check refresh token in redis: %w", err))
	}

	// Hapus token lama dari Redis setelah digunakan.
	if err := s.redis.Del(ctx, tokenString).Err(); err != nil {
		return "", "", apperror.NewInternalError(fmt.Errorf("could not delete old refresh token: %w", err))
	}

	// 3. Ambil claims dan User ID dari 'subject'
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", apperror.NewUnauthorizedError("invalid token claims")
	}

	userIDStr, ok := claims["sub"].(string)
	if !ok || userIDStr != userIDStrFromRedis {
		return "", "", apperror.NewUnauthorizedError("token user mismatch") // Safety check
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return "", "", apperror.NewUnauthorizedError("invalid user id format")
	}

	// 4. Cek apakah user masih ada di database
	user, err := s.userRepo.FindByIDWithRole(userID)
	if err != nil {
		// Jika user tidak ditemukan, kembalikan error unauthorized, bukan not found
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", apperror.NewUnauthorizedError("user for this token no longer exists")
		}
		return "", "", apperror.NewInternalError(err)
	}
	if user.RoleID == nil {
		return "", "", apperror.NewInternalError(fmt.Errorf("user %s has no role assigned", user.ID))
	}

	// 5. Create a new access token AND a new refresh token (Token Rotation)
	newAccessToken, _ := util.GenerateAccessToken(user.ID, *user.RoleID, s.jwtSecret)
	newRefreshToken, _ := util.GenerateRefreshToken(user.ID, s.jwtSecret)

	// 6. Store the new refresh token in Redis
	refreshTokenDuration := 7 * 24 * time.Hour
	if err := s.redis.Set(ctx, newRefreshToken, user.ID.String(), refreshTokenDuration).Err(); err != nil {
		return "", "", apperror.NewInternalError(fmt.Errorf("failed to store new refresh token in redis: %w", err))
	}

	return newAccessToken, newRefreshToken, nil
}

// LoginWithGoogle handles the user login or registration flow via Google OAuth.
func (s *authService) LoginWithGoogle(userInfo dto.GoogleUserInfo) (*dto.LoginResult, error) {
	// 1. Cek apakah user sudah ada dengan Google ID
	user, err := s.userRepo.FindByGoogleID(userInfo.ID)
	if err == nil {
		// User ditemukan, langsung login
		user, _ = s.userRepo.FindByIDWithRole(user.ID) // Muat ulang dengan role
		return s.createLoginResultForUser(user)
	}

	// Jika error bukan karena "not found", berarti ada masalah lain
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.NewInternalError(fmt.Errorf("error finding user by google id: %w", err))
	}

	// 2. User belum ada, cek berdasarkan email (mungkin sudah daftar manual)
	user, err = s.userRepo.FindByEmail(userInfo.Email)
	if err == nil {
		// User dengan email yang sama ditemukan, tautkan akunnya
		user.GoogleID = userInfo.ID
		user.AvatarURL = userInfo.Picture
		if err := s.userRepo.Update(user); err != nil {
			return nil, apperror.NewInternalError(fmt.Errorf("failed to link google account: %w", err))
		}
		user, _ = s.userRepo.FindByIDWithRole(user.ID) // Muat ulang dengan role
		return s.createLoginResultForUser(user)
	}

	// 3. User benar-benar baru, buat akun baru
	newUser := &model.User{
		Email:     userInfo.Email,
		Username:  generator.GenerateFromEmail(userInfo.Email), // Pastikan username unik
		GoogleID:  userInfo.ID,
		AvatarURL: userInfo.Picture,
	}

	// Temukan role default ("user") di database
	defaultRole, err := s.roleRepo.FindRoleByName("user")
	if err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to find default role: %w", err))
	}
	newUser.RoleID = &defaultRole.ID
	newUser.Role = defaultRole // Tautkan objek role secara langsung

	if err := s.userRepo.Create(newUser); err != nil { // Buat user baru dengan role yang sudah ditautkan
		return nil, apperror.NewInternalError(fmt.Errorf("failed to create user from google info: %w", err))
	}
	return s.createLoginResultForUser(newUser)
}

// createLoginResultForUser is an internal helper to generate access & refresh tokens,
// fetch permissions, and build the LoginResult DTO after a user is successfully authenticated.
func (s *authService) createLoginResultForUser(user *model.User) (*dto.LoginResult, error) {
	if user.RoleID == nil || user.Role == nil {
		return nil, apperror.NewInternalError(fmt.Errorf("user %s has no role assigned or role data is missing", user.Username))
	}

	accessToken, _ := util.GenerateAccessToken(user.ID, *user.RoleID, s.jwtSecret)
	refreshToken, _ := util.GenerateRefreshToken(user.ID, s.jwtSecret)

	refreshTokenDuration := 7 * 24 * time.Hour // 7 days
	if err := s.redis.Set(ctx, refreshToken, user.ID.String(), refreshTokenDuration).Err(); err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to store refresh token in redis: %w", err))
	}

	// Fetch and cache permissions for this user's role.
	permissions, err := s.authorizationService.GetAndCachePermissionsForRole(*user.RoleID)
	if err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to get permissions for role %s: %w", user.RoleID.String(), err))
	}

	// Create UserResponse DTO from the user model

	userResponse := &dto.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role.Name,
		AvatarURL: user.AvatarURL,
	}

	return &dto.LoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         userResponse,
		Permissions:  permissions,
	}, nil
}
