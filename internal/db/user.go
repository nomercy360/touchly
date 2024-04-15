package db

import "time"

type User struct {
	ID            int64     `db:"id" json:"id"`
	Email         string    `db:"email" json:"email"`
	PasswordHash  *string   `db:"password_hash" json:"-"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
	EmailVerified bool      `db:"email_verified" json:"email_verified"`
}

type OTP struct {
	ID        int64     `db:"id"`
	UserID    int64     `db:"user_id"`
	OTPCode   string    `db:"otp_code"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
	IsUsed    bool      `db:"is_used"`
}

func (s *storage) CreateUser(user User) (*User, error) {
	query := `
		INSERT INTO users
		   (email, password_hash, created_at, updated_at, email_verified)
		VALUES ($1, $2, NOW(), NOW(), $3)
	`

	err := s.pg.Get(&user, query, user.Email, user.PasswordHash, user.EmailVerified)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *storage) GetUserByEmail(email string) (*User, error) {
	var user User

	query := `
		SELECT id, email, password_hash, created_at, updated_at, email_verified
		FROM users
		WHERE email = $1
	`

	err := s.pg.Get(&user, query, email)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *storage) UpdateUserPassword(email, password string) error {
	query := `
		UPDATE users
		SET password_hash = $1
		WHERE email = $2
	`

	res, err := s.pg.Exec(query, password, email)

	if err != nil {
		return err
	}

	if rows, _ := res.RowsAffected(); rows == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *storage) SetOTPIsUsed(otpID int64) error {
	query := `
		UPDATE otps
		SET is_used = true
		WHERE id = $1
	`

	if _, err := s.pg.Exec(query, otpID); err != nil {
		return err
	}

	return nil
}

func (s *storage) UpdateUserVerified(userID int64) error {
	query := `
		UPDATE users
		SET email_verified = true
		WHERE id = $1
	`

	if _, err := s.pg.Exec(query, userID); err != nil {
		return err
	}

	return nil
}

func (s *storage) GetOTPByCode(code string, userID int64) (*OTP, error) {
	var otp OTP

	query := `
		SELECT id, user_id, otp_code, expires_at, created_at, is_used
		FROM otps
		WHERE otp_code = $1 AND user_id = $2
	`

	err := s.pg.Get(&otp, query, code, userID)

	if err != nil {
		return nil, err
	}

	return &otp, nil
}

func (s *storage) CreateOTP(otp OTP) (*OTP, error) {
	query := `
		INSERT INTO otps
		   (user_id, otp_code, expires_at, created_at, is_used)
		VALUES ($1, $2, $3, NOW(), false)
		RETURNING id
	`

	err := s.pg.Get(&otp.ID, query, otp.UserID, otp.OTPCode, otp.ExpiresAt)

	if err != nil {
		return nil, err
	}

	return &otp, nil
}

func (s *storage) GetUserByID(userID int64) (*User, error) {
	var user User

	query := `
		SELECT id, email, password_hash, created_at, updated_at, email_verified
		FROM users
		WHERE id = $1
	`

	err := s.pg.Get(&user, query, userID)

	if err != nil {
		return nil, err
	}

	return &user, nil
}