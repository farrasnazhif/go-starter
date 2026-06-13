package router

import (
	"net/http"
	"time"

	"github.com/farrasnazhif/go-starter/internal/handler"
	mw "github.com/farrasnazhif/go-starter/internal/middleware"
	"github.com/farrasnazhif/go-starter/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Config struct {
	FrontendURL     string
	JWTSecret       string
	RateLimit       int
	RegisterLimit   int
	RateLimitWindow time.Duration
}

func New(cfg Config, authHandler *handler.AuthHandler, userHandler *handler.UserHandler, userRepo repository.UserRepository) http.Handler {
	r := chi.NewRouter()

	r.Use(mw.CORS(cfg.FrontendURL))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	registerLimiter := mw.RateLimit(cfg.RegisterLimit, cfg.RateLimitWindow)
	apiLimiter := mw.RateLimit(cfg.RateLimit, cfg.RateLimitWindow)
	auth := mw.Auth(cfg.JWTSecret, userRepo)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"status":"ok"}`))
		})

		r.Route("/auth", func(r chi.Router) {
			r.With(registerLimiter).Post("/register", authHandler.Register)
			r.With(registerLimiter).Post("/otp/send", authHandler.SendOTP)
			r.With(registerLimiter).Post("/otp/verify", authHandler.VerifyOTP)
			r.With(apiLimiter).Post("/login", authHandler.Login)
		})

		r.Route("/users", func(r chi.Router) {
			r.Use(auth)
			r.Get("/me", userHandler.GetProfile)
			r.Get("/billing", userHandler.GetBilling)
		})
	})

	return r
}
