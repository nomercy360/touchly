package admin

import (
	"golang.org/x/crypto/bcrypt"
	"time"
	"touchly/internal/db"
	"touchly/internal/terrors"
)

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func (adm *admin) CreateUser(email, password string) (*db.User, error) {
	hash, err := hashPassword(password)
	if err != nil {
		return nil, terrors.InternalServerError(err, "invalid data")
	}

	now := time.Now()

	user := db.User{
		Email:           email,
		PasswordHash:    &hash,
		EmailVerifiedAt: &now,
	}

	res, err := adm.storage.CreateUser(user)

	if err != nil {
		return nil, terrors.InternalServerError(err, "could not create user")
	}

	return res, nil
}
