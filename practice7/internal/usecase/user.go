package usecase

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"practice7/internal/entity"
	"practice7/internal/usecase/repo"
	"practice7/pkg/mail"
	"practice7/utils"

	"github.com/google/uuid"
)

type UserUseCase struct {
	repo *repo.UserRepo
	mail *mail.Sender
}

func NewUserUseCase(r *repo.UserRepo, m *mail.Sender) *UserUseCase {
	return &UserUseCase{repo: r, mail: m}
}

func (u *UserUseCase) RegisterUser(user *entity.User) (*entity.User, string, error) {
	created, err := u.repo.RegisterUser(user)
	if err != nil {
		return nil, "", fmt.Errorf("register user: %w", err)
	}
	code, err := randomDigits(4)
	if err != nil {
		return nil, "", fmt.Errorf("code: %w", err)
	}
	hash, err := utils.HashPassword(code)
	if err != nil {
		return nil, "", fmt.Errorf("hash code: %w", err)
	}
	exp := time.Now().Add(15 * time.Minute)
	if err := u.repo.SetVerificationFields(created.ID, hash, &exp); err != nil {
		return nil, "", fmt.Errorf("save verification: %w", err)
	}
	if err := u.mail.SendVerificationCode(created.Email, code); err != nil {
		return nil, "", fmt.Errorf("send email: %w", err)
	}
	sessionID := uuid.New().String()
	return created, sessionID, nil
}

func (u *UserUseCase) LoginUser(in *entity.LoginUserDTO) (string, error) {
	userFromRepo, err := u.repo.LoginUser(in)
	if err != nil {
		return "", fmt.Errorf("login: %w", err)
	}
	if !utils.CheckPassword(userFromRepo.Password, in.Password) {
		return "", fmt.Errorf("invalid credentials")
	}
	token, err := utils.GenerateJWT(userFromRepo.ID, userFromRepo.Role)
	if err != nil {
		return "", fmt.Errorf("jwt: %w", err)
	}
	return token, nil
}

func (u *UserUseCase) GetMe(userID uuid.UUID) (*entity.User, error) {
	return u.repo.GetByID(userID)
}

func (u *UserUseCase) PromoteUser(_ uuid.UUID, targetID uuid.UUID) error {
	return u.repo.PromoteToAdmin(targetID)
}

func (u *UserUseCase) VerifyEmail(dto *entity.VerifyEmailDTO) error {
	user, err := u.repo.FindByEmail(dto.Email)
	if err != nil {
		return fmt.Errorf("user not found")
	}
	if user.Verified {
		return fmt.Errorf("already verified")
	}
	if user.VerifCodeHash == "" || user.VerifExpiresAt == nil || time.Now().After(*user.VerifExpiresAt) {
		return fmt.Errorf("code expired or missing")
	}
	if !utils.CheckPassword(user.VerifCodeHash, dto.Code) {
		return fmt.Errorf("invalid code")
	}
	return u.repo.ClearVerificationAndMarkVerified(user.ID)
}

func (u *UserUseCase) UserVerified(id uuid.UUID) (bool, error) {
	user, err := u.repo.GetByID(id)
	if err != nil {
		return false, err
	}
	return user.Verified, nil
}

func randomDigits(n int) (string, error) {
	const digits = "0123456789"
	b := make([]byte, n)
	for i := range b {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		b[i] = digits[idx.Int64()]
	}
	return string(b), nil
}
