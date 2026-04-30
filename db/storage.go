package db

import (
	"database/sql"
	"embed"
	"log"

	"github.com/PauloHInocencio/testers-admin-dashboard/db/database"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

/**
      sql.Open("pgx", databaseURL), Uses the pgx driver - a modern, pure-Go PostgreSQL driver.
	  Characteristics:
	  - Performance: Generally faster with better connection pooling
	  - Features: Supports more PostgreSQL-specific features (arrays, JSON, custom types, LISTEN/NOTIFY)
	  - Active development: Actively maintained and updated
	  - Import required: _ "github.com/jackc/pgx/v5/stdlib" (blank import to register driver)
	  - Driver name: "pgx"

	  sql.Open("postgres", databaseURL), Uses lib/pq - the older, traditional PostgreSQL driver.
	  Characteristics:
	  - Legacy: In maintenance mode (still works but minimal new features)
	  - Compatibility: Widely used in older codebases
	  - Import required: _ "github.com/lib/pq" (blank import to register driver)
	  - Driver name: "postgres"
*/

type Storage struct {
	DB      *sql.DB
	Queries *database.Queries
}

//go:embed migrations/*.sql
var migrationsFS embed.FS

func NewStorage(dbUrl string) *Storage {
	db, err := sql.Open("pgx", dbUrl)
	if err != nil {
		log.Fatalf("Can't connect to database: %s", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Couldn't ping database: %s", err)
	}

	return &Storage{
		DB:      db,
		Queries: database.New(db),
	}
}

func (s *Storage) MigrateUp() {
	goose.SetBaseFS(migrationsFS)

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("Error setting goose dialect: %s", err)
	}
	log.Println("Running database migrations...")
	if err := goose.Up(s.DB, "migrations"); err != nil {
		log.Fatalf("Error trying migrate up: %s", err)
	}
	log.Println("Migrations completed successfully")
}

func (s *Storage) MigrateDownToZero() {
	goose.SetBaseFS(migrationsFS)

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("Error setting goose dialect: %s", err)
	}
	log.Println("Running database down migrations...")
	if err := goose.DownTo(s.DB, "migrations", 0); err != nil {
		log.Fatalf("Error trying migrate down to zero: %s", err)
	}
	log.Println("Migrations completed successfully")
}
