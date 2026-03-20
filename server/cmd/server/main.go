package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/joho/godotenv"
	"github.com/resend/resend-go/v3"

	"github.com/trnahnh/katana-id/internal/auth"
	"github.com/trnahnh/katana-id/internal/check"
	"github.com/trnahnh/katana-id/internal/db"
	"github.com/trnahnh/katana-id/internal/health"
	"github.com/trnahnh/katana-id/util"
)

func main() {
	godotenv.Load()
	util.RequireEnvs()

	ctx := context.Background()
	queries, pool, err := db.Connect(ctx, os.Getenv("DB_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	if err := db.RunMigrations(os.Getenv("DB_URL")); err != nil {
		log.Fatal("Failed to run migration: ", err)
	}

	emailClient := resend.NewClient(os.Getenv("RESEND_API_KEY"))
	authHandler := &auth.Handler{Queries: queries, EmailClient: emailClient}

	checkHandler := &check.Handler{
		Store:        check.NewStore(),
		GitHubToken:  os.Getenv("GITHUB_TOKEN"),
		BraveAPIKey:  os.Getenv("BRAVE_API_KEY"),
		TwitterToken: os.Getenv("TWITTER_BEARER_TOKEN"),
	}

	r := chi.NewRouter()

	r.Use(cors.Handler(util.CorsOptions()))
	r.Use(httprate.Limit(60, 1*time.Minute))

	r.Get("/health", health.Health)

	r.Route("/auth", func(r chi.Router) {
		r.With(httprate.Limit(1, 1*time.Minute)).Post("/send-otp", authHandler.SendOTP)
		r.Post("/verify-otp", authHandler.VerifyOTP)
		r.Get("/me", authHandler.Me)
		r.Post("/logout", authHandler.Logout)
	})

	r.Route("/check", func(r chi.Router) {
		r.Use(authHandler.RequireAuth)
		r.Post("/", checkHandler.Check)
		r.Get("/{id}", checkHandler.Stream)
	})

	port := os.Getenv("PORT")
	log.Print("🍊 Server is starting on port ", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal(err)
	}
}
