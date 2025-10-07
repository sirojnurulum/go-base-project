package util

import (
	"go-base-project/internal/dto"
	"go-base-project/internal/model"
)

// MapUserToResponse converts a User model to UserResponse DTO
// This is a centralized mapping function to avoid duplication across services
func MapUserToResponse(user *model.User) *dto.UserResponse {
	if user == nil {
		return nil
	}

	response := &dto.UserResponse{
		ID:           user.ID,
		Username:     user.Username,
		Email:        user.Email,
		AvatarURL:    user.AvatarURL,
		RoleID:       user.RoleID,
		AuthProvider: user.AuthProvider,
	}

	// Get first active organization from many-to-many relationship
	// For now, we assume user has only one active organization
	if len(user.Organizations) > 0 {
		response.OrganizationID = &user.Organizations[0].ID
	}

	if user.Role != nil {
		response.Role = user.Role.Name
	}

	return response
}

// MapRoleToResponse converts a Role model to RoleResponse DTO
func MapRoleToResponse(role *model.Role) *dto.RoleResponse {
	if role == nil {
		return nil
	}

	// Extract permissions if loaded
	var permissions []string
	if role.Permissions != nil {
		permissions = make([]string, len(role.Permissions))
		for i, permission := range role.Permissions {
			permissions[i] = permission.Name
		}
	}

	// Extract organization types if loaded
	var organizationTypes []string
	if role.OrganizationTypes != nil {
		organizationTypes = make([]string, len(role.OrganizationTypes))
		for i, orgType := range role.OrganizationTypes {
			organizationTypes[i] = orgType.OrganizationType
		}
	}

	return &dto.RoleResponse{
		ID:                role.ID,
		Name:              role.Name,
		Description:       role.Description,
		Level:             role.Level,
		IsSystemRole:      role.IsSystemRole,
		PredefinedName:    role.PredefinedName,
		IsActive:          role.IsActive,
		OrganizationTypes: organizationTypes,
		Permissions:       permissions,
	}
}

// MapPermissionToResponse converts a Permission model to PermissionResponse DTO
func MapPermissionToResponse(permission *model.Permission) *dto.PermissionResponse {
	if permission == nil {
		return nil
	}

	return &dto.PermissionResponse{
		ID:          permission.ID,
		Name:        permission.Name,
		Description: permission.Description,
		CreatedAt:   permission.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   permission.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// MapOrganizationToResponse converts an Organization model to OrganizationResponse DTO
func MapOrganizationToResponse(org *model.Organization) *dto.OrganizationResponse {
	if org == nil {
		return nil
	}

	response := &dto.OrganizationResponse{
		ID:               org.ID,
		Name:             org.Name,
		Code:             org.Code,
		OrganizationType: org.OrganizationType,
		Description:      org.Description,
		CreatedBy:        org.CreatedBy,
		IsActive:         org.IsActive,
		CreatedAt:        org.CreatedAt,
		UpdatedAt:        org.UpdatedAt,
	}

	if org.ParentOrganizationID != nil {
		response.ParentOrganizationID = org.ParentOrganizationID
	}

	if org.ParentOrganization != nil {
		response.ParentOrganization = MapOrganizationToResponse(org.ParentOrganization)
	}

	if org.Creator.ID.String() != "00000000-0000-0000-0000-000000000000" {
		response.Creator = MapUserToResponse(&org.Creator)
	}

	return response
}

// ValidateAndSetPaginationParams validates and sets default pagination parameters
// This is a common utility function to avoid duplication across services
func ValidateAndSetPaginationParams(page, limit int) (int, int, int) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	offset := (page - 1) * limit
	return page, limit, offset
}
