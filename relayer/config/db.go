package config

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
func InitDB() (shouldClose bool, err error) {
	if DB != nil {
		// As the DB is already intialized, the caller should not close it.
		return shouldClose, nil
	}

    user := os.Getenv("DB_USER")
    password := os.Getenv("DB_PASSWORD")
    port := os.Getenv("DB_PORT")
    host := os.Getenv("DB_HOST")
    dbName := "sports_pulse"

    // Build connection string
    connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
        host, port, user, password, dbName)

    slog.Info("Connecting to PostgreSQL database", "host", host, "port", port, "dbname", dbName)

    DB, err = sql.Open("pgx", connStr)
    if err != nil {
        return shouldClose, fmt.Errorf("failed to open database connection: %w", err)
    }

    // Configure connection pool settings
    DB.SetMaxOpenConns(25)
    DB.SetMaxIdleConns(5)  
    DB.SetConnMaxLifetime(5 * time.Minute) 
    DB.SetConnMaxIdleTime(10 * time.Minute)

    // Verify the connection
    if err := DB.Ping(); err != nil {
        _ = DB.Close()
        DB = nil   // Set to nil so we know it's not usable
        return shouldClose, fmt.Errorf("failed to ping database: %w", err)
    }

    slog.Info("Successfully connected to PostgreSQL database")
    shouldClose = true
	return shouldClose, nil
}

// Close closes the database connection pool
func Close() error {
	if DB == nil {
		return nil // DB is already closed
	}

	err := DB.Close()
	if err == nil {
		DB = nil
	}

	return err
}
