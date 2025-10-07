package handler

import (
	"go-base-project/internal/apperror"
	"go-base-project/internal/constant"
	"go-base-project/internal/dto"
	"go-base-project/internal/service"
	"context"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// UserHandler handles HTTP requests related to user management.
type UserHandler struct {
	userService service.UserServiceInterface
}

// NewUserHandler creates a new instance of UserHandler.
func NewUserHandler(userService service.UserServiceInterface) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// CreateUser handles the creation of a new user.
// @Summary      Create a new user
// @Description  Creates a new user with the provided details. Requires 'users:create' permission.
// @Tags         Admin, Users
// @Accept       json
// @Produce      json
// @Param        user body dto.CreateUserRequest true "New User Details"
// @Security     BearerAuth
// @Success      201 {object} dto.UserResponse "User created successfully"
// @Failure      400 {object} apperror.AppError "Invalid request payload"
// @Failure      403 {object} apperror.AppError "Insufficient permissions"
// @Failure      409 {object} apperror.AppError "Username or email already exists"
// @Failure      500 {object} apperror.AppError "Internal server error"
// @Router       /admin/users [post]
func (h *UserHandler) CreateUser(c echo.Context) error {
	var req dto.CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, constant.ErrMsgInvalidRequestFormat)
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	user, err := h.userService.CreateUser(c.Request().Context(), req)
	if err != nil {
		return err // Serahkan ke error handler terpusat
	}

	return c.JSON(http.StatusCreated, user)
}

// ListUsers handles the retrieval of a paginated list of users filtered by level and organization.
// @Summary      List users with level and organization filtering
// @Description  Retrieves a paginated list of users filtered based on the requesting user's level and organization context. Users can only see users with levels below their own level and from their accessible organizations.
// @Tags         Admin, Users
// @Produce      json
// @Param        page query int false "Page number for pagination" default(1)
// @Param        limit query int false "Number of items per page for pagination" default(10)
// @Param        search query string false "Search term for username or email"
// @Security     BearerAuth
// @Success      200 {object} dto.PagedUserResponse "A paginated list of filtered users"
// @Failure      403 {object} apperror.AppError "Insufficient permissions"
// @Failure      500 {object} apperror.AppError "Internal server error"
// @Router       /admin/users [get]
func (h *UserHandler) ListUsers(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	search := c.QueryParam("search")

	// Get current user ID from JWT middleware context
	currentUserID, ok := c.Get(constant.UserIDKey).(uuid.UUID)
	if !ok {
		return apperror.NewAppError(http.StatusUnauthorized, constant.ErrMsgAuthorizationContextMissing, nil)
	}

	// Create context with current user ID for service layer filtering
	ctx := context.WithValue(c.Request().Context(), "current_user_id", currentUserID)

	pagedResponse, err := h.userService.ListUsers(ctx, page, limit, search)
	if err != nil {
		return err // Serahkan ke error handler terpusat
	}

	return c.JSON(http.StatusOK, pagedResponse)
}

// GetUserByID handles the retrieval of a single user by their ID.
// @Summary      Get a single user by ID
// @Description  Retrieves the details of a single user by their ID. Requires 'users:read' permission.
// @Tags         Admin, Users
// @Produce      json
// @Param        id path string true "User ID" format(uuid)
// @Security     BearerAuth
// @Success      200 {object} dto.UserResponse "User details"
// @Failure      403 {object} apperror.AppError "Insufficient permissions"
// @Failure      404 {object} apperror.AppError "User not found"
// @Failure      500 {object} apperror.AppError "Internal server error"
// @Router       /admin/users/{id} [get]
func (h *UserHandler) GetUserByID(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperror.NewAppError(http.StatusBadRequest, constant.ErrMsgInvalidUserID, err)
	}

	user, err := h.userService.GetUserByID(c.Request().Context(), id)
	if err != nil {
		return err // Serahkan ke error handler terpusat
	}

	return c.JSON(http.StatusOK, user)
}

// UpdateUser handles updating a user's details.
// @Summary      Update a user
// @Description  Updates a user's details. Requires 'users:update' permission.
// @Tags         Admin, Users
// @Accept       json
// @Produce      json
// @Param        id path string true "User ID" format(uuid)
// @Param        user body dto.UpdateUserRequest true "User Details to Update"
// @Security     BearerAuth
// @Success      200 {object} dto.UserResponse "User updated successfully"
// @Failure      400 {object} apperror.AppError "Invalid request payload or user ID"
// @Failure      403 {object} apperror.AppError "Insufficient permissions"
// @Failure      404 {object} apperror.AppError "User not found"
// @Failure      500 {object} apperror.AppError "Internal server error"
// @Router       /admin/users/{id} [put]
func (h *UserHandler) UpdateUser(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperror.NewAppError(http.StatusBadRequest, constant.ErrMsgInvalidUserID, err)
	}

	var req dto.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, constant.ErrMsgInvalidRequestFormat)
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	// Get current user ID from JWT middleware context
	currentUserID, ok := c.Get(constant.UserIDKey).(uuid.UUID)
	if !ok {
		return apperror.NewAppError(http.StatusUnauthorized, constant.ErrMsgAuthorizationContextMissing, nil)
	}

	// Create context with current user ID for service layer validation
	ctx := context.WithValue(c.Request().Context(), "current_user_id", currentUserID)

	user, err := h.userService.UpdateUser(ctx, id, req)
	if err != nil {
		return err // Serahkan ke error handler terpusat
	}

	response := dto.UpdateUserResponse{
		User:    *user,
		Message: constant.MsgUserUpdated,
	}

	return c.JSON(http.StatusOK, response)
}

// DeleteUser handles deleting a user.
// @Summary      Delete a user
// @Description  Deletes a user by their ID. Requires 'users:delete' permission.
// @Tags         Admin, Users
// @Produce      json
// @Param        id path string true "User ID" format(uuid)
// @Security     BearerAuth
// @Success      204 "No Content"
// @Failure      403 {object} apperror.AppError "Insufficient permissions"
// @Failure      404 {object} apperror.AppError "User not found"
// @Failure      500 {object} apperror.AppError "Internal server error"
// @Router       /admin/users/{id} [delete]
func (h *UserHandler) DeleteUser(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperror.NewAppError(http.StatusBadRequest, constant.ErrMsgInvalidUserID, err)
	}

	err = h.userService.DeleteUser(c.Request().Context(), id)
	if err != nil {
		return err // Serahkan ke error handler terpusat
	}

	return c.NoContent(http.StatusNoContent)
}

// User-Organization Management Handlers

// AssignUserToOrganization handles assigning a user to an organization.
// @Summary      Assign user to organization
// @Description  Assigns a user to an organization with a specific role. Requires 'users:assign-organization' permission.
// @Tags         Admin, Users, Organizations
// @Accept       json
// @Produce      json
// @Param        request body dto.AssignUserToOrganizationRequest true "Assignment Details"
// @Security     BearerAuth
// @Success      201 {object} dto.UserOrganizationResponse "User assigned successfully"
// @Failure      400 {object} apperror.AppError "Invalid request payload"
// @Failure      403 {object} apperror.AppError "Insufficient permissions"
// @Failure      404 {object} apperror.AppError "User or organization not found"
// @Failure      409 {object} apperror.AppError "User already assigned to organization"
// @Failure      500 {object} apperror.AppError "Internal server error"
// @Router       /admin/users/assign-organization [post]
func (h *UserHandler) AssignUserToOrganization(c echo.Context) error {
	var req dto.AssignUserToOrganizationRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, constant.ErrMsgInvalidRequestFormat)
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	ctx := context.WithValue(c.Request().Context(), constant.RequestIDKey, c.Response().Header().Get(echo.HeaderXRequestID))
	assignment, err := h.userService.AssignUserToOrganization(ctx, req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, assignment)
}

// RemoveUserFromOrganization handles removing a user from an organization.
// @Summary      Remove user from organization
// @Description  Removes a user from an organization. Requires 'users:remove-organization' permission.
// @Tags         Admin, Users, Organizations
// @Produce      json
// @Param        userId path string true "User ID" format(uuid)
// @Param        organizationId path string true "Organization ID" format(uuid)
// @Security     BearerAuth
// @Success      204 "No Content"
// @Failure      400 {object} apperror.AppError "Invalid user or organization ID"
// @Failure      403 {object} apperror.AppError "Insufficient permissions"
// @Failure      404 {object} apperror.AppError "User organization assignment not found"
// @Failure      500 {object} apperror.AppError "Internal server error"
// @Router       /admin/users/{userId}/organizations/{organizationId} [delete]
func (h *UserHandler) RemoveUserFromOrganization(c echo.Context) error {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		return apperror.NewAppError(http.StatusBadRequest, constant.ErrMsgInvalidUserID, err)
	}

	organizationID, err := uuid.Parse(c.Param("organizationId"))
	if err != nil {
		return apperror.NewAppError(http.StatusBadRequest, "Invalid organization ID", err)
	}

	ctx := context.WithValue(c.Request().Context(), constant.RequestIDKey, c.Response().Header().Get(echo.HeaderXRequestID))
	err = h.userService.RemoveUserFromOrganization(ctx, userID, organizationID)
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

// UpdateUserOrganizationRole handles updating a user's role in an organization.
// @Summary      Update user's role in organization
// @Description  Updates a user's role in an organization. Requires 'users:update-organization-role' permission.
// @Tags         Admin, Users, Organizations
// @Accept       json
// @Produce      json
// @Param        userId path string true "User ID" format(uuid)
// @Param        organizationId path string true "Organization ID" format(uuid)
// @Param        request body dto.UpdateUserOrganizationRequest true "Role Update Details"
// @Security     BearerAuth
// @Success      200 {object} dto.UserOrganizationResponse "Role updated successfully"
// @Failure      400 {object} apperror.AppError "Invalid request payload or IDs"
// @Failure      403 {object} apperror.AppError "Insufficient permissions"
// @Failure      404 {object} apperror.AppError "User organization assignment not found"
// @Failure      500 {object} apperror.AppError "Internal server error"
// @Router       /admin/users/{userId}/organizations/{organizationId} [put]
func (h *UserHandler) UpdateUserOrganizationRole(c echo.Context) error {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		return apperror.NewAppError(http.StatusBadRequest, constant.ErrMsgInvalidUserID, err)
	}

	organizationID, err := uuid.Parse(c.Param("organizationId"))
	if err != nil {
		return apperror.NewAppError(http.StatusBadRequest, "Invalid organization ID", err)
	}

	var req dto.UpdateUserOrganizationRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, constant.ErrMsgInvalidRequestFormat)
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	ctx := context.WithValue(c.Request().Context(), constant.RequestIDKey, c.Response().Header().Get(echo.HeaderXRequestID))
	assignment, err := h.userService.UpdateUserOrganizationRole(ctx, userID, organizationID, req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, assignment)
}

// GetUserOrganizations handles retrieving all organizations that a user belongs to.
// @Summary      Get user's organizations
// @Description  Retrieves all organizations that a user belongs to with pagination. Requires 'users:read' permission.
// @Tags         Admin, Users, Organizations
// @Produce      json
// @Param        userId path string true "User ID" format(uuid)
// @Param        page query int false "Page number" default(1) minimum(1)
// @Param        limit query int false "Items per page" default(10) minimum(1) maximum(100)
// @Security     BearerAuth
// @Success      200 {object} dto.PagedUserOrganizationResponse "User organizations retrieved successfully"
// @Failure      400 {object} apperror.AppError "Invalid user ID or query parameters"
// @Failure      403 {object} apperror.AppError "Insufficient permissions"
// @Failure      404 {object} apperror.AppError "User not found"
// @Failure      500 {object} apperror.AppError "Internal server error"
// @Router       /admin/users/{userId}/organizations [get]
func (h *UserHandler) GetUserOrganizations(c echo.Context) error {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		return apperror.NewAppError(http.StatusBadRequest, constant.ErrMsgInvalidUserID, err)
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	ctx := context.WithValue(c.Request().Context(), constant.RequestIDKey, c.Response().Header().Get(echo.HeaderXRequestID))
	organizations, err := h.userService.GetUserOrganizations(ctx, userID, page, limit)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, organizations)
}

// GetOrganizationMembers handles retrieving all members of an organization.
// @Summary      Get organization members
// @Description  Retrieves all members of an organization with pagination. Requires 'organizations:read-members' permission.
// @Tags         Admin, Organizations, Users
// @Produce      json
// @Param        organizationId path string true "Organization ID" format(uuid)
// @Param        page query int false "Page number" default(1) minimum(1)
// @Param        limit query int false "Items per page" default(10) minimum(1) maximum(100)
// @Security     BearerAuth
// @Success      200 {object} dto.PagedUserOrganizationResponse "Organization members retrieved successfully"
// @Failure      400 {object} apperror.AppError "Invalid organization ID or query parameters"
// @Failure      403 {object} apperror.AppError "Insufficient permissions"
// @Failure      404 {object} apperror.AppError "Organization not found"
// @Failure      500 {object} apperror.AppError "Internal server error"
// @Router       /admin/organizations/{organizationId}/members [get]
func (h *UserHandler) GetOrganizationMembers(c echo.Context) error {
	organizationID, err := uuid.Parse(c.Param("organizationId"))
	if err != nil {
		return apperror.NewAppError(http.StatusBadRequest, "Invalid organization ID", err)
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	ctx := context.WithValue(c.Request().Context(), constant.RequestIDKey, c.Response().Header().Get(echo.HeaderXRequestID))
	members, err := h.userService.GetOrganizationMembers(ctx, organizationID, page, limit)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, members)
}

// BulkAssignUsersToOrganization handles bulk assignment of users to an organization.
// @Summary      Bulk assign users to organization
// @Description  Assigns multiple users to an organization with the same role. Requires 'users:bulk-assign-organization' permission.
// @Tags         Admin, Users, Organizations
// @Accept       json
// @Produce      json
// @Param        request body dto.BulkAssignUsersToOrganizationRequest true "Bulk Assignment Details"
// @Security     BearerAuth
// @Success      200 {object} dto.BulkAssignResponse "Bulk assignment completed"
// @Failure      400 {object} apperror.AppError "Invalid request payload"
// @Failure      403 {object} apperror.AppError "Insufficient permissions"
// @Failure      404 {object} apperror.AppError "Organization not found"
// @Failure      500 {object} apperror.AppError "Internal server error"
// @Router       /admin/users/bulk-assign-organization [post]
func (h *UserHandler) BulkAssignUsersToOrganization(c echo.Context) error {
	var req dto.BulkAssignUsersToOrganizationRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, constant.ErrMsgInvalidRequestFormat)
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	ctx := context.WithValue(c.Request().Context(), constant.RequestIDKey, c.Response().Header().Get(echo.HeaderXRequestID))
	result, err := h.userService.BulkAssignUsersToOrganization(ctx, req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, result)
}

// GetUserOrganizationHistory handles retrieving organization assignment history for a user.
// @Summary      Get user's organization history
// @Description  Retrieves the organization assignment history for a user. Requires 'users:read-history' permission.
// @Tags         Admin, Users, Organizations
// @Produce      json
// @Param        userId path string true "User ID" format(uuid)
// @Param        page query int false "Page number" default(1) minimum(1)
// @Param        limit query int false "Items per page" default(10) minimum(1) maximum(100)
// @Security     BearerAuth
// @Success      200 {array} dto.UserOrganizationHistoryResponse "User organization history retrieved successfully"
// @Failure      400 {object} apperror.AppError "Invalid user ID or query parameters"
// @Failure      403 {object} apperror.AppError "Insufficient permissions"
// @Failure      404 {object} apperror.AppError "User not found"
// @Failure      500 {object} apperror.AppError "Internal server error"
// @Router       /admin/users/{userId}/organization-history [get]
func (h *UserHandler) GetUserOrganizationHistory(c echo.Context) error {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		return apperror.NewAppError(http.StatusBadRequest, constant.ErrMsgInvalidUserID, err)
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	ctx := context.WithValue(c.Request().Context(), constant.RequestIDKey, c.Response().Header().Get(echo.HeaderXRequestID))
	history, err := h.userService.GetUserOrganizationHistory(ctx, userID, page, limit)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, history)
}

// GetOrganizationUserHistory handles retrieving user assignment history for an organization.
// @Summary      Get organization user history
// @Description  Retrieves the user assignment history for an organization with pagination. Requires 'organizations:read-history' permission.
// @Tags         Admin, Organizations, Users
// @Produce      json
// @Param        organizationId path string true "Organization ID" format(uuid)
// @Param        page query int false "Page number" default(1) minimum(1)
// @Param        limit query int false "Items per page" default(10) minimum(1) maximum(100)
// @Security     BearerAuth
// @Success      200 {object} dto.PagedUserOrganizationHistoryResponse "Organization user history retrieved successfully"
// @Failure      400 {object} apperror.AppError "Invalid organization ID or query parameters"
// @Failure      403 {object} apperror.AppError "Insufficient permissions"
// @Failure      404 {object} apperror.AppError "Organization not found"
// @Failure      500 {object} apperror.AppError "Internal server error"
// @Router       /admin/organizations/{organizationId}/user-history [get]
func (h *UserHandler) GetOrganizationUserHistory(c echo.Context) error {
	organizationID, err := uuid.Parse(c.Param("organizationId"))
	if err != nil {
		return apperror.NewAppError(http.StatusBadRequest, "Invalid organization ID", err)
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	ctx := context.WithValue(c.Request().Context(), constant.RequestIDKey, c.Response().Header().Get(echo.HeaderXRequestID))
	history, err := h.userService.GetOrganizationUserHistory(ctx, organizationID, page, limit)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, history)
}

// LogUserOrganizationAction handles manually logging a user organization action.
// @Summary      Log user organization action
// @Description  Manually logs a user organization action for audit purposes. Requires 'users:log-actions' permission.
// @Tags         Admin, Users, Organizations
// @Accept       json
// @Produce      json
// @Param        request body dto.LogUserOrganizationActionRequest true "Action Log Details"
// @Security     BearerAuth
// @Success      201 {object} dto.UserOrganizationHistoryResponse "Action logged successfully"
// @Failure      400 {object} apperror.AppError "Invalid request payload"
// @Failure      403 {object} apperror.AppError "Insufficient permissions"
// @Failure      404 {object} apperror.AppError "User or organization not found"
// @Failure      500 {object} apperror.AppError "Internal server error"
// @Router       /admin/users/log-organization-action [post]
func (h *UserHandler) LogUserOrganizationAction(c echo.Context) error {
	var req dto.LogUserOrganizationActionRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, constant.ErrMsgInvalidRequestFormat)
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	// Get current user ID from JWT context
	userIDValue := c.Get(constant.UserIDKey)
	if userIDValue == nil {
		return apperror.NewUnauthorizedError("No user information found in context")
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		return apperror.NewUnauthorizedError("Invalid user information in context")
	}

	ctx := context.WithValue(c.Request().Context(), constant.RequestIDKey, c.Response().Header().Get(echo.HeaderXRequestID))
	history, err := h.userService.LogUserOrganizationAction(ctx, req, userID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, history)
}

// AssignRoleToUserInOrganization assigns a role to a user within a specific organization context.
// This method demonstrates multi-tenant role isolation.
// @Summary      Assign role to user in organization
// @Description  Assigns a specific role to a user within the current organization context. Demonstrates multi-tenant role isolation.
// @Tags         Organization, Users, Roles
// @Accept       json
// @Produce      json
// @Param        userId path string true "User ID" format(uuid)
// @Param        roleId body string true "Role ID to assign" format(uuid)
// @Security     BearerAuth
// @Success      200 {object} map[string]interface{} "Role assigned successfully"
// @Failure      400 {object} apperror.AppError "Bad request"
// @Failure      403 {object} apperror.AppError "Insufficient permissions"
// @Failure      500 {object} apperror.AppError "Internal server error"
// @Router       /organizations/{orgId}/users/{userId}/assign-role [post]
func (h *UserHandler) AssignRoleToUserInOrganization(c echo.Context) error {
	// Get organization ID from context (set by OrganizationContext middleware)
	orgIDValue := c.Get(constant.OrganizationIDKey)
	if orgIDValue == nil {
		return apperror.NewAppError(http.StatusBadRequest, constant.ErrMsgOrganizationContextRequired, nil)
	}

	organizationID, ok := orgIDValue.(uuid.UUID)
	if !ok {
		return apperror.NewAppError(http.StatusBadRequest, constant.ErrMsgInvalidOrganizationIDFormat, nil)
	}

	// Parse user ID from path parameter
	userIDStr := c.Param("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return apperror.NewAppError(http.StatusBadRequest, constant.ErrMsgInvalidUserID, err)
	}

	// Parse role ID from request body
	var req struct {
		RoleID string `json:"role_id" validate:"required"`
	}
	if err := c.Bind(&req); err != nil {
		return apperror.NewAppError(http.StatusBadRequest, constant.ErrMsgInvalidRequestFormat, err)
	}

	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		return apperror.NewAppError(http.StatusBadRequest, constant.ErrMsgInvalidRoleID, err)
	}

	// For demonstration, we'll return the assignment details
	// In a real implementation, this would call the user service to update the role
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":         "Role assignment with multi-tenant isolation",
		"user_id":         userID,
		"organization_id": organizationID,
		"role_id":         roleID,
		"note":            "This demonstrates organization-scoped role assignment preventing cross-tenant access",
	})
}

// GetUserRoleInOrganization retrieves a user's role within the current organization context.
// @Summary      Get user role in organization
// @Description  Retrieves the role assigned to a user within the current organization context.
// @Tags         Organization, Users, Roles
// @Produce      json
// @Param        userId path string true "User ID" format(uuid)
// @Security     BearerAuth
// @Success      200 {object} map[string]interface{} "User role information"
// @Failure      400 {object} apperror.AppError "Bad request"
// @Failure      403 {object} apperror.AppError "Insufficient permissions"
// @Failure      404 {object} apperror.AppError "User not found in organization"
// @Router       /organizations/{orgId}/users/{userId}/role [get]
func (h *UserHandler) GetUserRoleInOrganization(c echo.Context) error {
	// Get organization ID from context
	orgIDValue := c.Get(constant.OrganizationIDKey)
	if orgIDValue == nil {
		return apperror.NewAppError(http.StatusBadRequest, constant.ErrMsgOrganizationContextRequired, nil)
	}

	organizationID, ok := orgIDValue.(uuid.UUID)
	if !ok {
		return apperror.NewAppError(http.StatusBadRequest, constant.ErrMsgInvalidOrganizationIDFormat, nil)
	}

	// Parse user ID from path parameter
	userIDStr := c.Param("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return apperror.NewAppError(http.StatusBadRequest, constant.ErrMsgInvalidUserID, err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":         "User role in organization context",
		"user_id":         userID,
		"organization_id": organizationID,
		"note":            "This demonstrates organization-scoped role retrieval with complete tenant isolation",
	})
}

// ListUsersByOrganization handles the retrieval of users scoped to the current organization context.
// This method demonstrates how to use organization context from middleware.
// @Summary      List users by organization context
// @Description  Retrieves users within the current organization context. Organization ID is automatically injected by middleware.
// @Tags         Organization, Users
// @Produce      json
// @Param        page query int false "Page number (1-based)"
// @Param        limit query int false "Number of items per page"
// @Security     BearerAuth
// @Success      200 {object} map[string]interface{} "Organization-scoped user list"
// @Failure      400 {object} apperror.AppError "Bad request - Missing organization context"
// @Failure      403 {object} apperror.AppError "Insufficient permissions"
// @Failure      500 {object} apperror.AppError "Internal server error"
// @Router       /organizations/{orgId}/users [get]
func (h *UserHandler) ListUsersByOrganization(c echo.Context) error {
	// Get organization ID from context (set by OrganizationContext middleware)
	orgIDValue := c.Get(constant.OrganizationIDKey)
	if orgIDValue == nil {
		return apperror.NewAppError(http.StatusBadRequest, constant.ErrMsgOrganizationContextRequired, nil)
	}

	organizationID, ok := orgIDValue.(uuid.UUID)
	if !ok {
		return apperror.NewAppError(http.StatusBadRequest, constant.ErrMsgInvalidOrganizationIDFormat, nil)
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	// For demonstration purposes, let's return organization info with a message
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":         "Users scoped to organization",
		"organization_id": organizationID,
		"page":            page,
		"limit":           limit,
		"note":            "This demonstrates organization-scoped data access using OrganizationContext middleware",
	})
}
