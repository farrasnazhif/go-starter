package response

type UserWithToken struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	IsActive bool   `json:"is_active"`
	Role     string `json:"role"`
	Credits  int    `json:"credits"`
	Token    string `json:"token"`
}
