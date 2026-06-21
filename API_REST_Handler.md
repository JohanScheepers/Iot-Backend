# API Rest Handler

## 🛠️ Go Implementation: InfluxDB REST Handler

This implementation uses the InfluxDB v2 Go Client to query historical telemetry, downsample it server-side using Flux, and return a clean array of structured data points.

```
Go
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

// TelemetryPoint represents a single data point formatted cleanly for Flutter charts
type TelemetryPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// InfluxHandler handles time-series HTTP traffic
type InfluxHandler struct {
	Client influxdb2.Client
	Org    string
	Bucket string
}

// NewInfluxHandler creates a new handler instance
func NewInfluxHandler(client influxdb2.Client, org, bucket string) *InfluxHandler {
	return &InfluxHandler{Client: client, Org: org, Bucket: bucket}
}

// GetDeviceHistory handles requests for historical device telemetry
// GET /api/v1/devices/{id}/history?metric=temperature&range=30d&window=1h
func (h *InfluxHandler) GetDeviceHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// 1. Extract context variables passed down by your Firebase Auth Middleware
	// firebaseUID := ctx.Value("firebase_uid").(string) 
	// TODO: Run a quick MySQL validation here to verify if firebaseUID owns deviceID

	// 2. Parse Query Parameters
	deviceID := r.URL.Path[len("/api/v1/devices/") : len(r.URL.Path)-len("/history")] // Replace with your router path param logic
	metric := r.URL.Query().Get("metric")
	timeRange := r.URL.Query().Get("range")    // e.g., "24h", "7d", "30d"
	windowPeriod := r.URL.Query().Get("window") // e.g., "15m", "1h", "12h"

	// Fallback defaults to prevent malicious empty queries
	if metric == "" { metric = "temperature" }
	if timeRange == "" { timeRange = "7d" }
	if windowPeriod == "" { windowPeriod = "1h" }

	// 3. Build the Flux Query (Inject parameters safely)
	queryAPI := h.Client.QueryAPI(h.Org)
	fluxQuery := fmt.Sprintf(`
		from(bucket: "%s")
			|> range(start: -%s)
			|> filter(fn: (r) => r["_measurement"] == "sensor_telemetry")
			|> filter(fn: (r) => r["device_id"] == "%s")
			|> filter(fn: (r) => r["_field"] == "%s")
			|> aggregateWindow(every: %s, fn: mean, createEmpty: false)
			|> yield(name: "mean")`, 
		h.Bucket, timeRange, deviceID, metric, windowPeriod,
	)

	// 4. Execute the Query
	result, err := queryAPI.Query(ctx, fluxQuery)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to query InfluxDB: %v", err), http.StatusInternalServerError)
		return
	}
	defer result.Close()

	// 5. Parse Results into Data Structure
	var telemetryData []TelemetryPoint
	for result.Next() {
		val, ok := result.Record().Value().(float64)
		if !ok {
			// Handle cases where integer cast might fail if data type differs
			if intVal, ok := result.Record().Value().(int64); ok {
				val = float64(intVal)
			} else {
				continue
			}
		}
		
		telemetryData = append(telemetryData, TelemetryPoint{
			Timestamp: result.Record().Time(),
			Value:     val,
		})
	}

	if result.Err() != nil {
		http.Error(w, fmt.Sprintf("Error iterating query results: %v", result.Err()), http.StatusInternalServerError)
		return
	}

	// 6. Return Clean JSON Response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(telemetryData)
}
```

## 📱 The REST Response Structure

When your Flutter app makes a GET request to:
https://api.yourdomain.com/api/v1/devices/dev_48291/history?metric=temperature&range=24h&window=15m

The Go backend handles all the complex data windowing server-side and responds with an optimized, lightweight JSON payload:

```
JSON
[
  { "timestamp": "2026-06-21T10:00:00Z", "value": 22.4 },
  { "timestamp": "2026-06-21T10:15:00Z", "value": 22.6 },
  { "timestamp": "2026-06-21T10:30:00Z", "value": 23.1 },
  { "timestamp": "2026-06-21T10:45:00Z", "value": 22.9 }
]
```

## 🎨 Consuming This Data in Flutter

Because the JSON payload is structurally flat, parsing it into a Flutter reactive state model is fast and highly efficient.

1. Data Model

```
Dart
class TelemetryPoint {
  final DateTime timestamp;
  final double value;

  TelemetryPoint({required this.timestamp, required this.value});

  factory TelemetryPoint.fromJson(Map<String, dynamic> json) {
    return TelemetryPoint(
      timestamp: DateTime.parse(json['timestamp']),
      value: (json['value'] as num).toDouble(),
    );
  }
}
```

2. Integration with fl_chart
When passing this parsed list down to your UI layer, you can easily map the raw data points into continuous chart coordinates using line chart spot builde.

```
Dart
List<FlSpot> getChartSpots(List<TelemetryPoint> points) {
  return points.map((p) {
    // Map time to milliseconds since epoch for the X-axis mapping
    final x = p.timestamp.millisecondsSinceEpoch.toDouble();
    return FlSpot(x, p.value);
  }).toList();
}
```

## 💡 Engineering Best Practices for this API Layer

* **Enforce Strict Windowing Controls:** Never let the client pass an arbitrary window or range parameter unchecked. Maliciously long ranges (e.g., range=5y with window=1s) will cause InfluxDB out-of-memory errors. Always sanitize inputs inside your Go handler.

* **Streamline Date Formatting:** Go uses RFC3339 string parameters for time output by default. Flutter's native DateTime.parse() reads this exact format without requiring additional dependency overhead or complex parsing configurations.