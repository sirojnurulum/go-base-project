package repository

import (
	"go-base-project/internal/model"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type roleRepository struct {
	db *gorm.DB
}

// Create implements RoleRepository.
func (r *roleRepository) Create(ctx context.Context, role *model.Role) (*model.Role, error) {
	if err := r.db.WithContext(ctx).Create(role).Error; err != nil {
		return nil, err
	}
	return role, nil
}

// FindAll returns all roles with their permissions.
func (r *roleRepository) FindAll(ctx context.Context) ([]model.Role, error) {
	var roles []model.Role
	if err := r.db.WithContext(ctx).Preload("Permissions").Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

// FindAllActive returns all active roles with their permissions.
func (r *roleRepository) FindAllActive(ctx context.Context) ([]model.Role, error) {
	var roles []model.Role
	if err := r.db.WithContext(ctx).Preload("Permissions").Where("is_active = true").Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

// FindSystemRoles returns all system roles with their permissions.
func (r *roleRepository) FindSystemRoles(ctx context.Context) ([]model.Role, error) {
	var roles []model.Role
	if err := r.db.WithContext(ctx).Preload("Permissions").Where("is_system_role = true").Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

func NewRoleRepository(db *gorm.DB) RoleRepositoryInterface {
	return &roleRepository{db: db}
}

// FindByName finds a role by name (renamed from FindRoleByName for consistency).
func (r *roleRepository) FindByName(ctx context.Context, name string) (*model.Role, error) {
	var role model.Role
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

// FindByNameWithPermissions finds a role by name with permissions preloaded.
func (r *roleRepository) FindByNameWithPermissions(ctx context.Context, name string) (*model.Role, error) {
	var role model.Role
	if err := r.db.WithContext(ctx).Preload("Permissions").Where("name = ?", name).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Role, error) {
	var role model.Role
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

// FindByIDWithPermissions finds a role by ID with permissions preloaded.
func (r *roleRepository) FindByIDWithPermissions(ctx context.Context, id uuid.UUID) (*model.Role, error) {
	var role model.Role
	if err := r.db.WithContext(ctx).Preload("Permissions").Where("id = ?", id).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

// List returns paginated roles with search functionality.
func (r *roleRepository) List(ctx context.Context, offset, limit int, search string) ([]model.Role, error) {
	var roles []model.Role
	query := r.db.WithContext(ctx).Preload("Permissions")

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("name ILIKE ? OR description ILIKE ?", searchPattern, searchPattern)
	}

	err := query.Offset(offset).Limit(limit).Find(&roles).Error
	return roles, err
}

// Count returns the total count of roles with search filter.
func (r *roleRepository) Count(ctx context.Context, search string) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&model.Role{})

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("name ILIKE ? OR description ILIKE ?", searchPattern, searchPattern)
	}

	err := query.Count(&count).Error
	return count, err
}

// Update updates an existing role.
func (r *roleRepository) Update(ctx context.Context, role *model.Role) (*model.Role, error) {
	if err := r.db.WithContext(ctx).Save(role).Error; err != nil {
		return nil, err
	}
	return role, nil
}

// Delete soft deletes a role by ID.
func (r *roleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Role{}, id).Error
}

func (r *roleRepository) FindPermissionsByRoleID(ctx context.Context, roleID uuid.UUID) ([]string, error) {
	var permissions []string
	err := r.db.WithContext(ctx).Table("permissions").
		Select("permissions.name").
		Joins("join role_permissions on role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id = ?", roleID).
		Pluck("name", &permissions).Error
	return permissions, err
}

func (r *roleRepository) UpdateRolePermissions(ctx context.Context, roleID uuid.UUID, permissionNames []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var role model.Role
		if err := tx.First(&role, roleID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("role with id %s not found", roleID)
			}
			return err
		}

		var permissions []model.Permission
		if len(permissionNames) > 0 {
			if err := tx.Where("name IN ?", permissionNames).Find(&permissions).Error; err != nil {
				return fmt.Errorf("failed to find permissions: %w", err)
			}
		}

		// GORM's Association mode is cleaner and safer. It handles clearing old
		// associations and adding new ones automatically.
		if err := tx.Model(&role).Association("Permissions").Replace(&permissions); err != nil {
			return fmt.Errorf("failed to update role permissions: %w", err)
		}

		return nil
	})
}

// CreateRoleApproval creates a new role approval request.
func (r *roleRepository) CreateRoleApproval(ctx context.Context, approval *model.RoleApproval) (*model.RoleApproval, error) {
	if err := r.db.WithContext(ctx).Create(approval).Error; err != nil {
		return nil, err
	}
	return approval, nil
}

// FindAllRoleApprovals returns all role approval requests with user information.
func (r *roleRepository) FindAllRoleApprovals(ctx context.Context) ([]model.RoleApproval, error) {
	var approvals []model.RoleApproval
	if err := r.db.WithContext(ctx).
		Preload("RequestedByUser").
		Preload("Approver").
		Preload("Organization").
		Find(&approvals).Error; err != nil {
		return nil, err
	}
	return approvals, nil
}

// FindRoleApprovalsByStatus finds role approvals by status.
func (r *roleRepository) FindRoleApprovalsByStatus(ctx context.Context, status string) ([]model.RoleApproval, error) {
	var approvals []model.RoleApproval
	if err := r.db.WithContext(ctx).
		Preload("RequestedByUser").
		Preload("Approver").
		Preload("Organization").
		Where("status = ?", status).
		Find(&approvals).Error; err != nil {
		return nil, err
	}
	return approvals, nil
}

// FindRoleApprovalsByRequester finds role approvals by requester ID.
func (r *roleRepository) FindRoleApprovalsByRequester(ctx context.Context, requesterID uuid.UUID) ([]model.RoleApproval, error) {
	var approvals []model.RoleApproval
	if err := r.db.WithContext(ctx).
		Preload("RequestedByUser").
		Preload("Approver").
		Preload("Organization").
		Where("requested_by = ?", requesterID).
		Find(&approvals).Error; err != nil {
		return nil, err
	}
	return approvals, nil
}

// FindRoleApprovalByID finds a role approval request by ID.
func (r *roleRepository) FindRoleApprovalByID(ctx context.Context, id uuid.UUID) (*model.RoleApproval, error) {
	var approval model.RoleApproval
	if err := r.db.WithContext(ctx).
		Preload("RequestedByUser").
		Preload("Approver").
		Preload("Organization").
		Where("id = ?", id).
		First(&approval).Error; err != nil {
		return nil, err
	}
	return &approval, nil
}

// UpdateRoleApproval updates an existing role approval request.
func (r *roleRepository) UpdateRoleApproval(ctx context.Context, approval *model.RoleApproval) (*model.RoleApproval, error) {
	if err := r.db.WithContext(ctx).Save(approval).Error; err != nil {
		return nil, err
	}
	return approval, nil
}

// Permission CRUD operations

// GetAllPermissions returns all permissions from the database.
func (r *roleRepository) GetAllPermissions(ctx context.Context) ([]model.Permission, error) {
	var permissions []model.Permission
	if err := r.db.WithContext(ctx).Order("name ASC").Find(&permissions).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}

// CreatePermission creates a new permission in the database.
func (r *roleRepository) CreatePermission(ctx context.Context, permission *model.Permission) error {
	if err := r.db.WithContext(ctx).Create(permission).Error; err != nil {
		return err
	}
	return nil
}

// FindPermissionByName finds a permission by its name.
func (r *roleRepository) FindPermissionByName(ctx context.Context, name string) (*model.Permission, error) {
	var permission model.Permission
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&permission).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil for not found instead of error
		}
		return nil, err
	}
	return &permission, nil
}

// FindPermissionByID finds a permission by its ID.
func (r *roleRepository) FindPermissionByID(ctx context.Context, id uuid.UUID) (*model.Permission, error) {
	var permission model.Permission
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&permission).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil for not found instead of error
		}
		return nil, err
	}
	return &permission, nil
}

// UpdatePermission updates an existing permission.
func (r *roleRepository) UpdatePermission(ctx context.Context, permission *model.Permission) error {
	if err := r.db.WithContext(ctx).Save(permission).Error; err != nil {
		return err
	}
	return nil
}

// DeletePermission deletes a permission from the database.
func (r *roleRepository) DeletePermission(ctx context.Context, id uuid.UUID) error {
	// First check if permission is being used by any roles
	var count int64
	if err := r.db.WithContext(ctx).Table("role_permissions").
		Joins("JOIN permissions ON role_permissions.permission_id = permissions.id").
		Where("permissions.id = ?", id).
		Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return errors.New("cannot delete permission: it is currently assigned to one or more roles")
	}

	// Soft delete the permission
	if err := r.db.WithContext(ctx).Delete(&model.Permission{}, id).Error; err != nil {
		return err
	}
	return nil
}

// CheckPermissionExists checks if a permission with the given name exists.
func (r *roleRepository) CheckPermissionExists(ctx context.Context, name string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Permission{}).Where("name = ?", name).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// Organization-specific role methods implementation

// FindRolesByOrganizationType returns roles that are applicable to a specific organization type.
func (r *roleRepository) FindRolesByOrganizationType(ctx context.Context, organizationType string) ([]model.Role, error) {
	var roles []model.Role

	query := r.db.WithContext(ctx).
		Preload("Permissions").
		Preload("OrganizationTypes").
		Joins("INNER JOIN role_organization_types rot ON roles.id = rot.role_id").
		Where("rot.organization_type = ? AND roles.is_active = true", organizationType).
		Order("roles.level ASC, roles.name ASC")

	if err := query.Find(&roles).Error; err != nil {
		return nil, fmt.Errorf("failed to find roles for organization type %s: %w", organizationType, err)
	}

	return roles, nil
}

// CreateRoleOrganizationTypes creates organization type mappings for a role.
func (r *roleRepository) CreateRoleOrganizationTypes(ctx context.Context, roleID uuid.UUID, organizationTypes []string) error {
	if len(organizationTypes) == 0 {
		return nil
	}

	var roleOrgTypes []model.RoleOrganizationType
	for _, orgType := range organizationTypes {
		roleOrgTypes = append(roleOrgTypes, model.RoleOrganizationType{
			RoleID:           roleID,
			OrganizationType: orgType,
		})
	}

	if err := r.db.WithContext(ctx).Create(&roleOrgTypes).Error; err != nil {
		return fmt.Errorf("failed to create role organization types: %w", err)
	}

	return nil
}

// UpdateRoleOrganizationTypes updates organization type mappings for a role.
func (r *roleRepository) UpdateRoleOrganizationTypes(ctx context.Context, roleID uuid.UUID, organizationTypes []string) error {
	// Start transaction
	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Delete existing mappings
	if err := tx.Where("role_id = ?", roleID).Delete(&model.RoleOrganizationType{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete existing role organization types: %w", err)
	}

	// Create new mappings if any
	if len(organizationTypes) > 0 {
		var roleOrgTypes []model.RoleOrganizationType
		for _, orgType := range organizationTypes {
			roleOrgTypes = append(roleOrgTypes, model.RoleOrganizationType{
				RoleID:           roleID,
				OrganizationType: orgType,
			})
		}

		if err := tx.Create(&roleOrgTypes).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create new role organization types: %w", err)
		}
	}

	return tx.Commit().Error
}

// DeleteRoleOrganizationTypes deletes all organization type mappings for a role.
func (r *roleRepository) DeleteRoleOrganizationTypes(ctx context.Context, roleID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Where("role_id = ?", roleID).Delete(&model.RoleOrganizationType{}).Error; err != nil {
		return fmt.Errorf("failed to delete role organization types: %w", err)
	}
	return nil
}

// FindOrganizationTypesByRoleID returns all organization types for a specific role.
func (r *roleRepository) FindOrganizationTypesByRoleID(ctx context.Context, roleID uuid.UUID) ([]string, error) {
	var roleOrgTypes []model.RoleOrganizationType

	if err := r.db.WithContext(ctx).Where("role_id = ?", roleID).Find(&roleOrgTypes).Error; err != nil {
		return nil, fmt.Errorf("failed to find organization types for role %s: %w", roleID, err)
	}

	var organizationTypes []string
	for _, rot := range roleOrgTypes {
		organizationTypes = append(organizationTypes, rot.OrganizationType)
	}

	return organizationTypes, nil
}

// IsRoleSuperAdmin checks if a role is a super admin role.
// Super admin is identified by role name "super_admin" or predefined_name "super_admin"
// or by having level >= 100 (platform level).
func (r *roleRepository) IsRoleSuperAdmin(ctx context.Context, roleID uuid.UUID) (bool, error) {
	var role model.Role
	err := r.db.WithContext(ctx).
		Select("name, predefined_name, level").
		Where("id = ?", roleID).
		First(&role).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil // Role not found, definitely not super admin
		}
		return false, fmt.Errorf("failed to fetch role: %w", err)
	}

	// Check if role is super admin based on:
	// 1. Role name is "super_admin"
	// 2. Predefined name is "super_admin"
	// 3. Role level >= 100 (platform level)
	return role.Name == "super_admin" ||
		role.PredefinedName == "super_admin" ||
		role.Level >= 100, nil
}
