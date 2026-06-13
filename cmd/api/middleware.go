package main

import (
	"context"
	"net/http"
	"strings"
)

func (app *application) authTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			app.unauthorizedErrorResponse(w, "missing authorization header")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			app.unauthorizedErrorResponse(w, "invalid authorization header")
			return
		}

		token := parts[1]
		userID, err := app.validateJWTToken(token)
		if err != nil {
			app.unauthorizedErrorResponse(w, "invalid token")
			return
		}

		user, err := app.store.Users.GetByID(r.Context(), userID)
		if err != nil {
			app.unauthorizedErrorResponse(w, "invalid token")
			return
		}

		ctx := context.WithValue(r.Context(), userCtx, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) unauthorizedErrorResponse(w http.ResponseWriter, message string) {
	writeJSONError(w, http.StatusUnauthorized, message, "unauthorized")
}
