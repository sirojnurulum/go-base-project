package cache

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	// PermissionsCacheDuration adalah TTL default untuk cache izin peran.
	PermissionsCacheDuration = 15 * time.Minute
)

// GetRolePermissionsCacheKey menghasilkan kunci Redis untuk cache izin sebuah peran.
func GetRolePermissionsCacheKey(roleID uuid.UUID) string {
	return fmt.Sprintf("permissions:role:%s", roleID.String())
}
