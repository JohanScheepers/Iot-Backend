package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"iot_backend/internal/auth"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

type TelemetryPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

type TelemetryHandler struct {
	Influx influxdb2.Client
}

func (h *TelemetryHandler) GetDeviceHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Extract verified Firebase UID from context
	uid, ok := ctx.Value(auth.FirebaseUIDKey).(string)
	if !ok {
		http.Error(w, "Unauthorized Context", http.StatusUnauthorized)
		return
	}
	_ = uid // Use this to check device ownership in your MySQL DB later

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

	bucket := os.Getenv("INFLUX_BUCKET")
	org := os.Getenv("INFLUX_ORG")

	queryAPI := h.Influx.QueryAPI(org)
	fluxQuery := fmt.Sprintf(`
		from(bucket: "%s")
			|> range(start: -%s)
			|> filter(fn: (r) => r["_measurement"] == "sensor_telemetry")
			|> filter(fn: (r) => r["device_id"] == "%s")
			|> filter(fn: (r) => r["_field"] == "%s")
			|> aggregateWindow(every: %s, fn: mean, createEmpty: false)
			|> yield(name: "mean")`, bucket, timeRange, deviceID, metric, windowPeriod)

	result, err := queryAPI.Query(ctx, fluxQuery)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer result.Close()

	var data []TelemetryPoint
	for result.Next() {
		val, ok := result.Record().Value().(float64)
		if !ok { continue }
		data = append(data, TelemetryPoint{
			Timestamp: result.Record().Time(),
			Value:     val,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}