package handler

import (
	"beresin-backend/internal/apperror"
	"beresin-backend/internal/constant"
	"beresin-backend/internal/dto"
	"beresin-backend/internal/service"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// UserHandler handles HTTP requests related to user management.
type UserHandler struct {
	userService service.UserService
}

// NewUserHandler creates a new instance of UserHandler.
func NewUserHandler(userService service.UserService) *UserHandler {
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

	user, err := h.userService.CreateUser(req)
	if err != nil {
		return err // Serahkan ke error handler terpusat
	}

	return c.JSON(http.StatusCreated, user)
}

// ListUsers handles the retrieval of a paginated list of all users.
// @Summary      List all users
// @Description  Retrieves a paginated list of all users. Requires 'users:read' permission.
// @Tags         Admin, Users
// @Produce      json
// @Param        page query int false "Page number for pagination" default(1)
// @Param        limit query int false "Number of items per page for pagination" default(10)
// @Param        search query string false "Search term for username or email"
// @Security     BearerAuth
// @Success      200 {object} dto.PagedUserResponse "A paginated list of users"
// @Failure      403 {object} apperror.AppError "Insufficient permissions"
// @Failure      500 {object} apperror.AppError "Internal server error"
// @Router       /admin/users [get]
func (h *UserHandler) ListUsers(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	search := c.QueryParam("search")

	pagedResponse, err := h.userService.ListUsers(page, limit, search)
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

	user, err := h.userService.GetUserByID(id)
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

	user, err := h.userService.UpdateUser(id, req)
	if err != nil {
		return err // Serahkan ke error handler terpusat
	}

	return c.JSON(http.StatusOK, user)
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

	err = h.userService.DeleteUser(id)
	if err != nil {
		return err // Serahkan ke error handler terpusat
	}

	return c.NoContent(http.StatusNoContent)
}
