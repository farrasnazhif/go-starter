package store

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"time"
)

type OTPPurpose string

const (
	OTPPurposeRegistration OTPPurpose = "registration"
	OTPPurposeLogin        OTPPurpose = "login"
)

var ErrOTPExpired = errors.New("otp expired")

type OTPCode struct {
	ID         string     `json:"id"`
	Email      string     `json:"email"`
	Code       string     `json:"code"`
	Purpose    OTPPurpose `json:"purpose"`
	ExpiresAt  time.Time  `json:"expires_at"`
	CreatedAt  time.Time  `json:"created_at"`
	VerifiedAt *time.Time `json:"verified_at,omitempty"`
}

type OTPStore struct {
	db *sql.DB
}

func (s *OTPStore) Create(ctx context.Context, email string, purpose OTPPurpose, expiry time.Duration) (*OTPCode, error) {
	code, err := generateOTP()
	if err != nil {
		return nil, err
	}

	query := `
		INSERT INTO otp_codes (email, code, purpose, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	otp := &OTPCode{
		Email:     email,
		Code:      code,
		Purpose:   purpose,
		ExpiresAt: time.Now().Add(expiry),
	}

	err = s.db.QueryRowContext(ctx, query, email, code, purpose, otp.ExpiresAt).Scan(
		&otp.ID,
		&otp.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return otp, nil
}

func (s *OTPStore) Verify(ctx context.Context, email, code string, purpose OTPPurpose) (*OTPCode, error) {
	query := `
		SELECT id, email, code, purpose, expires_at, created_at, verified_at
		FROM otp_codes
		WHERE email = $1 AND code = $2 AND purpose = $3
		ORDER BY created_at DESC
		LIMIT 1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	var otp OTPCode
	var verifiedAt sql.NullTime
	err := s.db.QueryRowContext(ctx, query, email, code, purpose).Scan(
		&otp.ID,
		&otp.Email,
		&otp.Code,
		&otp.Purpose,
		&otp.ExpiresAt,
		&otp.CreatedAt,
		&verifiedAt,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	if verifiedAt.Valid {
		otp.VerifiedAt = &verifiedAt.Time
	}

	if !otp.ExpiresAt.After(time.Now()) {
		return nil, ErrOTPExpired
	}

	now := time.Now()
	if _, err := s.db.ExecContext(ctx, `UPDATE otp_codes SET verified_at = $1 WHERE id = $2`, now, otp.ID); err != nil {
		return nil, err
	}
	otp.VerifiedAt = &now

	return &otp, nil
}

func (s *OTPStore) IsVerified(ctx context.Context, email string, purpose OTPPurpose) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM otp_codes
			WHERE email = $1
				AND purpose = $2
				AND verified_at IS NOT NULL
				AND expires_at > $3
		)
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	var verified bool
	if err := s.db.QueryRowContext(ctx, query, email, purpose, time.Now()).Scan(&verified); err != nil {
		return false, err
	}
	return verified, nil
}

func (s *OTPStore) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM otp_codes WHERE id = $1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query, id)
	return err
}

func (s *OTPStore) CleanupExpired(ctx context.Context) error {
	query := `DELETE FROM otp_codes WHERE expires_at < $1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query, time.Now())
	return err
}

func generateOTP() (string, error) {
	const digits = "0123456789"
	const length = 6

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	for i := range bytes {
		bytes[i] = digits[bytes[i]%10]
	}

	return string(bytes), nil
}
