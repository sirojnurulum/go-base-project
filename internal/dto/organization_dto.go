package dto

import (
	"beresin-backend/internal/model"

	"github.com/google/uuid"
)

// CreateOrganizationRequest represents the request to create an organization
type CreateOrganizationRequest struct {
	Name        string                 `json:"name" validate:"required,min=3,max=255"`
	Code        string                 `json:"code" validate:"required,min=3,max=50,alphanum"`
	Type        model.OrganizationType `json:"type" validate:"required,oneof=platform company store"`
	Description string                 `json:"description" validate:"max=1000"`
	ParentID    *uuid.UUID             `json:"parent_id"`
}

// UpdateOrganizationRequest represents the request to update an organization
type UpdateOrganizationRequest struct {
	Name        string `json:"name" validate:"required,min=3,max=255"`
	Description string `json:"description" validate:"max=1000"`
}

// JoinOrganizationRequest represents the request to join an organization
type JoinOrganizationRequest struct {
	Code string `json:"code" validate:"required,min=3,max=50"`
}

// OrganizationResponse represents the response containing organization data
type OrganizationResponse struct {
	ID          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	Code        string                 `json:"code"`
	Type        model.OrganizationType `json:"type"`
	Description string                 `json:"description"`
	ParentID    *uuid.UUID             `json:"parent_id"`
	MemberCount int                    `json:"member_count"`
}

// OrganizationDetailResponse represents detailed organization response with members
type OrganizationDetailResponse struct {
	OrganizationResponse
	Parent   *OrganizationResponse  `json:"parent"`
	Children []OrganizationResponse `json:"children"`
	Members  []OrganizationMember   `json:"members"`
}

// OrganizationMember represents a member of an organization
type OrganizationMember struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Role     string    `json:"role"`
	JoinedAt string    `json:"joined_at"`
}
