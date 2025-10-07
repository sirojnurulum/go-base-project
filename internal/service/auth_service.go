package service

import (
	"go-base-project/internal/apperror"
	"go-base-project/internal/constant"
	"go-base-project/internal/dto"
	"go-base-project/internal/model"
	"go-base-project/internal/repository"
	"go-base-project/internal/util"
	"go-base-project/pkg/generator"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// authService implements the AuthService interface for authentication-related logic.
type authService struct {
	userRepo             repository.UserRepositoryInterface
	roleRepo             repository.RoleRepositoryInterface
	authorizationService AuthorizationServiceInterface
	redis                *redis.Client
	jwtSecret            string
}

// NewAuthService creates a new instance of authService.
func NewAuthService(userRepo repository.UserRepositoryInterface, roleRepo repository.RoleRepositoryInterface, authorizationService AuthorizationServiceInterface, redis *redis.Client, jwtSecret string) AuthServiceInterface {
	return &authService{
		userRepo:             userRepo,
		roleRepo:             roleRepo,
		authorizationService: authorizationService,
		redis:                redis,
		jwtSecret:            jwtSecret,
	}
}

// Logout invalidates a refresh token by deleting it from Redis.
func (s *authService) Logout(ctx context.Context, refreshToken string) error {
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
func (s *authService) Login(ctx context.Context, username, password string) (*dto.LoginResult, error) {
	ctx, span := otel.Tracer("authService").Start(ctx, "Login")
	defer span.End()

	// Input validation
	if username == "" || password == "" {
		return nil, apperror.NewValidationError("Username and password are required")
	}

	// 1. Find user by username with role (efficient single query)
	user, err := s.userRepo.FindByUsernameWithRole(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Warn().Str("username", username).Msg("Login attempt with non-existent username")
			return nil, apperror.NewUnauthorizedError(constant.ErrMsgInvalidCredentials)
		}
		log.Error().Err(err).Str("username", username).Msg("Failed to find user during login")
		return nil, apperror.NewInternalError(fmt.Errorf("failed to find user: %w", err))
	}

	// 2. Validate user and role data
	if user == nil {
		log.Warn().Str("username", username).Msg("User data is nil")
		return nil, apperror.NewUnauthorizedError(constant.ErrMsgInvalidCredentials)
	}
	if user.Role == nil {
		log.Error().Str("username", username).Str("user_id", user.ID.String()).Msg("User has no role assigned")
		return nil, apperror.NewUnauthorizedError("Account is not properly configured")
	}

	// 3. Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		log.Warn().Str("username", username).Msg("Invalid password attempt")
		return nil, apperror.NewUnauthorizedError(constant.ErrMsgInvalidCredentials)
	}

	// 4. Create comprehensive login result using repository data
	loginResult, err := s.createLoginResultForUser(ctx, user)
	if err != nil {
		log.Error().Err(err).Str("username", username).Str("user_id", user.ID.String()).Msg("Failed to create login result")
		return nil, err
	}

	log.Info().Str("username", username).Str("user_id", user.ID.String()).Str("role", user.Role.Name).Msg("User logged in successfully")
	return loginResult, nil
}

// RefreshToken validates a refresh token, then issues a new access token and a new refresh token (rotation).
func (s *authService) RefreshToken(ctx context.Context, tokenString string) (string, string, error) {
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
	user, err := s.userRepo.FindByIDWithRoleAndOrganizations(ctx, userID)
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
func (s *authService) LoginWithGoogle(ctx context.Context, userInfo dto.GoogleUserInfo) (*dto.LoginResult, error) {
	// 1. Cek apakah user sudah ada dengan Google ID
	user, err := s.userRepo.FindByGoogleID(ctx, userInfo.ID)
	if err == nil {
		// User ditemukan, langsung login
		user, _ = s.userRepo.FindByIDWithRoleAndOrganizations(ctx, user.ID) // Muat ulang dengan role dan organizations
		return s.createLoginResultForUser(ctx, user)
	}

	// Jika error bukan karena "not found", berarti ada masalah lain
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.NewInternalError(fmt.Errorf("error finding user by google id: %w", err))
	}

	// 2. User belum ada, cek berdasarkan email (mungkin sudah daftar manual)
	user, err = s.userRepo.FindByEmail(ctx, userInfo.Email)
	if err == nil {
		// User dengan email yang sama ditemukan, tautkan akunnya
		user.GoogleID = &userInfo.ID
		user.AvatarURL = userInfo.Picture
		user.AuthProvider = "google"
		if err := s.userRepo.Update(ctx, user); err != nil {
			return nil, apperror.NewInternalError(fmt.Errorf("failed to link google account: %w", err))
		}
		user, _ = s.userRepo.FindByIDWithRoleAndOrganizations(ctx, user.ID) // Muat ulang dengan role dan organizations
		return s.createLoginResultForUser(ctx, user)
	}

	// 3. User benar-benar baru, buat akun baru TANPA role dan organization
	newUser := &model.User{
		Email:        userInfo.Email,
		Username:     generator.GenerateFromEmail(userInfo.Email), // Pastikan username unik
		GoogleID:     &userInfo.ID,
		AvatarURL:    userInfo.Picture,
		AuthProvider: "google",
		// TIDAK assign role atau organization untuk user baru
		// RoleID akan tetap nil, user harus request role sendiri
	}

	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to create user from google info: %w", err))
	}

	// User baru tanpa role, return hasil dengan permissions kosong
	return s.createLoginResultForUserWithoutRole(ctx, newUser)
}

// createLoginResultForUser is an internal helper to generate access & refresh tokens,
// fetch permissions, and build the LoginResult DTO after a user is successfully authenticated.
func (s *authService) createLoginResultForUser(ctx context.Context, user *model.User) (*dto.LoginResult, error) {
	// Check if user has no role (handle new users or users with unassigned roles)
	if user.RoleID == nil {
		return s.createLoginResultForUserWithoutRole(ctx, user)
	}

	// Validate role data consistency
	if user.Role == nil {
		return nil, apperror.NewInternalError(fmt.Errorf("user %s has role ID but role data is missing", user.Username))
	}

	// Generate tokens with proper error handling
	accessToken, err := util.GenerateAccessToken(user.ID, *user.RoleID, s.jwtSecret)
	if err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to generate access token: %w", err))
	}

	refreshToken, err := util.GenerateRefreshToken(user.ID, s.jwtSecret)
	if err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to generate refresh token: %w", err))
	}

	// Store refresh token in Redis with proper expiration
	refreshTokenDuration := 7 * 24 * time.Hour // 7 days
	if err := s.redis.Set(ctx, refreshToken, user.ID.String(), refreshTokenDuration).Err(); err != nil {
		log.Error().Err(err).Str("user_id", user.ID.String()).Msg("Failed to store refresh token in Redis")
		return nil, apperror.NewInternalError(fmt.Errorf("failed to store refresh token: %w", err))
	}

	// Fetch and cache permissions using authorization service
	permissions, err := s.authorizationService.GetAndCachePermissionsForRole(ctx, *user.RoleID)
	if err != nil {
		log.Error().Err(err).Str("user_id", user.ID.String()).Str("role_id", user.RoleID.String()).Msg("Failed to fetch permissions")
		return nil, apperror.NewInternalError(fmt.Errorf("failed to get permissions for role: %w", err))
	}

	// Create structured UserResponse DTO using utility function
	userResponse := util.MapUserToResponse(user)

	return &dto.LoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         userResponse,
		Permissions:  permissions,
	}, nil
}

// createLoginResultForUserWithoutRole handles new users without roles
// These users need to complete organization joining and role request process
func (s *authService) createLoginResultForUserWithoutRole(ctx context.Context, user *model.User) (*dto.LoginResult, error) {
	// Generate tokens even for users without roles (they still need to authenticate)
	// Use a zero UUID for role_id in token since user has no role yet
	zeroRoleID := uuid.Nil
	accessToken, err := util.GenerateAccessToken(user.ID, zeroRoleID, s.jwtSecret)
	if err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to generate access token: %w", err))
	}

	refreshToken, err := util.GenerateRefreshToken(user.ID, s.jwtSecret)
	if err != nil {
		return nil, apperror.NewInternalError(fmt.Errorf("failed to generate refresh token: %w", err))
	}

	// Store refresh token in Redis
	refreshTokenDuration := 7 * 24 * time.Hour // 7 days
	if err := s.redis.Set(ctx, refreshToken, user.ID.String(), refreshTokenDuration).Err(); err != nil {
		log.Error().Err(err).Str("user_id", user.ID.String()).Msg("Failed to store refresh token in Redis")
		return nil, apperror.NewInternalError(fmt.Errorf("failed to store refresh token: %w", err))
	}

	// Create user response - no role information
	userResponse := util.MapUserToResponse(user)

	return &dto.LoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         userResponse,
		Permissions:  []string{}, // Empty permissions - user needs to request role
	}, nil
}

// GetUserWithPermissions retrieves a user by ID along with their role permissions
func (s *authService) GetUserWithPermissions(ctx context.Context, userID string) (*dto.LoginResult, error) {
	// Parse user ID as UUID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperror.NewValidationError("invalid user ID format")
	}

	// Get user with role and organization information
	user, err := s.userRepo.FindByIDWithRoleAndOrganizations(ctx, userUUID)
	if err != nil {
		return nil, apperror.NewNotFoundError("user")
	}

	// Get permissions for the user's role
	var permissions []string
	if user.Role != nil && user.RoleID != nil {
		permissions, err = s.roleRepo.FindPermissionsByRoleID(ctx, *user.RoleID)
		if err != nil {
			log.Error().Err(err).Str("user_id", userID).Str("role_id", user.RoleID.String()).Msg("Failed to get permissions")
			return nil, apperror.NewInternalError(fmt.Errorf("failed to get permissions for role: %w", err))
		}
	}

	// Create UserResponse DTO using utility function
	userResponse := util.MapUserToResponse(user)

	return &dto.LoginResult{
		// Don't include access/refresh tokens for this method
		User:        userResponse,
		Permissions: permissions,
	}, nil
}

// SwitchOrganizationContext switches the user's organization context and returns a new access token
func (s *authService) SwitchOrganizationContext(ctx context.Context, userID, roleID uuid.UUID, organizationID string) (*dto.SwitchOrganizationResult, error) {
	// Parse organization ID
	orgUUID, err := uuid.Parse(organizationID)
	if err != nil {
		return nil, apperror.NewValidationError("invalid organization ID format")
	}

	// Check if user has access to the specified organization
	hasAccess, err := s.authorizationService.CheckUserOrganizationAccess(ctx, userID, orgUUID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Str("organization_id", organizationID).Msg("Failed to check user organization access")
		return nil, apperror.NewInternalError(fmt.Errorf("failed to check organization access: %w", err))
	}

	if !hasAccess {
		return nil, apperror.NewForbiddenError("access denied to the specified organization")
	}

	// Generate new access token with organization context
	accessToken, err := util.GenerateAccessTokenWithOrganization(userID, roleID, &orgUUID, s.jwtSecret)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Str("organization_id", organizationID).Msg("Failed to generate access token with organization context")
		return nil, apperror.NewInternalError(fmt.Errorf("failed to generate access token: %w", err))
	}

	return &dto.SwitchOrganizationResult{
		AccessToken:    accessToken,
		OrganizationID: organizationID,
	}, nil
}
