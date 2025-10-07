package constant

// Organization type constants
const (
	OrganizationTypePlatform = "platform"
	OrganizationTypeHolding  = "holding"
	OrganizationTypeCompany  = "company"
	OrganizationTypeStore    = "store"
)

// Role level constants - Platform SaaS Multi-Tenant Architecture
const (
	// Level 100: Superadmin (only 1 user)
	RoleLevelSuperAdmin = 100 // Platform super admin - only one user allowed

	// Platform Level (76-99): Platform administrators, no organization required
	RoleLevelPlatformAdmin   = 99 // Platform admin - can see all holdings/companies
	RoleLevelPlatformManager = 76 // Platform manager - can see all holdings/companies

	// Holding Level (51-75): Multi-business owners managing multiple companies
	RoleLevelHoldingOwner   = 75 // Holding company owner/admin
	RoleLevelHoldingManager = 51 // Holding manager

	// Company Level (26-50): Individual business managers
	RoleLevelCompanyOwner   = 50 // Company owner/admin
	RoleLevelCompanyManager = 26 // Company manager

	// Store Level (1-25): Branch/outlet operators
	RoleLevelStoreManager = 25 // Store manager
	RoleLevelStoreStaff   = 1  // Store staff/regular user
)

// Role approval status constants
const (
	RoleApprovalStatusPending  = "pending"
	RoleApprovalStatusApproved = "approved"
	RoleApprovalStatusRejected = "rejected"
)

// Scan type constants
const (
	ScanTypeShip      = "ship"
	ScanTypeCancelAdd = "cancel_add"
)

// Scan result constants
const (
	ScanResultSuccess          = "success"
	ScanResultCancelledWarning = "cancelled_warning"
	ScanResultDuplicate        = "duplicate"
	ScanResultFailed           = "failed"
)
