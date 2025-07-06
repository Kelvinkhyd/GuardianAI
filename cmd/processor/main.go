package main

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"        // <--- ADDED THIS IMPORT
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    kafkalib "github.com/segmentio/kafka-go"

    "github.com/Kelvinkhyd/GuardianAI/internal/config"
    "github.com/Kelvinkhyd/GuardianAI/internal/database"
    "github.com/Kelvinkhyd/GuardianAI/internal/kafka"
    "github.com/Kelvinkhyd/GuardianAI/internal/models"
    "github.com/Kelvinkhyd/GuardianAI/internal/repository"
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
    consumerGroupID := "guardianai-alert-processor-group"
    kafkaConsumer := kafka.NewConsumer(cfg.KafkaBrokers, cfg.KafkaTopic, consumerGroupID)
    defer kafkaConsumer.Close()

    // Define the AI Service URL (accessible within the main function scope)
    // Use Docker service name for inter-container communication: guardianai_ai_service
    aiServiceURL := "http://localhost:8000/analyze-alert"
    log.Printf("AI Service URL: %s", aiServiceURL)


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
    kafkaConsumer.ConsumeMessages(ctx, func(message kafkalib.Message) error {
        var alert models.SecurityAlert // This will be the original alert from Kafka
        err := json.Unmarshal(message.Value, &alert)
        if err != nil {
            log.Printf("ERROR Processor: Failed to unmarshal alert JSON from Kafka: %v", err)
            return nil // Do not reprocess this bad message, skip it
        }

        log.Printf("Processor: Received alert ID: %s, Source: %s from Kafka for analysis.", alert.ID, alert.Source)

        // --- Step 1: Send alert to AI Service for analysis ---
        alertJSON, err := json.Marshal(alert)
        if err != nil {
            log.Printf("ERROR Processor: Failed to marshal alert for AI service: %v", err)
            return err // Re-queue if marshaling failed (shouldn't happen often)
        }

        req, err := http.NewRequestWithContext(context.Background(), "POST", aiServiceURL, bytes.NewBuffer(alertJSON))
        if err != nil {
            log.Printf("ERROR Processor: Failed to create AI service request: %v", err)
            return err // Re-queue
        }
        req.Header.Set("Content-Type", "application/json")

        client := &http.Client{Timeout: 10 * time.Second} // Set a timeout for AI service call
        resp, err := client.Do(req)
        if err != nil {
            log.Printf("ERROR Processor: Failed to call AI service for alert %s: %v", alert.ID, err)
            return err // Re-queue if AI service is unreachable or responds with error
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {
            bodyBytes, _ := ioutil.ReadAll(resp.Body)
            log.Printf("ERROR Processor: AI service returned non-OK status %d for alert %s: %s", resp.StatusCode, alert.ID, string(bodyBytes))
            return fmt.Errorf("AI service error: status %d", resp.StatusCode) // <--- fmt.Errorf requires fmt import
        }

        var analyzedAlert models.SecurityAlert // Use models.SecurityAlert as it now has AI fields
        err = json.NewDecoder(resp.Body).Decode(&analyzedAlert)
        if err != nil {
            log.Printf("ERROR Processor: Failed to unmarshal AI service response for alert %s: %v", alert.ID, err)
            return err // Re-queue
        }

        log.Printf("Processor: AI analysis complete for alert ID: %s. Predicted Severity: %s, Risk Score: %.2f",
            analyzedAlert.ID, analyzedAlert.PredictedSeverity, analyzedAlert.RiskScore)

        // --- Step 2: Update alert in database with AI results ---
        // The AI service returns the full alert with AI fields populated.
        // We set status to 'analyzed' after AI processing.
        analyzedAlert.Status = "analyzed"
        err = alertRepo.UpdateAlertWithAIResults(context.Background(), &analyzedAlert) // Pass the full analyzedAlert
        if err != nil {
            log.Printf("ERROR Processor: Failed to update alert %s with AI results in DB: %v", alert.ID, err)
            return err // Re-queue if DB update failed
        }
        log.Printf("Processor: Alert ID: %s updated in DB with AI results and status 'analyzed'.", alert.ID)

        return nil // Message processed successfully
    })

    log.Println("GuardianAI Alert Processor stopped.")
}