package repository

import (
	"context"
	"fmt"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

type TelemetryRepo struct {
	Client influxdb2.Client
	Org    string
	Bucket string
}

func NewTelemetryRepo(client influxdb2.Client, org, bucket string) *TelemetryRepo {
	return &TelemetryRepo{
		Client: client,
		Org:    org,
		Bucket: bucket,
	}
}

func (r *TelemetryRepo) QueryHistory(ctx context.Context, params HistoryParams) ([]TelemetryPoint, error) {
	queryAPI := r.Client.QueryAPI(r.Org)

	fluxQuery := fmt.Sprintf(`
		from(bucket: "%s")
			|> range(start: -%s)
			|> filter(fn: (r) => r["_measurement"] == "sensor_telemetry")
			|> filter(fn: (r) => r["device_id"] == "%s")
			|> filter(fn: (r) => r["_field"] == "%s")
			|> aggregateWindow(every: %s, fn: mean, createEmpty: false)
			|> yield(name: "mean")`, r.Bucket, params.Range, params.DeviceID, params.Metric, params.Window)

	result, err := queryAPI.Query(ctx, fluxQuery)
	if err != nil {
		return nil, err
	}
	defer result.Close()

	var data []TelemetryPoint
	for result.Next() {
		val, ok := result.Record().Value().(float64)
		if !ok {
			continue
		}
		data = append(data, TelemetryPoint{
			Timestamp: result.Record().Time(),
			Value:     val,
		})
	}
	if result.Err() != nil {
		return nil, result.Err()
	}

	return data, nil
}

func (r *TelemetryRepo) WriteTelemetry(ctx context.Context, payload TelemetryPayload) error {
	writeAPI := r.Client.WriteAPIBlocking(r.Org, r.Bucket)

	// Convert float fields to Influx fields map
	fields := make(map[string]interface{})
	for k, v := range payload.Fields {
		fields[k] = v
	}

	p := influxdb2.NewPoint(
		"sensor_telemetry",
		map[string]string{
			"device_id":     payload.DeviceID,
			"hardware_type": payload.HardwareType,
		},
		fields,
		payload.Timestamp,
	)

	return writeAPI.WritePoint(ctx, p)
}
