package repository

import (
	"go-base-project/internal/dto"
	"go-base-project/internal/model"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type organizationRepository struct {
	db *gorm.DB
}

// NewOrganizationRepository creates a new organization repository instance
func NewOrganizationRepository(db *gorm.DB) OrganizationRepositoryInterface {
	return &organizationRepository{db: db}
}

// Create creates a new organization
func (r *organizationRepository) Create(ctx context.Context, org *model.Organization) (*model.Organization, error) {
	if err := r.db.WithContext(ctx).Create(org).Error; err != nil {
		return nil, err
	}
	return org, nil
}

// FindByID finds an organization by ID with relationships
func (r *organizationRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Organization, error) {
	var org model.Organization
	err := r.db.WithContext(ctx).
		Preload("ParentOrganization").
		Preload("ChildOrganizations").
		Where("id = ?", id).
		First(&org).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}

// FindByIDWithDetails finds an organization by ID with detailed relationships
func (r *organizationRepository) FindByIDWithDetails(ctx context.Context, id uuid.UUID) (*model.Organization, error) {
	var org model.Organization
	err := r.db.WithContext(ctx).
		Preload("ParentOrganization").
		Preload("ChildOrganizations").
		Preload("UserOrganizations").
		Preload("UserOrganizations.User").
		Where("id = ?", id).
		First(&org).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}

// FindByCode finds an organization by code
func (r *organizationRepository) FindByCode(ctx context.Context, code string) (*model.Organization, error) {
	var org model.Organization
	err := r.db.WithContext(ctx).
		Preload("ParentOrganization").
		Preload("ChildOrganizations").
		Where("code = ?", code).
		First(&org).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}

// Update updates an existing organization
func (r *organizationRepository) Update(ctx context.Context, org *model.Organization) (*model.Organization, error) {
	if err := r.db.WithContext(ctx).Save(org).Error; err != nil {
		return nil, err
	}
	return org, nil
}

// Delete soft deletes an organization
func (r *organizationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Organization{}).Error
}

// FindAll finds all organizations with filters
func (r *organizationRepository) FindAll(ctx context.Context, filters map[string]interface{}) ([]model.Organization, error) {
	var orgs []model.Organization
	query := r.db.WithContext(ctx).Preload("ParentOrganization").Preload("ChildOrganizations")

	for key, value := range filters {
		query = query.Where(fmt.Sprintf("%s = ?", key), value)
	}

	err := query.Find(&orgs).Error
	return orgs, err
}

// List finds organizations with pagination and search
func (r *organizationRepository) List(ctx context.Context, page, limit int, search string, filters map[string]interface{}) ([]model.Organization, int64, error) {
	var orgs []model.Organization
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Organization{})

	// Apply filters
	for key, value := range filters {
		if key == "accessible_org_ids" {
			// Special handling for organization ID filtering
			if orgIDs, ok := value.([]uuid.UUID); ok && len(orgIDs) > 0 {
				query = query.Where("id IN ?", orgIDs)
			}
		} else {
			query = query.Where(fmt.Sprintf("%s = ?", key), value)
		}
	}

	// Apply search
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("name ILIKE ? OR code ILIKE ? OR description ILIKE ?",
			searchPattern, searchPattern, searchPattern)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	offset := (page - 1) * limit
	err := query.
		Preload("ParentOrganization").
		Preload("ChildOrganizations").
		Offset(offset).
		Limit(limit).
		Find(&orgs).Error

	return orgs, total, err
}

// Count counts organizations with filters
func (r *organizationRepository) Count(ctx context.Context, filters map[string]interface{}) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&model.Organization{})

	for key, value := range filters {
		if key == "accessible_org_ids" {
			// Special handling for organization ID filtering
			if orgIDs, ok := value.([]uuid.UUID); ok && len(orgIDs) > 0 {
				query = query.Where("id IN ?", orgIDs)
			}
		} else {
			query = query.Where(fmt.Sprintf("%s = ?", key), value)
		}
	}

	err := query.Count(&count).Error
	return count, err
}

// FindByType finds organizations by type
func (r *organizationRepository) FindByType(ctx context.Context, orgType string) ([]model.Organization, error) {
	var orgs []model.Organization
	err := r.db.WithContext(ctx).
		Preload("ParentOrganization").
		Preload("ChildOrganizations").
		Where("type = ?", orgType).
		Find(&orgs).Error
	return orgs, err
}

// FindByParent finds direct children of a parent organization
func (r *organizationRepository) FindByParent(ctx context.Context, parentID uuid.UUID) ([]model.Organization, error) {
	var orgs []model.Organization
	err := r.db.WithContext(ctx).
		Preload("ParentOrganization").
		Preload("ChildOrganizations").
		Where("parent_organization_id = ?", parentID).
		Find(&orgs).Error
	return orgs, err
}

// FindRootOrganizations finds all root organizations (no parent)
func (r *organizationRepository) FindRootOrganizations(ctx context.Context) ([]model.Organization, error) {
	var orgs []model.Organization
	err := r.db.WithContext(ctx).
		Preload("ParentOrganization").
		Preload("ChildOrganizations").
		Where("parent_organization_id IS NULL").
		Find(&orgs).Error
	return orgs, err
}

// FindActiveOrganizations finds all active organizations
func (r *organizationRepository) FindActiveOrganizations(ctx context.Context) ([]model.Organization, error) {
	var orgs []model.Organization
	err := r.db.WithContext(ctx).
		Preload("ParentOrganization").
		Preload("ChildOrganizations").
		Where("status = ?", "active").
		Find(&orgs).Error
	return orgs, err
}

// GetOrganizationHierarchy gets organization hierarchy starting from given organization
func (r *organizationRepository) GetOrganizationHierarchy(ctx context.Context, rootID uuid.UUID) ([]model.Organization, error) {
	var hierarchy []model.Organization

	// Get the root organization
	var rootOrg model.Organization
	err := r.db.WithContext(ctx).
		Preload("ParentOrganization").
		Preload("ChildOrganizations").
		Where("id = ?", rootID).
		First(&rootOrg).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("root organization not found")
		}
		return nil, err
	}

	hierarchy = append(hierarchy, rootOrg)

	// Get all descendants recursively
	children, err := r.GetChildrenRecursive(ctx, rootID)
	if err != nil {
		return nil, err
	}

	hierarchy = append(hierarchy, children...)
	return hierarchy, nil
}

// GetChildrenRecursive gets all children recursively for a parent organization
func (r *organizationRepository) GetChildrenRecursive(ctx context.Context, parentID uuid.UUID) ([]model.Organization, error) {
	var allChildren []model.Organization

	// Get direct children
	var directChildren []model.Organization
	err := r.db.WithContext(ctx).
		Preload("ParentOrganization").
		Where("parent_organization_id = ?", parentID).
		Find(&directChildren).Error
	if err != nil {
		return nil, err
	}

	// Add direct children to result
	allChildren = append(allChildren, directChildren...)

	// Recursively get children of each direct child
	for _, child := range directChildren {
		grandChildren, err := r.GetChildrenRecursive(ctx, child.ID)
		if err != nil {
			return nil, err
		}
		allChildren = append(allChildren, grandChildren...)
	}

	return allChildren, nil
}

// GetParentChain gets the parent chain for an organization
func (r *organizationRepository) GetParentChain(ctx context.Context, orgID uuid.UUID) ([]model.Organization, error) {
	var parentChain []model.Organization

	// Start with the given organization
	var org model.Organization
	err := r.db.WithContext(ctx).
		Preload("ParentOrganization").
		Where("id = ?", orgID).
		First(&org).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("organization not found")
		}
		return nil, err
	}

	// Traverse up the parent chain
	currentParentID := org.ParentOrganizationID
	for currentParentID != nil {
		var parent model.Organization
		err := r.db.WithContext(ctx).
			Preload("ParentOrganization").
			Where("id = ?", *currentParentID).
			First(&parent).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				break // Parent not found, end chain
			}
			return nil, err
		}

		parentChain = append(parentChain, parent)
		currentParentID = parent.ParentOrganizationID

		// Prevent infinite loops (circular references)
		if len(parentChain) > 50 {
			return nil, errors.New("circular reference detected in parent chain")
		}
	}

	return parentChain, nil
}

// AddUserToOrganization adds a user to an organization with a role
func (r *organizationRepository) AddUserToOrganization(ctx context.Context, userOrg *model.UserOrganization) error {
	return r.db.WithContext(ctx).Create(userOrg).Error
}

// RemoveUserFromOrganization removes a user from an organization
func (r *organizationRepository) RemoveUserFromOrganization(ctx context.Context, userID, orgID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND organization_id = ?", userID, orgID).
		Delete(&model.UserOrganization{}).Error
}

// UpdateUserOrganizationRole updates a user's role in an organization
func (r *organizationRepository) UpdateUserOrganizationRole(ctx context.Context, userID, orgID uuid.UUID, role string) error {
	return r.db.WithContext(ctx).
		Model(&model.UserOrganization{}).
		Where("user_id = ? AND organization_id = ?", userID, orgID).
		Update("role", role).Error
}

// FindUserOrganizations finds all organizations for a user
func (r *organizationRepository) FindUserOrganizations(ctx context.Context, userID uuid.UUID) ([]model.UserOrganization, error) {
	var userOrgs []model.UserOrganization
	err := r.db.WithContext(ctx).
		Preload("Organization").
		Preload("User").
		Where("user_id = ?", userID).
		Find(&userOrgs).Error
	return userOrgs, err
}

// FindOrganizationUsers finds all users for an organization
func (r *organizationRepository) FindOrganizationUsers(ctx context.Context, orgID uuid.UUID) ([]model.UserOrganization, error) {
	var userOrgs []model.UserOrganization
	err := r.db.WithContext(ctx).
		Preload("Organization").
		Preload("User").
		Where("organization_id = ?", orgID).
		Find(&userOrgs).Error
	return userOrgs, err
}

// FindActiveUserOrganizations finds all active organizations for a user
func (r *organizationRepository) FindActiveUserOrganizations(ctx context.Context, userID uuid.UUID) ([]model.UserOrganization, error) {
	var userOrgs []model.UserOrganization
	err := r.db.WithContext(ctx).
		Preload("Organization").
		Preload("User").
		Where("user_id = ? AND status = ?", userID, "active").
		Find(&userOrgs).Error
	return userOrgs, err
}

// FindActiveOrganizationUsers finds all active users for an organization
func (r *organizationRepository) FindActiveOrganizationUsers(ctx context.Context, orgID uuid.UUID) ([]model.UserOrganization, error) {
	var userOrgs []model.UserOrganization
	err := r.db.WithContext(ctx).
		Preload("Organization").
		Preload("User").
		Where("organization_id = ? AND status = ?", orgID, "active").
		Find(&userOrgs).Error
	return userOrgs, err
}

// CheckCodeExists checks if an organization code already exists
func (r *organizationRepository) CheckCodeExists(ctx context.Context, code string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Organization{}).
		Where("code = ?", code).
		Count(&count).Error
	return count > 0, err
}

// CheckCodeExistsExcluding checks if an organization code exists excluding a specific ID
func (r *organizationRepository) CheckCodeExistsExcluding(ctx context.Context, code string, excludeID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Organization{}).
		Where("code = ? AND id != ?", code, excludeID).
		Count(&count).Error
	return count > 0, err
}

// GetAllExistingCodes gets all existing organization codes
func (r *organizationRepository) GetAllExistingCodes(ctx context.Context) ([]string, error) {
	var codes []string
	err := r.db.WithContext(ctx).
		Model(&model.Organization{}).
		Pluck("code", &codes).Error
	return codes, err
}

// GetOrganizationStats gets organization statistics
func (r *organizationRepository) GetOrganizationStats(ctx context.Context, orgID uuid.UUID) (*dto.OrganizationStatsResponse, error) {
	var stats dto.OrganizationStatsResponse

	// Get basic organization info
	var org model.Organization
	err := r.db.WithContext(ctx).Where("id = ?", orgID).First(&org).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("organization not found")
		}
		return nil, err
	}

	// Count direct children
	var directChildren int64
	err = r.db.WithContext(ctx).
		Model(&model.Organization{}).
		Where("parent_organization_id = ?", orgID).
		Count(&directChildren).Error
	if err != nil {
		return nil, err
	}
	stats.DirectChildren = int(directChildren)

	// Count total descendants (recursive)
	descendants, err := r.GetChildrenRecursive(ctx, orgID)
	if err != nil {
		return nil, err
	}
	stats.TotalDescendants = len(descendants)

	// Count total members in organization
	var totalMembers int64
	err = r.db.WithContext(ctx).
		Model(&model.UserOrganization{}).
		Where("organization_id = ?", orgID).
		Count(&totalMembers).Error
	if err != nil {
		return nil, err
	}
	stats.TotalMembers = int(totalMembers)

	// Count active members in organization
	var activeMembers int64
	err = r.db.WithContext(ctx).
		Model(&model.UserOrganization{}).
		Where("organization_id = ? AND status = ?", orgID, "active").
		Count(&activeMembers).Error
	if err != nil {
		return nil, err
	}
	stats.ActiveMembers = int(activeMembers)

	// Calculate organization level (depth from root)
	parentChain, err := r.GetParentChain(ctx, orgID)
	if err != nil {
		return nil, err
	}
	stats.OrganizationLevel = len(parentChain)

	return &stats, nil
}

// GetHierarchyStats gets overall hierarchy statistics
func (r *organizationRepository) GetHierarchyStats(ctx context.Context) (*dto.HierarchyStatsResponse, error) {
	var stats dto.HierarchyStatsResponse

	// Count total organizations
	var totalOrgs int64
	err := r.db.WithContext(ctx).
		Model(&model.Organization{}).
		Count(&totalOrgs).Error
	if err != nil {
		return nil, err
	}
	stats.TotalOrganizations = int(totalOrgs)

	// Count active organizations
	var activeOrgs int64
	err = r.db.WithContext(ctx).
		Model(&model.Organization{}).
		Where("status = ?", "active").
		Count(&activeOrgs).Error
	if err != nil {
		return nil, err
	}
	stats.ActiveOrganizations = int(activeOrgs)

	// Count by type
	var typeCounts []struct {
		Type  string
		Count int64
	}

	err = r.db.WithContext(ctx).
		Model(&model.Organization{}).
		Select("type, COUNT(*) as count").
		Group("type").
		Scan(&typeCounts).Error
	if err != nil {
		return nil, err
	}

	// Map type counts
	for _, tc := range typeCounts {
		switch tc.Type {
		case "holding":
			stats.HoldingCount = int(tc.Count)
		case "company":
			stats.CompanyCount = int(tc.Count)
		case "store":
			stats.StoreCount = int(tc.Count)
		}
	}

	// Calculate max depth
	maxDepthInt64 := int64(0)
	rootOrgs, err := r.FindRootOrganizations(ctx)
	if err != nil {
		return nil, err
	}

	for _, root := range rootOrgs {
		depth, err := r.calculateMaxDepth(ctx, root.ID, 1)
		if err != nil {
			return nil, err
		}
		if depth > maxDepthInt64 {
			maxDepthInt64 = depth
		}
	}
	stats.MaxDepth = int(maxDepthInt64)

	return &stats, nil
}

// calculateMaxDepth calculates the maximum depth in the hierarchy from a given root
func (r *organizationRepository) calculateMaxDepth(ctx context.Context, orgID uuid.UUID, currentDepth int64) (int64, error) {
	children, err := r.FindByParent(ctx, orgID)
	if err != nil {
		return 0, err
	}

	if len(children) == 0 {
		return currentDepth, nil
	}

	maxChildDepth := currentDepth
	for _, child := range children {
		childDepth, err := r.calculateMaxDepth(ctx, child.ID, currentDepth+1)
		if err != nil {
			return 0, err
		}
		if childDepth > maxChildDepth {
			maxChildDepth = childDepth
		}
	}

	return maxChildDepth, nil
}
