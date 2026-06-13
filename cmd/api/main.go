package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/farrasnazhif/go-starter/internal/database"
	"github.com/farrasnazhif/go-starter/internal/handler"
	"github.com/farrasnazhif/go-starter/internal/mailer"
	"github.com/farrasnazhif/go-starter/internal/repository"
	"github.com/farrasnazhif/go-starter/internal/router"
	"github.com/farrasnazhif/go-starter/internal/service"
)

func main() {
	// Config
	addr := getEnv("ADDR", ":8080")
	jwtSecret := getEnv("JWT_SECRET", "your-secret-key")
	frontendURL := getEnv("FRONTEND_URL", "http://localhost:3000")
	env := getEnv("ENV", "development")

	// Database
	db, err := database.New(database.Config{
		Addr:         getEnv("DB_ADDR", "postgres://admin:adminpassword@localhost/go-starter?sslmode=disable"),
		MaxOpenConns: 30,
		MaxIdleConns: 30,
		MaxIdleTime:  "15m",
	})
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}
	defer db.Close()
	log.Println("database connection pool established")

	// Repositories
	userRepo := repository.NewUserRepository(db)
	otpRepo := repository.NewOTPRepository(db)

	// Mailer
	m := mailer.NewResend(getEnv("RESEND_API_KEY", ""), getEnv("FROM_EMAIL", "onboarding@resend.dev"))

	// Services
	authSvc := service.NewAuthService(userRepo, otpRepo, m, jwtSecret, env)
	userSvc := service.NewUserService(userRepo)

	// Handlers
	authHandler := handler.NewAuthHandler(authSvc)
	userHandler := handler.NewUserHandler(userSvc)

	// Router
	mux := router.New(router.Config{
		FrontendURL:     frontendURL,
		JWTSecret:       jwtSecret,
		RateLimit:       30,
		RegisterLimit:   5,
		RateLimitWindow: time.Minute,
	}, authHandler, userHandler, userRepo)

	// Server
	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  10 * time.Second,
		IdleTimeout:  time.Minute,
	}

	log.Printf("server starting on %s [env=%s]", addr, env)
	log.Fatal(srv.ListenAndServe())
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
