package handler

import (
	"beresin-backend/internal/apperror"
	"beresin-backend/internal/config"
	"beresin-backend/internal/constant"
	"beresin-backend/internal/dto"
	"beresin-backend/internal/service"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
	"github.com/rs/zerolog/log"
)

const refreshTokenCookiePath = "/api/auth"

// AuthHandler handles HTTP requests related to authentication.
type AuthHandler struct {
	authService       service.AuthService
	googleOauthConfig *oauth2.Config
	frontendURL       string
	cfg               config.Config
}

// NewAuthHandler creates a new instance of AuthHandler.
func NewAuthHandler(authService service.AuthService, oauthConfig *oauth2.Config, cfg config.Config) *AuthHandler {
	return &AuthHandler{
		authService:       authService,
		googleOauthConfig: oauthConfig,
		frontendURL:       cfg.FrontendURL,
		cfg:               cfg,
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

	loginResult, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		return err // Serahkan ke error handler terpusat
	}

	// Atur refresh token di dalam cookie HttpOnly yang aman
	cookie := new(http.Cookie)
	cookie.Name = "refresh_token"
	cookie.Value = loginResult.RefreshToken
	cookie.Expires = time.Now().Add(7 * 24 * time.Hour)
	cookie.Path = refreshTokenCookiePath // Batasi scope cookie ke path otentikasi
	cookie.HttpOnly = true
	// cookie.Secure = true // Aktifkan di produksi (membutuhkan HTTPS)
	cookie.SameSite = http.SameSiteLaxMode

	c.SetCookie(cookie)

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
	newAccessToken, newRefreshToken, err := h.authService.RefreshToken(refreshToken)
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
	// Kita harus membuat cookie baru secara eksplisit karena atribut seperti Path, HttpOnly, dll.
	// tidak dikirim oleh browser dalam request, sehingga tidak ada di 'cookie' yang kita baca.
	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    newRefreshToken,
		Expires:  time.Now().Add(7 * 24 * time.Hour), // Set ulang masa berlaku
		Path:     refreshTokenCookiePath,             // Set path secara eksplisit
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

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
	url := h.googleOauthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
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
	token, err := h.googleOauthConfig.Exchange(context.Background(), code)
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
	loginResult, err := h.authService.LoginWithGoogle(userInfo)
	if err != nil {
		log.Error().Err(err).Msg("LoginWithGoogle service failed")
		return c.Redirect(http.StatusTemporaryRedirect, errorRedirectURL)
	}

	// Atur refresh token di cookie
	cookie := new(http.Cookie)
	cookie.Name = "refresh_token"
	cookie.Value = loginResult.RefreshToken
	cookie.Expires = time.Now().Add(7 * 24 * time.Hour)
	cookie.Path = refreshTokenCookiePath
	cookie.HttpOnly = true
	cookie.SameSite = http.SameSiteLaxMode
	c.SetCookie(cookie)

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
	if err := h.authService.Logout(cookie.Value); err != nil {
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
