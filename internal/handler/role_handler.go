package handler

import (
	"go-base-project/internal/apperror"
	"go-base-project/internal/constant"
	"go-base-project/internal/dto"
	"go-base-project/internal/service"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// RoleHandler handles HTTP requests related to role management.
type RoleHandler struct {
	roleService service.RoleServiceInterface
}

// NewRoleHandler creates a new instance of RoleHandler.
func NewRoleHandler(roleService service.RoleServiceInterface) *RoleHandler {
	return &RoleHandler{
		roleService: roleService,
	}
}

// CreateRole handles the creation of a new role.
// @Summary      Create a new role
// @Description  Creates a new role. Requires 'roles:create' permission.
// @Tags         Admin, Roles
// @Accept       json
// @Produce      json
// @Param        role body dto.CreateRoleRequest true "New Role Details"
// @Security     BearerAuth
// @Success      201 {object} dto.RoleResponse "Role created successfully"
// @Failure      400 {object} apperror.AppError "Invalid request payload"
// @Failure      403 {object} apperror.AppError "Insufficient permissions"
// @Failure      409 {object} apperror.AppError "Role with that name already exists"
// @Failure      500 {object} apperror.AppError "Internal server error"
// @Router       /admin/roles [post]
func (h *RoleHandler) CreateRole(c echo.Context) error {
	var req dto.CreateRoleRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, constant.ErrMsgInvalidRequestFormat)
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	// Get current user's role ID from JWT context
	roleIDValue := c.Get(constant.RoleIDKey)
	roleID, ok := roleIDValue.(uuid.UUID)
	if !ok {
		return echo.NewHTTPError(http.StatusForbidden, "Role not found in token")
	}

	// Get current user's role to determine their level
	currentRole, err := h.roleService.GetRoleByID(c.Request().Context(), roleID)
	if err != nil {
		return err
	}

	// Create role with hierarchical validation based on user's level
	role, err := h.roleService.CreateRole(c.Request().Context(), req, currentRole.Level)
	if err != nil {
		return err // Cukup kembalikan error, biarkan middleware yang menangani
	}

	return c.JSON(http.StatusCreated, role)
}

// ListRoles handles the retrieval of roles with hierarchical access control.
// @Summary      List roles based on user level
// @Description  Retrieves all roles below the current user's level (hierarchical access control). Requires 'roles:read' permission.
// @Tags         Admin, Roles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {array} dto.RoleResponse "Roles retrieved successfully"
// @Failure      403 {object} apperror.AppError "Insufficient permissions"
// @Failure      500 {object} apperror.AppError "Internal server error"
// @Router       /admin/roles [get]
func (h *RoleHandler) ListRoles(c echo.Context) error {
	// Get current user's role ID from JWT context
	roleIDValue := c.Get(constant.RoleIDKey)
	roleID, ok := roleIDValue.(uuid.UUID)
	if !ok {
		return echo.NewHTTPError(http.StatusForbidden, "Role not found in token")
	}

	// Get current user's role to determine their level
	currentRole, err := h.roleService.GetRoleByID(c.Request().Context(), roleID)
	if err != nil {
		return err
	}

	// Get roles below current user's level using hierarchical filtering
	roles, err := h.roleService.ListRoles(c.Request().Context(), currentRole.Level)
	if err != nil {
		return err // Let centralized error handler deal with it
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"roles": roles,
	})
}

// UpdateRole handles updating a role's basic information (name, description, level).
// @Summary      Update role information
// @Description  Updates a role's basic information with hierarchical validation. Requires 'roles:update' permission.
// @Tags         Admin, Roles
// @Accept       json
// @Produce      json
// @Param        id path string true "Role ID"
// @Param        role body dto.UpdateRoleRequest true "Updated role information"
// @Success      200 {object} dto.RoleResponse "Role updated successfully"
// @Failure      400 {object} apperror.AppError "Invalid request payload or validation error"
// @Failure      403 {object} apperror.AppError "Insufficient permissions"
// @Failure      404 {object} apperror.AppError "Role not found"
// @Failure      409 {object} apperror.AppError "Role name conflict"
// @Failure      500 {object} apperror.AppError "Internal server error"
// @Router       /admin/roles/{id} [put]
func (h *RoleHandler) UpdateRole(c echo.Context) error {
	// Parse role ID from URL parameter
	roleIDStr := c.Param("id")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid role ID format")
	}

	// Parse request body
	var req dto.UpdateRoleRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, constant.ErrMsgInvalidRequestFormat)
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	// Get current user's role ID from JWT context
	currentRoleIDValue := c.Get(constant.RoleIDKey)
	currentRoleID, ok := currentRoleIDValue.(uuid.UUID)
	if !ok {
		return echo.NewHTTPError(http.StatusForbidden, "Role not found in token")
	}

	// Get current user's role to determine their level
	currentRole, err := h.roleService.GetRoleByID(c.Request().Context(), currentRoleID)
	if err != nil {
		return err
	}

	// Update role with hierarchical validation based on user's level
	updatedRole, err := h.roleService.UpdateRole(c.Request().Context(), roleID, req, currentRole.Level)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, updatedRole)
}

// UpdateRolePermissions handles updating permissions for a role.
// @Summary      Update permissions for a role
// @Description  Updates the list of permissions associated with a specific role. This action requires 'roles:assign' permission.
// @Tags         Admin, Roles
// @Accept       json
// @Produce      json
// @Param        id path string true "Role ID" format(uuid)
// @Param        permissions body dto.UpdateRolePermissionsRequest true "List of permission names"
// @Security     BearerAuth
// @Success      200 {object} map[string]string "Role permissions updated successfully"
// @Failure      400 {object} apperror.AppError "Invalid role ID format or validation failed"
// @Failure      403 {object} apperror.AppError "Insufficient permissions"
// @Failure      404 {object} apperror.AppError "Role not found"
// @Failure      500 {object} apperror.AppError "Internal server error"
// @Router       /admin/roles/{id}/permissions [put]
func (h *RoleHandler) UpdateRolePermissions(c echo.Context) error {
	roleIDStr := c.Param("id")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		return apperror.NewAppError(http.StatusBadRequest, constant.ErrMsgInvalidRoleID, err)
	}

	var req dto.UpdateRolePermissionsRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, constant.ErrMsgInvalidRequestFormat)
	}
	if err := c.Validate(&req); err != nil {
		return err // Error sudah dalam format HTTPError dari custom validator
	}

	if err := h.roleService.UpdateRolePermissions(c.Request().Context(), roleID, req.PermissionNames); err != nil {
		return err // Serahkan ke error handler terpusat
	}

	return c.JSON(http.StatusOK, map[string]string{"message": constant.MsgRolePermsUpdated})
}

// DISABLED: Role approval functionality
// CreateRoleApprovalRequest handles the creation of a new role approval request.
// @Summary      Create a role approval request
// @Description  Creates a new role approval request. Users can request new roles to be created.
// @Tags         Roles, Approval
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateRoleApprovalRequest true "Role Approval Request Details"
// @Security     BearerAuth
// @Success      201 {object} dto.RoleApprovalResponse "Role approval request created successfully"
// @Failure      400 {object} apperror.AppError "Invalid request payload"
// @Failure      401 {object} apperror.AppError "Unauthorized"
// @Failure      500 {object} apperror.AppError "Internal server error"
// @Router       /roles/approval-requests [post]
/*
func (h *RoleHandler) CreateRoleApprovalRequest(c echo.Context) error {
	var req dto.CreateRoleApprovalRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, constant.ErrMsgInvalidRequestFormat)
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	// Get user ID from context (set by auth middleware)
	userID, ok := c.Get(constant.UserIDKey).(uuid.UUID)
	if !ok {
		return apperror.NewAppError(http.StatusUnauthorized, "User ID not found in context", nil)
	}

	approval, err := h.roleService.CreateRoleApprovalRequest(c.Request().Context(), req, userID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, approval)
}
*/

/*
// ListRoleApprovalRequests handles the retrieval of all role approval requests.
// @Summary      List all role approval requests
// @Description  Retrieves all role approval requests. Requires 'roles:approve' permission.
// @Tags         Admin, Roles, Approval
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {array} dto.RoleApprovalResponse "Role approval requests retrieved successfully"
// @Failure      403 {object} apperror.AppError "Insufficient permissions"
// @Failure      500 {object} apperror.AppError "Internal server error"
// @Router       /admin/roles/approval-requests [get]
func (h *RoleHandler) ListRoleApprovalRequests(c echo.Context) error {
	approvals, err := h.roleService.ListRoleApprovalRequests(c.Request().Context())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"approval_requests": approvals,
	})
}
*/

/*
// ApproveRejectRoleRequest handles approving or rejecting a role approval request.
// @Summary      Approve or reject a role request
// @Description  Approves or rejects a role approval request. Requires 'roles:approve' permission.
// @Tags         Admin, Roles, Approval
// @Accept       json
// @Produce      json
// @Param        id path string true "Approval Request ID" format(uuid)
// @Param        decision body dto.ApprovalDecisionRequest true "Approval Decision"
// @Security     BearerAuth
// @Success      200 {object} dto.RoleApprovalResponse "Role request processed successfully"
// @Failure      400 {object} apperror.AppError "Invalid request ID format or validation failed"
// @Failure      403 {object} apperror.AppError "Insufficient permissions"
// @Failure      404 {object} apperror.AppError "Approval request not found"
// @Failure      409 {object} apperror.AppError "Request already processed"
// @Failure      500 {object} apperror.AppError "Internal server error"
// @Router       /admin/roles/approval-requests/{id}/decision [put]
func (h *RoleHandler) ApproveRejectRoleRequest(c echo.Context) error {
	approvalIDStr := c.Param("id")
	approvalID, err := uuid.Parse(approvalIDStr)
	if err != nil {
		return apperror.NewAppError(http.StatusBadRequest, "Invalid approval request ID format", err)
	}

	var req dto.ApprovalDecisionRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, constant.ErrMsgInvalidRequestFormat)
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	// Get user ID from context (set by auth middleware)
	approverID, ok := c.Get(constant.UserIDKey).(uuid.UUID)
	if !ok {
		return apperror.NewAppError(http.StatusUnauthorized, "User ID not found in context", nil)
	}

	approval, err := h.roleService.ApproveRejectRoleRequest(c.Request().Context(), approvalID, req, approverID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, approval)
}
*/

// GetPredefinedRoleOptions handles the retrieval of predefined role options.
// @Summary      Get predefined role options
// @Description  Retrieves available predefined role templates for role creation based on user's level (hierarchical access control). Requires 'roles:create' permission.
// @Tags         Admin, Roles
// @Accept       json
// @Produce      json
// @Success      200 {array} dto.PredefinedRoleOption "Predefined role options retrieved successfully"
// @Failure      403 {object} apperror.AppError "Insufficient permissions"
// @Failure      500 {object} apperror.AppError "Internal server error"
// @Router       /roles/predefined-options [get]
func (h *RoleHandler) GetPredefinedRoleOptions(c echo.Context) error {
	// Get current user's role ID from JWT context
	roleIDValue := c.Get(constant.RoleIDKey)
	roleID, ok := roleIDValue.(uuid.UUID)

	// For users without role (new users from OAuth), provide basic role options
	if !ok || roleID == uuid.Nil {
		// Get basic predefined role options for new users (level 0 means all available options)
		options, err := h.roleService.GetPredefinedRoleOptions(c.Request().Context(), 0)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"predefined_roles": options,
		})
	}

	// Get current user's role to determine their level
	currentRole, err := h.roleService.GetRoleByID(c.Request().Context(), roleID)
	if err != nil {
		return err
	}

	// Get predefined role options below current user's level using hierarchical filtering
	options, err := h.roleService.GetPredefinedRoleOptions(c.Request().Context(), currentRole.Level)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"predefined_roles": options,
	})
}

// Permission Management Handlers

// ListPermissions handles the retrieval of all permissions.
// @Summary      List all permissions
// @Description  Retrieves all available permissions in the system. Requires 'permissions:read' permission.
// @Tags         Admin, Permissions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {array} dto.PermissionResponse "Permissions retrieved successfully"
// @Failure      403 {object} apperror.AppError "Insufficient permissions"
// @Failure      500 {object} apperror.AppError "Internal server error"
// @Router       /admin/permissions [get]
func (h *RoleHandler) ListPermissions(c echo.Context) error {
	permissions, err := h.roleService.ListPermissions(c.Request().Context())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, permissions)
}

// CreatePermission handles the creation of a new permission.
// @Summary      Create a new permission
// @Description  Creates a new permission in the system. Requires 'permissions:create' permission.
// @Tags         Admin, Permissions
// @Accept       json
// @Produce      json
// @Param        permission body dto.CreatePermissionRequest true "New Permission Details"
// @Security     BearerAuth
// @Success      201 {object} dto.PermissionResponse "Permission created successfully"
// @Failure      400 {object} apperror.AppError "Invalid request payload"
// @Failure      403 {object} apperror.AppError "Insufficient permissions"
// @Failure      409 {object} apperror.AppError "Permission with that name already exists"
// @Failure      500 {object} apperror.AppError "Internal server error"
// @Router       /admin/permissions [post]
func (h *RoleHandler) CreatePermission(c echo.Context) error {
	var req dto.CreatePermissionRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, constant.ErrMsgInvalidRequestFormat)
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	permission, err := h.roleService.CreatePermission(c.Request().Context(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, permission)
}

// UpdatePermission handles updating an existing permission.
// @Summary      Update a permission
// @Description  Updates an existing permission. Requires 'permissions:update' permission.
// @Tags         Admin, Permissions
// @Accept       json
// @Produce      json
// @Param        id path string true "Permission ID"
// @Param        permission body dto.UpdatePermissionRequest true "Updated Permission Details"
// @Security     BearerAuth
// @Success      200 {object} dto.PermissionResponse "Permission updated successfully"
// @Failure      400 {object} apperror.AppError "Invalid request payload"
// @Failure      403 {object} apperror.AppError "Insufficient permissions"
// @Failure      404 {object} apperror.AppError "Permission not found"
// @Failure      409 {object} apperror.AppError "Permission name already exists"
// @Failure      500 {object} apperror.AppError "Internal server error"
// @Router       /admin/permissions/{id} [put]
func (h *RoleHandler) UpdatePermission(c echo.Context) error {
	permissionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, constant.ErrMsgInvalidRequestFormat)
	}

	var req dto.UpdatePermissionRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, constant.ErrMsgInvalidRequestFormat)
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	permission, err := h.roleService.UpdatePermission(c.Request().Context(), permissionID, req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, permission)
}

// DeletePermission handles deleting a permission.
// @Summary      Delete a permission
// @Description  Deletes a permission from the system. Cannot delete if permission is assigned to any roles. Requires 'permissions:delete' permission.
// @Tags         Admin, Permissions
// @Accept       json
// @Produce      json
// @Param        id path string true "Permission ID"
// @Security     BearerAuth
// @Success      200 {object} map[string]string "Permission deleted successfully"
// @Failure      400 {object} apperror.AppError "Invalid permission ID"
// @Failure      403 {object} apperror.AppError "Insufficient permissions"
// @Failure      404 {object} apperror.AppError "Permission not found"
// @Failure      409 {object} apperror.AppError "Cannot delete permission assigned to roles"
// @Failure      500 {object} apperror.AppError "Internal server error"
// @Router       /admin/permissions/{id} [delete]
func (h *RoleHandler) DeletePermission(c echo.Context) error {
	permissionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, constant.ErrMsgInvalidRequestFormat)
	}

	err = h.roleService.DeletePermission(c.Request().Context(), permissionID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Permission deleted successfully",
	})
}

// Organization-specific Role Handlers

// GetRolesForOrganizationType handles retrieving roles applicable to a specific organization type.
// @Summary      Get roles for organization type
// @Description  Retrieves all roles that are applicable to a specific organization type. Requires 'roles:read' permission.
// @Tags         Admin, Roles, Organizations
// @Produce      json
// @Param        organization_type query string true "Organization Type" Enums(platform, holding, company, store)
// @Security     BearerAuth
// @Success      200 {array} dto.RoleResponse "Roles retrieved successfully"
// @Failure      400 {object} apperror.AppError "Invalid organization type"
// @Failure      403 {object} apperror.AppError "Insufficient permissions"
// @Failure      500 {object} apperror.AppError "Internal server error"
// @Router       /admin/roles/organization-types [get]
func (h *RoleHandler) GetRolesForOrganizationType(c echo.Context) error {
	organizationType := c.QueryParam("organization_type")
	if organizationType == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "organization_type query parameter is required")
	}

	// Get current user's role ID and level from JWT context
	roleIDValue := c.Get(constant.RoleIDKey)
	if roleIDValue == nil {
		return apperror.NewUnauthorizedError("No role information found in context")
	}

	roleID, ok := roleIDValue.(uuid.UUID)
	if !ok {
		return apperror.NewUnauthorizedError("Invalid role information in context")
	}

	// Get user's role details to determine their level
	userRole, err := h.roleService.GetRoleByID(c.Request().Context(), roleID)
	if err != nil {
		return err
	}

	roles, err := h.roleService.GetRolesForOrganizationType(c.Request().Context(), organizationType, userRole.Level)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, roles)
}
