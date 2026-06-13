package main

import (
	"net/http"
	"time"

	"github.com/farrasnazhif/go-starter/internal/store"
)

type userKey string

type userProfileResponse struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
	Role     string `json:"role"`
	Credits  int    `json:"credits"`
}

type userBillingResponse struct {
	UserID             string     `json:"userId"`
	Email              string     `json:"email"`
	Username           string     `json:"username"`
	Role               string     `json:"role"`
	SubscriptionStatus string     `json:"subscription_status"`
	CreatedAt          *time.Time `json:"created_at"`
	ExpiredAt          *time.Time `json:"expired_at"`
}

const userCtx userKey = "user"

// getUserHandler godoc
//
//	@Summary		Get the authenticated user's profile
//	@Description	Return the authenticated user's profile information
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Success		200	{object}	userProfileResponse
//	@Failure		401	{object}	error
//	@Failure		500	{object}	error
//	@Router			/users/me [get]
func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromCtx(r)

	response := userProfileResponse{
		ID:       user.ID,
		Email:    user.Email,
		Username: user.Username,
		IsActive: user.IsActive,
		Role:     user.Role,
		Credits:  user.Credits,
	}

	if err := app.jsonResponse(w, http.StatusOK, "User retrieved successfully", response); err != nil {
		app.internalServerError(w, r, err)
		return
	}

}

// getUserBillingHandler godoc
//
//	@Summary		Get the authenticated user's billing profile
//	@Description	Return the authenticated user's billing information
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Success		200	{object}	userBillingResponse
//	@Failure		401	{object}	error
//	@Failure		500	{object}	error
//	@Router			/users/billing [get]
func (app *application) getUserBillingHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromCtx(r)

	response := userBillingResponse{
		UserID:             user.ID,
		Email:              user.Email,
		Username:           user.Username,
		Role:               user.Role,
		SubscriptionStatus: user.SubscriptionStatus,
		CreatedAt:          user.SubscriptionCreatedAtPtr(),
		ExpiredAt:          user.ExpiredAtPtr(),
	}

	if err := app.jsonResponse(w, http.StatusOK, "User billing retrieved successfully", response); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// func (app *application) usersContextMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		id := chi.URLParam(r, "userID")

// 		ctx := r.Context()

// 		user, err := app.store.Users.GetByID(ctx, id)

// 		if err != nil {
// 			switch {
// 			case errors.Is(err, store.ErrNotFound):
// 				app.notFoundResponse(w, r, err)
// 			default:
// 				app.internalServerError(w, r, err)
// 			}
// 			return
// 		}

// 		ctx = context.WithValue(ctx, userCtx, user)
// 		next.ServeHTTP(w, r.WithContext(ctx))
// 	})
// }

func getUserFromCtx(r *http.Request) *store.User {
	user, _ := r.Context().Value(userCtx).(*store.User)
	return user
}
