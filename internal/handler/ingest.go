package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"iot_backend/internal/repository"
)

type IngestHandler struct {
	TelemetryRepo *repository.TelemetryRepo
	DeviceRepo    *repository.DeviceRepo
}

func NewIngestHandler(tRepo *repository.TelemetryRepo, dRepo *repository.DeviceRepo) *IngestHandler {
	return &IngestHandler{
		TelemetryRepo: tRepo,
		DeviceRepo:    dRepo,
	}
}

type IngestRequest struct {
	DeviceID     string             `json:"device_id"`
	HardwareType string             `json:"hardware_type"`
	Fields       map[string]float64 `json:"fields"`
	Timestamp    string             `json:"timestamp,omitempty"`
}

func (h *IngestHandler) IngestTelemetry(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req IngestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if req.DeviceID == "" {
		http.Error(w, "Missing device_id", http.StatusBadRequest)
		return
	}
	if req.HardwareType == "" {
		http.Error(w, "Missing hardware_type", http.StatusBadRequest)
		return
	}
	if len(req.Fields) == 0 {
		http.Error(w, "Missing fields", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// 1. Verify device is active
	active, err := h.DeviceRepo.IsActive(ctx, req.DeviceID)
	if err != nil {
		http.Error(w, "Database error checking device status", http.StatusInternalServerError)
		return
	}
	if !active {
		http.Error(w, "Forbidden: Device is inactive or not provisioned", http.StatusForbidden)
		return
	}

	// 2. Parse timestamp or default to now
	ts := time.Now()
	if req.Timestamp != "" {
		if parsedTs, err := time.Parse(time.RFC3339, req.Timestamp); err == nil {
			ts = parsedTs
		}
	}

	payload := repository.TelemetryPayload{
		DeviceID:     req.DeviceID,
		HardwareType: req.HardwareType,
		Fields:       req.Fields,
		Timestamp:    ts,
	}

	// 3. Write point to InfluxDB
	if err := h.TelemetryRepo.WriteTelemetry(ctx, payload); err != nil {
		http.Error(w, "Error writing telemetry data", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"status":"success","message":"Telemetry ingested"}`))
}
