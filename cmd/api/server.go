package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	mw "schoolmanagementapi/internal/api/middlewares"
	"schoolmanagementapi/internal/api/repositories/sqlconnect"
	"schoolmanagementapi/internal/api/router"
	"schoolmanagementapi/internal/pkg/utils"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env for local development.
	// In production (e.g. Render), environment variables are provided
	// by the hosting platform, so it's fine if .env doesn't exist.
	_ = godotenv.Load("cmd/api/.env")

	// Database connection
	db, err := sqlconnect.ConnectDB()
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	defer db.Close()

	// Resolve server port.
	// Render provides PORT automatically.
	port := os.Getenv("PORT")
	if port == "" {
		port = os.Getenv("API_PORT")
	}
	if port == "" {
		port = "8080"
	}

	addr := ":" + port

	// Configure middlewares
	rl := mw.NewRateLimiter(5, time.Minute)

	HPPOptions := mw.HPPOptions{
		CheckQuery:                  true,
		CheckBody:                   true,
		CheckBodyOnlyForContentType: "application/json",
		Whitelist:                   []string{"page", "limit", "sort", "order", "search", "category", "status", "type", "from", "to", "start_date", "end_date", "min", "max", "min_amount", "max_amount", "token"},
	}

	r := router.MainRouter()

	JWTMiddleware := mw.ExcludeMiddlewares(
		mw.JWTMiddleware,
		"/",
		"/execs/login",
		"/execs/forgotpassword",
		"/execs/reset/resetpassword",
	)

	SecureMux := utils.ApplyMiddlewares(
		r,
		mw.Compression,
		mw.SecurityHeaders,
		mw.HppMiddleware(HPPOptions),
		mw.XSSMiddleware,
		JWTMiddleware,
		mw.ResponseTimeMiddleware,
		rl.RateLimiterMiddleware,
		mw.Cors,
	)

	// Create server
	server := &http.Server{
		Addr:    addr,
		Handler: SecureMux,
	}

	// Use HTTPS locally if certificate files exist.
	cert := os.Getenv("CERT_FILE")
	key := os.Getenv("KEY_FILE")

	if cert != "" && key != "" {
		if _, err := os.Stat(cert); err == nil {
			if _, err := os.Stat(key); err == nil {

				server.TLSConfig = &tls.Config{
					MinVersion: tls.VersionTLS12,
				}

				fmt.Printf("HTTPS server running on https://localhost%s\n", addr)
				log.Fatal(server.ListenAndServeTLS(cert, key))
			}
		}
	}

	// Production / Render
	fmt.Printf("HTTP server running on port %s\n", port)
	log.Fatal(server.ListenAndServe())
}
