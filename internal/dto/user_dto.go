package dto

import "github.com/google/uuid"

// UserResponse adalah DTO untuk data publik seorang user.
// Ini menyembunyikan detail implementasi seperti password hash.
type UserResponse struct {
	ID        uuid.UUID `json:"id" example:"a1b2c3d4-e5f6-7890-1234-567890abcdef"`
	Username  string    `json:"username" example:"johndoe"`
	Email     string    `json:"email" example:"john.doe@example.com"`
	Role      string    `json:"role" example:"user"`
	AvatarURL string    `json:"avatar_url" example:"https://example.com/avatar.png"`
}

// CreateUserRequest adalah DTO untuk membuat user baru.
type CreateUserRequest struct {
	Username string    `json:"username" validate:"required,min=3" example:"newuser"`
	Email    string    `json:"email" validate:"required,email" example:"new.user@example.com"`
	Password string    `json:"password" validate:"required,min=8" example:"strongpassword123"`
	RoleID   uuid.UUID `json:"role_id" validate:"required" example:"b1c2d3e4-f5g6-7890-1234-567890abcdef"`
}

// UpdateUserRequest adalah DTO untuk memperbarui user yang ada.
type UpdateUserRequest struct {
	Username string    `json:"username" validate:"omitempty,min=3" example:"updateduser"`
	Email    string    `json:"email" validate:"omitempty,email" example:"updated.user@example.com"`
	RoleID   uuid.UUID `json:"role_id" validate:"omitempty" example:"c1d2e3f4-g5h6-7890-1234-567890abcdef"`
}

// PagedUserResponse adalah DTO untuk response daftar user dengan metadata paginasi.
type PagedUserResponse struct {
	Users      []UserResponse `json:"users"`
	Page       int            `json:"page" example:"1"`
	Limit      int            `json:"limit" example:"10"`
	Total      int64          `json:"total" example:"100"`
	TotalPages int            `json:"total_pages" example:"10"`
}
