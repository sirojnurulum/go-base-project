package service_test

import (
	"beresin-backend/internal/apperror"
	"beresin-backend/internal/dto"
	"beresin-backend/internal/model"
	"beresin-backend/internal/service"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestCreateUserService(t *testing.T) {
	mockRepo := new(MockUserRepository)
	userService := service.NewUserService(mockRepo)

	t.Run("Success", func(t *testing.T) {
		req := dto.CreateUserRequest{
			Username: "newuser",
			Email:    "new@example.com",
			Password: "password123",
			RoleID:   uuid.New(),
		}

		mockRepo.On("FindByUsername", req.Username).Return(nil, gorm.ErrRecordNotFound).Once()
		mockRepo.On("FindByEmail", req.Email).Return(nil, gorm.ErrRecordNotFound).Once()
		mockRepo.On("Create", mock.AnythingOfType("*model.User")).Return(nil).Once()
		mockRepo.On("FindByIDWithRole", mock.AnythingOfType("uuid.UUID")).Return(&model.User{
			ID:       uuid.New(),
			Username: req.Username,
			Email:    req.Email,
			Role:     &model.Role{Name: "user"},
		}, nil).Once()

		res, err := userService.CreateUser(req)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, req.Username, res.Username)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Username Exists", func(t *testing.T) {
		req := dto.CreateUserRequest{Username: "existinguser"}
		mockRepo.On("FindByUsername", req.Username).Return(&model.User{}, nil).Once()

		_, err := userService.CreateUser(req)

		appErr, ok := err.(*apperror.AppError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusConflict, appErr.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Email Exists", func(t *testing.T) {
		req := dto.CreateUserRequest{Username: "newuser", Email: "existing@example.com"}
		mockRepo.On("FindByUsername", req.Username).Return(nil, gorm.ErrRecordNotFound).Once()
		mockRepo.On("FindByEmail", req.Email).Return(&model.User{}, nil).Once()

		_, err := userService.CreateUser(req)

		appErr, ok := err.(*apperror.AppError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusConflict, appErr.Code)
		mockRepo.AssertExpectations(t)
	})
}

func TestListUsersService(t *testing.T) {
	mockRepo := new(MockUserRepository)
	userService := service.NewUserService(mockRepo)

	t.Run("Success", func(t *testing.T) {
		users := []model.User{
			{ID: uuid.New(), Username: "user1", Email: "user1@example.com", Role: &model.Role{Name: "user"}},
		}
		mockRepo.On("List", 0, 10, "").Return(users, nil).Once()
		mockRepo.On("Count", "").Return(int64(1), nil).Once()

		res, err := userService.ListUsers(1, 10, "")

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, 1, len(res.Users))
		assert.Equal(t, int64(1), res.Total)
		mockRepo.AssertExpectations(t)
	})
}

func TestGetUserByIDService(t *testing.T) {
	mockRepo := new(MockUserRepository)
	userService := service.NewUserService(mockRepo)

	t.Run("Success", func(t *testing.T) {
		userID := uuid.New()
		dbUser := &model.User{ID: userID, Username: "testuser", Role: &model.Role{Name: "admin"}}
		mockRepo.On("FindByIDWithRole", userID).Return(dbUser, nil).Once()

		res, err := userService.GetUserByID(userID)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, userID, res.ID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		userID := uuid.New()
		mockRepo.On("FindByIDWithRole", userID).Return(nil, gorm.ErrRecordNotFound).Once()

		_, err := userService.GetUserByID(userID)

		appErr, ok := err.(*apperror.AppError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusNotFound, appErr.Code)
		mockRepo.AssertExpectations(t)
	})
}

func TestUpdateUserService(t *testing.T) {
	mockRepo := new(MockUserRepository)
	userService := service.NewUserService(mockRepo)

	t.Run("Success", func(t *testing.T) {
		userID := uuid.New()
		req := dto.UpdateUserRequest{Username: "updateduser"}
		originalUser := &model.User{ID: userID, Username: "original"}
		updatedUser := &model.User{ID: userID, Username: "updateduser", Role: &model.Role{Name: "user"}}

		mockRepo.On("FindByIDWithRole", userID).Return(originalUser, nil).Once()
		mockRepo.On("Update", mock.AnythingOfType("*model.User")).Return(nil).Once()
		mockRepo.On("FindByIDWithRole", userID).Return(updatedUser, nil).Once()

		res, err := userService.UpdateUser(userID, req)

		assert.NoError(t, err)
		assert.Equal(t, "updateduser", res.Username)
		mockRepo.AssertExpectations(t)
	})
}

func TestDeleteUserService(t *testing.T) {
	mockRepo := new(MockUserRepository)
	userService := service.NewUserService(mockRepo)

	t.Run("Success", func(t *testing.T) {
		userID := uuid.New()
		mockRepo.On("FindByID", userID).Return(&model.User{ID: userID}, nil).Once()
		mockRepo.On("Delete", userID).Return(nil).Once()

		err := userService.DeleteUser(userID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		userID := uuid.New()
		mockRepo.On("FindByID", userID).Return(nil, gorm.ErrRecordNotFound).Once()

		err := userService.DeleteUser(userID)

		appErr, ok := err.(*apperror.AppError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusNotFound, appErr.Code)
		mockRepo.AssertExpectations(t)
	})
}
