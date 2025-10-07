package dto

// LoginRequest adalah DTO (Data Transfer Object) untuk request login.
type LoginRequest struct {
	Username string `json:"username" validate:"required" example:"admin"`
	Password string `json:"password" validate:"required" example:"password"`
}

// LoginResponse adalah DTO untuk response login yang dikirim ke client.
type LoginResponse struct {
	AccessToken string       `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	Message     string       `json:"message" example:"Login successful"`
	User        UserResponse `json:"user"` // Menggunakan UserResponse DTO
	Permissions []string     `json:"permissions"`
}

// RefreshTokenResponse adalah DTO untuk response refresh token.
type RefreshTokenResponse struct {
	AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// GoogleUserInfo adalah DTO untuk menyimpan informasi dari Google.
type GoogleUserInfo struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Picture string `json:"picture"`
	Name    string `json:"name"`
}

// LoginResult adalah DTO internal yang dikembalikan oleh service ke handler.
// Ini memisahkan data mentah (termasuk refresh token) dari response API publik.
type LoginResult struct {
	AccessToken  string
	RefreshToken string
	User         *UserResponse
	Permissions  []string
}

// SwitchOrganizationRequest adalah DTO untuk request switch organization context.
type SwitchOrganizationRequest struct {
	OrganizationID string `json:"organization_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// SwitchOrganizationResponse adalah DTO untuk response switch organization context.
type SwitchOrganizationResponse struct {
	AccessToken    string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	OrganizationID string `json:"organization_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Message        string `json:"message" example:"Organization context switched successfully"`
}

// SwitchOrganizationResult adalah DTO internal untuk hasil switch organization.
type SwitchOrganizationResult struct {
	AccessToken    string
	OrganizationID string
}
