package handler

import (
	"beresin-backend/internal/constant"
	"beresin-backend/internal/dto"
	"beresin-backend/internal/service"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// OrganizationHandler handles HTTP requests related to organization management
type OrganizationHandler struct {
	organizationService service.OrganizationServiceInterface
}

// NewOrganizationHandler creates a new instance of OrganizationHandler
func NewOrganizationHandler(organizationService service.OrganizationServiceInterface) *OrganizationHandler {
	return &OrganizationHandler{
		organizationService: organizationService,
	}
}

// CreateOrganization handles organization creation (admin only)
// @Summary      Create a new organization
// @Description  Creates a new organization. Requires 'organizations:create' permission.
// @Tags         Admin, Organizations
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateOrganizationRequest true "Organization creation request"
// @Success      201 {object} model.Organization
// @Failure      400 {object} map[string]interface{}
// @Failure      401 {object} map[string]interface{}
// @Failure      403 {object} map[string]interface{}
// @Failure      500 {object} map[string]interface{}
// @Router       /admin/organizations [post]
func (h *OrganizationHandler) CreateOrganization(c echo.Context) error {
	var req dto.CreateOrganizationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid request body",
		})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		})
	}

	organization, err := h.organizationService.CreateOrganization(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, organization)
}

// GetAllOrganizations handles listing all organizations
// @Summary      Get all organizations
// @Description  Retrieves all organizations. Requires 'organizations:read' permission.
// @Tags         Organizations
// @Produce      json
// @Success      200 {array} dto.OrganizationResponse
// @Failure      401 {object} map[string]interface{}
// @Failure      403 {object} map[string]interface{}
// @Failure      500 {object} map[string]interface{}
// @Router       /api/organizations [get]
func (h *OrganizationHandler) GetAllOrganizations(c echo.Context) error {
	organizations, err := h.organizationService.GetAllOrganizations()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to retrieve organizations",
		})
	}

	return c.JSON(http.StatusOK, organizations)
}

// GetOrganizationByID handles retrieving an organization by ID
// @Summary      Get organization by ID
// @Description  Retrieves detailed organization information by ID. Requires 'organizations:read' permission.
// @Tags         Organizations
// @Produce      json
// @Param        id path string true "Organization ID"
// @Success      200 {object} dto.OrganizationDetailResponse
// @Failure      400 {object} map[string]interface{}
// @Failure      401 {object} map[string]interface{}
// @Failure      403 {object} map[string]interface{}
// @Failure      404 {object} map[string]interface{}
// @Failure      500 {object} map[string]interface{}
// @Router       /api/organizations/{id} [get]
func (h *OrganizationHandler) GetOrganizationByID(c echo.Context) error {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid organization ID format",
		})
	}

	organization, err := h.organizationService.GetOrganizationByID(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"error": "Organization not found",
		})
	}

	return c.JSON(http.StatusOK, organization)
}

// GetOrganizationByCode handles retrieving an organization by code
// @Summary      Get organization by code
// @Description  Retrieves organization information by code.
// @Tags         Organizations
// @Produce      json
// @Param        code path string true "Organization Code"
// @Success      200 {object} model.Organization
// @Failure      404 {object} map[string]interface{}
// @Failure      500 {object} map[string]interface{}
// @Router       /api/organizations/code/{code} [get]
func (h *OrganizationHandler) GetOrganizationByCode(c echo.Context) error {
	code := c.Param("code")
	if code == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Organization code is required",
		})
	}

	organization, err := h.organizationService.GetOrganizationByCode(code)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"error": "Organization not found",
		})
	}

	return c.JSON(http.StatusOK, organization)
}

// UpdateOrganization handles organization updates (admin only)
// @Summary      Update organization
// @Description  Updates organization information. Requires 'organizations:update' permission.
// @Tags         Admin, Organizations
// @Accept       json
// @Produce      json
// @Param        id path string true "Organization ID"
// @Param        request body dto.UpdateOrganizationRequest true "Organization update request"
// @Success      200 {object} model.Organization
// @Failure      400 {object} map[string]interface{}
// @Failure      401 {object} map[string]interface{}
// @Failure      403 {object} map[string]interface{}
// @Failure      404 {object} map[string]interface{}
// @Failure      500 {object} map[string]interface{}
// @Router       /admin/organizations/{id} [put]
func (h *OrganizationHandler) UpdateOrganization(c echo.Context) error {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid organization ID format",
		})
	}

	var req dto.UpdateOrganizationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid request body",
		})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		})
	}

	organization, err := h.organizationService.UpdateOrganization(id, &req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, organization)
}

// DeleteOrganization handles organization deletion (admin only)
// @Summary      Delete organization
// @Description  Deletes an organization. Requires 'organizations:delete' permission.
// @Tags         Admin, Organizations
// @Param        id path string true "Organization ID"
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} map[string]interface{}
// @Failure      401 {object} map[string]interface{}
// @Failure      403 {object} map[string]interface{}
// @Failure      404 {object} map[string]interface{}
// @Failure      500 {object} map[string]interface{}
// @Router       /admin/organizations/{id} [delete]
func (h *OrganizationHandler) DeleteOrganization(c echo.Context) error {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid organization ID format",
		})
	}

	err = h.organizationService.DeleteOrganization(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Organization deleted successfully",
	})
}

// JoinOrganization handles user joining an organization
// @Summary      Join organization
// @Description  Allows a user to join an organization using organization code.
// @Tags         Organizations
// @Accept       json
// @Produce      json
// @Param        request body dto.JoinOrganizationRequest true "Join organization request"
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} map[string]interface{}
// @Failure      401 {object} map[string]interface{}
// @Failure      404 {object} map[string]interface{}
// @Failure      500 {object} map[string]interface{}
// @Router       /api/organizations/join [post]
func (h *OrganizationHandler) JoinOrganization(c echo.Context) error {
	// Get user ID from JWT context
	userID, ok := c.Get(constant.UserIDKey).(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": "User not authenticated",
		})
	}

	var req dto.JoinOrganizationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid request body",
		})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		})
	}

	err := h.organizationService.JoinOrganization(userID, &req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Successfully joined organization",
	})
}

// LeaveOrganization handles user leaving an organization
// @Summary      Leave organization
// @Description  Allows a user to leave an organization.
// @Tags         Organizations
// @Param        id path string true "Organization ID"
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} map[string]interface{}
// @Failure      401 {object} map[string]interface{}
// @Failure      404 {object} map[string]interface{}
// @Failure      500 {object} map[string]interface{}
// @Router       /api/organizations/{id}/leave [delete]
func (h *OrganizationHandler) LeaveOrganization(c echo.Context) error {
	// Get user ID from JWT context
	userID, ok := c.Get(constant.UserIDKey).(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": "User not authenticated",
		})
	}

	idParam := c.Param("id")
	organizationID, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid organization ID format",
		})
	}

	err = h.organizationService.LeaveOrganization(userID, organizationID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Successfully left organization",
	})
}

// GetUserOrganizations handles retrieving user's organizations
// @Summary      Get user organizations
// @Description  Retrieves all organizations that the current user belongs to.
// @Tags         Organizations
// @Produce      json
// @Success      200 {array} dto.OrganizationResponse
// @Failure      401 {object} map[string]interface{}
// @Failure      500 {object} map[string]interface{}
// @Router       /api/users/me/organizations [get]
func (h *OrganizationHandler) GetUserOrganizations(c echo.Context) error {
	// Get user ID from JWT context
	userID, ok := c.Get(constant.UserIDKey).(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": "User not authenticated",
		})
	}

	organizations, err := h.organizationService.GetUserOrganizations(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to retrieve user organizations",
		})
	}

	return c.JSON(http.StatusOK, organizations)
}

// GetOrganizationMembers handles retrieving organization members (admin only)
// @Summary      Get organization members
// @Description  Retrieves all members of an organization. Requires 'organizations:manage_members' permission.
// @Tags         Admin, Organizations
// @Produce      json
// @Param        id path string true "Organization ID"
// @Success      200 {array} dto.OrganizationMember
// @Failure      400 {object} map[string]interface{}
// @Failure      401 {object} map[string]interface{}
// @Failure      403 {object} map[string]interface{}
// @Failure      404 {object} map[string]interface{}
// @Failure      500 {object} map[string]interface{}
// @Router       /admin/organizations/{id}/members [get]
func (h *OrganizationHandler) GetOrganizationMembers(c echo.Context) error {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid organization ID format",
		})
	}

	members, err := h.organizationService.GetOrganizationMembers(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to retrieve organization members",
		})
	}

	return c.JSON(http.StatusOK, members)
}
