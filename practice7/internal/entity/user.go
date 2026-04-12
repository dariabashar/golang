package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID               uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey"`
	Username         string     `json:"username" gorm:"uniqueIndex;not null"`
	Email            string     `json:"email" gorm:"uniqueIndex;not null"`
	Password         string     `json:"-" gorm:"not null"`
	Role             string     `json:"role" gorm:"not null;default:user"`
	Verified         bool       `json:"verified" gorm:"not null;default:false"`
	VerifCodeHash    string     `json:"-" gorm:"column:verif_code_hash"`
	VerifExpiresAt   *time.Time `json:"-" gorm:"column:verif_expires_at"`
}

func (User) TableName() string { return "users" }

func (u *User) BeforeCreate(_ *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
