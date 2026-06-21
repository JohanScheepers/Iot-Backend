package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                    string
	Env                     string // "development" | "production"
	FirebaseCredentialsPath string
	MySQLDSN                string
	InfluxURL               string
	InfluxToken             string
	InfluxOrg               string
	InfluxBucket            string
	
	// MQTT (Phase 6)
	MQTTBrokerURL string
	MQTTClientID  string
	MQTTUsername  string
	MQTTPassword  string
}

func Load() *Config {
	// Load .env file if it exists (ignore error if not present)
	_ = godotenv.Load()

	cfg := &Config{
		Port:                    getEnv("PORT", "8080"),
		Env:                     getEnv("ENV", "development"),
		FirebaseCredentialsPath: getEnv("FIREBASE_CREDENTIALS_PATH", "./configs/firebase-service-account.json"),
		MySQLDSN:                os.Getenv("MYSQL_DSN"),
		InfluxURL:               os.Getenv("INFLUX_URL"),
		InfluxToken:             os.Getenv("INFLUX_TOKEN"),
		InfluxOrg:               os.Getenv("INFLUX_ORG"),
		InfluxBucket:            os.Getenv("INFLUX_BUCKET"),
		MQTTBrokerURL:           getEnv("MQTT_BROKER_URL", "tcp://localhost:1883"),
		MQTTClientID:            getEnv("MQTT_CLIENT_ID", "iot_backend_api"),
		MQTTUsername:            os.Getenv("MQTT_USERNAME"),
		MQTTPassword:            os.Getenv("MQTT_PASSWORD"),
	}

	// Validate required fields
	validateRequired("MYSQL_DSN", cfg.MySQLDSN)
	validateRequired("INFLUX_URL", cfg.InfluxURL)
	validateRequired("INFLUX_TOKEN", cfg.InfluxToken)
	validateRequired("INFLUX_ORG", cfg.InfluxOrg)
	validateRequired("INFLUX_BUCKET", cfg.InfluxBucket)

	return cfg
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func validateRequired(key, val string) {
	if val == "" {
		panic(fmt.Sprintf("Configuration error: environment variable %s is required but not set", key))
	}
}
