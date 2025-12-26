package postgres

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// Client represents a PostgreSQL/SQLite database client
type Client struct {
	DB     *sql.DB
	DBType string
}

// NewClient creates a new database client
func NewClient() (*Client, error) {
	dbType := os.Getenv("DB_TYPE")
	if dbType == "" {
		dbType = "sqlite" // Default to SQLite for MVP
	}

	var db *sql.DB
	var err error

	if dbType == "postgres" {
		connStr := os.Getenv("DATABASE_URL")
		if connStr == "" {
			connStr = "postgres://user:password@localhost/dbname?sslmode=disable"
		}
		db, err = sql.Open("postgres", connStr)
	} else {
		dbPath := os.Getenv("SQLITE_PATH")
		if dbPath == "" {
			dbPath = "./data/platform.db"
		}
		db, err = sql.Open("sqlite3", dbPath)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	return &Client{DB: db, DBType: dbType}, nil
}

// Close closes the database connection
func (c *Client) Close() error {
	return c.DB.Close()
}

