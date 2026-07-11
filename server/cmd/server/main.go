package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/joho/godotenv"
	"github.com/resend/resend-go/v3"

	"github.com/trnahnh/katana-id/internal/auth"
	"github.com/trnahnh/katana-id/internal/brand"
	"github.com/trnahnh/katana-id/internal/check"
	"github.com/trnahnh/katana-id/internal/db"
	"github.com/trnahnh/katana-id/internal/gemini"
	"github.com/trnahnh/katana-id/internal/health"
	"github.com/trnahnh/katana-id/util"
)

func main() {
	godotenv.Load()
	util.RequireEnvs()

	ctx := context.Background()
	client, err := db.Connect(ctx, os.Getenv("DB_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	serverURL := os.Getenv("SERVER_URL")
	emailClient := resend.NewClient(os.Getenv("RESEND_API_KEY"))
	authHandler := &auth.Handler{
		DB:                 client,
		EmailClient:        emailClient,
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		GitHubClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		GitHubClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		ServerURL:          serverURL,
		FrontendURL:        os.Getenv("FRONTEND_URL"),
		// Browsers refuse Secure cookies over http://localhost, so only set the
		// flag when the server is actually served over https.
		SecureCookies: strings.HasPrefix(serverURL, "https"),
	}

	// Gemini is optional: when the key is absent or the client fails to build,
	// the brand handler falls back to a local name generator.
	var geminiClient *gemini.Client
	if gc, err := gemini.New(ctx, os.Getenv("GEMINI_API_KEY"), os.Getenv("GEMINI_MODEL")); err != nil {
		log.Print("⚠️  Gemini disabled: ", err)
	} else {
		geminiClient = gc
		log.Print("✨ Gemini brand generation enabled")
	}

	brandHandler := &brand.Handler{
		DB:     client,
		Gemini: geminiClient,
		Store:  check.NewStore(),
		CheckOpts: check.Options{
			GitHubToken:  os.Getenv("GITHUB_TOKEN"),
			BraveAPIKey:  os.Getenv("BRAVE_API_KEY"),
			TwitterToken: os.Getenv("TWITTER_BEARER_TOKEN"),
		},
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
		r.Get("/google", authHandler.GoogleLogin)
		r.Get("/google/callback", authHandler.GoogleCallback)
		r.Get("/github", authHandler.GitHubLogin)
		r.Get("/github/callback", authHandler.GitHubCallback)
	})

	r.Route("/generate", func(r chi.Router) {
		r.Use(authHandler.RequireAuth)
		r.Post("/", brandHandler.Generate)
		r.Get("/{id}/stream", brandHandler.Stream)
	})

	r.Route("/kits", func(r chi.Router) {
		r.Use(authHandler.RequireAuth)
		r.Get("/{id}/pdf", brandHandler.PDF)
	})

	port := os.Getenv("PORT")
	log.Print("🍊 Server is starting on port ", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal(err)
	}
}
