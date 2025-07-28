package handler_test

import (
	"beresin-backend/internal/apperror"
	"beresin-backend/internal/config"
	"beresin-backend/internal/constant"
	"beresin-backend/internal/dto"
	"beresin-backend/internal/handler"
	"beresin-backend/internal/validator"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/oauth2"
)

// MockAuthService is a mock for the AuthService interface.
type MockAuthService struct {
	mock.Mock
}

// Implement the AuthService interface for the mock
func (m *MockAuthService) Login(username, password string) (*dto.LoginResult, error) {
	args := m.Called(username, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.LoginResult), args.Error(1)
}

func (m *MockAuthService) RefreshToken(tokenString string) (string, string, error) {
	args := m.Called(tokenString)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockAuthService) LoginWithGoogle(userInfo dto.GoogleUserInfo) (*dto.LoginResult, error) {
	args := m.Called(userInfo)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.LoginResult), args.Error(1)
}

func (m *MockAuthService) Logout(refreshToken string) error {
	args := m.Called(refreshToken)
	return args.Error(0)
}

func TestLogin(t *testing.T) {
	// Setup
	mockService := new(MockAuthService)
	cfg := config.Config{} // Dummy config
	authHandler := handler.NewAuthHandler(mockService, &oauth2.Config{}, cfg)
	e := echo.New()
	e.Validator = validator.NewCustomValidator()

	t.Run("Success", func(t *testing.T) {
		reqDTO := dto.LoginRequest{
			Username: "testuser",
			Password: "password",
		}
		jsonReq, _ := json.Marshal(reqDTO)

		loginResult := &dto.LoginResult{
			AccessToken:  "fake_access_token",
			RefreshToken: "fake_refresh_token",
			User: &dto.UserResponse{
				ID:       uuid.New(),
				Username: "testuser",
				Email:    "test@example.com",
				Role:     "user",
			},
			Permissions: []string{"dashboard:view"},
		}

		mockService.On("Login", reqDTO.Username, reqDTO.Password).Return(loginResult, nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(string(jsonReq)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(t, authHandler.Login(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)

			var resBody dto.LoginResponse
			err := json.Unmarshal(rec.Body.Bytes(), &resBody)
			assert.NoError(t, err)
			assert.Equal(t, loginResult.AccessToken, resBody.AccessToken)
			assert.Equal(t, loginResult.User.Username, resBody.User.Username)
			assert.Equal(t, constant.MsgLoginSuccess, resBody.Message)

			cookie := rec.Result().Cookies()[0]
			assert.Equal(t, "refresh_token", cookie.Name)
			assert.Equal(t, loginResult.RefreshToken, cookie.Value)
			assert.True(t, cookie.HttpOnly)
			assert.Equal(t, "/api/auth", cookie.Path)
		}
		mockService.AssertExpectations(t)
	})

	t.Run("Invalid Credentials", func(t *testing.T) {
		reqDTO := dto.LoginRequest{Username: "wronguser", Password: "wrongpassword"}
		jsonReq, _ := json.Marshal(reqDTO)
		expectedErr := apperror.NewUnauthorizedError(constant.ErrMsgInvalidCredentials)
		mockService.On("Login", reqDTO.Username, reqDTO.Password).Return(nil, expectedErr).Once()

		req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(string(jsonReq)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := authHandler.Login(c)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockService.AssertExpectations(t)
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		reqDTO := dto.LoginRequest{Username: "testuser", Password: "password"}
		jsonReq, _ := json.Marshal(reqDTO)
		expectedErr := apperror.NewInternalError(errors.New("database error"))
		mockService.On("Login", reqDTO.Username, reqDTO.Password).Return(nil, expectedErr).Once()

		req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(string(jsonReq)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := authHandler.Login(c)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockService.AssertExpectations(t)
	})
}

// TestGoogleLoginFlow menguji seluruh alur Google OAuth2 dari perspektif handler.
func TestGoogleLoginFlow(t *testing.T) {
	// Setup
	mockService := new(MockAuthService)
	cfg := config.Config{
		FrontendURL: "http://localhost:3000",
	}

	// 1. Mock server Google untuk endpoint Token dan UserInfo
	mockGoogleServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "mock_google_access_token",
				"token_type":   "Bearer",
			})
		} else if r.URL.Path == "/userinfo" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(dto.GoogleUserInfo{
				ID:      "google-user-123",
				Email:   "test.user@google.com",
				Picture: "http://example.com/pic.jpg",
				Name:    "Test User",
			})
		} else {
			http.NotFound(w, r)
		}
	}))
	defer mockGoogleServer.Close()

	// 2. Intercept panggilan HTTP ke URL user info Google yang di-hardcode di handler.
	// Ini adalah teknik untuk menguji kode yang memiliki dependensi eksternal yang tidak dapat dikonfigurasi.
	// Transport asli disimpan untuk dipulihkan setelah tes selesai.
	transportToRestore := http.DefaultClient.Transport

	// http.DefaultClient memiliki Transport nil secara default, dan metode http.Client.Do
	// akan menggunakan http.DefaultTransport jika Transport-nya nil. Kita perlu meniru perilaku ini.
	effectiveTransport := http.DefaultClient.Transport
	if effectiveTransport == nil {
		effectiveTransport = http.DefaultTransport
	}

	http.DefaultClient.Transport = &mockTransport{
		mockUserInfoURL:     "https://www.googleapis.com/oauth2/v2/userinfo",
		redirectedServerURL: mockGoogleServer.URL + "/userinfo",
		originalTransport:   effectiveTransport,
	}
	defer func() {
		http.DefaultClient.Transport = transportToRestore
	}()

	// Konfigurasi oauth2 client untuk menggunakan mock server kita
	oauthConfig := &oauth2.Config{
		ClientID:     "fake-client-id",
		ClientSecret: "fake-client-secret",
		RedirectURL:  "http://localhost:8080/api/auth/google/callback",
		Scopes:       []string{"profile", "email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: mockGoogleServer.URL + "/token", // Arahkan ke mock server
		},
	}

	authHandler := handler.NewAuthHandler(mockService, oauthConfig, cfg)
	e := echo.New()

	t.Run("GoogleLogin - Success Redirect", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/auth/google/login", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := authHandler.GoogleLogin(c)

		assert.NoError(t, err, "Handler seharusnya tidak mengembalikan error")
		assert.Equal(t, http.StatusTemporaryRedirect, rec.Code, "Harus redirect sementara")
		location := rec.Header().Get("Location")
		assert.Contains(t, location, oauthConfig.Endpoint.AuthURL, "URL harus mengarah ke endpoint otentikasi Google")
		assert.Contains(t, location, "client_id=fake-client-id", "URL harus mengandung client_id")
		assert.Contains(t, location, "redirect_uri=", "URL harus mengandung redirect_uri")
	})

	t.Run("GoogleCallback - Success", func(t *testing.T) {
		expectedUserInfo := dto.GoogleUserInfo{
			ID:      "google-user-123",
			Email:   "test.user@google.com",
			Picture: "http://example.com/pic.jpg",
			Name:    "Test User",
		}
		loginResult := &dto.LoginResult{
			AccessToken:  "final_access_token",
			RefreshToken: "final_refresh_token",
		}
		mockService.On("LoginWithGoogle", expectedUserInfo).Return(loginResult, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/auth/google/callback?code=fake-auth-code", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := authHandler.GoogleCallback(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusPermanentRedirect, rec.Code, "Harus redirect permanen ke frontend")

		// Assertion 1: URL redirect ke frontend benar dan mengandung access token
		locationURL, _ := url.Parse(rec.Header().Get("Location"))
		assert.Equal(t, "http", locationURL.Scheme)
		assert.Equal(t, "localhost:3000", locationURL.Host)
		assert.Equal(t, "/auth/google/callback", locationURL.Path)
		assert.Equal(t, loginResult.AccessToken, locationURL.Query().Get("access_token"), "Access token harus ada di query param")

		// Assertion 2: Cookie refresh token diatur dengan benar
		cookie := rec.Result().Cookies()[0]
		assert.Equal(t, "refresh_token", cookie.Name, "Nama cookie harus 'refresh_token'")
		assert.Equal(t, loginResult.RefreshToken, cookie.Value, "Value cookie harus sesuai")
		assert.True(t, cookie.HttpOnly, "Cookie harus HttpOnly")
		assert.Equal(t, "/api/auth", cookie.Path, "Path cookie harus dibatasi")

		mockService.AssertExpectations(t)
	})

	t.Run("GoogleCallback - Service Returns Error", func(t *testing.T) {
		expectedUserInfo := dto.GoogleUserInfo{
			ID:      "google-user-123",
			Email:   "test.user@google.com",
			Picture: "http://example.com/pic.jpg",
			Name:    "Test User",
		}
		mockService.On("LoginWithGoogle", expectedUserInfo).Return(nil, errors.New("internal service error")).Once()

		req := httptest.NewRequest(http.MethodGet, "/auth/google/callback?code=fake-auth-code", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := authHandler.GoogleCallback(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusTemporaryRedirect, rec.Code, "Harus redirect ke halaman error frontend")
		assert.Equal(t, "http://localhost:3000/login?error=true", rec.Header().Get("Location"))
		mockService.AssertExpectations(t)
	})
}

// mockTransport adalah helper untuk mencegat panggilan HTTP keluar.
type mockTransport struct {
	mockUserInfoURL     string
	redirectedServerURL string
	originalTransport   http.RoundTripper
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Jika request ditujukan ke URL user info Google, alihkan ke mock server kita.
	if strings.HasPrefix(req.URL.String(), m.mockUserInfoURL) {
		newReqURL, _ := url.Parse(m.redirectedServerURL)
		req.URL = newReqURL
	}
	// Untuk semua request lain (misal, token exchange), teruskan ke transport asli.
	return m.originalTransport.RoundTrip(req)
}

func TestLogout(t *testing.T) {
	mockService := new(MockAuthService)
	cfg := config.Config{} // Dummy config
	authHandler := handler.NewAuthHandler(mockService, &oauth2.Config{}, cfg)
	e := echo.New()

	t.Run("Success", func(t *testing.T) {
		refreshToken := "valid_refresh_token"
		mockService.On("Logout", refreshToken).Return(nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
		req.AddCookie(&http.Cookie{Name: "refresh_token", Value: refreshToken})
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(t, authHandler.Logout(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)

			var resBody map[string]string
			json.Unmarshal(rec.Body.Bytes(), &resBody)
			assert.Equal(t, constant.MsgLogoutSuccess, resBody["message"])

			cookie := rec.Result().Cookies()[0]
			assert.Equal(t, "refresh_token", cookie.Name)
			assert.Equal(t, "", cookie.Value)
			assert.True(t, cookie.Expires.Before(time.Now()))
		}
		mockService.AssertExpectations(t)
	})

	t.Run("No Cookie", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(t, authHandler.Logout(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			var resBody map[string]string
			json.Unmarshal(rec.Body.Bytes(), &resBody)
			assert.Equal(t, constant.MsgAlreadyLoggedOut, resBody["message"])
		}
	})
}

func TestRefreshToken(t *testing.T) {
	mockService := new(MockAuthService)
	cfg := config.Config{} // Dummy config
	authHandler := handler.NewAuthHandler(mockService, &oauth2.Config{}, cfg)
	e := echo.New()

	t.Run("Success", func(t *testing.T) {
		oldRefreshToken := "valid_refresh_token"
		newAccessToken := "new_access_token"
		newRefreshToken := "new_refresh_token"

		mockService.On("RefreshToken", oldRefreshToken).Return(newAccessToken, newRefreshToken, nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
		req.AddCookie(&http.Cookie{Name: "refresh_token", Value: oldRefreshToken, Path: "/api/auth"})
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(t, authHandler.RefreshToken(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)

			var resBody dto.RefreshTokenResponse
			json.Unmarshal(rec.Body.Bytes(), &resBody)
			assert.Equal(t, newAccessToken, resBody.AccessToken)

			cookie := rec.Result().Cookies()[0]
			assert.Equal(t, "refresh_token", cookie.Name)
			assert.Equal(t, newRefreshToken, cookie.Value)
			assert.Equal(t, "/api/auth", cookie.Path)
		}
		mockService.AssertExpectations(t)
	})

	t.Run("Invalid Token", func(t *testing.T) {
		invalidToken := "invalid_refresh_token"
		expectedErr := apperror.NewUnauthorizedError("invalid token")
		mockService.On("RefreshToken", invalidToken).Return("", "", expectedErr).Once()

		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
		req.AddCookie(&http.Cookie{Name: "refresh_token", Value: invalidToken})
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := authHandler.RefreshToken(c)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)

		// Check that the cookie was cleared even on error
		cookie := rec.Result().Cookies()[0]
		assert.Equal(t, "refresh_token", cookie.Name)
		assert.Equal(t, "", cookie.Value)
		assert.True(t, cookie.Expires.Before(time.Now()))
		mockService.AssertExpectations(t)
	})
}
