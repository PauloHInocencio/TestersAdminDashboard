package main

import (
	"fmt"
	"log"
	"os"

	"github.com/PauloHInocencio/testers-admin-dashboard/api"
	"github.com/PauloHInocencio/testers-admin-dashboard/db"
)

func main() {
	// Get SSL mode from environment (default to require)
	sslMode := os.Getenv("DB_SSLMODE")

	dbUrl := fmt.Sprintf(
		"postgres://%s:%s@%s:5432/%s?sslmode=%s&sslrootcert=/app/certs/root.crt",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_NAME"),
		sslMode,
	)

	log.Printf("Connection to database with SSL mode: %s", sslMode)

	storage := db.NewStorage(dbUrl)
	storage.MigrateUp()

	// Get server port from environment (default to 8080)
	serverPort := os.Getenv("PORT")
	if serverPort == "" {
		serverPort = "8080"
	}

	server := api.NewServer(serverPort, storage)
	log.Printf("Starting Testers Admin Dashboard API on port %s", serverPort)

	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
