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

// MockRoleService is a mock for the RoleService interface.
type MockRoleService struct {
	mock.Mock
}

func (m *MockRoleService) CreateRole(req dto.CreateRoleRequest) (*dto.RoleResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.RoleResponse), args.Error(1)
}

func (m *MockRoleService) UpdateRolePermissions(roleID uuid.UUID, permissionNames []string) error {
	args := m.Called(roleID, permissionNames)
	return args.Error(0)
}

func TestCreateRole(t *testing.T) {
	mockService := new(MockRoleService)
	roleHandler := handler.NewRoleHandler(mockService)
	e := echo.New()
	e.Validator = validator.NewCustomValidator()

	t.Run("Success", func(t *testing.T) {
		reqDTO := dto.CreateRoleRequest{Name: "new-role", Description: "A new role"}
		jsonReq, _ := json.Marshal(reqDTO)
		resDTO := &dto.RoleResponse{ID: uuid.New(), Name: "new-role", Description: "A new role"}

		mockService.On("CreateRole", reqDTO).Return(resDTO, nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/admin/roles", strings.NewReader(string(jsonReq)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(t, roleHandler.CreateRole(c)) {
			assert.Equal(t, http.StatusCreated, rec.Code)
			var responseBody dto.RoleResponse
			json.Unmarshal(rec.Body.Bytes(), &responseBody)
			assert.Equal(t, resDTO.Name, responseBody.Name)
		}
		mockService.AssertExpectations(t)
	})

	t.Run("Conflict", func(t *testing.T) {
		reqDTO := dto.CreateRoleRequest{Name: "existing-role"}
		jsonReq, _ := json.Marshal(reqDTO)
		expectedErr := apperror.NewConflictError("role with name 'existing-role' already exists")

		mockService.On("CreateRole", reqDTO).Return(nil, expectedErr).Once()

		req := httptest.NewRequest(http.MethodPost, "/admin/roles", strings.NewReader(string(jsonReq)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := roleHandler.CreateRole(c)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockService.AssertExpectations(t)
	})
}

func TestUpdateRolePermissions(t *testing.T) {
	mockService := new(MockRoleService)
	roleHandler := handler.NewRoleHandler(mockService)
	e := echo.New()
	e.Validator = validator.NewCustomValidator()

	t.Run("Success", func(t *testing.T) {
		roleID := uuid.New()
		reqDTO := dto.UpdateRolePermissionsRequest{PermissionNames: []string{"users:read"}}
		jsonReq, _ := json.Marshal(reqDTO)

		mockService.On("UpdateRolePermissions", roleID, reqDTO.PermissionNames).Return(nil).Once()

		req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(string(jsonReq)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(roleID.String())

		if assert.NoError(t, roleHandler.UpdateRolePermissions(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			var responseBody map[string]string
			json.Unmarshal(rec.Body.Bytes(), &responseBody)
			assert.Equal(t, constant.MsgRolePermsUpdated, responseBody["message"])
		}
		mockService.AssertExpectations(t)
	})

	t.Run("Role Not Found", func(t *testing.T) {
		roleID := uuid.New()
		reqDTO := dto.UpdateRolePermissionsRequest{PermissionNames: []string{"users:read"}}
		jsonReq, _ := json.Marshal(reqDTO)
		expectedErr := apperror.NewNotFoundError("role")

		mockService.On("UpdateRolePermissions", roleID, reqDTO.PermissionNames).Return(expectedErr).Once()

		req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(string(jsonReq)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(roleID.String())

		err := roleHandler.UpdateRolePermissions(c)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockService.AssertExpectations(t)
	})
}
