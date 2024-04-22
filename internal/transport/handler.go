package transport

import (
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	api2 "touchly/internal/api"
	"touchly/internal/db"
	"touchly/internal/terrors"
)

type transport struct {
	api       api
	admin     admin
	jwtSecret string
}

type admin interface {
	CreateUser(email, password string) (*db.User, error)
}

type api interface {
	LoginUser(email, password string) (*string, error)
	VerifyOTP(email, code string) error
	SendOTP(email string) error
	SetPassword(email, password string) error
	GetUserByID(userID int64) (*db.User, error)

	ListContacts(tagIDs []int, search string, lat float64, lng float64, radius int, page, pageSize int) (db.ContactsPage, error)
	CreateContact(userID int64, contact db.Contact) (*db.Contact, error)
	GetContact(id int64) (*db.Contact, error)
	UpdateContact(userID int64, contact db.Contact) error
	DeleteContact(userID, id int64) error

	CreateContactAddress(userID int64, address db.Address) (*db.Address, error)

	ListTags() ([]db.Tag, error)
	CreateTag(tag db.Tag) (*db.Tag, error)
	DeleteTag(id int64) error

	ListSavedContacts(userID int64) ([]db.Contact, error)
	SaveContact(userID, contactID int64) error
	DeleteSavedContact(userID, contactID int64) error

	GetPresignedURL(userID int64, filename string) (*api2.UploadURL, error)
}

func New(api api, admin admin, jwtSecret string) *transport {
	return &transport{api: api, admin: admin, jwtSecret: jwtSecret}
}

// WriteError responds to a HTTP request with an error.
func WriteError(r *http.Request, w http.ResponseWriter, err error) {
	var terror *terrors.Error

	if errors.As(err, &terror) {
		logError(r.URL.Path, terror)
		WriteJSON(w, terror.Code, map[string]string{"error": terror.Msg})

		return
	}

	log.Printf("err: %v", err)
	WriteJSON(w, http.StatusInternalServerError, terrors.InternalServerError)
}

func logError(path string, err *terrors.Error) {
	if err.Err != nil {
		log.Printf("path: %s, code: %d, msg: %s, err: %v", path, err.Code, err.Msg, err.Err)
	} else {
		log.Printf("path: %s, code: %d, msg: %s", path, err.Code, err.Msg)
	}
}

// WriteJSON writes a JSON response to a HTTP request.
func WriteJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if _, err := w.Write(response); err != nil {
		log.Printf("failed to write response: %v", err)
	}

	return
}

func WriteOK(w http.ResponseWriter) {
	WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type HealthStatus struct {
	Status string `json:"status"`
}

func (tr *transport) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, HealthStatus{Status: "ok"})
}

func (tr *transport) RegisterRoutes(r chi.Router) {
	r.Get("/health", tr.HealthCheckHandler)

	r.Mount("/api", ApiRoutes(tr))

	r.Mount("/admin", AdminRoutes(tr))
}

func ApiRoutes(tr *transport) http.Handler {
	r := chi.NewRouter()

	r.Post("/login", tr.LoginUserHandler)
	r.Post("/otp", tr.SendOTPHandler)
	r.Post("/otp-verify", tr.VerifyOTPHandler)
	r.Post("/set-password", tr.SetPasswordHandler)

	r.Get("/contacts", tr.ListContactsHandler)
	r.Get("/contacts/{id}", tr.GetContactHandler)

	r.Get("/tags", tr.ListTagsHandler)

	r.Group(func(r chi.Router) {
		r.Use(WithAuth("secret"))

		r.Get("/me", tr.GetMeHandler)
		r.Post("/contacts", tr.CreateContactHandler)

		r.Post("/tags", tr.CreateTagHandler)
		r.Delete("/tags/{id}", tr.DeleteTagHandler)

		r.Get("/contacts/saved", tr.ListSavedContactsHandler)
		r.Post("/contacts/{id}/save", tr.SaveContactHandler)
		r.Delete("/contacts/{id}/save", tr.DeleteSavedContactHandler)
		r.Post("/contacts/{id}/address", tr.CreateContactAddressHandler)

		r.Post("/uploads/get-url", tr.GetUploadURLHandler)
	})

	return r
}

func AdminRoutes(tr *transport) http.Handler {
	r := chi.NewRouter()

	r.Use(WithAdminAuth("secret"))

	r.Post("/users", tr.CreateUserHandler)

	return r
}
