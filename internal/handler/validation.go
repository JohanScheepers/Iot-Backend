package handler

import (
	"errors"
)

var (
	AllowedMetrics = map[string]bool{
		"temperature":     true,
		"humidity":        true,
		"battery_voltage": true,
		"rssi":            true,
	}
	AllowedRanges = map[string]bool{
		"1h":  true,
		"6h":  true,
		"24h": true,
		"7d":  true,
		"14d": true,
		"30d": true,
	}
	AllowedWindows = map[string]bool{
		"1m":  true,
		"5m":  true,
		"15m": true,
		"1h":  true,
		"6h":  true,
		"12h": true,
	}
)

func ValidateHistoryParams(metric, timeRange, window string) error {
	if !AllowedMetrics[metric] {
		return errors.New("invalid metric parameter")
	}
	if !AllowedRanges[timeRange] {
		return errors.New("invalid range parameter")
	}
	if !AllowedWindows[window] {
		return errors.New("invalid window parameter")
	}
	return nil
}
