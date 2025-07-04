package config

import (
    "log"
    "os"
)

// Config holds application-wide configuration settings.
type Config struct {
    DatabaseURL string
    ServerPort  string
}

// LoadConfig reads configuration from environment variables.
func LoadConfig() *Config {
    dbURL := os.Getenv("DATABASE_URL")
    if dbURL == "" {
        // Default for local development if not set
        dbURL = "postgres://user:password@localhost:5432/guardianai_db?sslmode=disable"
        log.Println("DATABASE_URL environment variable not set, using default for local development.")
    }

    serverPort := os.Getenv("SERVER_PORT")
    if serverPort == "" {
        serverPort = ":8080"
        log.Println("SERVER_PORT environment variable not set, using default :8080.")
    }

    return &Config{
        DatabaseURL: dbURL,
        ServerPort:  serverPort,
    }
}