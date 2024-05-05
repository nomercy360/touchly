package transport

import (
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"net/http"
	api2 "touchly/internal/api"
	"touchly/internal/db"
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
	ListMyContacts(userID int64) (db.ContactsPage, error)

	ListContacts(userID int64, tagIDs []int, search string, lat float64, lng float64, radius int, page, pageSize int) (db.ContactsPage, error)
	CreateContact(userID int64, contact db.Contact) (*db.Contact, error)
	GetContact(userID, id int64) (*db.Contact, error)
	UpdateContact(userID, contactID int64, contact api2.UpdateContactRequest) (*db.Contact, error)
	UpdateContactVisibility(userID, contactID int64, visibility db.ContactVisibility) error
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

type HealthStatus struct {
	Status string `json:"status"`
}

func (tr *transport) HealthCheckHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, HealthStatus{Status: "ok"})
}

func (tr *transport) RegisterRoutes(e *echo.Echo) {
	e.GET("/health", tr.HealthCheckHandler)

	a := e.Group("/api")

	config := echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(svc.JWTClaims)
		},
		SigningKey: []byte("secret"),
	}

	a.Use(echojwt.WithConfig(config))

	a.POST("/login", tr.LoginUserHandler)
	a.POST("/otp", tr.SendOTPHandler)
	a.POST("/otp-verify", tr.VerifyOTPHandler)
	a.POST("/set-password", tr.SetPasswordHandler)
	a.GET("/tags", tr.ListTagsHandler)
	a.POST("/contacts", tr.CreateContactHandler)
	a.GET("/contacts", tr.ListContactsHandler)
	a.GET("/contacts/{id}", tr.GetContactHandler)
	a.PUT("/contacts/{id}", tr.UpdateContactHandler)
	a.PUT("/contacts/{id}/visibility", tr.UpdateContactVisibilityHandler)
	a.POST("/contacts/{id}/address", tr.CreateContactAddressHandler)
	a.GET("/me", tr.GetMeHandler)
	a.POST("/contacts/{id}/save", tr.SaveContactHandler)
	a.DELETE("/contacts/{id}/save", tr.DeleteSavedContactHandler)
	a.GET("/contacts/saved", tr.ListSavedContactsHandler)
	a.POST("/tags", tr.CreateTagHandler)
	a.DELETE("/tags/{id}", tr.DeleteTagHandler)
	a.POST("/uploads/get-url", tr.GetUploadURLHandler)

	adm := e.Group("/admin")
	adm.Use(WithAdminAuth("secret"))

	adm.POST("/users", tr.CreateUserHandler)

	// e.Mount("/admin", AdminRoutes(tr))
}

func ApiRoutes(tr *transport) http.Handler {
	r := chi.NewRouter()

	r.Post("/login", tr.LoginUserHandler)
	r.Post("/otp", tr.SendOTPHandler)
	r.Post("/otp-verify", tr.VerifyOTPHandler)
	r.Post("/set-password", tr.SetPasswordHandler)

	r.Get("/tags", tr.ListTagsHandler)

	r.Group(func(r chi.Router) {
		r.Use(WithAuth("secret", true))

		r.Get("/me", tr.GetMeHandler)
		r.Post("/contacts", tr.CreateContactHandler)
		r.Get("/me/contacts", tr.ListMyContactsHandler)

		r.Post("/tags", tr.CreateTagHandler)
		r.Delete("/tags/{id}", tr.DeleteTagHandler)

		r.Get("/contacts/saved", tr.ListSavedContactsHandler)
		r.Post("/contacts/{id}/save", tr.SaveContactHandler)
		r.Delete("/contacts/{id}/save", tr.DeleteSavedContactHandler)

		r.Post("/contacts/{id}/address", tr.CreateContactAddressHandler)
		r.Put("/contacts/{id}", tr.UpdateContactHandler)
		r.Put("/contacts/{id}/visibility", tr.UpdateContactVisibilityHandler)

		r.Post("/uploads/get-url", tr.GetUploadURLHandler)
	})

	r.Group(func(r chi.Router) {
		r.Use(WithAuth("secret", false))

		r.Get("/contacts", tr.ListContactsHandler)
		r.Get("/contacts/{id}", tr.GetContactHandler)
	})

	return r
}

func AdminRoutes(tr *transport) http.Handler {
	r := chi.NewRouter()

	r.Use(WithAdminAuth("secret"))

	r.Post("/users", tr.CreateUserHandler)

	return r
}
