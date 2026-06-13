package response

import "time"

type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
	Role     string `json:"role"`
	Credits  int    `json:"credits"`
}

type UserBilling struct {
	UserID             string     `json:"userId"`
	Email              string     `json:"email"`
	Username           string     `json:"username"`
	Role               string     `json:"role"`
	SubscriptionStatus string     `json:"subscription_status"`
	CreatedAt          *time.Time `json:"created_at"`
	ExpiredAt          *time.Time `json:"expired_at"`
}
