package main

import (
	"log"
	"net/http"
	"os"

	"iot_backend/internal/auth"
	"iot_backend/internal/database"
	"iot_backend/internal/handler"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	credPath := os.Getenv("FIREBASE_CREDENTIALS_PATH")
	if credPath == "" {
		credPath = "./configs/firebase-service-account.json"
	}

	// 1. Initialize DB pools
	dbHub := database.InitDBs()
	defer dbHub.MySQL.Close()
	defer dbHub.Influx.Close()

	// 2. Initialize Middlewares and Handlers
	authMiddleware := auth.NewAuthMiddleware(credPath)
	telemetryHandler := &handler.TelemetryHandler{Influx: dbHub.Influx}

	// 3. Register HTTP routes
	mux := http.NewServeMux()
	
	// Secure individual endpoints by wrapping the handler with the auth middleware function
	mux.HandleFunc("/api/v1/devices/", authMiddleware.Secure(telemetryHandler.GetDeviceHistory))

	log.Printf("🚀 Server running smoothly on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server crash: %v", err)
	}
}