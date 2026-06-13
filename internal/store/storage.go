package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrNotFound          = errors.New("Route not found")
	ErrConflict          = errors.New("Resource conflicted")
	QueryTimeoutDuration = time.Second * 5
)

type Storage struct {
	Users interface {
		Create(ctx context.Context, tx *sql.Tx, user *User) error
		GetByID(ctx context.Context, id string) (*User, error)
		GetByEmail(ctx context.Context, email string) (*User, error)
		VerifyPassword(ctx context.Context, email, password string) (*User, error)
		Update(ctx context.Context, user *User) error
		Delete(ctx context.Context, userID string) error
	}
	OTP interface {
		Create(ctx context.Context, email string, purpose OTPPurpose, expiry time.Duration) (*OTPCode, error)
		Verify(ctx context.Context, email, code string, purpose OTPPurpose) (*OTPCode, error)
		IsVerified(ctx context.Context, email string, purpose OTPPurpose) (bool, error)
		Delete(ctx context.Context, id string) error
		CleanupExpired(ctx context.Context) error
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Users: &UserStore{db},
		OTP:   &OTPStore{db},
	}
}

func withTx(db *sql.DB, ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
