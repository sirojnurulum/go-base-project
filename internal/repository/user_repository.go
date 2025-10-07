package repository

import (
	"go-base-project/internal/model"
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository instance
func NewUserRepository(db *gorm.DB) UserRepositoryInterface {
	return &userRepository{db: db}
}

// Create implements UserRepository.
func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// Update implements UserRepository.
func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByGoogleID(ctx context.Context, googleID string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("google_id = ?", googleID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByUsernameWithRole mencari pengguna berdasarkan username dan memuat relasi Role.
func (r *userRepository) FindByUsernameWithRole(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Preload("Role").Where("username = ?", username).First(&user).Error
	return &user, err
}

func (r *userRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	return &user, err
}

// FindByIDWithRole mencari pengguna berdasarkan ID dan memuat relasi Role.
func (r *userRepository) FindByIDWithRole(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Preload("Role").Where("id = ?", id).First(&user).Error
	return &user, err
}

// FindByIDWithRoleAndOrganizations mencari pengguna berdasarkan ID dan memuat relasi Role dan Organizations.
func (r *userRepository) FindByIDWithRoleAndOrganizations(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).
		Preload("Role").
		Preload("Organizations").
		Where("id = ?", id).
		First(&user).Error
	return &user, err
}

func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	return &user, err
}

// FindByOrganizationID mencari pengguna berdasarkan organization ID dengan paginasi.
func (r *userRepository) FindByOrganizationID(ctx context.Context, organizationID uuid.UUID, offset, limit int) ([]model.User, error) {
	var users []model.User
	err := r.db.WithContext(ctx).
		Preload("Role").
		Joins("JOIN user_organizations ON users.id = user_organizations.user_id").
		Where("user_organizations.organization_id = ? AND user_organizations.is_active = true", organizationID).
		Offset(offset).
		Limit(limit).
		Find(&users).Error
	return users, err
}

// List mengambil daftar pengguna dengan paginasi.
func (r *userRepository) List(ctx context.Context, offset, limit int, search string) ([]model.User, error) {
	var users []model.User
	query := r.db.WithContext(ctx).Preload("Role")
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("username ILIKE ? OR email ILIKE ?", searchPattern, searchPattern)
	}
	err := query.Offset(offset).Limit(limit).Find(&users).Error
	return users, err
}

// Count menghitung total jumlah pengguna dengan filter pencarian.
func (r *userRepository) Count(ctx context.Context, search string) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&model.User{})
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("username ILIKE ? OR email ILIKE ?", searchPattern, searchPattern)
	}
	err := query.Count(&count).Error
	return count, err
}

// Delete menghapus pengguna berdasarkan ID.
func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.User{}, id).Error
}

// CountUsersByRoleLevel menghitung jumlah user dengan role level tertentu.
func (r *userRepository) CountUsersByRoleLevel(ctx context.Context, level int) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.User{}).
		Joins("JOIN roles ON users.role_id = roles.id").
		Where("roles.level = ?", level).
		Count(&count).Error
	return count, err
}

// User-Organization Management Implementation

// CreateUserOrganization membuat user-organization relationship baru.
func (r *userRepository) CreateUserOrganization(ctx context.Context, userOrg *model.UserOrganization) (*model.UserOrganization, error) {
	err := r.db.WithContext(ctx).Create(userOrg).Error
	if err != nil {
		return nil, err
	}
	return userOrg, nil
}

// FindUserOrganization mencari user-organization relationship spesifik.
func (r *userRepository) FindUserOrganization(ctx context.Context, userID, organizationID uuid.UUID) (*model.UserOrganization, error) {
	var userOrg model.UserOrganization
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Organization").
		Where("user_id = ? AND organization_id = ?", userID, organizationID).
		First(&userOrg).Error
	return &userOrg, err
}

// FindUserOrganizationWithRole mencari user-organization relationship dengan role information.
func (r *userRepository) FindUserOrganizationWithRole(ctx context.Context, userID, organizationID uuid.UUID) (*model.UserOrganization, error) {
	var userOrg model.UserOrganization
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Organization").
		Preload("Role").
		Where("user_id = ? AND organization_id = ?", userID, organizationID).
		First(&userOrg).Error
	return &userOrg, err
}

// FindUserOrganizations mencari semua organization yang diikuti user dengan paginasi.
func (r *userRepository) FindUserOrganizations(ctx context.Context, userID uuid.UUID, offset, limit int) ([]model.UserOrganization, error) {
	var userOrgs []model.UserOrganization
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Organization").
		Preload("Role").
		Where("user_id = ? AND is_active = true", userID).
		Offset(offset).
		Limit(limit).
		Find(&userOrgs).Error
	return userOrgs, err
}

// FindOrganizationMembers mencari semua member dalam organization dengan paginasi.
func (r *userRepository) FindOrganizationMembers(ctx context.Context, organizationID uuid.UUID, offset, limit int) ([]model.UserOrganization, error) {
	var userOrgs []model.UserOrganization
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Organization").
		Where("organization_id = ? AND is_active = true", organizationID).
		Offset(offset).
		Limit(limit).
		Find(&userOrgs).Error
	return userOrgs, err
}

// FindOrganizationMembersWithRoles mencari semua member dalam organization dengan role information.
func (r *userRepository) FindOrganizationMembersWithRoles(ctx context.Context, organizationID uuid.UUID, offset, limit int) ([]model.UserOrganization, error) {
	var userOrgs []model.UserOrganization
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Organization").
		Preload("Role").
		Where("organization_id = ? AND is_active = true", organizationID).
		Offset(offset).
		Limit(limit).
		Find(&userOrgs).Error
	return userOrgs, err
}

// CountUserOrganizations menghitung total organization yang diikuti user.
func (r *userRepository) CountUserOrganizations(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.UserOrganization{}).
		Where("user_id = ? AND is_active = true", userID).
		Count(&count).Error
	return count, err
}

// CountOrganizationMembers menghitung total member dalam organization.
func (r *userRepository) CountOrganizationMembers(ctx context.Context, organizationID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.UserOrganization{}).
		Where("organization_id = ? AND is_active = true", organizationID).
		Count(&count).Error
	return count, err
}

// UpdateUserOrganization memperbarui user-organization relationship.
func (r *userRepository) UpdateUserOrganization(ctx context.Context, userOrg *model.UserOrganization) (*model.UserOrganization, error) {
	err := r.db.WithContext(ctx).Save(userOrg).Error
	if err != nil {
		return nil, err
	}
	return userOrg, nil
}

// UpdateUserOrganizationRole memperbarui role user dalam organization tertentu.
func (r *userRepository) UpdateUserOrganizationRole(ctx context.Context, userID, organizationID uuid.UUID, roleID *uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&model.UserOrganization{}).
		Where("user_id = ? AND organization_id = ?", userID, organizationID).
		Update("role_id", roleID).Error
}

// DeleteUserOrganization menghapus user-organization relationship.
func (r *userRepository) DeleteUserOrganization(ctx context.Context, userID, organizationID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND organization_id = ?", userID, organizationID).
		Delete(&model.UserOrganization{}).Error
}

// BulkCreateUserOrganizations membuat multiple user-organization relationships sekaligus.
func (r *userRepository) BulkCreateUserOrganizations(ctx context.Context, userOrgs []model.UserOrganization) ([]model.UserOrganization, []error) {
	var created []model.UserOrganization
	var errors []error

	// Use transaction for bulk operations
	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, userOrg := range userOrgs {
		if err := tx.Create(&userOrg).Error; err != nil {
			errors = append(errors, err)
		} else {
			created = append(created, userOrg)
		}
	}

	if len(errors) > 0 {
		tx.Rollback()
		return nil, errors
	}

	if err := tx.Commit().Error; err != nil {
		return nil, []error{err}
	}

	return created, nil
}

// User Organization History Implementation

// CreateUserOrganizationHistory creates a new user organization history record.
func (r *userRepository) CreateUserOrganizationHistory(ctx context.Context, history *model.UserOrganizationHistory) (*model.UserOrganizationHistory, error) {
	if err := r.db.WithContext(ctx).Create(history).Error; err != nil {
		return nil, err
	}
	return history, nil
}

// FindUserOrganizationHistory retrieves user organization history with pagination.
func (r *userRepository) FindUserOrganizationHistory(ctx context.Context, userID uuid.UUID, offset, limit int) ([]model.UserOrganizationHistory, error) {
	var history []model.UserOrganizationHistory

	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Organization").
		Preload("Actor").
		Where("user_id = ?", userID).
		Order("action_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&history).Error

	return history, err
}

// FindOrganizationUserHistory retrieves organization user history with pagination.
func (r *userRepository) FindOrganizationUserHistory(ctx context.Context, organizationID uuid.UUID, offset, limit int) ([]model.UserOrganizationHistory, error) {
	var history []model.UserOrganizationHistory

	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Organization").
		Preload("Actor").
		Where("organization_id = ?", organizationID).
		Order("action_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&history).Error

	return history, err
}

// CountUserOrganizationHistory counts total history records for a user.
func (r *userRepository) CountUserOrganizationHistory(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.UserOrganizationHistory{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}

// CountOrganizationUserHistory counts total history records for an organization.
func (r *userRepository) CountOrganizationUserHistory(ctx context.Context, organizationID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.UserOrganizationHistory{}).
		Where("organization_id = ?", organizationID).
		Count(&count).Error
	return count, err
}

// FindUserOrganizationHistoryByAction retrieves user organization history filtered by action.
func (r *userRepository) FindUserOrganizationHistoryByAction(ctx context.Context, userID uuid.UUID, action string, offset, limit int) ([]model.UserOrganizationHistory, error) {
	var history []model.UserOrganizationHistory

	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Organization").
		Preload("Actor").
		Where("user_id = ? AND action = ?", userID, action).
		Order("action_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&history).Error

	return history, err
}

// ListWithFilters retrieves users with level and organization filtering
func (r *userRepository) ListWithFilters(ctx context.Context, offset, limit int, search string, maxLevel int, organizationIDs []uuid.UUID) ([]model.User, error) {
	var users []model.User
	query := r.db.WithContext(ctx).
		Preload("Role").
		Select("DISTINCT users.*")

	// Apply search filter if provided
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("users.username ILIKE ? OR users.email ILIKE ?", searchPattern, searchPattern)
	}

	// Level filtering: only show users with roles having level < maxLevel
	// This ensures Platform Admin (level 99) cannot see Super Admin (level 100)
	query = query.Joins("LEFT JOIN roles ON users.role_id = roles.id").
		Where("roles.level < ? OR roles.level IS NULL", maxLevel)

	// Organization filtering for non-platform users
	if len(organizationIDs) > 0 {
		// For organization-level users, only show users from their accessible organizations
		// PLUS platform-level users (who don't have organization assignments)
		subQuery := r.db.Table("user_organizations").
			Select("DISTINCT user_id").
			Where("organization_id IN ? AND is_active = true", organizationIDs)

		query = query.Where("users.id IN (?) OR roles.level >= 76", subQuery)
	}

	err := query.Offset(offset).Limit(limit).Find(&users).Error
	return users, err
}

// CountWithFilters counts users with level and organization filtering
func (r *userRepository) CountWithFilters(ctx context.Context, search string, maxLevel int, organizationIDs []uuid.UUID) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).
		Model(&model.User{}).
		Select("COUNT(DISTINCT users.id)")

	// Apply search filter if provided
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("users.username ILIKE ? OR users.email ILIKE ?", searchPattern, searchPattern)
	}

	// Level filtering: only show users with roles having level < maxLevel
	query = query.Joins("LEFT JOIN roles ON users.role_id = roles.id").
		Where("roles.level < ? OR roles.level IS NULL", maxLevel)

	// Organization filtering for non-platform users
	if len(organizationIDs) > 0 {
		// For organization-level users, only show users from their accessible organizations
		// PLUS platform-level users (who don't have organization assignments)
		subQuery := r.db.Table("user_organizations").
			Select("DISTINCT user_id").
			Where("organization_id IN ? AND is_active = true", organizationIDs)

		query = query.Where("users.id IN (?) OR roles.level >= 76", subQuery)
	}

	err := query.Count(&count).Error
	return count, err
}
