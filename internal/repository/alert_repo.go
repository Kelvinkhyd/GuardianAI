package repository

import (
    "context"
    "database/sql"
    "fmt"
    "time"

    "github.com/Kelvinkhyd/GuardianAI/internal/models" // Make sure this path is correct
)

// AlertRepository defines the interface for alert data operations.
type AlertRepository interface {
    CreateAlert(ctx context.Context, alert *models.SecurityAlert) error
    GetAlertByID(ctx context.Context, id string) (*models.SecurityAlert, error)
    GetAllAlerts(ctx context.Context, limit, offset int) ([]models.SecurityAlert, error)
    UpdateAlertStatus(ctx context.Context, id, status string) error // Ensure this method exists and matches
    UpdateAlertWithAIResults(ctx context.Context, alert *models.SecurityAlert) error // The new method
}

// pgAlertRepository implements AlertRepository for PostgreSQL.
type pgAlertRepository struct {
    db *sql.DB
}

// NewPgAlertRepository creates a new instance of PgAlertRepository.
func NewPgAlertRepository(db *sql.DB) AlertRepository {
    return &pgAlertRepository{db: db}
}

// CreateAlert inserts a new security alert into the database.
func (r *pgAlertRepository) CreateAlert(ctx context.Context, alert *models.SecurityAlert) error {
    query := `
        INSERT INTO alerts (
            id, source, timestamp, severity, category, title, description,
            source_ip, target_ip, hostname, username, file_hash, status,
            predicted_severity, risk_score, recommended_action, ai_model_version
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
        )`
    _, err := r.db.ExecContext(ctx, query,
        alert.ID, alert.Source, alert.Timestamp, alert.Severity, alert.Category,
        alert.Title, alert.Description, alert.SourceIP, alert.TargetIP,
        alert.Hostname, alert.Username, alert.FileHash, alert.Status,
        alert.PredictedSeverity, alert.RiskScore, alert.RecommendedAction, alert.AIModelVersion)
    if err != nil {
        return fmt.Errorf("failed to create alert: %w", err)
    }
    return nil
}

// GetAlertByID retrieves a single alert by its ID.
func (r *pgAlertRepository) GetAlertByID(ctx context.Context, id string) (*models.SecurityAlert, error) {
    query := `SELECT
        id, source, timestamp, severity, category, title, description,
        source_ip, target_ip, hostname, username, file_hash, status, created_at,
        predicted_severity, risk_score, recommended_action, ai_model_version
        FROM alerts WHERE id = $1`

    var alert models.SecurityAlert
    var createdAt time.Time
    // Use sql.Null* for columns that can be NULL in the database
    var predictedSeverity sql.NullString
    var riskScore sql.NullFloat64
    var recommendedAction sql.NullString
    var aiModelVersion sql.NullString

    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &alert.ID, &alert.Source, &alert.Timestamp, &alert.Severity, &alert.Category,
        &alert.Title, &alert.Description, &alert.SourceIP, &alert.TargetIP,
        &alert.Hostname, &alert.Username, &alert.FileHash, &alert.Status, &createdAt,
        &predictedSeverity, &riskScore, &recommendedAction, &aiModelVersion)

    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil // Alert not found
        }
        return nil, fmt.Errorf("failed to get alert by ID %s: %w", id, err)
    }

    // Assign nullable types to actual struct fields
    if predictedSeverity.Valid { alert.PredictedSeverity = predictedSeverity.String }
    if riskScore.Valid { alert.RiskScore = riskScore.Float64 }
    if recommendedAction.Valid { alert.RecommendedAction = recommendedAction.String }
    if aiModelVersion.Valid { alert.AIModelVersion = aiModelVersion.String }

    alert.CreatedAt = createdAt // Assign created_at to the struct field
    return &alert, nil
}

// GetAllAlerts retrieves a list of alerts with pagination.
func (r *pgAlertRepository) GetAllAlerts(ctx context.Context, limit, offset int) ([]models.SecurityAlert, error) {
    query := `SELECT
        id, source, timestamp, severity, category, title, description,
        source_ip, target_ip, hostname, username, file_hash, status, created_at,
        predicted_severity, risk_score, recommended_action, ai_model_version
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
        // Use sql.Null* for columns that can be NULL in the database
        var predictedSeverity sql.NullString
        var riskScore sql.NullFloat64
        var recommendedAction sql.NullString
        var aiModelVersion sql.NullString

        err := rows.Scan(
            &alert.ID, &alert.Source, &alert.Timestamp, &alert.Severity, &alert.Category,
            &alert.Title, &alert.Description, &alert.SourceIP, &alert.TargetIP,
            &alert.Hostname, &alert.Username, &alert.FileHash, &alert.Status, &createdAt,
            &predictedSeverity, &riskScore, &recommendedAction, &aiModelVersion)
        if err != nil {
            return nil, fmt.Errorf("failed to scan alert row: %w", err)
        }

        if predictedSeverity.Valid { alert.PredictedSeverity = predictedSeverity.String }
        if riskScore.Valid { alert.RiskScore = riskScore.Float64 }
        if recommendedAction.Valid { alert.RecommendedAction = recommendedAction.String }
        if aiModelVersion.Valid { alert.AIModelVersion = aiModelVersion.String }

        alert.CreatedAt = createdAt // Assign created_at to the struct field
        alerts = append(alerts, alert)
    }

    if err = rows.Err(); err != nil {
        return nil, fmt.Errorf("row iteration error: %w", err)
    }

    return alerts, nil
}

// UpdateAlertStatus updates the status of a specific alert by its ID.
func (r *pgAlertRepository) UpdateAlertStatus(ctx context.Context, id, status string) error {
    query := `UPDATE alerts SET status = $1 WHERE id = $2`
    res, err := r.db.ExecContext(ctx, query, status, id)
    if err != nil {
        return fmt.Errorf("failed to update alert status for ID %s: %w", id, err)
    }
    rowsAffected, err := res.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get rows affected after updating status for ID %s: %w", id, err)
    }
    if rowsAffected == 0 {
        return fmt.Errorf("no alert found with ID %s to update status", id)
    }
    return nil
}

// UpdateAlertWithAIResults updates an alert's AI-generated fields and status in the database.
func (r *pgAlertRepository) UpdateAlertWithAIResults(ctx context.Context, alert *models.SecurityAlert) error {
    query := `
        UPDATE alerts SET
            status = $1,
            predicted_severity = $2,
            risk_score = $3,
            recommended_action = $4,
            ai_model_version = $5
        WHERE id = $6`
    res, err := r.db.ExecContext(ctx, query,
        alert.Status, alert.PredictedSeverity, alert.RiskScore,
        alert.RecommendedAction, alert.AIModelVersion, alert.ID)
    if err != nil {
        return fmt.Errorf("failed to update alert %s with AI results: %w", alert.ID, err)
    }
    rowsAffected, err := res.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get rows affected for alert %s update: %w", alert.ID, err)
    }
    if rowsAffected == 0 {
        return fmt.Errorf("no alert found with ID %s to update with AI results", alert.ID)
    }
    return nil
}