package handler

import (
	"beresin-backend/internal/apperror"
	"beresin-backend/internal/constant"
	"beresin-backend/internal/dto"
	"beresin-backend/internal/service"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// RoleHandler handles HTTP requests related to role management.
type RoleHandler struct {
	roleService service.RoleService
}

// NewRoleHandler creates a new instance of RoleHandler.
func NewRoleHandler(roleService service.RoleService) *RoleHandler {
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

	role, err := h.roleService.CreateRole(req)
	if err != nil {
		return err // Cukup kembalikan error, biarkan middleware yang menangani
	}

	return c.JSON(http.StatusCreated, role)
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

	if err := h.roleService.UpdateRolePermissions(roleID, req.PermissionNames); err != nil {
		return err // Serahkan ke error handler terpusat
	}

	return c.JSON(http.StatusOK, map[string]string{"message": constant.MsgRolePermsUpdated})
}
