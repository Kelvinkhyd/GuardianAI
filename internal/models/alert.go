package models

import (
    "time"
)

// SecurityAlert represents a generic security alert structure.
type SecurityAlert struct {
    ID               string    `json:"id"`
    Source           string    `json:"source"`
    Timestamp        time.Time `json:"timestamp"`
    Severity         string    `json:"severity"`
    Category         string    `json:"category"`
    Title            string    `json:"title"`
    Description      string    `json:"description,omitempty"`
    SourceIP         string    `json:"source_ip,omitempty"`
    TargetIP         string    `json:"target_ip,omitempty"`
    Hostname         string    `json:"hostname,omitempty"`
    Username         string    `json:"username,omitempty"`
    FileHash         string    `json:"file_hash,omitempty"`
    Status           string    `json:"status"` // e.g., "new", "processed", "triaged"
    CreatedAt        time.Time `json:"created_at"` // This is usually set by DB, not client

    // New AI/ML related fields
    PredictedSeverity string    `json:"predicted_severity,omitempty"`  // AI's predicted severity
    RiskScore         float64   `json:"risk_score,omitempty"`        // Numerical risk score from AI
    RecommendedAction string    `json:"recommended_action,omitempty"` // AI's recommended action
    AIModelVersion    string    `json:"ai_model_version,omitempty"`   // Version of AI model used
}