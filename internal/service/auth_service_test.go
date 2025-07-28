package service_test

import (
	"beresin-backend/internal/apperror"
	"beresin-backend/internal/model"
	"beresin-backend/internal/service"
	"beresin-backend/internal/util"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const testJWTSecret = "test-secret-key-for-auth-service"

func TestLoginAuthService(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockAuthzService := new(MockAuthorizationService)
	redisClient, mockRedis := redismock.NewClientMock()
	authService := service.NewAuthService(mockRepo, mockRoleRepo, mockAuthzService, redisClient, testJWTSecret)

	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	t.Run("Success", func(t *testing.T) {
		roleID := uuid.New()
		user := &model.User{
			ID:       uuid.New(),
			Username: "testuser",
			Password: string(hashedPassword),
			RoleID:   &roleID,
			Role:     &model.Role{ID: roleID, Name: "user"},
		}
		permissions := []string{"dashboard:view"}
		predictableRefreshToken := "a-very-predictable-refresh-token"

		mockRepo.On("FindByUsernameWithRole", "testuser").Return(user, nil).Once()
		mockAuthzService.On("GetAndCachePermissionsForRole", roleID).Return(permissions, nil).Once()

		// Gunakan test hook untuk mengatur refresh token yang dapat diprediksi dan pastikan itu dibersihkan
		util.SetTestRefreshToken(predictableRefreshToken)
		defer util.SetTestRefreshToken("")

		// Sekarang kita bisa mengatur kunci yang tepat untuk mock redis
		mockRedis.ExpectSet(predictableRefreshToken, user.ID.String(), 7*24*time.Hour).SetVal("OK")

		result, err := authService.Login("testuser", password)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, user.Username, result.User.Username)
		assert.NotEmpty(t, result.AccessToken)
		assert.Equal(t, predictableRefreshToken, result.RefreshToken)
		assert.Equal(t, permissions, result.Permissions)
		assert.NoError(t, mockRedis.ExpectationsWereMet())
		mockAuthzService.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("User Not Found", func(t *testing.T) {
		mockRepo.On("FindByUsernameWithRole", "nonexistent").Return(nil, gorm.ErrRecordNotFound).Once()

		_, err := authService.Login("nonexistent", "password")

		// Periksa apakah error adalah tipe AppError dan memiliki kode yang benar
		appErr, ok := err.(*apperror.AppError)
		assert.True(t, ok, "Error should be of type AppError")
		assert.Equal(t, 401, appErr.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Wrong Password", func(t *testing.T) {
		user := &model.User{
			ID:       uuid.New(),
			Username: "testuser",
			Password: string(hashedPassword),
			Role:     &model.Role{ID: uuid.New(), Name: "user"},
		}
		mockRepo.On("FindByUsernameWithRole", "testuser").Return(user, nil).Once()

		_, err := authService.Login("testuser", "wrongpassword")

		appErr, ok := err.(*apperror.AppError)
		assert.True(t, ok, "Error should be of type AppError")
		assert.Equal(t, 401, appErr.Code)
		mockRepo.AssertExpectations(t)
	})
}

func TestRefreshTokenAuthService(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockAuthzService := new(MockAuthorizationService)
	redisClient, mockRedis := redismock.NewClientMock()
	authService := service.NewAuthService(mockRepo, mockRoleRepo, mockAuthzService, redisClient, testJWTSecret)

	userID := uuid.New()
	roleID := uuid.New()
	oldRefreshToken, _ := util.GenerateRefreshToken(userID, testJWTSecret)

	t.Run("Success", func(t *testing.T) {
		// Pastikan mock user memiliki data Role yang lengkap
		user := &model.User{ID: userID, RoleID: &roleID, Role: &model.Role{ID: roleID, Name: "user"}}
		predictableNewRefreshToken := "a-new-predictable-refresh-token"

		mockRedis.ExpectGet(oldRefreshToken).SetVal(userID.String())
		mockRedis.ExpectDel(oldRefreshToken).SetVal(1)
		// Gunakan FindByIDWithRole untuk mencocokkan implementasi service
		mockRepo.On("FindByIDWithRole", userID).Return(user, nil).Once()

		// Gunakan test hook untuk token *baru* dan pastikan itu dibersihkan
		util.SetTestRefreshToken(predictableNewRefreshToken)
		defer util.SetTestRefreshToken("")

		// Atur kunci yang tepat untuk token baru
		mockRedis.ExpectSet(predictableNewRefreshToken, userID.String(), 7*24*time.Hour).SetVal("OK")

		newAccess, newRefresh, err := authService.RefreshToken(oldRefreshToken)

		assert.NoError(t, err)
		assert.NotEmpty(t, newAccess)
		assert.Equal(t, predictableNewRefreshToken, newRefresh)
		assert.NotEqual(t, oldRefreshToken, newRefresh)
		assert.NoError(t, mockRedis.ExpectationsWereMet())
		mockRepo.AssertExpectations(t)
	})

	t.Run("Token Not Found in Redis", func(t *testing.T) {
		// Gunakan token yang valid secara sintaksis agar lolos dari parsing JWT.
		someToken, _ := util.GenerateRefreshToken(userID, testJWTSecret)
		mockRedis.ExpectGet(someToken).SetErr(redis.Nil)

		_, _, err := authService.RefreshToken(someToken)

		appErr, ok := err.(*apperror.AppError)
		assert.True(t, ok, "Error should be of type AppError")
		assert.Equal(t, 401, appErr.Code)
		assert.Contains(t, appErr.Message, "refresh token not found or already used")
		assert.NoError(t, mockRedis.ExpectationsWereMet())
	})

	t.Run("User Not Found in DB", func(t *testing.T) {
		// Gunakan token yang valid secara sintaksis agar lolos dari parsing JWT.
		someToken, _ := util.GenerateRefreshToken(userID, testJWTSecret)
		mockRedis.ExpectGet(someToken).SetVal(userID.String())
		mockRedis.ExpectDel(someToken).SetVal(1)
		// Gunakan FindByIDWithRole untuk mencocokkan implementasi service
		mockRepo.On("FindByIDWithRole", userID).Return(nil, gorm.ErrRecordNotFound).Once()

		_, _, err := authService.RefreshToken(someToken)

		appErr, ok := err.(*apperror.AppError)
		assert.True(t, ok, "Error should be of type AppError")
		assert.Equal(t, 401, appErr.Code) // Service harus mengembalikan Unauthorized, bukan NotFound
		assert.NoError(t, mockRedis.ExpectationsWereMet())
		mockRepo.AssertExpectations(t)
	})
}
func TestLogoutAuthService(t *testing.T) {
	redisClient, mockRedis := redismock.NewClientMock()
	authService := service.NewAuthService(nil, nil, nil, redisClient, testJWTSecret)

	t.Run("Success", func(t *testing.T) {
		// Gunakan token yang valid secara sintaksis agar lolos dari parsing JWT.
		token, _ := util.GenerateRefreshToken(uuid.New(), testJWTSecret)
		mockRedis.ExpectDel(token).SetVal(1)

		err := authService.Logout(token)

		assert.NoError(t, err)
		assert.NoError(t, mockRedis.ExpectationsWereMet())
	})
}
