package service

import (
	"beresin-backend/internal/dto"
	"context"

	"github.com/google/uuid"
)

// UserService mendefinisikan kontrak untuk layanan manajemen pengguna.
type UserService interface {
	CreateUser(req dto.CreateUserRequest) (*dto.UserResponse, error)
	ListUsers(page, limit int, search string) (*dto.PagedUserResponse, error)
	GetUserByID(id uuid.UUID) (*dto.UserResponse, error)
	UpdateUser(id uuid.UUID, req dto.UpdateUserRequest) (*dto.UserResponse, error)
	UpdateUserWithContext(ctx context.Context, currentUserID, targetUserID uuid.UUID, req dto.UpdateUserRequest) (*dto.UserResponse, error)
	DeleteUser(id uuid.UUID) error
}
