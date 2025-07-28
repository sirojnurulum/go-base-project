package service

import (
	"beresin-backend/internal/dto"

	"github.com/google/uuid"
)

// UserService mendefinisikan kontrak untuk layanan manajemen pengguna.
type UserService interface {
	CreateUser(req dto.CreateUserRequest) (*dto.UserResponse, error)
	ListUsers(page, limit int, search string) (*dto.PagedUserResponse, error)
	GetUserByID(id uuid.UUID) (*dto.UserResponse, error)
	UpdateUser(id uuid.UUID, req dto.UpdateUserRequest) (*dto.UserResponse, error)
	DeleteUser(id uuid.UUID) error
}
