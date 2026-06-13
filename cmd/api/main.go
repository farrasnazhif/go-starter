package main

import (
	"time"

	"github.com/farrasnazhif/go-starter/internal/db"
	"github.com/farrasnazhif/go-starter/internal/env"
	"github.com/farrasnazhif/go-starter/internal/mailer"
	"github.com/farrasnazhif/go-starter/internal/store"
	"go.uber.org/zap"
)

const version = "0.0.1"

//	@title			go-starter API
//	@version		0.0.1
//	@description	User management and authentication API
//	@termsOfService	https://go-starter.example.com/terms

//	@contact.name	API Support
//	@contact.url	https://go-starter.example.com/support
//	@contact.email	support@go-starter.example.com

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath					/api/v1
// @schemes					http https
//
// @securityDefinitions.apikey	ApiKeyAuth
// @in							header
// @name						Authorization
// @description				JWT Bearer token for API authentication
func main() {
	frontendURL := env.GetString("FRONTEND_URL", "http://localhost:3000")
	cfg := config{
		addr:        env.GetString("ADDR", ":8080"),
		apiURL:      env.GetString("EXTERNAL_URL", "localhost:8080"),
		frontendURL: frontendURL,
		db: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost/go-starter?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		env: env.GetString("ENV", "development"),
		jwt: jwtConfig{
			secret: env.GetString("JWT_SECRET", "your-secret-key"),
		},
		mail: mailConfig{
			exp:       time.Hour * 24 * 3,
			fromEmail: env.GetString("FROM_EMAIL", "onboarding@resend.dev"),
			resend: resendConfig{
				apiKey: env.GetString("RESEND_API_KEY", ""),
			},
		},
		rateLimiter: DefaultRateLimitConfig(),
	}

	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	db, err := db.New(
		cfg.db.addr,
		cfg.db.maxOpenConns,
		cfg.db.maxIdleConns,
		cfg.db.maxIdleTime,
	)
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()
	logger.Info("database connection pool established")

	store := store.NewStorage(db)
	mailer := mailer.NewResend(cfg.mail.resend.apiKey, cfg.mail.fromEmail)

	app := &application{
		config: cfg,
		store:  store,
		logger: logger,
		mailer: mailer,
	}

	mux := app.mount()
	logger.Fatal(app.run(mux))
}
