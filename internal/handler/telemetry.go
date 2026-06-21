package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"iot_backend/internal/auth"
	"iot_backend/internal/repository"
)

type TelemetryHandler struct {
	TelemetryRepo *repository.TelemetryRepo
	DeviceRepo    *repository.DeviceRepo
	UserRepo      *repository.UserRepo
}

func NewTelemetryHandler(tRepo *repository.TelemetryRepo, dRepo *repository.DeviceRepo, uRepo *repository.UserRepo) *TelemetryHandler {
	return &TelemetryHandler{
		TelemetryRepo: tRepo,
		DeviceRepo:    dRepo,
		UserRepo:      uRepo,
	}
}

func (h *TelemetryHandler) GetDeviceHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Extract verified Firebase UID from context
	uid, ok := ctx.Value(auth.FirebaseUIDKey).(string)
	if !ok {
		http.Error(w, "Unauthorized Context", http.StatusUnauthorized)
		return
	}

	// Simple URL parsing /api/v1/devices/{id}/history
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid device route parsing", http.StatusBadRequest)
		return
	}
	deviceID := pathParts[4]

	// Parse parameters
	metric := r.URL.Query().Get("metric")
	timeRange := r.URL.Query().Get("range")
	windowPeriod := r.URL.Query().Get("window")

	if metric == "" { metric = "temperature" }
	if timeRange == "" { timeRange = "7d" }
	if windowPeriod == "" { windowPeriod = "1h" }

	// Validate query parameters to prevent injection or invalid values
	if err := ValidateHistoryParams(metric, timeRange, windowPeriod); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 1. Get internal user details from Firebase UID
	user, err := h.UserRepo.GetByFirebaseUID(ctx, uid)
	if err != nil {
		http.Error(w, "Forbidden: User record not found", http.StatusForbidden)
		return
	}

	// 2. Verify device ownership
	owned, err := h.DeviceRepo.IsOwnedBy(ctx, user.ID, deviceID)
	if err != nil {
		http.Error(w, "Database error checking device ownership", http.StatusInternalServerError)
		return
	}
	if !owned {
		http.Error(w, "Forbidden: User does not own this device", http.StatusForbidden)
		return
	}

	params := repository.HistoryParams{
		DeviceID: deviceID,
		Metric:   metric,
		Range:    timeRange,
		Window:   windowPeriod,
	}

	data, err := h.TelemetryRepo.QueryHistory(ctx, params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}