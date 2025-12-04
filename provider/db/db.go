package db

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver via database/sql
)

var DB *sql.DB

// Init initializes the database connection pool
func Init() error {
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	port := os.Getenv("DB_PORT")
	host := os.Getenv("DB_HOST")
	dbName := "chiliz_chain_pulse"

	// Build connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbName)

	slog.Info("Connecting to PostgreSQL database", "host", host, "port", port, "dbname", dbName)

	var err error
	DB, err = sql.Open("pgx", connStr)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool settings
	DB.SetMaxOpenConns(25)                  // Maximum number of open connections
	DB.SetMaxIdleConns(5)                   // Maximum number of idle connections
	DB.SetConnMaxLifetime(5 * time.Minute)  // Maximum connection lifetime
	DB.SetConnMaxIdleTime(10 * time.Minute) // Maximum idle connection time

	// Verify the connection
	if err := DB.Ping(); err != nil {
		DB.Close() // Close the failed connection
		DB = nil   // Set to nil so we know it's not usable
		return fmt.Errorf("failed to ping database: %w", err)
	}

	slog.Info("Successfully connected to PostgreSQL database")
	return nil
}

// Close closes the database connection pool
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
