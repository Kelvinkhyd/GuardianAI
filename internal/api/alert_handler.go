package api

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "strconv" // For parsing limit/offset
    "time"

    "github.com/gorilla/mux"

    "github.com/Kelvinkhyd/GuardianAI/internal/kafka" // Import kafka package
    "github.com/Kelvinkhyd/GuardianAI/internal/models"
    "github.com/Kelvinkhyd/GuardianAI/internal/repository" // Import repository
)

// Handler holds dependencies for our API handlers.
type Handler struct {
    AlertRepo repository.AlertRepository
    KafkaProducer *kafka.Producer // Add Kafka producer
}

// NewHandler creates a new Handler instance.
func NewHandler(ar repository.AlertRepository, kp *kafka.Producer) *Handler {
    return &Handler{AlertRepo: ar, KafkaProducer: kp}
}

// HandleAlerts receives incoming security alerts via HTTP POST and stores them, then publishes to Kafka.
func (h *Handler) HandleAlerts(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Only POST requests are accepted", http.StatusMethodNotAllowed)
        return
    }

    var alert models.SecurityAlert
    err := json.NewDecoder(r.Body).Decode(&alert)
    if err != nil {
        http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
        return
    }

    if alert.ID == "" {
        alert.ID = fmt.Sprintf("alert-%d", time.Now().UnixNano())
    }
    alert.Status = "new"
    // It's good practice to ensure the timestamp is set if not provided or to current time.
    // If the client provides a timestamp, use it. Otherwise, set it to now.
    if alert.Timestamp.IsZero() {
        alert.Timestamp = time.Now()
    }

    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()

    // 1. Save to Database
    err = h.AlertRepo.CreateAlert(ctx, &alert)
    if err != nil {
        log.Printf("ERROR: Failed to save alert to DB: %v", err)
        http.Error(w, "Failed to process alert: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // 2. Publish to Kafka
    alertJSON, err := json.Marshal(alert)
    if err != nil {
        log.Printf("ERROR: Failed to marshal alert for Kafka: %v", err)
        // Still return 202 because DB save succeeded, but log the Kafka issue.
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusAccepted)
        json.NewEncoder(w).Encode(map[string]string{"message": "Alert received and saved to DB, but failed to publish to Kafka", "alert_id": alert.ID})
        return
    }

    // Use alert ID as key for Kafka message to ensure order for a specific alert (if partitions are by key)
    err = h.KafkaProducer.PublishMessage(ctx, []byte(alert.ID), alertJSON)
    if err != nil {
        log.Printf("ERROR: Failed to publish alert %s to Kafka: %v", alert.ID, err)
        // Similar to above, DB save succeeded, so return Accepted.
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusAccepted)
        json.NewEncoder(w).Encode(map[string]string{"message": "Alert received and saved to DB, but failed to publish to Kafka", "alert_id": alert.ID})
        return
    }

    log.Printf("Received, saved, and published alert to Kafka: ID=%s, Source=%s", alert.ID, alert.Source)

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusAccepted)
    json.NewEncoder(w).Encode(map[string]string{"message": "Alert received, saved, and published for processing", "alert_id": alert.ID})
}


// GetAlerts retrieves a list of security alerts from the database.
// Supports pagination via query parameters: /alerts?limit=10&offset=0
func (h *Handler) GetAlerts(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Only GET requests are accepted", http.StatusMethodNotAllowed)
        return
    }

    limitStr := r.URL.Query().Get("limit")
    offsetStr := r.URL.Query().Get("offset")

    limit, err := strconv.Atoi(limitStr)
    if err != nil || limit <= 0 {
        limit = 10 // Default limit
    }
    offset, err := strconv.Atoi(offsetStr)
    if err != nil || offset < 0 {
        offset = 0 // Default offset
    }

    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()

    alerts, err := h.AlertRepo.GetAllAlerts(ctx, limit, offset)
    if err != nil {
        log.Printf("ERROR: Failed to retrieve alerts from DB: %v", err)
        http.Error(w, "Failed to retrieve alerts: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(alerts)
}

// GetAlertByID retrieves a single security alert by its ID.
// Usage: /alerts/{id}
func (h *Handler) GetAlertByID(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        // Mux already handles this with .Methods("GET"), but it's good practice.
        http.Error(w, "Only GET requests are accepted", http.StatusMethodNotAllowed)
        return
    }

    // Use mux.Vars to extract the 'id' variable from the path
    vars := mux.Vars(r)
    alertID := vars["id"] // "id" matches the {id} in the route definition

    if alertID == "" {
        http.Error(w, "Alert ID is missing from the URL path", http.StatusBadRequest)
        return
    }

    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()

    alert, err := h.AlertRepo.GetAlertByID(ctx, alertID)
    if err != nil {
        log.Printf("ERROR: Failed to retrieve alert by ID %s: %v", alertID, err)
        http.Error(w, "Failed to retrieve alert: "+err.Error(), http.StatusInternalServerError)
        return
    }
    if alert == nil {
        http.Error(w, "Alert not found", http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(alert)
}

// You can now remove the splitPath helper function, as it's no longer needed.
// func splitPath(path string) []string { ... }
// Helper to split URL path for basic routing (to be replaced by Gorilla Mux)
func splitPath(path string) []string {
    var parts []string
    for _, p := range ([]byte)(path) {
        if p == '/' {
            if len(parts) > 0 && len(parts[len(parts)-1]) > 0 {
                parts = append(parts, "")
            }
        } else {
            if len(parts) == 0 {
                parts = append(parts, string(p))
            } else {
                parts[len(parts)-1] += string(p)
            }
        }
    }
    if len(parts) > 0 && len(parts[len(parts)-1]) == 0 {
        parts = parts[:len(parts)-1]
    }
    return parts
}