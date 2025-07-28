package repository

import (
	"beresin-backend/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *model.User) error
	FindByUsernameWithRole(username string) (*model.User, error)
	FindByUsername(username string) (*model.User, error)
	FindByEmail(email string) (*model.User, error)
	FindByGoogleID(googleID string) (*model.User, error)
	FindByIDWithRole(id uuid.UUID) (*model.User, error)
	FindByID(id uuid.UUID) (*model.User, error)
	List(offset, limit int, search string) ([]model.User, error)
	Count(search string) (int64, error)
	Update(user *model.User) error
	Delete(id uuid.UUID) error
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}
