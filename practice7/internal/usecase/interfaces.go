package usecase

import (
	"practice7/internal/entity"

	"github.com/google/uuid"
)

type UserInterface interface {
	RegisterUser(user *entity.User) (*entity.User, string, error)
	LoginUser(user *entity.LoginUserDTO) (string, error)
	GetMe(userID uuid.UUID) (*entity.User, error)
	PromoteUser(actorID, targetID uuid.UUID) error
	VerifyEmail(dto *entity.VerifyEmailDTO) error
	UserVerified(id uuid.UUID) (bool, error)
}
