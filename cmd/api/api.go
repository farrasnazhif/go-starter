package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/farrasnazhif/go-starter/docs"
	"github.com/farrasnazhif/go-starter/internal/mailer"
	"github.com/farrasnazhif/go-starter/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"
)

type application struct {
	config config
	store  store.Storage
	logger *zap.SugaredLogger
	mailer mailer.Client
}

type config struct {
	addr        string
	db          dbConfig
	env         string
	apiURL      string
	mail        mailConfig
	frontendURL string
	rateLimiter RateLimitConfig
	jwt         jwtConfig
}

type jwtConfig struct {
	secret string
}

type mailConfig struct {
	resend    resendConfig
	fromEmail string
	exp       time.Duration
}

type resendConfig struct {
	apiKey string
}

type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{
			app.config.frontendURL,
			"http://localhost:3000",
			"http://127.0.0.1:3000",
		},
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "Cache-Control"},
		ExposedHeaders:   []string{"Link", "Content-Disposition"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", app.healthCheckHandler)

		docsURL := fmt.Sprintf("%s/swagger/doc.json", app.config.addr)
		r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL(docsURL)))

		r.Route("/users", func(r chi.Router) {
			r.With(app.authTokenMiddleware).Get("/me", app.getUserHandler)
			r.With(app.authTokenMiddleware).Get("/billing", app.getUserBillingHandler)
		})

		r.Route("/auth", func(r chi.Router) {
			r.Route("/otp", func(r chi.Router) {
				r.With(app.registerRateLimiter()).Post("/send", app.sendRegistrationOTPHandler)
				r.With(app.registerRateLimiter()).Post("/verify", app.verifyRegistrationOTPHandler)
			})
			r.With(app.registerRateLimiter()).Post("/register", app.registerUserHandler)
			r.With(app.apiRateLimiter()).Post("/login", app.loginUserHandler)
		})
	})

	return r
}

func (app *application) run(mux http.Handler) error {
	docs.SwaggerInfo.Version = version
	docs.SwaggerInfo.Host = app.config.apiURL
	docs.SwaggerInfo.BasePath = "/api/v1"

	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: time.Second * 300,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}

	app.logger.Infow("Server has started at", "addr", app.config.addr, "env", app.config.env)

	server := NewServer(srv, app)
	return server.ListenAndServeWithGracefulShutdown()
}
