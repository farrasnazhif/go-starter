package main

import (
	"net/http"
	"time"

	"github.com/go-chi/httprate"
)

// RateLimitConfig holds configuration for rate limiting
type RateLimitConfig struct {
	RegisterLimit   int
	ActivateLimit   int
	GeneralAPILimit int
	WindowDuration  time.Duration
}

// DefaultRateLimitConfig returns default rate limit settings
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		RegisterLimit:   5,  // 5 requests
		ActivateLimit:   10, // 10 requests
		GeneralAPILimit: 30, // 30 requests
		WindowDuration:  time.Minute,
	}
}

// registerRateLimiter returns a rate limiter middleware for registration endpoint
func (app *application) registerRateLimiter() func(http.Handler) http.Handler {
	return httprate.LimitByIP(
		app.config.rateLimiter.RegisterLimit,
		app.config.rateLimiter.WindowDuration,
	)
}

// activateRateLimiter returns a rate limiter middleware for activation endpoint
// func (app *application) activateRateLimiter() func(http.Handler) http.Handler {
// 	return httprate.LimitByIP(
// 		app.config.rateLimiter.ActivateLimit,
// 		app.config.rateLimiter.WindowDuration,
// 	)
// }

// apiRateLimiter returns a rate limiter middleware for general API endpoints
func (app *application) apiRateLimiter() func(http.Handler) http.Handler {
	return httprate.LimitByIP(
		app.config.rateLimiter.GeneralAPILimit,
		app.config.rateLimiter.WindowDuration,
	)
}
