package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/farrasnazhif/go-starter/internal/database"
	_ "github.com/farrasnazhif/go-starter/docs"
	"github.com/farrasnazhif/go-starter/internal/handler"
	"github.com/farrasnazhif/go-starter/internal/mailer"
	"github.com/farrasnazhif/go-starter/internal/paypal"
	"github.com/farrasnazhif/go-starter/internal/repository"
	"github.com/farrasnazhif/go-starter/internal/router"
	"github.com/farrasnazhif/go-starter/internal/service"
)

//	@title			Go Starter API
//	@version		1.0
//	@description	User authentication, profile management, and PayPal billing API

//	@host		localhost:8080
//	@BasePath	/api/v1
//	@schemes	http https

//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Bearer token (e.g. "Bearer xxx")

func main() {
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
	billingRepo := repository.NewBillingRepository(db)

	// PayPal
	paypalClient := paypal.NewClient(paypal.Config{
		ClientID:  getEnv("PAYPAL_CLIENT_ID", ""),
		Secret:    getEnv("PAYPAL_CLIENT_SECRET", ""),
		BaseURL:   getEnv("PAYPAL_BASE_URL", "https://api-m.sandbox.paypal.com"),
		PlanID:    getEnv("PAYPAL_PRO_PLAN_ID", ""),
		WebhookID: getEnv("PAYPAL_WEBHOOK_ID", ""),
		ReturnURL: getEnv("PAYPAL_RETURN_URL", frontendURL+"/billing/success"),
		CancelURL: getEnv("PAYPAL_CANCEL_URL", frontendURL+"/billing/cancel"),
	})

	// Mailer
	m := mailer.NewResend(getEnv("RESEND_API_KEY", ""), getEnv("FROM_EMAIL", "onboarding@resend.dev"))

	// Services
	authSvc := service.NewAuthService(userRepo, otpRepo, m, jwtSecret, env)
	userSvc := service.NewUserService(userRepo)
	billingSvc := service.NewBillingService(billingRepo, userRepo, paypalClient)

	// Handlers
	authHandler := handler.NewAuthHandler(authSvc)
	userHandler := handler.NewUserHandler(userSvc)
	billingHandler := handler.NewBillingHandler(billingSvc)

	// Router
	mux := router.New(router.Config{
		FrontendURL:     frontendURL,
		JWTSecret:       jwtSecret,
		RateLimit:       30,
		RegisterLimit:   5,
		RateLimitWindow: time.Minute,
	}, authHandler, userHandler, billingHandler, userRepo)

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
