package model

import (
	"time"

	"github.com/google/uuid"
)

// Permission merepresentasikan izin individual untuk suatu aksi.
type Permission struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v7()" json:"id"`
	Name        string    `gorm:"type:varchar(100);unique;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
