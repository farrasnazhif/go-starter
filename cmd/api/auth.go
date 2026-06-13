package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/farrasnazhif/go-starter/internal/mailer"
	"github.com/farrasnazhif/go-starter/internal/store"
	"github.com/golang-jwt/jwt"
)

type RegisterUserPayload struct {
	Username string `json:"username" validate:"required,max=100"`
	Email    string `json:"email" validate:"required,email,max=225"`
	Password string `json:"password" validate:"required,min=3,max=72"`
}

type SendOTPPayload struct {
	Email string `json:"email" validate:"required,email"`
}

type VerifyOTPPayload struct {
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code" validate:"required,len=6"`
}

type LoginPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type UserWithToken struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	IsActive bool   `json:"is_active"`
	Role     string `json:"role"`
	Credits  int    `json:"credits"`
	Token    string `json:"token"`
}

type authUserData struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	IsActive bool   `json:"is_active"`
	Role     string `json:"role"`
	Credits  int    `json:"credits"`
	Token    string `json:"token"`
}

type otpVerifyResponse struct {
	Message  string `json:"message"`
	Verified bool   `json:"verified"`
}

type registerResponse struct {
	Message string       `json:"message"`
	Data    authUserData `json:"data"`
}

func (app *application) generateJWTToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(app.config.jwt.secret))
}

func (app *application) validateJWTToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(app.config.jwt.secret), nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok := claims["user_id"].(string)
		if !ok {
			return "", fmt.Errorf("invalid user_id in token")
		}
		return userID, nil
	}

	return "", fmt.Errorf("invalid token")
}

// registerUserHandler godoc
//
//	@Summary		Register a user
//	@Description	Create a user account and send OTP for verification
//	@Tags			authentication
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		RegisterUserPayload	true	"User credentials"
//	@Success		201		{object}	map[string]any		"User registered"
//	@Failure		400		{object}	error
//	@Failure		500		{object}	error
//	@Router			/auth/register [post]
func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload RegisterUserPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()

	_, err := app.store.Users.GetByEmail(ctx, payload.Email)
	if err == nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"status":  http.StatusBadRequest,
			"message": "email already registered",
		})
		return
	}
	if err != store.ErrNotFound {
		app.internalServerError(w, r, err)
		return
	}

	user := &store.User{
		Username: payload.Username,
		Email:    payload.Email,
		IsActive: false,
	}

	if err := user.Password.Set(payload.Password); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.store.Users.Create(ctx, nil, user); err != nil {
		switch err {
		case store.ErrDuplicateEmail:
			writeJSON(w, http.StatusBadRequest, map[string]any{
				"status":  http.StatusBadRequest,
				"message": "email already registered",
			})
		case store.ErrDuplicateUsername:
			app.badRequestResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	// Generate and send OTP
	otp, err := app.store.OTP.Create(ctx, payload.Email, store.OTPPurposeRegistration, 10*time.Minute)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	vars := struct {
		OTPCode string
	}{
		OTPCode: otp.Code,
	}

	isProdEnv := app.config.env == "production"
	err = app.mailer.Send(mailer.UserOTP, "User", payload.Email, vars, !isProdEnv)
	if err != nil {
		app.logger.Errorw("error sending OTP email", "error", err)
		app.internalServerError(w, r, err)
		return
	}

	if err := writeJSON(w, http.StatusCreated, map[string]any{
		"message": "User registered successfully. OTP has been sent to your email.",
		"email":   payload.Email,
	}); err != nil {
		app.internalServerError(w, r, err)
	}
}

// sendRegistrationOTPHandler godoc
//
//	@Summary		Send OTP for user registration
//	@Description	Send OTP for user registration
//	@Tags			authentication
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		SendOTPPayload		true	"Email for OTP"
//	@Success		200		{object}	map[string]string	"OTP sent"
//	@Failure		400		{object}	error
//	@Failure		500		{object}	error
//	@Router			/auth/otp/send [post]
func (app *application) sendRegistrationOTPHandler(w http.ResponseWriter, r *http.Request) {
	var payload SendOTPPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()

	// Check if user already exists
	_, err := app.store.Users.GetByEmail(ctx, payload.Email)
	if err == nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"status":  http.StatusBadRequest,
			"message": "email already registered",
		})
		return
	} else if err != store.ErrNotFound {
		app.internalServerError(w, r, err)
		return
	}

	// Create OTP
	otp, err := app.store.OTP.Create(ctx, payload.Email, store.OTPPurposeRegistration, 10*time.Minute)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	// Send OTP email
	vars := struct {
		OTPCode string
	}{
		OTPCode: otp.Code,
	}

	isProdEnv := app.config.env == "production"
	err = app.mailer.Send(mailer.UserOTP, "User", payload.Email, vars, !isProdEnv)
	if err != nil {
		app.logger.Errorw("error sending OTP email", "error", err)
		app.internalServerError(w, r, err)
		return
	}

	if err := writeJSON(w, http.StatusOK, map[string]string{"message": "OTP sent successfully"}); err != nil {
		app.internalServerError(w, r, err)
	}
}

// verifyRegistrationOTPHandler godoc
//
//	@Summary		Verify OTP and activate user
//	@Description	Verify OTP and activate user account
//	@Tags			authentication
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		VerifyOTPPayload	true	"Email and OTP code"
//	@Success		200		{object}	UserWithToken		"OTP verified and user activated"
//	@Failure		400		{object}	error
//	@Failure		500		{object}	error
//	@Router			/auth/otp/verify [post]
func (app *application) verifyRegistrationOTPHandler(w http.ResponseWriter, r *http.Request) {
	var payload VerifyOTPPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()

	_, err := app.store.OTP.Verify(ctx, payload.Email, payload.Code, store.OTPPurposeRegistration)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			writeJSON(w, http.StatusBadRequest, map[string]any{
				"status":  http.StatusBadRequest,
				"message": "invalid otp",
			})
		case store.ErrOTPExpired:
			writeJSON(w, http.StatusBadRequest, map[string]any{
				"status":  http.StatusBadRequest,
				"message": "otp expired",
			})
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	// Fetch user and activate
	user, err := app.store.Users.GetByEmail(ctx, payload.Email)
	if err != nil {
		if err == store.ErrNotFound {
			app.badRequestResponse(w, r, fmt.Errorf("user not found"))
		} else {
			app.internalServerError(w, r, err)
		}
		return
	}

	// Activate user
	user.IsActive = true
	if err := app.store.Users.Update(ctx, user); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	// Generate JWT token
	token, err := app.generateJWTToken(user.ID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	userWithToken := UserWithToken{
		ID:       user.ID,
		Email:    user.Email,
		Username: user.Username,
		IsActive: user.IsActive,
		Role:     user.Role,
		Credits:  user.Credits,
		Token:    token,
	}

	if err := app.jsonResponse(w, http.StatusOK, "OTP verified successfully. Account activated", userWithToken); err != nil {
		app.internalServerError(w, r, err)
	}
}

// loginHandler godoc
//
//	@Summary		Login user with email and password
//	@Description	Login user with email and password
//	@Tags			authentication
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		LoginPayload	true	"Email and password"
//	@Success		200		{object}	UserWithToken	"User logged in"
//	@Failure		400		{object}	error
//	@Failure		500		{object}	error
//	@Router			/auth/login [post]
func (app *application) loginUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload LoginPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()

	// Verify password
	user, err := app.store.Users.VerifyPassword(ctx, payload.Email, payload.Password)
	if err != nil {
		if err.Error() == "invalid password" || err == store.ErrNotFound {
			app.badRequestResponse(w, r, fmt.Errorf("invalid credentials"))
		} else {
			app.internalServerError(w, r, err)
		}
		return
	}

	if !user.IsActive {
		app.badRequestResponse(w, r, fmt.Errorf("account not activated"))
		return
	}

	// Generate JWT token
	token, err := app.generateJWTToken(user.ID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	userWithToken := UserWithToken{
		ID:       user.ID,
		Email:    user.Email,
		Username: user.Username,
		IsActive: user.IsActive,
		Role:     user.Role,
		Credits:  user.Credits,
		Token:    token,
	}

	if err := app.jsonResponse(w, http.StatusOK, "Login successful", userWithToken); err != nil {
		app.internalServerError(w, r, err)
	}
}
