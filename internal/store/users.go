package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail    = errors.New("a user with that email already exists")
	ErrDuplicateUsername = errors.New("a user with that username already exists")
)

type User struct {
	ID                    string       `json:"id"`
	Username              string       `json:"username"`
	Email                 string       `json:"email"`
	Password              password     `json:"-"`
	CreatedAt             string       `json:"created_at"`
	SubscriptionCreatedAt sql.NullTime `json:"-"`
	ExpiredAt             sql.NullTime `json:"-"`
	IsActive              bool         `json:"is_active"`
	Role                  string       `json:"role"`
	Credits               int          `json:"credits"`
	SubscriptionStatus    string       `json:"subscription_status,omitempty"`
}

type password struct {
	text *string
	hash []byte
}

func (p *password) Set(text string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	p.text = &text
	p.hash = hash
	return nil
}

func (u *User) SubscriptionCreatedAtPtr() *time.Time {
	if !u.SubscriptionCreatedAt.Valid {
		return nil
	}
	return &u.SubscriptionCreatedAt.Time
}

func (u *User) ExpiredAtPtr() *time.Time {
	if !u.ExpiredAt.Valid {
		return nil
	}
	return &u.ExpiredAt.Time
}

type UserStore struct {
	db *sql.DB
}

func (s *UserStore) Create(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `
		INSERT INTO users (username, email, password, is_active)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(ctx, query, user.Username, user.Email, user.Password.hash, user.IsActive).
		Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case err.Error() == `pq: duplicate key value violates unique constraint "users_username_key"`:
			return ErrDuplicateUsername
		default:
			return err
		}
	}
	return nil
}

func (s *UserStore) GetByID(ctx context.Context, id string) (*User, error) {
	query := `
	SELECT id, email, username, created_at, subscription_created_at, expired_at, is_active, role, credits, subscription_status
	FROM users WHERE id = $1
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	var user User
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Username, &user.CreatedAt,
		&user.SubscriptionCreatedAt, &user.ExpiredAt,
		&user.IsActive, &user.Role, &user.Credits, &user.SubscriptionStatus,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (s *UserStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `
	SELECT id, email, username, password, created_at, subscription_created_at, expired_at, is_active, role, credits, subscription_status
	FROM users WHERE email = $1
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	var user User
	err := s.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Username, &user.Password.hash, &user.CreatedAt,
		&user.SubscriptionCreatedAt, &user.ExpiredAt,
		&user.IsActive, &user.Role, &user.Credits, &user.SubscriptionStatus,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (s *UserStore) VerifyPassword(ctx context.Context, email, password string) (*User, error) {
	user, err := s.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword(user.Password.hash, []byte(password)); err != nil {
		return nil, errors.New("invalid password")
	}
	return user, nil
}

func (s *UserStore) Update(ctx context.Context, user *User) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		query := `UPDATE users SET username = $1, email = $2, password = $3, is_active = $4 WHERE id = $5`
		ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
		defer cancel()
		_, err := tx.ExecContext(ctx, query, user.Username, user.Email, user.Password.hash, user.IsActive, user.ID)
		return err
	})
}

func (s *UserStore) Delete(ctx context.Context, userID string) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		query := `DELETE FROM users WHERE id = $1`
		ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
		defer cancel()
		_, err := tx.ExecContext(ctx, query, userID)
		return err
	})
}
