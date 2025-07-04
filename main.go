package main

import (
    "log"
    "net/http"

    "github.com/gorilla/mux"

    "github.com/Kelvinkhyd/GuardianAI/internal/api"
    "github.com/Kelvinkhyd/GuardianAI/internal/config"
    "github.com/Kelvinkhyd/GuardianAI/internal/database"
    "github.com/Kelvinkhyd/GuardianAI/internal/repository"
)

func main() {
    cfg := config.LoadConfig() // Load configuration

    // Establish database connection
    dbConn, err := database.NewDBConnection(cfg.DatabaseURL)
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    defer dbConn.Close() // Ensure database connection is closed when main exits

    // Initialize repository
    alertRepo := repository.NewPgAlertRepository(dbConn.DB)

    // Initialize API handlers with the repository
    apiHandler := api.NewHandler(alertRepo)

    // Create a new Gorilla Mux router
    router := mux.NewRouter()

    // Define routes using the Mux router
    router.HandleFunc("/alerts", apiHandler.HandleAlerts).Methods("POST")
    router.HandleFunc("/alerts", apiHandler.GetAlerts).Methods("GET")
    router.HandleFunc("/alerts/{id}", apiHandler.GetAlertByID).Methods("GET")

    // Attach the Mux router to the HTTP server
    log.Printf("GuardianAI API server starting on port %s", cfg.ServerPort)
    err = http.ListenAndServe(cfg.ServerPort, router) // Pass the router here
    if err != nil {
        log.Fatalf("Server failed to start: %v", err)
    }
}