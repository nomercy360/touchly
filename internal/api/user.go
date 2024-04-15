package api

import (
	"bytes"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	"math/rand"
	"strings"
	"time"
	"touchly/internal/db"
	"touchly/internal/services"
)

func GenerateOTPCode() string {
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)

	digits := "0123456789"
	otpLength := 4
	var otpCode strings.Builder

	for i := 0; i < otpLength; i++ {
		randomIndex := r.Intn(len(digits))
		otpCode.WriteByte(digits[randomIndex])
	}

	return otpCode.String()
}

func (api *api) LoginUser(email, password string) (*string, error) {
	if email == "" || password == "" {
		return nil, errors.New("invalid request")
	}

	user, err := api.storage.GetUserByEmail(email)
	if err != nil {
		return nil, err
	}

	if !user.EmailVerified || user.PasswordHash == nil {
		return nil, errors.New("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password))
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": user.ID,
		"exp":    jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
	})

	if tokenString, err := token.SignedString([]byte("secret")); err != nil {
		return nil, err
	} else {
		return &tokenString, nil
	}
}

var (
	ErrorInvalidRequest = errors.New("invalid request")
	ErrorUserNotFound   = errors.New("user not found")
)

func (api *api) SendOTP(email string) error {
	if email == "" {
		return ErrorInvalidRequest
	}

	user, err := api.storage.GetUserByEmail(email)

	if err != nil {
		if errors.As(err, &db.ErrNotFound) {
			u := db.User{
				Email:         email,
				EmailVerified: false,
			}

			if user, err = api.storage.CreateUser(u); err != nil {
				return err
			}

		} else {
			return err
		}
	}

	// Generate OTP (implement your own logic for OTP generation)
	otpCode := GenerateOTPCode()

	// Set OTP expiration time (e.g., 10 minutes from now)
	expiresAt := time.Now().Add(10 * time.Minute)

	otp := db.OTP{
		UserID:    user.ID,
		OTPCode:   otpCode,
		ExpiresAt: expiresAt,
	}

	if _, err := api.storage.CreateOTP(otp); err != nil {
		return err
	}

	if err := api.SendOTPEmail(email, otpCode); err != nil {
		return err
	}

	return nil
}

func (api *api) SendOTPEmail(recipientEmail, otpCode string) error {
	type Context struct {
		OTPCode string
	}

	tmpl, err := template.ParseFiles("templates/otp.gohtml")

	if err != nil {
		return err
	}

	var tpl bytes.Buffer
	err = tmpl.Execute(&tpl, Context{OTPCode: otpCode})

	if err != nil {
		return err
	}

	message := &services.MailMessage{
		To:            recipientEmail,
		Subject:       "Your OTP code",
		MessageStream: "outbound",
		From:          "hi@mxksim.dev",
		HtmlBody:      tpl.String(),
	}

	err = api.emailClient.SendEmail(message)

	if err != nil {
		return err
	}

	return nil
}

func (api *api) VerifyOTP(email, otpCode string) error {
	if email == "" || otpCode == "" {
		return ErrorInvalidRequest
	}

	user, err := api.storage.GetUserByEmail(email)

	if err != nil {
		return err
	}

	otp, err := api.storage.GetOTPByCode(otpCode, user.ID)

	if err != nil {
		return err
	}

	if otp.IsUsed {
		return errors.New("OTP is already used")
	}

	if err := api.storage.SetOTPIsUsed(otp.ID); err != nil {
		return err
	}

	if err := api.storage.UpdateUserVerified(user.ID); err != nil {
		return err
	}

	return nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (api *api) SetPassword(email, password string) error {
	if email == "" || password == "" {
		return ErrorInvalidRequest
	}

	hashedPassword, err := hashPassword(password)
	if err != nil {
		return err
	}

	if err := api.storage.UpdateUserPassword(email, hashedPassword); err != nil {
		return err
	}

	return nil
}

func (api *api) GetUserByID(userID int64) (*db.User, error) {
	user, err := api.storage.GetUserByID(userID)

	if err != nil {
		return nil, err
	}

	return user, nil
}
