package constant

// User-facing messages for API responses.
const (
	// Success Messages
	MsgLoginSuccess     = "Login successful"
	MsgLogoutSuccess    = "Logged out successfully"
	MsgAlreadyLoggedOut = "Already logged out"
	MsgRolePermsUpdated = "Role permissions updated successfully"
	MsgUserDeleted      = "User deleted successfully"
	MsgWelcomeAdmin     = "Welcome to the admin dashboard!"
	MsgAuthenticated    = "authenticated"
	MsgStatusOK         = "ok"

	// Error Messages
	ErrMsgInvalidCredentials      = "invalid credentials"
	ErrMsgUserNotFound            = "user not found"
	ErrMsgInvalidUserID           = "Invalid user ID format"
	ErrMsgInvalidRoleID           = "Invalid role ID format"
	ErrMsgInvalidRequestFormat    = "Invalid request format"
	ErrMsgFailedReadRefreshToken  = "Failed to read refresh token"
	ErrMsgUnauthorized            = "unauthorized"
	ErrMsgInsufficientPermissions = "insufficient permissions"
	ErrMsgRoleNotFoundInToken     = "role not found in token"
	ErrMsgInvalidRoleIDFormat     = "invalid role ID format"
	ErrMsgFailedTokenClaims       = "failed to get token claims"
	ErrMsgMissingAuthHeader       = "missing authorization header"
	ErrMsgInvalidOrExpiredToken   = "invalid or expired token"
)
