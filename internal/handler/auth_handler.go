package handler

import (
	"go-base-project/internal/apperror"
	"go-base-project/internal/config"
	"go-base-project/internal/constant"
	"go-base-project/internal/dto"
	"go-base-project/internal/service"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
)

const refreshTokenCookiePath = "/api/auth"

// AuthHandler handles HTTP requests related to authentication.
type AuthHandler struct {
	authService service.AuthServiceInterface
	oauthConfig *oauth2.Config
	cfg         config.Config
	jwtSecret   string
	frontendURL string
}

// NewAuthHandler creates a new instance of AuthHandler.
func NewAuthHandler(authService service.AuthServiceInterface, oauthConfig *oauth2.Config, cfg config.Config) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		oauthConfig: oauthConfig,
		cfg:         cfg,
		jwtSecret:   cfg.JWTSecret,
		frontendURL: cfg.FrontendURL,
	}
}

// Login
// @Summary      User Login
// @Description  Authenticates a user and returns an access token. The refresh token is set in an HttpOnly cookie.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        credentials body dto.LoginRequest true "Login Credentials"
// @Success      200 {object} dto.LoginResponse "Login successful"
// @Failure      400 {object} apperror.AppError "Invalid request payload"
// @Failure      401 {object} apperror.AppError "Invalid credentials"
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c echo.Context) error {
	var req dto.LoginRequest

	// Bind dan validasi request dalam satu langkah
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, constant.ErrMsgInvalidRequestFormat)
	}
	if err := c.Validate(&req); err != nil {
		return err // Error sudah dalam format HTTPError dari custom validator
	}

	loginResult, err := h.authService.Login(c.Request().Context(), req.Username, req.Password)
	if err != nil {
		return err // Serahkan ke error handler terpusat
	}

	// Atur refresh token di dalam cookie HttpOnly yang aman
	h.setRefreshTokenCookie(c, loginResult.RefreshToken)

	return c.JSON(http.StatusOK, dto.LoginResponse{
		AccessToken: loginResult.AccessToken,
		User:        *loginResult.User,
		Message:     constant.MsgLoginSuccess,
		Permissions: loginResult.Permissions,
	})
}

// RefreshToken
// @Summary      Refresh Access Token
// @Description  Generates a new access token using a valid refresh token from the cookie.
// @Tags         Auth
// @Produce      json
// @Success      200 {object} dto.RefreshTokenResponse "Token refreshed successfully"
// @Failure      401 {object} apperror.AppError "Unauthorized or invalid refresh token"
// @Router       /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c echo.Context) error {
	cookie, err := c.Cookie("refresh_token")
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, constant.ErrMsgUnauthorized)
	}

	refreshToken := cookie.Value
	// Terima tiga nilai balik dari service: token akses baru, token refresh baru, dan error.
	newAccessToken, newRefreshToken, err := h.authService.RefreshToken(c.Request().Context(), refreshToken)
	if err != nil {
		// Juga, hapus cookie yang mungkin tidak valid lagi.
		c.SetCookie(&http.Cookie{
			Name:    "refresh_token",
			Value:   "",
			Path:    refreshTokenCookiePath,
			Expires: time.Unix(0, 0),
		})
		return err // Serahkan ke error handler terpusat
	}

	// Atur refresh token yang BARU ke dalam cookie. Ini adalah inti dari token rotation.
	h.setRefreshTokenCookie(c, newRefreshToken)

	return c.JSON(http.StatusOK, dto.RefreshTokenResponse{
		AccessToken: newAccessToken,
	})
}

// GoogleLogin
// @Summary      Google Login
// @Description  Redirects the user to Google's authentication page.
// @Tags         Auth
// @Router       /auth/google/login [get]
func (h *AuthHandler) GoogleLogin(c echo.Context) error {
	// Dapatkan URL otentikasi dari konfigurasi OAuth2
	url := h.oauthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
	return c.Redirect(http.StatusTemporaryRedirect, url)
}

// GoogleCallback handles the callback from Google after successful authentication.
// @Summary      Google Callback
// @Description  Handles the callback from Google after successful authentication. This endpoint is not intended to be called directly by users.
// @Tags         Auth
// @Router       /auth/google/callback [get]
func (h *AuthHandler) GoogleCallback(c echo.Context) error {
	errorRedirectURL := fmt.Sprintf("%s/login?error=true", h.frontendURL)

	// Ambil kode otentikasi dari query parameter
	code := c.QueryParam("code")
	if code == "" {
		return c.Redirect(http.StatusTemporaryRedirect, errorRedirectURL)
	}

	// Tukar kode dengan token
	token, err := h.oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		log.Error().Err(err).Msg("Google OAuth exchange failed")
		return c.Redirect(http.StatusTemporaryRedirect, errorRedirectURL)
	}

	// Dapatkan informasi pengguna dari Google API
	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user info from Google")
		return c.Redirect(http.StatusTemporaryRedirect, errorRedirectURL)
	}
	defer response.Body.Close()

	var userInfo dto.GoogleUserInfo
	if err := json.NewDecoder(response.Body).Decode(&userInfo); err != nil {
		log.Error().Err(err).Msg("Failed to decode Google user info")
		return c.Redirect(http.StatusTemporaryRedirect, errorRedirectURL)
	}

	// Lakukan proses login/registrasi di service
	loginResult, err := h.authService.LoginWithGoogle(c.Request().Context(), userInfo)
	if err != nil {
		log.Error().Err(err).Msg("LoginWithGoogle service failed")
		return c.Redirect(http.StatusTemporaryRedirect, errorRedirectURL)
	}

	// Atur refresh token di cookie
	h.setRefreshTokenCookie(c, loginResult.RefreshToken)

	// Redirect ke frontend dengan access token
	redirectURL := fmt.Sprintf("%s/auth/google/callback?access_token=%s", h.frontendURL, loginResult.AccessToken)
	return c.Redirect(http.StatusPermanentRedirect, redirectURL)
}

// Logout handles user logout by invalidating the refresh token.
// @Summary      Logout user
// @Description  Invalidates the refresh token and clears the session cookie.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Success      200 {object} map[string]string "Logout successful message"
// @Failure      400 {object} apperror.AppError "Failed to read refresh token"
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c echo.Context) error {
	cookie, err := c.Cookie("refresh_token")
	if err != nil {
		// If the cookie is not found, the user is already effectively logged out.
		if errors.Is(err, http.ErrNoCookie) {
			return c.JSON(http.StatusOK, map[string]string{"message": constant.MsgAlreadyLoggedOut})
		}
		return apperror.NewAppError(http.StatusBadRequest, constant.ErrMsgFailedReadRefreshToken, err)
	}

	// Invalidate the token in the backend (e.g., delete from Redis)
	if err := h.authService.Logout(c.Request().Context(), cookie.Value); err != nil {
		// The service layer already logs the error. We proceed to clear the cookie.
	}

	// Clear the cookie on the client side by setting an expired one.
	expiredCookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     refreshTokenCookiePath,
		Expires:  time.Unix(0, 0), // Expire immediately
		HttpOnly: true,
		Secure:   h.cfg.Env == "production", // Use secure cookies in production
		SameSite: http.SameSiteStrictMode,
	}
	c.SetCookie(expiredCookie)

	return c.JSON(http.StatusOK, map[string]string{"message": constant.MsgLogoutSuccess})
}

// GetCurrentUser
// @Summary      Get Current User Information
// @Description  Returns the current authenticated user's information with their permissions.
// @Tags         Auth
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} dto.LoginResponse "Current user information"
// @Failure      401 {object} apperror.AppError "Unauthorized"
// @Router       /auth/me [get]
func (h *AuthHandler) GetCurrentUser(c echo.Context) error {
	// Get user ID from JWT context
	userID := c.Get(constant.UserIDKey)
	if userID == nil {
		return apperror.NewAppError(http.StatusUnauthorized, constant.ErrMsgUnauthorized, nil)
	}

	// Get user information with permissions from service
	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		return apperror.NewAppError(http.StatusUnauthorized, "Invalid user ID format", nil)
	}

	loginResult, err := h.authService.GetUserWithPermissions(c.Request().Context(), userUUID.String())
	if err != nil {
		log.Error().Err(err).Msg("Failed to get current user information")
		return err
	}

	// Return user information with permissions (similar to login response)
	response := &dto.LoginResponse{
		Message:     "User information retrieved successfully",
		User:        *loginResult.User,
		Permissions: loginResult.Permissions,
		// Note: We don't return the access token here since the user already has it
	}

	return c.JSON(http.StatusOK, response)
}

// SwitchOrganization
// @Summary      Switch Organization Context
// @Description  Switches the user's organization context and returns a new access token with the organization context
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.SwitchOrganizationRequest true "Organization switch request"
// @Success      200 {object} dto.SwitchOrganizationResponse "Organization switched successfully"
// @Failure      400 {object} apperror.AppError "Bad request"
// @Failure      401 {object} apperror.AppError "Unauthorized"
// @Failure      403 {object} apperror.AppError "Forbidden - No access to organization"
// @Router       /auth/switch-organization [post]
func (h *AuthHandler) SwitchOrganization(c echo.Context) error {
	// Get user ID and role ID from JWT context
	userID := c.Get(constant.UserIDKey)
	roleID := c.Get(constant.RoleIDKey)

	if userID == nil || roleID == nil {
		return apperror.NewAppError(http.StatusUnauthorized, constant.ErrMsgUnauthorized, nil)
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		return apperror.NewAppError(http.StatusUnauthorized, "Invalid user ID format", nil)
	}

	roleUUID, ok := roleID.(uuid.UUID)
	if !ok {
		return apperror.NewAppError(http.StatusUnauthorized, "Invalid role ID format", nil)
	}

	// Parse request body
	var req dto.SwitchOrganizationRequest
	if err := c.Bind(&req); err != nil {
		return apperror.NewAppError(http.StatusBadRequest, constant.ErrMsgInvalidRequestFormat, err)
	}

	// Switch organization context and get new token
	switchResult, err := h.authService.SwitchOrganizationContext(c.Request().Context(), userUUID, roleUUID, req.OrganizationID)
	if err != nil {
		var appError *apperror.AppError
		if errors.As(err, &appError) {
			return appError
		}
		log.Error().Err(err).Msg("Failed to switch organization context")
		return apperror.NewAppError(http.StatusInternalServerError, "Failed to switch organization context", err)
	}

	// Return new access token with organization context
	response := &dto.SwitchOrganizationResponse{
		AccessToken:    switchResult.AccessToken,
		OrganizationID: switchResult.OrganizationID,
		Message:        "Organization context switched successfully",
	}

	return c.JSON(http.StatusOK, response)
}

// setRefreshTokenCookie is a helper method to set refresh token cookie with consistent settings
func (h *AuthHandler) setRefreshTokenCookie(c echo.Context, refreshToken string) {
	cookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		Path:     refreshTokenCookiePath,
		HttpOnly: true,
		Secure:   h.cfg.Env == "production",
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(cookie)
}
