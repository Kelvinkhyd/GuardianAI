-- scripts/init_db.sql
CREATE TABLE IF NOT EXISTS alerts (
    id VARCHAR(255) PRIMARY KEY,
    source VARCHAR(255) NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    severity VARCHAR(50) NOT NULL,
    category VARCHAR(255) NOT NULL,
    title VARCHAR(512) NOT NULL,
    description TEXT,
    source_ip VARCHAR(100),
    target_ip VARCHAR(100),
    hostname VARCHAR(255),
    username VARCHAR(255),
    file_hash VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'new',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Add new columns for AI/ML integration (Run these if the table already exists)
ALTER TABLE alerts ADD COLUMN IF NOT EXISTS predicted_severity VARCHAR(50);
ALTER TABLE alerts ADD COLUMN IF NOT EXISTS risk_score NUMERIC(5,3); -- e.g., 0.000 to 1.000
ALTER TABLE alerts ADD COLUMN IF NOT EXISTS recommended_action TEXT;
ALTER TABLE alerts ADD COLUMN IF NOT EXISTS ai_model_version VARCHAR(100);