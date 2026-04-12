package repo

import (
	"fmt"
	"time"

	"practice7/internal/entity"
	"practice7/pkg/postgres"

	"github.com/google/uuid"
)

type UserRepo struct {
	PG *postgres.Postgres
}

func NewUserRepo(pg *postgres.Postgres) *UserRepo {
	return &UserRepo{PG: pg}
}

func (u *UserRepo) RegisterUser(user *entity.User) (*entity.User, error) {
	if err := u.PG.Conn.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (u *UserRepo) LoginUser(in *entity.LoginUserDTO) (*entity.User, error) {
	var out entity.User
	if err := u.PG.Conn.Where("username = ?", in.Username).First(&out).Error; err != nil {
		return nil, fmt.Errorf("username not found: %w", err)
	}
	return &out, nil
}

func (u *UserRepo) GetByID(id uuid.UUID) (*entity.User, error) {
	var out entity.User
	if err := u.PG.Conn.Where("id = ?", id).First(&out).Error; err != nil {
		return nil, err
	}
	return &out, nil
}

func (u *UserRepo) PromoteToAdmin(id uuid.UUID) error {
	res := u.PG.Conn.Model(&entity.User{}).Where("id = ?", id).Update("role", "admin")
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

func (u *UserRepo) SetVerificationFields(userID uuid.UUID, codeHash string, exp *time.Time) error {
	return u.PG.Conn.Model(&entity.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"verif_code_hash":  codeHash,
		"verif_expires_at": exp,
		"verified":         false,
	}).Error
}

func (u *UserRepo) ClearVerificationAndMarkVerified(userID uuid.UUID) error {
	return u.PG.Conn.Model(&entity.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"verif_code_hash":  "",
		"verif_expires_at": nil,
		"verified":         true,
	}).Error
}

func (u *UserRepo) FindByEmail(email string) (*entity.User, error) {
	var out entity.User
	if err := u.PG.Conn.Where("email = ?", email).First(&out).Error; err != nil {
		return nil, err
	}
	return &out, nil
}
