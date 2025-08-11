package util

// ErrorMessages contains user-friendly error messages
var ErrorMessages = map[string]string{
	// User management security messages
	"cannot_change_own_role":          "Users cannot change their own role",
	"insufficient_authority_level":    "Insufficient authority to assign this role level",
	"only_superadmin_can_modify":      "Only super administrators can modify super administrator accounts",
	"only_superadmin_can_assign":      "Only super administrators can assign super administrator role",
	"authorization_context_not_found": "Authorization context not found",
	"current_user_no_role":            "Current user has no role assigned",
	"user_not_found":                  "User not found",
	"role_not_found":                  "Role not found",
	"organization_not_found":          "Organization not found",
	"organization_code_exists":        "Organization code already exists",
	"invalid_organization_hierarchy":  "Invalid organization hierarchy",
	"cannot_delete_org_with_children": "Cannot delete organization with children",
	"cannot_delete_org_with_members":  "Cannot delete organization with members",
	"user_already_member":             "User is already a member of this organization",
	"user_not_member":                 "User is not a member of this organization",
}

// GetUserFriendlyError returns a user-friendly error message
func GetUserFriendlyError(key string) string {
	if msg, exists := ErrorMessages[key]; exists {
		return msg
	}
	return "An unexpected error occurred"
}

// StructuredError represents a structured API error response
type StructuredError struct {
	Error   string            `json:"error"`
	Code    int               `json:"code,omitempty"`
	Details map[string]string `json:"details,omitempty"`
}

// NewStructuredError creates a new structured error
func NewStructuredError(message string, code int, details map[string]string) *StructuredError {
	return &StructuredError{
		Error:   message,
		Code:    code,
		Details: details,
	}
}
