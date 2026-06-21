package repository

import (
	"context"
	"database/sql"
)

type DeviceRepo struct {
	DB *sql.DB
}

func NewDeviceRepo(db *sql.DB) *DeviceRepo {
	return &DeviceRepo{DB: db}
}

func (r *DeviceRepo) GetByID(ctx context.Context, deviceID string) (*Device, error) {
	var d Device
	query := "SELECT id, mac_address, hardware_type, firmware_version, installation_date, status FROM devices WHERE id = ?"
	err := r.DB.QueryRowContext(ctx, query, deviceID).Scan(
		&d.ID,
		&d.MACAddress,
		&d.HardwareType,
		&d.FirmwareVersion,
		&d.InstallationDate,
		&d.Status,
	)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *DeviceRepo) IsOwnedBy(ctx context.Context, userID int, deviceID string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM user_devices WHERE user_id = ? AND device_id = ?)"
	err := r.DB.QueryRowContext(ctx, query, userID, deviceID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *DeviceRepo) IsActive(ctx context.Context, deviceID string) (bool, error) {
	var status string
	query := "SELECT status FROM devices WHERE id = ?"
	err := r.DB.QueryRowContext(ctx, query, deviceID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return status == "active", nil
}
