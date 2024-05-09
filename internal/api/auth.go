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
	"touchly/internal/terrors"
)

type JWTClaims struct {
	jwt.RegisteredClaims
	UserID int64 `json:"uid"`
}

func generateOTPCode() string {
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

func GenerateJWT(secret string, uid int64) (string, error) {
	claims := &JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour * 30)),
		},
		UserID: uid,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	t, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return t, nil
}

func (api *api) SetPassword(email, password string) error {
	if email == "" || password == "" {
		return terrors.InvalidRequest(nil, "email and password are required")
	}

	hashedPassword, err := hashPassword(password)
	if err != nil {
		return terrors.InternalServerError(err, "failed to hash password")
	}

	if err := api.storage.UpdateUserPassword(email, hashedPassword); err != nil {
		return terrors.InternalServerError(err, "failed to set password")
	}

	return nil
}

func (api *api) VerifyOTP(email, otpCode string) error {
	if email == "" || otpCode == "" {
		return terrors.InvalidRequest(nil, "email and OTP code are required")
	}

	user, err := api.storage.GetUserByEmail(email)

	if err != nil {
		if db.IsNoRowsError(err) {
			return terrors.InvalidRequest(nil, "user not found")
		} else {
			return terrors.InternalServerError(err, "failed to get user")
		}
	}

	otp, err := api.storage.GetOTPByCode(otpCode, user.ID)

	if err != nil {
		return terrors.InternalServerError(err, "invalid OTP")
	}

	if otp.IsUsed {
		return terrors.InvalidRequest(nil, "OTP is already used")
	}

	if err := api.storage.SetOTPIsUsed(otp.ID); err != nil {
		return terrors.InternalServerError(err, "failed to update OTP")
	}

	if err := api.storage.UpdateUserVerified(user.ID); err != nil {
		return terrors.InternalServerError(err, "failed to update user")
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

func (api *api) SendOTP(email string) error {
	if email == "" {
		return terrors.InvalidRequest(nil, "email is required")
	}

	user, err := api.storage.GetUserByEmail(email)

	if err != nil {
		if db.IsNoRowsError(err) {
			u := db.User{
				Email:           email,
				EmailVerifiedAt: nil,
			}

			user, err = api.storage.CreateUser(u)

			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// Generate OTP (implement your own logic for OTP generation)
	otpCode := generateOTPCode()

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

func (api *api) LoginUser(email, password string) (*string, error) {
	if email == "" || password == "" {
		return nil, errors.New("invalid request")
	}

	user, err := api.storage.GetUserByEmail(email)
	if err != nil {
		return nil, err
	}

	if user.EmailVerifiedAt == nil || user.PasswordHash == nil {
		return nil, errors.New("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password))
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	token, err := GenerateJWT(api.jwtSecret, user.ID)

	if err != nil {
		return nil, err
	}

	return &token, nil
}
