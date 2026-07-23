package main

import (
	"crypto/tls"
	"embed"
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

//go:embed .env
var envFile embed.FS

func loadEnvFromEmbeddedFile() {
	content, err := envFile.ReadFile(".env")
	if err != nil {
		log.Fatalf("Error reading .env file: %v", err)
	}
	tempFile, err := os.CreateTemp("", ".env")
	if err != nil {
		log.Fatalf("Error creating temp .env file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	_, err = tempFile.Write(content)
	if err != nil {
		log.Fatalf("Error writing to temp .env file: %v", err)
	}
	err = tempFile.Close()
	if err != nil {
		log.Fatalf("Error closing temp file: %v", err)
	}
	err = godotenv.Load(tempFile.Name())
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

func main() {
	// Only in Development, for running source code...
	// err := godotenv.Load()
	// if err != nil {
	// 	return
	// }

	// load env variables from the embedded .env file
	loadEnvFromEmbeddedFile()

	db, err := sqlconnect.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	cert := os.Getenv("CERT_FILE")
	key := os.Getenv("KEY_FILE")

	port := os.Getenv("API_PORT")

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS10,
	}

	_ = mw.NewRateLimiter(5, time.Minute)

	HPPOptions := mw.HPPOptions{
		CheckQuery:                  true,
		CheckBody:                   true,
		CheckBodyOnlyForContentType: "application/json",
		Whitelist:                   []string{"page", "limit", "sort", "order", "search", "category", "status", "type", "from", "to", "start_date", "end_date", "min", "max", "min_amount", "max_amount", "token"},
	}
	router := router.MainRouter()
	JWTMiddleware := mw.ExcludeMiddlewares(mw.JWTMiddleware, "/execs/login", "/execs/forgotpassword", "/execs/reset/resetpassword")
	SecureMux := utils.ApplyMiddlewares(router, mw.Compression, mw.SecurityHeaders, mw.HppMiddleware(HPPOptions), mw.XSSMiddleware, JWTMiddleware, mw.ResponseTimeMiddleware, mw.Cors)

	// Create Custom Server
	server := &http.Server{
		Addr:      port,
		Handler:   SecureMux,
		TLSConfig: tlsConfig,
	}
	fmt.Println("Server is running on port", port)

	err = server.ListenAndServeTLS(cert, key)
	if err != nil {
		log.Fatal("Cannot start TLS server", err)
	}
}

// We use HTTP methods because we want to standardize API architecture.
