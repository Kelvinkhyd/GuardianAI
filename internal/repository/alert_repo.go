package repository

import (
    "context"
    "database/sql"
    "fmt"
    "time"

    "github.com/Kelvinkhyd/GuardianAI/internal/models"
)

// AlertRepository defines the interface for alert data operations.
type AlertRepository interface {
    CreateAlert(ctx context.Context, alert *models.SecurityAlert) error
    GetAlertByID(ctx context.Context, id string) (*models.SecurityAlert, error)
    GetAllAlerts(ctx context.Context, limit, offset int) ([]models.SecurityAlert, error)
    UpdateAlertStatus(ctx context.Context, id, status string) error
}

// pgAlertRepository implements AlertRepository for PostgreSQL.
type pgAlertRepository struct {
    db *sql.DB
}

// NewPgAlertRepository creates a new PostgreSQL alert repository.
func NewPgAlertRepository(db *sql.DB) AlertRepository {
    return &pgAlertRepository{db: db}
}

// CreateAlert inserts a new security alert into the database.
func (r *pgAlertRepository) CreateAlert(ctx context.Context, alert *models.SecurityAlert) error {
    query := `
        INSERT INTO alerts (
            id, source, timestamp, severity, category, title, description,
            source_ip, target_ip, hostname, username, file_hash, status
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
        )`
    _, err := r.db.ExecContext(ctx, query,
        alert.ID, alert.Source, alert.Timestamp, alert.Severity, alert.Category,
        alert.Title, alert.Description, alert.SourceIP, alert.TargetIP,
        alert.Hostname, alert.Username, alert.FileHash, alert.Status)
    if err != nil {
        return fmt.Errorf("failed to create alert: %w", err)
    }
    return nil
}

// GetAlertByID retrieves a single alert by its ID.
func (r *pgAlertRepository) GetAlertByID(ctx context.Context, id string) (*models.SecurityAlert, error) {
    query := `SELECT
        id, source, timestamp, severity, category, title, description,
        source_ip, target_ip, hostname, username, file_hash, status, created_at
        FROM alerts WHERE id = $1`

    var alert models.SecurityAlert
    var createdAt time.Time // To capture created_at for potential display

    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &alert.ID, &alert.Source, &alert.Timestamp, &alert.Severity, &alert.Category,
        &alert.Title, &alert.Description, &alert.SourceIP, &alert.TargetIP,
        &alert.Hostname, &alert.Username, &alert.FileHash, &alert.Status, &createdAt) // Scan into createdAt

    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil // Alert not found
        }
        return nil, fmt.Errorf("failed to get alert by ID %s: %w", id, err)
    }
    // You could add alert.CreatedAt = createdAt if you extend the model
    return &alert, nil
}

// GetAllAlerts retrieves a list of alerts with pagination.
func (r *pgAlertRepository) GetAllAlerts(ctx context.Context, limit, offset int) ([]models.SecurityAlert, error) {
    query := `SELECT
        id, source, timestamp, severity, category, title, description,
        source_ip, target_ip, hostname, username, file_hash, status, created_at
        FROM alerts ORDER BY created_at DESC LIMIT $1 OFFSET $2`

    rows, err := r.db.QueryContext(ctx, query, limit, offset)
    if err != nil {
        return nil, fmt.Errorf("failed to get all alerts: %w", err)
    }
    defer rows.Close()

    var alerts []models.SecurityAlert
    for rows.Next() {
        var alert models.SecurityAlert
        var createdAt time.Time
        err := rows.Scan(
            &alert.ID, &alert.Source, &alert.Timestamp, &alert.Severity, &alert.Category,
            &alert.Title, &alert.Description, &alert.SourceIP, &alert.TargetIP,
            &alert.Hostname, &alert.Username, &alert.FileHash, &alert.Status, &createdAt)
        if err != nil {
            return nil, fmt.Errorf("failed to scan alert row: %w", err)
        }
        alerts = append(alerts, alert)
    }

    if err = rows.Err(); err != nil {
        return nil, fmt.Errorf("row iteration error: %w", err)
    }

    return alerts, nil
}

// UpdateAlertStatus updates the status of an alert.
func (r *pgAlertRepository) UpdateAlertStatus(ctx context.Context, id, status string) error {
    query := `UPDATE alerts SET status = $1 WHERE id = $2`
    res, err := r.db.ExecContext(ctx, query, status, id)
    if err != nil {
        return fmt.Errorf("failed to update alert status: %w", err)
    }
    rowsAffected, err := res.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get rows affected: %w", err)
    }
    if rowsAffected == 0 {
        return fmt.Errorf("no alert found with ID %s to update", id)
    }
    return nil
}