# ai_service/main.py
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from typing import Optional
import datetime
import logging

logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

# Define the input model for an alert (matches our Go SecurityAlert for relevant fields)
class SecurityAlertInput(BaseModel):
    id: str
    source: str
    timestamp: datetime.datetime
    severity: str
    category: str
    title: str
    description: Optional[str] = None
    source_ip: Optional[str] = None
    target_ip: Optional[str] = None
    hostname: Optional[str] = None
    username: Optional[str] = None
    file_hash: Optional[str] = None
    status: Optional[str] = None # Current status from Go, likely 'new'

# Define the output model for the analyzed alert
# This includes the original fields plus new AI-generated fields
class AnalyzedAlertOutput(SecurityAlertInput):
    predicted_severity: str
    risk_score: float # e.g., 0.0 to 1.0
    recommended_action: str
    ai_model_version: str = "v1.0.0_dummy_model"

app = FastAPI(
    title="GuardianAI Alert Analysis Service",
    description="Analyzes security alerts and provides AI-driven insights.",
    version="1.0.0"
)

@app.get("/")
async def read_root():
    return {"message": "GuardianAI Alert Analysis Service is running!"}

@app.post("/analyze-alert", response_model=AnalyzedAlertOutput)
async def analyze_alert(alert_input: SecurityAlertInput):
    logger.info(f"Received alert for analysis: ID={alert_input.id}, Category={alert_input.category}, Severity={alert_input.severity}")

    # --- Placeholder AI Logic ---
    # In a real scenario, you'd load and run a trained ML model here.
    # For demonstration, we'll use simple rules to mock predictions.

    predicted_severity = alert_input.severity # Default to original severity
    risk_score = 0.5 # Default risk score
    recommended_action = "Investigate immediately."

    if "login" in alert_input.category.lower() and alert_input.severity.lower() == "high":
        predicted_severity = "critical"
        risk_score = 0.9
        recommended_action = "Isolate user account and review audit logs."
    elif "malware" in alert_input.category.lower() or "virus" in alert_input.title.lower():
        predicted_severity = "critical"
        risk_score = 0.95
        recommended_action = "Quarantine host, analyze malware signature."
    elif "network anomaly" in alert_input.category.lower():
        predicted_severity = "high"
        risk_score = 0.75
        recommended_action = "Block source IP, review firewall logs."

    # Add more sophisticated rules or load a model here later
    # For example, using scikit-learn:
    # from sklearn.ensemble import RandomForestClassifier
    # model = load_model("path/to/your/model.pkl")
    # features = preprocess_alert(alert_input)
    # prediction = model.predict(features)
    # predicted_severity = map_prediction_to_severity(prediction)

    analyzed_alert = AnalyzedAlertOutput(
        id=alert_input.id,
        source=alert_input.source,
        timestamp=alert_input.timestamp,
        severity=alert_input.severity, # Keep original, but add predicted
        category=alert_input.category,
        title=alert_input.title,
        description=alert_input.description,
        source_ip=alert_input.source_ip,
        target_ip=alert_input.target_ip,
        hostname=alert_input.hostname,
        username=alert_input.username,
        file_hash=alert_input.file_hash,
        status=alert_input.status,
        predicted_severity=predicted_severity,
        risk_score=risk_score,
        recommended_action=recommended_action
    )
    logger.info(f"Analyzed alert ID={alert_input.id}. Predicted Severity: {predicted_severity}, Risk Score: {risk_score}")
    return analyzed_alert

if __name__ == "__main__":
    import uvicorn
    # Running via uvicorn directly for local testing
    uvicorn.run(app, host="0.0.0.0", port=8000)