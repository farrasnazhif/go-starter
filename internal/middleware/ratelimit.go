package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/httprate"
)

func RateLimit(limit int, window time.Duration) func(http.Handler) http.Handler {
	return httprate.LimitByIP(limit, window)
}
