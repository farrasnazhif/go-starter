package mailer

import "embed"

const (
	FromName            = "GoSocial"
	maxRetires          = 3
	UserWelcomeTemplate = "user_invitation.tmpl"
	UserOTP             = "user_otp.tmpl"
)

//go:embed templates/*.tmpl
var FS embed.FS

type Client interface {
	Send(templateFile, username, email string, data any, isSandbox bool) error
}
