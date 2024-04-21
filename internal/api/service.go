package api

import (
	"log"
	"time"
	"touchly/internal/db"
	"touchly/internal/services"
)

type s3Client interface {
	GetPresignedURL(objectKey string, duration time.Duration) (string, error)
}

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
	s3Client    s3Client
	logger      *log.Logger
}

func NewApi(storage storage, emailClient emailClient, s3Client s3Client) *api {
	return &api{
		storage:     storage,
		emailClient: emailClient,
		s3Client:    s3Client,
		logger:      log.New(log.Writer(), "api: ", log.LstdFlags),
	}
}
