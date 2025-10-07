package service

import (
	"beresin-backend/internal/apperror"
	"beresin-backend/internal/dto"
	"beresin-backend/internal/model"
	"beresin-backend/internal/repository"
	"beresin-backend/internal/util"
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
) // authService implements the AuthService interface for authentication-related logic.
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

// Login authenticates a user with username and password.
func (s *authService) Login(ctx context.Context, username, password string) (*dto.LoginResult, error) {
	// Find user by username
	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NewClientError("Invalid credentials")
		}
		return nil, err
	}

	// Check if user is active
	if !user.IsActive {
		return nil, apperror.NewClientError("Account is deactivated")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, apperror.NewClientError("Invalid credentials")
	}

	// Generate access token and refresh token
	accessToken, refreshToken, err := s.generateTokens(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	// Get user with permissions
	userWithPermissions, err := s.GetUserWithPermissions(ctx, user.ID.String())
	if err != nil {
		return nil, err
	}

	return &dto.LoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         userWithPermissions.User,
		Permissions:  userWithPermissions.Permissions,
	}, nil
}

// RefreshToken generates new access and refresh tokens from a valid refresh token.
func (s *authService) RefreshToken(ctx context.Context, tokenString string) (newAccessToken string, newRefreshToken string, err error) {
	// Check if refresh token exists in Redis
	userIDStr, err := s.redis.Get(ctx, tokenString).Result()
	if err != nil {
		if err == redis.Nil {
			return "", "", apperror.NewClientError("Invalid or expired refresh token")
		}
		return "", "", err
	}

	// Parse user ID
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return "", "", err
	}

	// Delete the old refresh token from Redis
	s.redis.Del(ctx, tokenString)

	// Generate new tokens
	newAccessToken, newRefreshToken, err = s.generateTokens(ctx, userID)
	if err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}

// Logout invalidates a refresh token by deleting it from Redis.
func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	err := s.redis.Del(ctx, refreshToken).Err()
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete refresh token from Redis")
		return err
	}
	return nil
}

// LoginWithGoogle authenticates a user with Google OAuth.
func (s *authService) LoginWithGoogle(ctx context.Context, userInfo dto.GoogleUserInfo) (*dto.LoginResult, error) {
	// Find user by Google ID
	user, err := s.userRepo.FindByGoogleID(ctx, userInfo.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Check if user exists by email
			existingUser, emailErr := s.userRepo.FindByEmail(ctx, userInfo.Email)
			if emailErr != nil && !errors.Is(emailErr, gorm.ErrRecordNotFound) {
				return nil, emailErr
			}

			if existingUser != nil {
				// Update existing user with Google ID
				existingUser.GoogleID = &userInfo.ID
				if err := s.userRepo.Update(ctx, existingUser); err != nil {
					return nil, err
				}
				user = existingUser
			} else {
				// Create new user
				user = &model.User{
					ID:       uuid.New(),
					Email:    userInfo.Email,
					Name:     userInfo.Name,
					Username: userInfo.Email, // Use email as username
					GoogleID: &userInfo.ID,
					IsActive: true,
				}
				if err := s.userRepo.Create(ctx, user); err != nil {
					return nil, err
				}
			}
		} else {
			return nil, err
		}
	}

	// Check if user is active
	if !user.IsActive {
		return nil, apperror.NewClientError("Account is deactivated")
	}

	// Generate tokens
	accessToken, refreshToken, err := s.generateTokens(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	// Get user with permissions
	userWithPermissions, err := s.GetUserWithPermissions(ctx, user.ID.String())
	if err != nil {
		return nil, err
	}

	return &dto.LoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         userWithPermissions.User,
		Permissions:  userWithPermissions.Permissions,
	}, nil
}

// GetUserWithPermissions retrieves user with permissions.
func (s *authService) GetUserWithPermissions(ctx context.Context, userID string) (*dto.LoginResult, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperror.NewClientError("Invalid user ID format")
	}

	// Find user with role and organizations
	user, err := s.userRepo.FindByIDWithRoleAndOrganizations(ctx, userUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NewClientError("User not found")
		}
		return nil, err
	}

	// Check if user is active
	if !user.IsActive {
		return nil, apperror.NewClientError("Account is deactivated")
	}

	// Get permissions for the user's role
	var permissions []string
	if user.Role != nil && user.RoleID != nil {
		permissions, err = s.roleRepo.FindPermissionsByRoleID(ctx, *user.RoleID)
		if err != nil {
			return nil, err
		}
	}

	// Create UserResponse DTO using utility function
	userResponse := util.MapUserToResponse(user)

	return &dto.LoginResult{
		User:        userResponse,
		Permissions: permissions,
	}, nil
}

// SwitchOrganizationContext switches the user's organization context.
func (s *authService) SwitchOrganizationContext(ctx context.Context, userID, roleID uuid.UUID, organizationID string) (*dto.SwitchOrganizationResult, error) {
	// Parse organization ID
	orgUUID, err := uuid.Parse(organizationID)
	if err != nil {
		return nil, apperror.NewClientError("Invalid organization ID format")
	}

	// Check if user has access to the specified organization
	hasAccess, err := s.authorizationService.CheckUserOrganizationAccess(ctx, userID, orgUUID)
	if err != nil {
		return nil, err
	}

	if !hasAccess {
		return nil, apperror.NewClientError("User does not have access to this organization")
	}

	// Generate new access token with the new context
	accessToken, _, err := s.generateTokens(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &dto.SwitchOrganizationResult{
		AccessToken:    accessToken,
		OrganizationID: organizationID,
	}, nil
}

// generateTokens generates both access and refresh tokens for a user.
func (s *authService) generateTokens(ctx context.Context, userID uuid.UUID) (accessToken, refreshToken string, err error) {
	// Generate access token (expires in 15 minutes)
	accessClaims := jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     time.Now().Add(time.Minute * 15).Unix(),
		"iat":     time.Now().Unix(),
		"type":    "access",
	}

	accessTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err = accessTokenObj.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", "", err
	}

	// Generate refresh token (expires in 7 days)
	refreshClaims := jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(),
		"iat":     time.Now().Unix(),
		"type":    "refresh",
	}

	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err = refreshTokenObj.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", "", err
	}

	// Store refresh token in Redis with expiration
	err = s.redis.Set(ctx, refreshToken, userID.String(), time.Hour*24*7).Err()
	if err != nil {
		log.Error().Err(err).Msg("Failed to store refresh token in Redis")
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// ValidateAccessToken validates an access token and returns the user ID.
func (s *authService) ValidateAccessToken(tokenString string) (uuid.UUID, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return uuid.Nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if tokenType, ok := claims["type"].(string); !ok || tokenType != "access" {
			return uuid.Nil, errors.New("invalid token type")
		}

		userIDStr, ok := claims["user_id"].(string)
		if !ok {
			return uuid.Nil, errors.New("invalid user ID in token")
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return uuid.Nil, err
		}

		return userID, nil
	}

	return uuid.Nil, errors.New("invalid token")
}
