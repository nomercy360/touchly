package transport

import (
	"encoding/json"
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
func (tr *transport) LoginUserHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginUserRequest

	if err := decodeRequest(r, &req); err != nil {
		_ = WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	token, err := tr.api.LoginUser(req.Email, req.Password)
	if err != nil {
		_ = WriteError(w, http.StatusUnauthorized, err.Error())
		return
	}

	_ = WriteJSON(w, http.StatusOK, map[string]string{"token": *token})
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
func (tr *transport) VerifyOTPHandler(w http.ResponseWriter, r *http.Request) {
	var req VerifyOTPRequest

	if err := decodeRequest(r, &req); err != nil {
		_ = WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	err := tr.api.VerifyOTP(req.Email, req.OTP)
	if err != nil {
		_ = WriteError(w, http.StatusUnauthorized, err.Error())
		return
	}

	_ = WriteJSON(w, http.StatusOK, nil)
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
func (tr *transport) SendOTPHandler(w http.ResponseWriter, r *http.Request) {
	var req SendOTPRequest

	if err := decodeRequest(r, &req); err != nil {
		_ = WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	err := tr.api.SendOTP(req.Email)
	if err != nil {
		_ = WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_ = WriteJSON(w, http.StatusOK, map[string]string{"message": "OK"})
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
func (tr *transport) SetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var req SetPasswordRequest

	if err := decodeRequest(r, &req); err != nil {
		_ = WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	err := tr.api.SetPassword(req.Email, req.Password)
	if err != nil {
		_ = WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_ = WriteJSON(w, http.StatusOK, map[string]string{"message": "OK"})
}

// GetMeHandler godoc
// @Summary      Get user
// @Description  get user
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200  {object}   db.User
// @Security     JWT
// @Router       /api/me [get]
func (tr *transport) GetMeHandler(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromRequest(r)

	user, err := tr.api.GetUserByID(userID)

	if err != nil {
		_ = WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_ = WriteJSON(w, http.StatusOK, user)
}
