package handler

import (
	"errors"
	"github.com/go-playground/validator"
	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
	"strconv"
	"time"
	api2 "touchly/internal/api"
	"touchly/internal/db"
)

type transport struct {
	api       api
	admin     admin
	jwtSecret string
}

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
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
	CreateContact(userID int64, contact api2.CreateContactRequest) (*db.Contact, error)
	GetContact(userID, id int64) (*db.Contact, error)
	UpdateContact(userID, contactID int64, contact api2.UpdateContactRequest) (*db.Contact, error)
	UpdateContactVisibility(userID, contactID int64, visibility db.ContactVisibility) error
	DeleteContact(userID, id int64) error

	CreateContactAddress(userID, contactID int64, address api2.CreateAddressRequest) (*db.Address, error)

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
	e.Validator = &CustomValidator{validator: validator.New()}

	e.GET("/health", tr.HealthCheckHandler)

	a := e.Group("/api")

	config := echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(api2.JWTClaims)
		},
		SigningKey:             []byte(tr.jwtSecret),
		ContinueOnIgnoredError: true,
		ErrorHandler: func(c echo.Context, err error) error {
			var extErr *echojwt.TokenExtractionError
			if !errors.As(err, &extErr) {
				return echo.NewHTTPError(http.StatusUnauthorized, "auth is invalid")
			}

			claims := &api2.JWTClaims{
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour * 30)),
				},
				UserID: 0,
			}

			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

			c.Set("user", token)

			return nil
		},
	}

	a.Use(echojwt.WithConfig(config))

	a.POST("/login", tr.LoginUserHandler)
	a.POST("/otp", tr.SendOTPHandler)
	a.POST("/otp-verify", tr.VerifyOTPHandler)
	a.POST("/set-password", tr.SetPasswordHandler)
	a.GET("/tags", tr.ListTagsHandler)
	a.POST("/contacts", tr.CreateContactHandler)
	a.GET("/contacts", tr.ListContactsHandler)
	a.GET("/contacts/:id", tr.GetContactHandler)
	a.PUT("/contacts/:id", tr.UpdateContactHandler)
	a.PUT("/contacts/:id/visibility", tr.UpdateContactVisibilityHandler)
	a.POST("/contacts/:id/address", tr.CreateContactAddressHandler)
	a.GET("/me", tr.GetMeHandler)
	a.GET("/me/contacts", tr.ListMyContactsHandler)
	a.GET("/me/saved-contacts", tr.ListSavedContactsHandler)
	a.POST("/contacts/:id/save", tr.SaveContactHandler)
	a.DELETE("/contacts/:id/save", tr.DeleteSavedContactHandler)
	a.POST("/tags", tr.CreateTagHandler)
	a.DELETE("/tags/:id", tr.DeleteTagHandler)
	a.POST("/uploads/get-url", tr.GetUploadURLHandler)

	adm := e.Group("/admin")
	adm.Use(middleware.KeyAuth(tr.AdminKeyValidator))

	adm.POST("/users", tr.CreateUserHandler)
}

func (tr *transport) AdminKeyValidator(key string, c echo.Context) (bool, error) {
	switch key {
	case tr.jwtSecret:
		return true, nil
	default:
		return false, nil
	}
}

func getID(c echo.Context) (int64, error) {
	return strconv.ParseInt(c.Param("id"), 10, 64)
}
