package constant

// User-facing messages for API responses.
const (
	// Success Messages
	MsgLoginSuccess     = "Login successful"
	MsgLogoutSuccess    = "Logged out successfully"
	MsgAlreadyLoggedOut = "Already logged out"
	MsgRolePermsUpdated = "Role permissions updated successfully"
	MsgUserDeleted      = "User deleted successfully"
	MsgUserUpdated      = "User updated successfully"
	MsgWelcomeAdmin     = "Welcome to the admin dashboard!"
	MsgAuthenticated    = "authenticated"
	MsgStatusOK         = "ok"

	// Role Approval Messages
	MsgRoleApprovalCreated = "Role approval request created successfully"
	MsgRoleRequestApproved = "Role request approved successfully"
	MsgRoleRequestRejected = "Role request rejected successfully"

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

	// User Management Security Messages
	ErrMsgCannotChangeOwnRole         = "Users cannot change their own role"
	ErrMsgInsufficientAuthorityLevel  = "Insufficient authority to assign this role level"
	ErrMsgOnlySuperAdminCanModify     = "Only super administrators can modify super administrator accounts"
	ErrMsgOnlySuperAdminCanAssign     = "Only super administrators can assign super administrator role"
	ErrMsgAuthorizationContextMissing = "Authorization context not found"
	ErrMsgCurrentUserHasNoRole        = "Current user has no role assigned"

	// Role Approval Error Messages
	ErrMsgInvalidApprovalID        = "Invalid approval request ID format"
	ErrMsgApprovalNotFound         = "Approval request not found"
	ErrMsgApprovalAlreadyProcessed = "Approval request has already been processed"

	// Organization Context Error Messages
	ErrMsgUserNotFoundInContext       = "User not found in request context"
	ErrMsgOrganizationContextRequired = "Organization context is required for this operation"
	ErrMsgOrganizationAccessDenied    = "Access denied to the specified organization"
	ErrMsgInvalidOrganizationIDFormat = "Invalid organization ID format"
)
