package database

import (
    "database/sql"
    "log"

    _ "github.com/lib/pq" // PostgreSQL driver
)

// DB holds our database connection pool.
type DB struct {
    *sql.DB
}

// NewDBConnection establishes a new PostgreSQL database connection.
func NewDBConnection(databaseURL string) (*DB, error) {
    db, err := sql.Open("postgres", databaseURL)
    if err != nil {
        return nil, err
    }

    // Ping the database to ensure connection is established
    err = db.Ping()
    if err != nil {
        return nil, err
    }

    db.SetMaxOpenConns(25) // Max open connections
    db.SetMaxIdleConns(25) // Max idle connections
    // db.SetConnMaxLifetime(5 * time.Minute) // Connection max lifetime (uncomment if needed)

    log.Println("Successfully connected to the database!")
    return &DB{db}, nil
}