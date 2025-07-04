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

    "github.com/Kelvinkhyd/GuardianAI/internal/models"
    "github.com/Kelvinkhyd/GuardianAI/internal/repository" // Import repository
)

// Handler holds dependencies for our API handlers.
type Handler struct {
    AlertRepo repository.AlertRepository
}

// NewHandler creates a new Handler instance.
func NewHandler(ar repository.AlertRepository) *Handler {
    return &Handler{AlertRepo: ar}
}

// HandleAlerts receives incoming security alerts via HTTP POST and stores them.
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
        alert.ID = fmt.Sprintf("alert-%d", time.Now().UnixNano()) // Simple ID generation
    }
    alert.Status = "new" // Initial status, will be processed by orchestration layer later
    alert.Timestamp = time.Now() // Ensure timestamp is set if not provided or to current time

    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second) // Set a timeout for DB operations
    defer cancel()

    err = h.AlertRepo.CreateAlert(ctx, &alert)
    if err != nil {
        log.Printf("ERROR: Failed to save alert to DB: %v", err)
        http.Error(w, "Failed to process alert: "+err.Error(), http.StatusInternalServerError)
        return
    }

    log.Printf("Received and saved new alert: ID=%s, Source=%s, Category=%s, Severity=%s, Hostname=%s",
        alert.ID, alert.Source, alert.Category, alert.Severity, alert.Hostname)

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusAccepted)
    json.NewEncoder(w).Encode(map[string]string{"message": "Alert received and saved", "alert_id": alert.ID})
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