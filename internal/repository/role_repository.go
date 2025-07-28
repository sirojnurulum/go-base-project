package repository

import (
	"beresin-backend/internal/model"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type roleRepository struct {
	db *gorm.DB
}

// Create implements RoleRepository.
func (r *roleRepository) Create(role *model.Role) (*model.Role, error) {
	if err := r.db.Create(role).Error; err != nil {
		return nil, err
	}
	return role, nil
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) FindRoleByName(name string) (*model.Role, error) {
	var role model.Role
	if err := r.db.Where("name = ?", name).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) FindPermissionsByRoleID(roleID uuid.UUID) ([]string, error) {
	var permissions []string
	err := r.db.Table("permissions").
		Select("permissions.name").
		Joins("join role_permissions on role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id = ?", roleID).
		Pluck("name", &permissions).Error
	return permissions, err
}

func (r *roleRepository) UpdateRolePermissions(roleID uuid.UUID, permissionNames []string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
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
