package models

import (
    "time"
)

// SecurityAlert represents a generic security alert received from a SIEM or other source.
type SecurityAlert struct {
    ID          string    `json:"id"`           // Unique ID for the alert
    Source      string    `json:"source"`       // e.g., "Splunk", "Elastic", "CrowdStrike"
    Timestamp   time.Time `json:"timestamp"`    // When the alert was generated
    Severity    string    `json:"severity"`     // e.g., "low", "medium", "high", "critical"
    Category    string    `json:"category"`     // e.g., "Malware", "Suspicious Login", "Data Exfiltration"
    Title       string    `json:"title"`        // Short, descriptive title
    Description string    `json:"description"`  // Detailed description of the alert
    SourceIP    string    `json:"source_ip"`    // IP address related to the alert (e.g., attacker IP)
    TargetIP    string    `json:"target_ip"`    // IP of the affected asset
    Hostname    string    `json:"hostname"`     // Hostname of the affected asset
    Username    string    `json:"username"`     // User account involved (if any)
    FileHash    string    `json:"file_hash"`    // Hash of a malicious file (if any)
    Status      string    `json:"status"`       // e.g., "new", "triaged", "in_progress", "resolved"
    // Add more fields as needed for specific alert types
}