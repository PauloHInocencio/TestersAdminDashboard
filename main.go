package main

import (
	"fmt"
	"os"

	"github.com/PauloHInocencio/testers-admin-dashboard/db"
)

func main() {
	dbUrl := fmt.Sprintf(
		"postgres://%s:%s@%s:5432/%s?sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_NAME"),
	)

	storage := db.NewStorage(dbUrl)
	storage.MigrateUp()

}
