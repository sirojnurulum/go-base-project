package service_test

import (
	"beresin-backend/internal/dto"
	"beresin-backend/internal/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock for the UserRepository interface.
// It is shared across tests in the service_test package.
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *model.User) error {
	args := m.Called(user)
	if user != nil {
		user.ID = uuid.New()
	}
	return args.Error(0)
}

func (m *MockUserRepository) List(offset, limit int, search string) ([]model.User, error) {
	args := m.Called(offset, limit, search)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.User), args.Error(1)
}

func (m *MockUserRepository) Count(search string) (int64, error) {
	args := m.Called(search)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) FindByID(id uuid.UUID) (*model.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) FindByUsername(username string) (*model.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) FindByEmail(email string) (*model.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) FindByGoogleID(googleID string) (*model.User, error) {
	args := m.Called(googleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) FindByIDWithRole(id uuid.UUID) (*model.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) FindByUsernameWithRole(username string) (*model.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) Update(user *model.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

// MockAuthService is a mock for the AuthService interface.
type MockAuthService struct {
	mock.Mock
}

// Implement the AuthService interface for the mock
func (m *MockAuthService) Login(username, password string) (*dto.LoginResult, error) {
	args := m.Called(username, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.LoginResult), args.Error(1)
}

func (m *MockAuthService) RefreshToken(tokenString string) (string, string, error) {
	args := m.Called(tokenString)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockAuthService) LoginWithGoogle(userInfo dto.GoogleUserInfo) (*dto.LoginResult, error) {
	args := m.Called(userInfo)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.LoginResult), args.Error(1)
}

func (m *MockAuthService) Logout(refreshToken string) error {
	args := m.Called(refreshToken)
	return args.Error(0)
}

// MockRoleRepository is a mock for the RoleRepository interface.
type MockRoleRepository struct {
	mock.Mock
}

func (m *MockRoleRepository) FindRoleByName(name string) (*model.Role, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Role), args.Error(1)
}

// MockAuthorizationService is a mock for the AuthorizationService interface.
type MockAuthorizationService struct {
	mock.Mock
}

func (m *MockAuthorizationService) CheckPermission(roleID uuid.UUID, requiredPermission string) (bool, error) {
	args := m.Called(roleID, requiredPermission)
	return args.Bool(0), args.Error(1)
}

func (m *MockAuthorizationService) InvalidateRolePermissionsCache(roleID uuid.UUID) error {
	args := m.Called(roleID)
	return args.Error(0)
}

func (m *MockAuthorizationService) GetAndCachePermissionsForRole(roleID uuid.UUID) ([]string, error) {
	args := m.Called(roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockRoleRepository) FindPermissionsByRoleID(roleID uuid.UUID) ([]string, error) {
	args := m.Called(roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockRoleRepository) UpdateRolePermissions(roleID uuid.UUID, permissionNames []string) error {
	args := m.Called(roleID, permissionNames)
	return args.Error(0)
}

func (m *MockRoleRepository) Create(role *model.Role) (*model.Role, error) {
	args := m.Called(role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Role), args.Error(1)
}
