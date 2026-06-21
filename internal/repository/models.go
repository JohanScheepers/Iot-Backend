package repository

import (
	"time"
)

type User struct {
	ID          int       `json:"id"`
	FirebaseUID string    `json:"firebase_uid"`
	Email       string    `json:"email"`
	CompanyID   *int      `json:"company_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type Device struct {
	ID               string     `json:"id"`
	MACAddress       string     `json:"mac_address"`
	HardwareType     string     `json:"hardware_type"`
	FirmwareVersion  *string    `json:"firmware_version,omitempty"`
	InstallationDate *time.Time `json:"installation_date,omitempty"`
	Status           string     `json:"status"` // "active" etc.
}

type TelemetryPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

type TelemetryPayload struct {
	DeviceID     string             `json:"device_id"`
	HardwareType string             `json:"hardware_type"`
	Fields       map[string]float64 `json:"fields"`
	Timestamp    time.Time          `json:"timestamp"`
}

type HistoryParams struct {
	DeviceID string
	Metric   string // "temperature", "humidity", etc.
	Range    string // "24h", "7d", "30d"
	Window   string // "15m", "1h", "12h"
}
