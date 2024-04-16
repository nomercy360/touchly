package api

import (
	"touchly/internal/db"
	"touchly/internal/services"
)

type storage interface {
	CreateUser(user db.User) (*db.User, error)
	GetUserByEmail(email string) (*db.User, error)
	UpdateUserPassword(email, password string) error
	GetUserByID(userID int64) (*db.User, error)
	SetOTPIsUsed(otpID int64) error
	UpdateUserVerified(userID int64) error
	GetOTPByCode(code string, userID int64) (*db.OTP, error)
	CreateOTP(otp db.OTP) (*db.OTP, error)

	CreateContact(contact db.Contact) (*db.Contact, error)
	DeleteContact(userID, id int64) error
	UpdateContact(contact db.Contact) error
	ListContacts(tagIDs []int, search string, page, pageSize int) (db.ContactsPage, error)
	GetContact(id int64) (*db.Contact, error)
	SaveContact(userID, contactID int64) error
	DeleteSavedContact(userID, contactID int64) error
	ListSavedContacts(userID int64) ([]db.Contact, error)

	ListTags() ([]db.Tag, error)
	CreateTag(tag db.Tag) (*db.Tag, error)
	DeleteTag(id int64) error
}

type emailClient interface {
	SendEmail(message *services.MailMessage) error
}

type api struct {
	storage     storage
	emailClient emailClient
}

func NewApi(storage storage, emailClient emailClient) *api {
	return &api{
		storage:     storage,
		emailClient: emailClient,
	}
}
