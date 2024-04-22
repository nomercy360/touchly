package admin

import (
	"touchly/internal/db"
)

type storage interface {
	CreateUser(user db.User) (*db.User, error)
}

type admin struct {
	storage storage
}

func NewAdmin(storage storage) *admin {
	return &admin{
		storage: storage,
	}
}
