package main

import (
    "log"
    "net/http"

    "github.com/Kelvinkhyd/GuardianAI/internal/api" // Adjust this path!
)

func main() {
    // Define our HTTP route. When a POST request comes to /alerts, call api.HandleAlerts
    http.HandleFunc("/alerts", api.HandleAlerts)

    port := ":8080"
    log.Printf("GuardianAI API server starting on port %s", port)
    // Start the HTTP server
    err := http.ListenAndServe(port, nil)
    if err != nil {
        log.Fatalf("Server failed to start: %v", err)
    }
}