package dto

import "github.com/google/uuid"

// UserResponse adalah DTO untuk data publik seorang user.
// Ini menyembunyikan detail implementasi seperti password hash.
type UserResponse struct {
	ID             uuid.UUID  `json:"id" example:"a1b2c3d4-e5f6-7890-1234-567890abcdef"`
	Username       string     `json:"username" example:"johndoe"`
	Email          string     `json:"email" example:"john.doe@example.com"`
	Role           string     `json:"role" example:"user"`
	RoleID         *uuid.UUID `json:"role_id,omitempty" example:"b1c2d3e4-f5g6-7890-1234-567890abcdef"`
	OrganizationID *uuid.UUID `json:"organization_id,omitempty" example:"c1d2e3f4-g5h6-7890-1234-567890abcdef"`
	AvatarURL      string     `json:"avatar_url" example:"https://example.com/avatar.png"`
	AuthProvider   string     `json:"auth_provider" example:"local"` // Authentication method
}

// CreateUserRequest adalah DTO untuk membuat user baru.
type CreateUserRequest struct {
	Username     string    `json:"username" validate:"required,min=3" example:"newuser"`
	Email        string    `json:"email" validate:"required,email" example:"new.user@example.com"`
	Password     string    `json:"password" validate:"omitempty,min=8" example:"strongpassword123"` // Optional for OAuth users
	RoleID       uuid.UUID `json:"role_id" validate:"required" example:"b1c2d3e4-f5g6-7890-1234-567890abcdef"`
	AuthProvider string    `json:"auth_provider" validate:"omitempty,oneof=local google" example:"local"` // Authentication method
	GoogleID     *string   `json:"google_id" validate:"omitempty" example:"110539596352895004866"`        // Google OAuth ID
}

// UpdateUserRequest adalah DTO untuk memperbarui user yang ada.
type UpdateUserRequest struct {
	Username string     `json:"username" validate:"omitempty,min=3" example:"updateduser"`
	Email    string     `json:"email" validate:"omitempty,email" example:"updated.user@example.com"`
	RoleID   *uuid.UUID `json:"role_id" validate:"omitempty" example:"c1d2e3f4-g5h6-7890-1234-567890abcdef"`
}

// UpdateUserResponse adalah DTO untuk response update user dengan pesan informatif.
type UpdateUserResponse struct {
	User    UserResponse `json:"user"`
	Message string       `json:"message" example:"User updated successfully"`
}

// PagedUserResponse adalah DTO untuk response daftar user dengan metadata paginasi.
type PagedUserResponse struct {
	Users      []UserResponse `json:"users"`
	Page       int            `json:"page" example:"1"`
	Limit      int            `json:"limit" example:"10"`
	Total      int64          `json:"total" example:"100"`
	TotalPages int            `json:"total_pages" example:"10"`
}

// User-Organization Management DTOs

// AssignUserToOrganizationRequest adalah DTO untuk assign user ke organization.
type AssignUserToOrganizationRequest struct {
	UserID         uuid.UUID  `json:"user_id" validate:"required" example:"a1b2c3d4-e5f6-7890-1234-567890abcdef"`
	OrganizationID uuid.UUID  `json:"organization_id" validate:"required" example:"b1c2d3e4-f5g6-7890-1234-567890abcdef"`
	RoleID         *uuid.UUID `json:"role_id" example:"c1d2e3f4-g5h6-7890-1234-567890abcdef"`
	IsActive       bool       `json:"is_active" example:"true"`
}

// UpdateUserOrganizationRequest adalah DTO untuk update user organization assignment.
type UpdateUserOrganizationRequest struct {
	RoleID   *uuid.UUID `json:"role_id" example:"c1d2e3f4-g5h6-7890-1234-567890abcdef"`
	IsActive bool       `json:"is_active" example:"true"`
}

// BulkAssignUsersToOrganizationRequest adalah DTO untuk bulk assign users ke organization.
type BulkAssignUsersToOrganizationRequest struct {
	UserIDs        []uuid.UUID `json:"user_ids" validate:"required,min=1" example:"[\"a1b2c3d4-e5f6-7890-1234-567890abcdef\", \"b2c3d4e5-f6g7-8901-2345-678901bcdefg\"]"`
	OrganizationID uuid.UUID   `json:"organization_id" validate:"required" example:"c1d2e3f4-g5h6-7890-1234-567890abcdef"`
	RoleID         *uuid.UUID  `json:"role_id" example:"d1e2f3g4-h5i6-7890-1234-567890abcdef"`
}

// BulkAssignResponse adalah DTO untuk response bulk assign operations.
type BulkAssignResponse struct {
	SuccessCount int                        `json:"success_count" example:"5"`
	FailureCount int                        `json:"failure_count" example:"1"`
	Assignments  []UserOrganizationResponse `json:"assignments,omitempty"`
	Errors       []BulkAssignError          `json:"errors,omitempty"`
	Message      string                     `json:"message" example:"5 users assigned successfully, 1 failed"`
}

// BulkAssignError adalah DTO untuk error dalam bulk operations.
type BulkAssignError struct {
	UserID uuid.UUID `json:"user_id" example:"d3e4f5g6-h7i8-9012-3456-789012cdefgh"`
	Error  string    `json:"error" example:"User already assigned to organization"`
}

// UserOrganizationHistoryResponse adalah DTO untuk organization assignment history.
type UserOrganizationHistoryResponse struct {
	ID             uuid.UUID `json:"id" example:"e4f5g6h7-i8j9-0123-4567-890123defghi"`
	UserID         uuid.UUID `json:"user_id" example:"a1b2c3d4-e5f6-7890-1234-567890abcdef"`
	OrganizationID uuid.UUID `json:"organization_id" example:"b1c2d3e4-f5g6-7890-1234-567890abcdef"`
	Action         string    `json:"action" example:"assigned"`
	PreviousRole   string    `json:"previous_role,omitempty" example:"member"`
	NewRole        string    `json:"new_role,omitempty" example:"admin"`
	ActionBy       uuid.UUID `json:"action_by" example:"f5g6h7i8-j9k0-1234-5678-901234efghij"`
	ActionAt       string    `json:"action_at" example:"2024-01-15T10:30:00Z"`
	Reason         string    `json:"reason,omitempty" example:"Promoted to admin role"`
}

// PagedUserOrganizationResponse adalah DTO untuk paginated user-organization relationships.
type PagedUserOrganizationResponse struct {
	UserOrganizations []UserOrganizationResponse `json:"user_organizations"`
	Page              int                        `json:"page" example:"1"`
	Limit             int                        `json:"limit" example:"10"`
	Total             int64                      `json:"total" example:"25"`
	TotalPages        int                        `json:"total_pages" example:"3"`
}

// PagedUserOrganizationHistoryResponse adalah DTO untuk paginated user-organization history.
type PagedUserOrganizationHistoryResponse struct {
	History    []UserOrganizationHistoryResponse `json:"history"`
	Page       int                               `json:"page" example:"1"`
	Limit      int                               `json:"limit" example:"10"`
	Total      int64                             `json:"total" example:"25"`
	TotalPages int                               `json:"total_pages" example:"3"`
}

// LogUserOrganizationActionRequest adalah DTO untuk manually logging user organization actions.
type LogUserOrganizationActionRequest struct {
	UserID         uuid.UUID `json:"user_id" validate:"required" example:"a1b2c3d4-e5f6-7890-1234-567890abcdef"`
	OrganizationID uuid.UUID `json:"organization_id" validate:"required" example:"b1c2d3e4-f5g6-7890-1234-567890abcdef"`
	Action         string    `json:"action" validate:"required,oneof=assigned removed role_updated status_changed" example:"assigned"`
	PreviousRole   string    `json:"previous_role,omitempty" example:"user"`
	NewRole        string    `json:"new_role,omitempty" example:"admin"`
	PreviousStatus *bool     `json:"previous_status,omitempty" example:"false"`
	NewStatus      *bool     `json:"new_status,omitempty" example:"true"`
	Reason         string    `json:"reason,omitempty" example:"Manual role update by administrator"`
}
