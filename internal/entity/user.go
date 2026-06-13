package entity

import (
	"database/sql"
	"time"
)

type User struct {
	ID                    string
	Username              string
	Email                 string
	PasswordHash          []byte
	CreatedAt             string
	SubscriptionCreatedAt sql.NullTime
	ExpiredAt             sql.NullTime
	IsActive              bool
	Role                  string
	Credits               int
	SubscriptionStatus    string
	PayPalSubscriptionID  string
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
