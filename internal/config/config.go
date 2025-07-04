package config

import (
    "log"
    "os"
)

// Config holds application-wide configuration settings.
type Config struct {
    DatabaseURL  string
    ServerPort   string
    KafkaBrokers []string // Add Kafka brokers
    KafkaTopic   string   // Add Kafka topic name
}

// LoadConfig reads configuration from environment variables.
func LoadConfig() *Config {
    dbURL := os.Getenv("DATABASE_URL")
    if dbURL == "" {
        dbURL = "postgres://user:password@localhost:5432/guardianai_db?sslmode=disable"
        log.Println("DATABASE_URL environment variable not set, using default for local development.")
    }

    serverPort := os.Getenv("SERVER_PORT")
    if serverPort == "" {
        serverPort = ":8080"
        log.Println("SERVER_PORT environment variable not set, using default :8080.")
    }

    // Kafka Configuration
    kafkaBrokersStr := os.Getenv("KAFKA_BROKERS")
    if kafkaBrokersStr == "" {
        kafkaBrokersStr = "localhost:9092" // Default for local Docker Kafka
        log.Println("KAFKA_BROKERS environment variable not set, using default localhost:9092.")
    }
    kafkaBrokers := []string{kafkaBrokersStr} // Simple for now, can be split by comma later

    kafkaTopic := os.Getenv("KAFKA_TOPIC")
    if kafkaTopic == "" {
        kafkaTopic = "security_alerts" // Default topic name
        log.Println("KAFKA_TOPIC environment variable not set, using default 'security_alerts'.")
    }

    return &Config{
        DatabaseURL:  dbURL,
        ServerPort:   serverPort,
        KafkaBrokers: kafkaBrokers,
        KafkaTopic:   kafkaTopic,
    }
}