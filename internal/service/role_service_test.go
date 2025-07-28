package service_test

import (
	"beresin-backend/internal/apperror"
	"beresin-backend/internal/service"
	"errors"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateRolePermissionsService(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRoleRepo := new(MockRoleRepository)
		mockAuthzService := new(MockAuthorizationService)
		roleService := service.NewRoleService(mockRoleRepo, mockAuthzService)

		roleID := uuid.New()
		permissions := []string{"users:read", "users:create"}

		mockRoleRepo.On("UpdateRolePermissions", roleID, permissions).Return(nil).Once()
		mockAuthzService.On("InvalidateRolePermissionsCache", roleID).Return(nil).Once()

		err := roleService.UpdateRolePermissions(roleID, permissions)

		assert.NoError(t, err)
		mockRoleRepo.AssertExpectations(t)
		mockAuthzService.AssertExpectations(t)
	})

	t.Run("Repository Fails", func(t *testing.T) {
		mockRoleRepo := new(MockRoleRepository)
		mockAuthzService := new(MockAuthorizationService)
		roleService := service.NewRoleService(mockRoleRepo, mockAuthzService)

		roleID := uuid.New()
		permissions := []string{"users:read"}
		repoError := errors.New("db error")

		mockRoleRepo.On("UpdateRolePermissions", roleID, permissions).Return(repoError).Once()

		err := roleService.UpdateRolePermissions(roleID, permissions)

		appErr, ok := err.(*apperror.AppError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusInternalServerError, appErr.Code)
		mockRoleRepo.AssertExpectations(t)
		// InvalidateRolePermissionsCache should not be called
		mockAuthzService.AssertNotCalled(t, "InvalidateRolePermissionsCache", mock.Anything)
	})

	t.Run("Cache Invalidation Fails", func(t *testing.T) {
		// This test verifies that even if cache invalidation fails, the operation
		// is considered successful from the caller's perspective, as the DB change is the critical part.
		// The error should be logged internally.
		mockRoleRepo := new(MockRoleRepository)
		mockAuthzService := new(MockAuthorizationService)
		roleService := service.NewRoleService(mockRoleRepo, mockAuthzService)

		roleID := uuid.New()
		permissions := []string{"users:read"}
		cacheError := errors.New("redis error")

		mockRoleRepo.On("UpdateRolePermissions", roleID, permissions).Return(nil).Once()
		mockAuthzService.On("InvalidateRolePermissionsCache", roleID).Return(cacheError).Once()

		err := roleService.UpdateRolePermissions(roleID, permissions)

		assert.NoError(t, err) // The function should still return nil
		mockRoleRepo.AssertExpectations(t)
		mockAuthzService.AssertExpectations(t)
	})
}
