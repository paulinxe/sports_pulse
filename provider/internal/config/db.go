package config

import (
    "database/sql"
    "fmt"
    "log/slog"
    "os"
    "time"

    _ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver via database/sql
)

// InitDB creates and configures a database connection pool.
// The caller is responsible for calling db.Close() when done.
func InitDB() (*sql.DB, error) {
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	port := os.Getenv("DB_PORT")
	host := os.Getenv("DB_HOST")
	dbName := "sports_pulse"

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbName)

	slog.Info("Connecting to PostgreSQL database", "host", host, "port", port, "dbname", dbName)

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(10 * time.Minute)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	slog.Info("Successfully connected to PostgreSQL database")
	return db, nil
}
