package main

import (
	"log"
	"net/http"

	"iot_backend/internal/auth"
	"iot_backend/internal/config"
	"iot_backend/internal/database"
	"iot_backend/internal/handler"
	"iot_backend/internal/repository"
)

func main() {
	// 0. Load configuration
	cfg := config.Load()

	// 1. Initialize DB pools
	dbHub := database.InitDBs(cfg)
	defer dbHub.MySQL.Close()
	defer dbHub.Influx.Close()

	// 2. Initialize Repositories
	userRepo := repository.NewUserRepo(dbHub.MySQL)
	deviceRepo := repository.NewDeviceRepo(dbHub.MySQL)
	telemetryRepo := repository.NewTelemetryRepo(dbHub.Influx, cfg.InfluxOrg, cfg.InfluxBucket)

	// 3. Initialize Middlewares and Handlers
	authMiddleware := auth.NewAuthMiddleware(cfg.FirebaseCredentialsPath)
	telemetryHandler := handler.NewTelemetryHandler(telemetryRepo, deviceRepo, userRepo)
	ingestHandler := handler.NewIngestHandler(telemetryRepo, deviceRepo)

	// 3. Register HTTP routes
	mux := http.NewServeMux()
	
	// Secure individual endpoints by wrapping the handler with the auth middleware function
	mux.HandleFunc("/api/v1/devices/", authMiddleware.Secure(telemetryHandler.GetDeviceHistory))
	mux.HandleFunc("/api/v1/telemetry", ingestHandler.IngestTelemetry)

	log.Printf("🚀 Server running smoothly on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatalf("Server crash: %v", err)
	}
}