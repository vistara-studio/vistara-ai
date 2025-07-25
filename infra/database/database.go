package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
	"github.com/vistara-studio/vistara-ai/infra/config"
)

// Database struct holds the database connection
type Database struct {
	DB *sql.DB
}

// NewConnection creates a new database connection
func NewConnection(cfg *config.Config) (*Database, error) {
	// Connect to database
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.DBMaxConnections)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Hour)

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Successfully connected to PostgreSQL database")

	return &Database{DB: db}, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	if d.DB != nil {
		return d.DB.Close()
	}
	return nil
}

// Ping tests the database connection
func (d *Database) Ping() error {
	if d.DB == nil {
		return fmt.Errorf("database connection is nil")
	}
	return d.DB.Ping()
}

// GetDB returns the underlying sql.DB instance
func (d *Database) GetDB() *sql.DB {
	return d.DB
}
