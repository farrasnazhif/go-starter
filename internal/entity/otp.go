package entity

import "time"

type OTPPurpose string

const (
	OTPPurposeRegistration OTPPurpose = "registration"
	OTPPurposeLogin        OTPPurpose = "login"
)

type OTPCode struct {
	ID         string
	Email      string
	Code       string
	Purpose    OTPPurpose
	ExpiresAt  time.Time
	CreatedAt  time.Time
	VerifiedAt *time.Time
}
