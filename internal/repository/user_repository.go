package repository

import (
	"beresin-backend/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

// Create implements UserRepository.
func (r *userRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

// Update implements UserRepository.
func (r *userRepository) Update(user *model.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByGoogleID(googleID string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("google_id = ?", googleID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByUsernameWithRole mencari pengguna berdasarkan username dan memuat relasi Role.
func (r *userRepository) FindByUsernameWithRole(username string) (*model.User, error) {
	var user model.User
	err := r.db.Preload("Role").Where("username = ?", username).First(&user).Error
	return &user, err
}

func (r *userRepository) FindByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Where("username = ?", username).First(&user).Error
	return &user, err
}

// FindByIDWithRole mencari pengguna berdasarkan ID dan memuat relasi Role.
func (r *userRepository) FindByIDWithRole(id uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.db.Preload("Role").Where("id = ?", id).First(&user).Error
	return &user, err
}

func (r *userRepository) FindByID(id uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.db.Where("id = ?", id).First(&user).Error
	return &user, err
}

// List mengambil daftar pengguna dengan paginasi.
func (r *userRepository) List(offset, limit int, search string) ([]model.User, error) {
	var users []model.User
	query := r.db.Preload("Role")
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("username ILIKE ? OR email ILIKE ?", searchPattern, searchPattern)
	}
	err := query.Offset(offset).Limit(limit).Find(&users).Error
	return users, err
}

// Count menghitung total jumlah pengguna dengan filter pencarian.
func (r *userRepository) Count(search string) (int64, error) {
	var count int64
	query := r.db.Model(&model.User{})
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("username ILIKE ? OR email ILIKE ?", searchPattern, searchPattern)
	}
	err := query.Count(&count).Error
	return count, err
}

// Delete menghapus pengguna berdasarkan ID.
func (r *userRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.User{}, id).Error
}
