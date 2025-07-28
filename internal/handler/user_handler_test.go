package handler_test

import (
	"beresin-backend/internal/apperror"
	"beresin-backend/internal/constant"
	"beresin-backend/internal/dto"
	"beresin-backend/internal/handler"
	"beresin-backend/internal/validator"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserService is a mock for the UserService interface.
type MockUserService struct {
	mock.Mock
}

// Implement the UserService interface for the mock
func (m *MockUserService) CreateUser(req dto.CreateUserRequest) (*dto.UserResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.UserResponse), args.Error(1)
}

func (m *MockUserService) ListUsers(page, limit int, search string) (*dto.PagedUserResponse, error) {
	args := m.Called(page, limit, search)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PagedUserResponse), args.Error(1)
}

func (m *MockUserService) GetUserByID(id uuid.UUID) (*dto.UserResponse, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.UserResponse), args.Error(1)
}

func (m *MockUserService) UpdateUser(id uuid.UUID, req dto.UpdateUserRequest) (*dto.UserResponse, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.UserResponse), args.Error(1)
}

func (m *MockUserService) DeleteUser(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func TestCreateUser(t *testing.T) {
	// Setup
	mockService := new(MockUserService)
	userHandler := handler.NewUserHandler(mockService)
	e := echo.New()
	e.Validator = validator.NewCustomValidator() // Important for validating requests

	t.Run("Success", func(t *testing.T) {
		// Prepare request and expected response
		roleID := uuid.New()
		reqDTO := dto.CreateUserRequest{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "password123",
			RoleID:   roleID,
		}
		jsonReq, _ := json.Marshal(reqDTO)

		resDTO := &dto.UserResponse{
			ID:       uuid.New(),
			Username: "testuser",
			Email:    "test@example.com",
			Role:     "user",
		}

		// Setup mock
		mockService.On("CreateUser", reqDTO).Return(resDTO, nil).Once()

		// Perform request
		req := httptest.NewRequest(http.MethodPost, "/admin/users", strings.NewReader(string(jsonReq)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Assertions
		if assert.NoError(t, userHandler.CreateUser(c)) {
			assert.Equal(t, http.StatusCreated, rec.Code)

			var responseUser dto.UserResponse
			err := json.Unmarshal(rec.Body.Bytes(), &responseUser)
			assert.NoError(t, err)
			assert.Equal(t, resDTO.Username, responseUser.Username)
		}

		// Verify that the mock was called
		mockService.AssertExpectations(t)
	})

	t.Run("Conflict - Username Exists", func(t *testing.T) {
		// Prepare request
		roleID := uuid.New()
		reqDTO := dto.CreateUserRequest{
			Username: "existinguser",
			Email:    "new@example.com",
			Password: "password123",
			RoleID:   roleID,
		}
		jsonReq, _ := json.Marshal(reqDTO)
		expectedErr := apperror.NewConflictError("username already exists")

		// Setup mock
		mockService.On("CreateUser", reqDTO).Return(nil, expectedErr).Once()

		// Perform request
		req := httptest.NewRequest(http.MethodPost, "/admin/users", strings.NewReader(string(jsonReq)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Assertions
		err := userHandler.CreateUser(c)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)

		mockService.AssertExpectations(t)
	})
}

func TestListUsers(t *testing.T) {
	mockService := new(MockUserService)
	userHandler := handler.NewUserHandler(mockService)
	e := echo.New()

	t.Run("Success", func(t *testing.T) {
		pagedResponse := &dto.PagedUserResponse{
			Users: []dto.UserResponse{
				{ID: uuid.New(), Username: "user1", Email: "user1@example.com", Role: "user"},
			},
			Page:       1,
			Limit:      10,
			Total:      1,
			TotalPages: 1,
		}

		mockService.On("ListUsers", 1, 10, "search_term").Return(pagedResponse, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/admin/users?page=1&limit=10&search=search_term", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(t, userHandler.ListUsers(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			var responseBody dto.PagedUserResponse
			err := json.Unmarshal(rec.Body.Bytes(), &responseBody)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(responseBody.Users))
			assert.Equal(t, pagedResponse.Total, responseBody.Total)
		}
		mockService.AssertExpectations(t)
	})
}

func TestGetUserByID(t *testing.T) {
	// Setup
	mockService := new(MockUserService)
	userHandler := handler.NewUserHandler(mockService)
	e := echo.New()

	t.Run("Success", func(t *testing.T) {
		userID := uuid.New()
		resDTO := &dto.UserResponse{
			ID:       userID,
			Username: "testuser",
			Email:    "test@example.com",
			Role:     "user",
		}

		mockService.On("GetUserByID", userID).Return(resDTO, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/", nil) // Path doesn't matter here
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(userID.String())

		if assert.NoError(t, userHandler.GetUserByID(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			var responseUser dto.UserResponse
			err := json.Unmarshal(rec.Body.Bytes(), &responseUser)
			assert.NoError(t, err)
			assert.Equal(t, resDTO.ID, responseUser.ID)
		}
		mockService.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		userID := uuid.New()
		expectedErr := apperror.NewNotFoundError("user")
		mockService.On("GetUserByID", userID).Return(nil, expectedErr).Once()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(userID.String())

		err := userHandler.GetUserByID(c)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockService.AssertExpectations(t)
	})

	t.Run("Invalid ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("not-a-uuid")

		err := userHandler.GetUserByID(c)
		assert.Error(t, err)
		// We check the type and code because the underlying error from uuid.Parse is complex.
		appErr, ok := err.(*apperror.AppError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, appErr.Code)
		assert.Equal(t, constant.ErrMsgInvalidUserID, appErr.Message)
	})
}

func TestUpdateUser(t *testing.T) {
	mockService := new(MockUserService)
	userHandler := handler.NewUserHandler(mockService)
	e := echo.New()
	e.Validator = validator.NewCustomValidator()

	t.Run("Success", func(t *testing.T) {
		userID := uuid.New()
		reqDTO := dto.UpdateUserRequest{
			Username: "updateduser",
			Email:    "updated@example.com",
		}
		jsonReq, _ := json.Marshal(reqDTO)

		resDTO := &dto.UserResponse{
			ID:       userID,
			Username: "updateduser",
			Email:    "updated@example.com",
			Role:     "user",
		}

		mockService.On("UpdateUser", userID, reqDTO).Return(resDTO, nil).Once()

		req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(string(jsonReq)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(userID.String())

		if assert.NoError(t, userHandler.UpdateUser(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			var responseUser dto.UserResponse
			err := json.Unmarshal(rec.Body.Bytes(), &responseUser)
			assert.NoError(t, err)
			assert.Equal(t, resDTO.Username, responseUser.Username)
		}
		mockService.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		userID := uuid.New()
		reqDTO := dto.UpdateUserRequest{Username: "any"}
		jsonReq, _ := json.Marshal(reqDTO)
		expectedErr := apperror.NewNotFoundError("user")

		mockService.On("UpdateUser", userID, reqDTO).Return(nil, expectedErr).Once()

		req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(string(jsonReq)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(userID.String())

		err := userHandler.UpdateUser(c)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockService.AssertExpectations(t)
	})
}

func TestDeleteUser(t *testing.T) {
	mockService := new(MockUserService)
	userHandler := handler.NewUserHandler(mockService)
	e := echo.New()

	t.Run("Success", func(t *testing.T) {
		userID := uuid.New()
		mockService.On("DeleteUser", userID).Return(nil).Once()

		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(userID.String())

		if assert.NoError(t, userHandler.DeleteUser(c)) {
			assert.Equal(t, http.StatusNoContent, rec.Code)
		}
		mockService.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		userID := uuid.New()
		expectedErr := apperror.NewNotFoundError("user")
		mockService.On("DeleteUser", userID).Return(expectedErr).Once()

		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(userID.String())

		err := userHandler.DeleteUser(c)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockService.AssertExpectations(t)
	})
}
