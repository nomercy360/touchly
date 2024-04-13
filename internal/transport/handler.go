package transport

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"net/http"
	"touchly/internal/db"
)

type transport struct {
	api       api
	jwtSecret string
}

type api interface {
	LoginUser(email, password string) (*string, error)
	VerifyOTP(email, code string) error
	SendOTP(email string) error
	SetPassword(email, password string) error
	GetUserByID(userID int64) (*db.User, error)

	ListContacts() ([]db.Contact, error)
	CreateContact(userID int64, contact db.Contact) (*db.Contact, error)
	GetContact(id int64) (*db.Contact, error)
	UpdateContact(userID int64, contact db.Contact) error
	DeleteContact(userID, id int64) error

	ListTags() ([]db.Tag, error)
	CreateTag(tag db.Tag) (*db.Tag, error)
	DeleteTag(id int64) error

	ListSavedContacts(userID int64) ([]db.Contact, error)
	SaveContact(userID, contactID int64) error
	DeleteSavedContact(userID, contactID int64) error

	ListAddresses() ([]db.Address, error)
}

func New(api api, jwtSecret string) *transport {
	return &transport{api: api, jwtSecret: jwtSecret}
}

// WriteError responds to a HTTP request with an error.
func WriteError(w http.ResponseWriter, code int, message string) error {
	err := WriteJSON(w, code, map[string]string{"error": message})
	if err != nil {
		return err
	}

	return nil
}

// WriteJSON writes a JSON response to a HTTP request.
func WriteJSON(w http.ResponseWriter, code int, payload interface{}) error {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err := w.Write(response)
	if err != nil {
		return err
	}

	return nil
}

type HealthStatus struct {
	Status string `json:"status"`
}

func (tr *transport) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	_ = WriteJSON(w, http.StatusOK, HealthStatus{Status: "ok"})
}

func (tr *transport) RegisterRoutes(r chi.Router) {
	r.Get("/health", tr.HealthCheckHandler)

	r.Mount("/api", ApiRoutes(tr))
}

func ApiRoutes(tr *transport) http.Handler {
	r := chi.NewRouter()

	// r.Use(WithAuth("secret"))

	r.Post("/login", tr.LoginUserHandler)
	r.Post("/otp", tr.SendOTPHandler)
	r.Post("/otp/verify", tr.VerifyOTPHandler)
	r.Post("/password", tr.SetPasswordHandler)

	r.Get("/me", tr.GetMeHandler)
	r.Get("/contacts", tr.ListContactsHandler)
	r.Post("/contacts", tr.CreateContactHandler)
	r.Get("/contacts/{id}", tr.GetContactHandler)

	r.Get("/tags", tr.ListTagsHandler)
	r.Post("/tags", tr.CreateTagHandler)
	r.Delete("/tags/{id}", tr.DeleteTagHandler)

	r.Get("/contacts/saved", tr.ListSavedContactsHandler)
	r.Post("/contacts/{id}/save", tr.SaveContactHandler)
	r.Delete("/contacts/{id}/save", tr.DeleteSavedContactHandler)

	r.Get("/address", tr.ListAddressesHandler)

	return r
}
