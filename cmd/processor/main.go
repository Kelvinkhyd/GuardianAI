package main

import (
    "context"
    "encoding/json"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    kafkalib "github.com/segmentio/kafka-go" // <--- Changed to 'kafkalib' alias

    "github.com/Kelvinkhyd/GuardianAI/internal/config"
    "github.com/Kelvinkhyd/GuardianAI/internal/kafka" // This is YOUR internal/kafka package
    "github.com/Kelvinkhyd/GuardianAI/internal/models"
    "github.com/Kelvinkhyd/GuardianAI/internal/repository"
    "github.com/Kelvinkhyd/GuardianAI/internal/database"
)

func main() {
    cfg := config.LoadConfig()

    // Establish database connection (for updating alert status later)
    dbConn, err := database.NewDBConnection(cfg.DatabaseURL)
    if err != nil {
        log.Fatalf("Processor failed to connect to database: %v", err)
    }
    defer dbConn.Close()

    alertRepo := repository.NewPgAlertRepository(dbConn.DB)


    // Initialize Kafka Consumer
    // Group ID is important for consumer groups to distribute partitions
    consumerGroupID := "guardianai-alert-processor-group"
    // This kafka.NewConsumer now correctly refers to YOUR internal/kafka package
    kafkaConsumer := kafka.NewConsumer(cfg.KafkaBrokers, cfg.KafkaTopic, consumerGroupID)
    defer kafkaConsumer.Close()

    log.Println("GuardianAI Alert Processor starting...")

    // Context for graceful shutdown
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Handle OS signals for graceful shutdown
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        sig := <-sigChan
        log.Printf("Received signal %v, shutting down...", sig)
        cancel() // Cancel the context to stop the consumer loop
    }()

    // Start consuming messages
    // Changed 'kafka.Message' to 'kafkalib.Message' here
    kafkaConsumer.ConsumeMessages(ctx, func(message kafkalib.Message) error {
        var alert models.SecurityAlert
        err := json.Unmarshal(message.Value, &alert)
        if err != nil {
            log.Printf("ERROR Processor: Failed to unmarshal alert JSON: %v", err)
            // In a real system, send to a dead-letter queue or log for manual review
            return nil // Do not reprocess this bad message
        }

        log.Printf("Processor: Received alert ID: %s, Source: %s, Category: %s, Value: %s",
            alert.ID, alert.Source, alert.Category, string(message.Value))

        // --- Placeholder for Alert Processing Logic ---
        // This is where the AI analysis, enrichment, and orchestration will happen.
        // For now, we'll just simulate some work and update the status in DB.

        processingCtx, processingCancel := context.WithTimeout(context.Background(), 3*time.Second)
        defer processingCancel()

        log.Printf("Processor: Simulating AI analysis and enrichment for alert ID: %s...", alert.ID)
        time.Sleep(2 * time.Second) // Simulate work

        // Update alert status in database
        newStatus := "processed" // Or "analyzed", "triaged", based on AI outcome
        err = alertRepo.UpdateAlertStatus(processingCtx, alert.ID, newStatus)
        if err != nil {
            log.Printf("ERROR Processor: Failed to update status for alert %s to %s: %v", alert.ID, newStatus, err)
            return err // Re-queue if DB update failed (Kafka will retry on non-nil error)
        }
        log.Printf("Processor: Alert ID: %s status updated to '%s'", alert.ID, newStatus)

        return nil // Message processed successfully
    })

    log.Println("GuardianAI Alert Processor stopped.")
}