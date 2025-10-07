package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateOrganizationRequest represents the request payload for creating an organization
type CreateOrganizationRequest struct {
	Name                 string     `json:"name" validate:"required,min=3,max=100"`
	OrganizationType     string     `json:"type" validate:"required,oneof=platform holding company store"`
	ParentOrganizationID *uuid.UUID `json:"parentOrganizationId,omitempty"`
	Description          string     `json:"description,omitempty" validate:"max=500"`
}

// UpdateOrganizationRequest represents the request payload for updating an organization
type UpdateOrganizationRequest struct {
	Name        string `json:"name,omitempty" validate:"omitempty,min=3,max=100"`
	Description string `json:"description,omitempty" validate:"max=500"`
	IsActive    *bool  `json:"is_active,omitempty"`
}

// OrganizationResponse represents the response payload for organization data
type OrganizationResponse struct {
	ID                   uuid.UUID              `json:"id"`
	Name                 string                 `json:"name"`
	Code                 string                 `json:"code"`
	OrganizationType     string                 `json:"organization_type"`
	ParentOrganizationID *uuid.UUID             `json:"parent_organization_id,omitempty"`
	ParentOrganization   *OrganizationResponse  `json:"parent_organization,omitempty"`
	Description          string                 `json:"description,omitempty"`
	CreatedBy            uuid.UUID              `json:"created_by"`
	Creator              *UserResponse          `json:"creator,omitempty"`
	IsActive             bool                   `json:"is_active"`
	CreatedAt            time.Time              `json:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at"`
	ChildOrganizations   []OrganizationResponse `json:"child_organizations,omitempty"`
	MemberCount          int                    `json:"member_count,omitempty"`
}

// JoinOrganizationRequest represents the request to join an existing organization
type JoinOrganizationRequest struct {
	OrganizationCode string `json:"organization_code" validate:"required,len=8"`
}

// CreateRoleWithOrganizationRequest extends CreateRoleApprovalRequest with organization context
type CreateRoleWithOrganizationRequest struct {
	CreateRoleApprovalRequest
	// For joining existing organization
	OrganizationCode string `json:"organization_code,omitempty" validate:"omitempty,len=8"`
	// For creating new organization
	IsNewOrganization         bool   `json:"is_new_organization"`
	RequestedOrganizationName string `json:"requested_organization_name,omitempty" validate:"required_if=IsNewOrganization true,omitempty,min=3,max=100"`
	RequestedOrganizationType string `json:"requested_organization_type,omitempty" validate:"required_if=IsNewOrganization true,omitempty,oneof=platform holding company store"`
	OrganizationDescription   string `json:"organization_description,omitempty" validate:"max=500"`
}

// RoleApprovalWithOrganizationResponse extends RoleApprovalResponse with organization context
type RoleApprovalWithOrganizationResponse struct {
	RoleApprovalResponse
	Organization              *OrganizationResponse `json:"organization,omitempty"`
	OrganizationCode          string                `json:"organization_code,omitempty"`
	IsNewOrganization         bool                  `json:"is_new_organization"`
	RequestedOrganizationName string                `json:"requested_organization_name,omitempty"`
	RequestedOrganizationType string                `json:"requested_organization_type,omitempty"`
}

// UserOrganizationResponse represents user-organization relationship
type UserOrganizationResponse struct {
	UserID         uuid.UUID            `json:"user_id"`
	OrganizationID uuid.UUID            `json:"organization_id"`
	Organization   OrganizationResponse `json:"organization"`
	RoleID         *uuid.UUID           `json:"role_id,omitempty"`
	Role           *RoleResponse        `json:"role,omitempty"`
	JoinedAt       time.Time            `json:"joined_at"`
	IsActive       bool                 `json:"is_active"`
}

// ListOrganizationsRequest represents query parameters for listing organizations
type ListOrganizationsRequest struct {
	OrganizationType     string     `query:"type" validate:"omitempty,oneof=platform holding company store"`
	ParentOrganizationID *uuid.UUID `query:"parent_id"`
	IsActive             *bool      `query:"active"`
	Search               string     `query:"search"`
	Page                 int        `query:"page" validate:"min=1"`
	Limit                int        `query:"limit" validate:"min=1,max=100"`
}

// OrganizationStatsResponse represents statistics for an organization
type OrganizationStatsResponse struct {
	TotalMembers      int `json:"total_members"`
	ActiveMembers     int `json:"active_members"`
	DirectChildren    int `json:"direct_children"`
	TotalDescendants  int `json:"total_descendants"`
	OrganizationLevel int `json:"organization_level"`
}

// HierarchyStatsResponse represents overall hierarchy statistics
type HierarchyStatsResponse struct {
	TotalOrganizations  int `json:"total_organizations"`
	ActiveOrganizations int `json:"active_organizations"`
	HoldingCount        int `json:"holding_count"`
	CompanyCount        int `json:"company_count"`
	StoreCount          int `json:"store_count"`
	MaxDepth            int `json:"max_depth"`
}

// CreateCompleteStructureRequest represents request for creating complete organization structure
type CreateCompleteStructureRequest struct {
	// User information
	UserID   uuid.UUID `json:"user_id" validate:"required"`
	UserRole string    `json:"user_role" validate:"required,oneof=holding_owner company_manager store_manager"`

	// Organization structure
	HoldingName string `json:"holding_name" validate:"required,min=3,max=100"`
	CompanyName string `json:"company_name" validate:"required,min=3,max=100"`
	StoreName   string `json:"store_name" validate:"required,min=3,max=100"`

	// Optional descriptions
	HoldingDescription string `json:"holding_description,omitempty"`
	CompanyDescription string `json:"company_description,omitempty"`
	StoreDescription   string `json:"store_description,omitempty"`
}

// CompleteOrganizationStructureResponse represents response after creating complete structure
type CompleteOrganizationStructureResponse struct {
	Message         string                     `json:"message"`
	HoldingID       uuid.UUID                  `json:"holding_id"`
	CompanyID       uuid.UUID                  `json:"company_id"`
	StoreID         uuid.UUID                  `json:"store_id"`
	UserAssignedTo  uuid.UUID                  `json:"user_assigned_to"`
	UserRole        string                     `json:"user_role"`
	Organizations   []OrganizationResponse     `json:"organizations"`
	UserMemberships []UserOrganizationResponse `json:"user_memberships"`
}

type OrganizationStatisticsResponse struct {
	PlatformCount int64 `json:"platform_count"`
	HoldingCount  int64 `json:"holding_count"`
	CompanyCount  int64 `json:"company_count"`
	StoreCount    int64 `json:"store_count"`
}

// PaginationInfo represents pagination metadata
type PaginationInfo struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// PaginatedOrganizationsResponse represents paginated organization list response
type PaginatedOrganizationsResponse struct {
	Organizations []OrganizationResponse `json:"organizations"`
	Pagination    PaginationInfo         `json:"pagination"`
}
