package handler

import (
	"encoding/json"
	"github.com/labstack/echo/v4"
	"net/http"
)

func decodeRequest(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

type LoginUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginUserHandler godoc
// @Summary      Login user
// @Description  login user
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        login body LoginUserRequest true "login"
// @Success      200  {object}   map[string]string
// @Router       /api/login [post]
func (tr *transport) LoginUserHandler(c echo.Context) error {
	var req LoginUserRequest

	if err := c.Bind(&req); err != nil {
		return err
	}

	token, err := tr.api.LoginUser(req.Email, req.Password)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]string{"token": *token})
}

type VerifyOTPRequest struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

// VerifyOTPHandler godoc
// @Summary      Verify OTP
// @Description  verify OTP
// @Tags         users
// @Accept       json
// @Produce      json
// @Param 	     verify body VerifyOTPRequest true "verify"
// @Success      200  {object}   nil
// @Router       /api/otp-verify [post]
func (tr *transport) VerifyOTPHandler(c echo.Context) error {
	var req VerifyOTPRequest

	if err := c.Bind(&req); err != nil {
		return err
	}

	err := tr.api.VerifyOTP(req.Email, req.OTP)
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

type SendOTPRequest struct {
	Email string `json:"email"`
}

// SendOTPHandler godoc
// @Summary      Send OTP
// @Description  send OTP
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        email body SendOTPRequest true "email"
// @Success      200  {object}   nil
// @Router       /api/otp [post]
func (tr *transport) SendOTPHandler(c echo.Context) error {
	var req SendOTPRequest

	if err := c.Bind(&req); err != nil {
		return err
	}

	err := tr.api.SendOTP(req.Email)
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

type SetPasswordRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// SetPasswordHandler godoc
// @Summary      Set password
// @Description  set password
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        password body SetPasswordRequest true "password"
// @Success      200  {object}   nil
// @Router       /api/set-password [post]
func (tr *transport) SetPasswordHandler(c echo.Context) error {
	var req SetPasswordRequest

	if err := c.Bind(req); err != nil {
		return err
	}

	err := tr.api.SetPassword(req.Email, req.Password)
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

// GetMeHandler godoc
// @Summary      Get user
// @Description  get user
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200  {object}   User
// @Security     JWT
// @Router       /api/me [get]
func (tr *transport) GetMeHandler(c echo.Context) error {
	userID := getUserID(c)

	user, err := tr.api.GetUserByID(userID)

	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, user)
}

type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func (tr *transport) CreateUserHandler(c echo.Context) error {
	var req CreateUserRequest

	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	res, err := tr.admin.CreateUser(req.Email, req.Password)

	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, res)
}
