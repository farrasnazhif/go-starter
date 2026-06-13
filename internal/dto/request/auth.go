package request

type Register struct {
	Username string `json:"username" validate:"required,max=100"`
	Email    string `json:"email" validate:"required,email,max=225"`
	Password string `json:"password" validate:"required,min=3,max=72"`
}

type SendOTP struct {
	Email string `json:"email" validate:"required,email"`
}

type VerifyOTP struct {
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code" validate:"required,len=6"`
}

type Login struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}
