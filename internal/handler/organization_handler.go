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

// OrganizationHandler handles HTTP requests related to organization management
type OrganizationHandler struct {
	orgService service.OrganizationServiceInterface
}

// NewOrganizationHandler creates a new instance of OrganizationHandler
func NewOrganizationHandler(orgService service.OrganizationServiceInterface) *OrganizationHandler {
	return &OrganizationHandler{
		orgService: orgService,
	}
}

// CreateOrganization handles organization creation
// @Summary      Create a new organization
// @Description  Creates a new organization. Requires platform level permissions.
// @Tags         Admin, Organizations
// @Accept       json
// @Produce      json
// @Param        organization body dto.CreateOrganizationRequest true "Organization Details"
// @Security     BearerAuth
// @Success      201 {object} dto.OrganizationResponse "Organization created successfully"
// @Failure      400 {object} apperror.AppError "Invalid request payload"
// @Failure      403 {object} apperror.AppError "Insufficient permissions"
// @Failure      500 {object} apperror.AppError "Internal server error"
// @Router       /admin/organizations [post]
func (h *OrganizationHandler) CreateOrganization(c echo.Context) error {
	var req dto.CreateOrganizationRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	// Get user ID from context
	userID, ok := c.Get(constant.UserIDKey).(uuid.UUID)
	if !ok {
		return apperror.NewAppError(http.StatusUnauthorized, "User ID not found in context", nil)
	}

	org, err := h.orgService.CreateOrganization(c.Request().Context(), req, userID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, org)
}

// GetOrganization handles organization retrieval by ID
// @Summary      Get organization by ID
// @Description  Retrieves organization details by ID
// @Tags         Organizations
// @Accept       json
// @Produce      json
// @Param        id path string true "Organization ID"
// @Security     BearerAuth
// @Success      200 {object} dto.OrganizationResponse "Organization retrieved successfully"
// @Failure      404 {object} apperror.AppError "Organization not found"
// @Failure      500 {object} apperror.AppError "Internal server error"
// @Router       /organizations/{id} [get]
func (h *OrganizationHandler) GetOrganization(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid organization ID format")
	}

	org, err := h.orgService.GetOrganizationByID(c.Request().Context(), id)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, org)
}

// GetOrganizationByCode handles organization retrieval by code
// @Summary      Get organization by code
// @Description  Retrieves organization details by code
// @Tags         Organizations
// @Accept       json
// @Produce      json
// @Param        code path string true "Organization Code"
// @Security     BearerAuth
// @Success      200 {object} dto.OrganizationResponse "Organization retrieved successfully"
// @Failure      404 {object} apperror.AppError "Organization not found"
// @Failure      500 {object} apperror.AppError "Internal server error"
// @Router       /organizations/code/{code} [get]
func (h *OrganizationHandler) GetOrganizationByCode(c echo.Context) error {
	code := c.Param("code")
	if code == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Organization code is required")
	}

	org, err := h.orgService.GetOrganizationByCode(c.Request().Context(), code)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, org)
}

// UpdateOrganization handles organization updates
// @Summary      Update organization
// @Description  Updates organization details. Requires admin permissions for the organization.
// @Tags         Admin, Organizations
// @Accept       json
// @Produce      json
// @Param        id path string true "Organization ID"
// @Param        organization body dto.UpdateOrganizationRequest true "Updated Organization Details"
// @Security     BearerAuth
// @Success      200 {object} dto.OrganizationResponse "Organization updated successfully"
// @Failure      400 {object} apperror.AppError "Invalid request payload"
// @Failure      403 {object} apperror.AppError "Insufficient permissions"
// @Failure      404 {object} apperror.AppError "Organization not found"
// @Failure      500 {object} apperror.AppError "Internal server error"
// @Router       /admin/organizations/{id} [put]
func (h *OrganizationHandler) UpdateOrganization(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid organization ID format")
	}

	var req dto.UpdateOrganizationRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	// Get user ID from context
	userID, ok := c.Get(constant.UserIDKey).(uuid.UUID)
	if !ok {
		return apperror.NewAppError(http.StatusUnauthorized, "User ID not found in context", nil)
	}

	org, err := h.orgService.UpdateOrganization(c.Request().Context(), id, req, userID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, org)
}

// DeleteOrganization handles organization deletion
// @Summary      Delete organization
// @Description  Deletes an organization. Requires platform level permissions.
// @Tags         Admin, Organizations
// @Accept       json
// @Produce      json
// @Param        id path string true "Organization ID"
// @Security     BearerAuth
// @Success      204 "Organization deleted successfully"
// @Failure      400 {object} apperror.AppError "Invalid request"
// @Failure      403 {object} apperror.AppError "Insufficient permissions"
// @Failure      404 {object} apperror.AppError "Organization not found"
// @Failure      500 {object} apperror.AppError "Internal server error"
// @Router       /admin/organizations/{id} [delete]
func (h *OrganizationHandler) DeleteOrganization(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid organization ID format")
	}

	// Get user ID from context
	userID, ok := c.Get(constant.UserIDKey).(uuid.UUID)
	if !ok {
		return apperror.NewAppError(http.StatusUnauthorized, "User ID not found in context", nil)
	}

	if err := h.orgService.DeleteOrganization(c.Request().Context(), id, userID); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

// ListOrganizations handles organization listing with filters
// @Summary      List organizations
// @Description  Retrieves organizations with optional filters
// @Tags         Organizations
// @Accept       json
// @Produce      json
// @Param        type query string false "Organization Type" Enums(platform, company, store)
// @Param        parent_id query string false "Parent Organization ID"
// @Param        active query boolean false "Is Active"
// @Param        page query int false "Page number" default(1)
// @Param        limit query int false "Items per page" default(10)
// @Security     BearerAuth
// @Success      200 {array} dto.OrganizationResponse "Organizations retrieved successfully"
// @Failure      400 {object} apperror.AppError "Invalid query parameters"
// @Failure      500 {object} apperror.AppError "Internal server error"
// @Router       /organizations [get]
func (h *OrganizationHandler) ListOrganizations(c echo.Context) error {
	var req dto.ListOrganizationsRequest

	// Parse query parameters
	req.OrganizationType = c.QueryParam("type")
	req.Search = c.QueryParam("search")

	if parentIDStr := c.QueryParam("parent_id"); parentIDStr != "" {
		parentID, err := uuid.Parse(parentIDStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid parent organization ID format")
		}
		req.ParentOrganizationID = &parentID
	}

	if activeStr := c.QueryParam("active"); activeStr != "" {
		active, err := strconv.ParseBool(activeStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid active parameter format")
		}
		req.IsActive = &active
	}

	// Parse pagination
	if pageStr := c.QueryParam("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			req.Page = 1
		} else {
			req.Page = page
		}
	} else {
		req.Page = 1
	}

	if limitStr := c.QueryParam("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 100 {
			req.Limit = 10
		} else {
			req.Limit = limit
		}
	} else {
		req.Limit = 10
	}

	// Get current user ID from JWT middleware context for filtering
	currentUserID, ok := c.Get(constant.UserIDKey).(uuid.UUID)
	if !ok {
		return apperror.NewAppError(http.StatusUnauthorized, constant.ErrMsgAuthorizationContextMissing, nil)
	}

	// Create context with current user ID for hierarchical filtering
	ctx := context.WithValue(c.Request().Context(), "current_user_id", currentUserID)

	response, err := h.orgService.ListOrganizations(ctx, req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, response)
}

// JoinOrganization handles user joining an organization
// @Summary      Join organization
// @Description  Allows a user to join an organization using organization code
// @Tags         Organizations, Users
// @Accept       json
// @Produce      json
// @Param        request body dto.JoinOrganizationRequest true "Join Organization Request"
// @Security     BearerAuth
// @Success      201 {object} dto.UserOrganizationResponse "Successfully joined organization"
// @Failure      400 {object} apperror.AppError "Invalid request payload"
// @Failure      404 {object} apperror.AppError "Organization not found"
// @Failure      409 {object} apperror.AppError "User already member of organization"
// @Failure      500 {object} apperror.AppError "Internal server error"
// @Router       /organizations/join [post]
func (h *OrganizationHandler) JoinOrganization(c echo.Context) error {
	var req dto.JoinOrganizationRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	// Get user ID from context
	userID, ok := c.Get(constant.UserIDKey).(uuid.UUID)
	if !ok {
		return apperror.NewAppError(http.StatusUnauthorized, "User ID not found in context", nil)
	}

	userOrg, err := h.orgService.JoinOrganization(c.Request().Context(), userID, req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, userOrg)
}

// LeaveOrganization handles user leaving an organization
// @Summary      Leave organization
// @Description  Allows a user to leave an organization
// @Tags         Organizations, Users
// @Accept       json
// @Produce      json
// @Param        id path string true "Organization ID"
// @Security     BearerAuth
// @Success      204 "Successfully left organization"
// @Failure      400 {object} apperror.AppError "Invalid organization ID"
// @Failure      404 {object} apperror.AppError "Organization not found"
// @Failure      500 {object} apperror.AppError "Internal server error"
// @Router       /organizations/{id}/leave [delete]
func (h *OrganizationHandler) LeaveOrganization(c echo.Context) error {
	idStr := c.Param("id")
	orgID, err := uuid.Parse(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid organization ID format")
	}

	// Get user ID from context
	userID, ok := c.Get(constant.UserIDKey).(uuid.UUID)
	if !ok {
		return apperror.NewAppError(http.StatusUnauthorized, "User ID not found in context", nil)
	}

	if err := h.orgService.LeaveOrganization(c.Request().Context(), userID, orgID); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

// GetUserOrganizations handles retrieving user's organizations
// @Summary      Get user organizations
// @Description  Retrieves all organizations for the current user
// @Tags         Organizations, Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {array} dto.UserOrganizationResponse "User organizations retrieved successfully"
// @Failure      500 {object} apperror.AppError "Internal server error"
// @Router       /users/me/organizations [get]
func (h *OrganizationHandler) GetUserOrganizations(c echo.Context) error {
	// Get user ID from context
	userID, ok := c.Get(constant.UserIDKey).(uuid.UUID)
	if !ok {
		return apperror.NewAppError(http.StatusUnauthorized, "User ID not found in context", nil)
	}

	userOrgs, err := h.orgService.GetUserOrganizations(c.Request().Context(), userID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, userOrgs)
}

// GetOrganizationMembers handles retrieving organization members
// @Summary      Get organization members
// @Description  Retrieves all members of an organization
// @Tags         Admin, Organizations
// @Accept       json
// @Produce      json
// @Param        id path string true "Organization ID"
// @Security     BearerAuth
// @Success      200 {array} dto.UserOrganizationResponse "Organization members retrieved successfully"
// @Failure      400 {object} apperror.AppError "Invalid organization ID"
// @Failure      403 {object} apperror.AppError "Insufficient permissions"
// @Failure      500 {object} apperror.AppError "Internal server error"
// @Router       /admin/organizations/{id}/members [get]
func (h *OrganizationHandler) GetOrganizationMembers(c echo.Context) error {
	idStr := c.Param("id")
	orgID, err := uuid.Parse(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid organization ID format")
	}

	members, err := h.orgService.GetOrganizationMembers(c.Request().Context(), orgID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, members)
}

// CreateCompleteOrganizationStructure handles creation of complete organization structure for new users
// @Summary      Create complete organization structure
// @Description  Creates holding->company->store structure and assigns user. Platform admin only.
// @Tags         Admin, Organizations
// @Accept       json
// @Produce      json
// @Param        structure body dto.CreateCompleteStructureRequest true "Complete Structure Details"
// @Security     BearerAuth
// @Success      201 {object} dto.CompleteOrganizationStructureResponse "Complete structure created successfully"
// @Failure      400 {object} apperror.AppError "Invalid request payload"
// @Failure      403 {object} apperror.AppError "Insufficient permissions - Platform level required"
// @Failure      500 {object} apperror.AppError "Internal server error"
// @Router       /admin/organizations/complete-structure [post]
func (h *OrganizationHandler) CreateCompleteOrganizationStructure(c echo.Context) error {
	var req dto.CreateCompleteStructureRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	// Get user ID from context (creator must be Platform level)
	userID, ok := c.Get(constant.UserIDKey).(uuid.UUID)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found")
	}

	result, err := h.orgService.CreateCompleteOrganizationStructure(c.Request().Context(), req, userID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, result)
}

// GetOrganizationStatistics handles retrieving organization statistics based on user level
// @Summary      Get organization statistics
// @Description  Retrieves organization statistics filtered by user level (platform shows all, company shows only stores, etc)
// @Tags         Organizations, Statistics
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} dto.OrganizationStatisticsResponse "Organization statistics retrieved successfully"
// @Failure      401 {object} apperror.AppError "Unauthorized"
// @Failure      500 {object} apperror.AppError "Internal server error"
// @Router       /organizations/statistics [get]
func (h *OrganizationHandler) GetOrganizationStatistics(c echo.Context) error {
	// Get user ID from context
	userID, ok := c.Get(constant.UserIDKey).(uuid.UUID)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
	}

	statistics, err := h.orgService.GetOrganizationStatistics(c.Request().Context(), userID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, statistics)
}
