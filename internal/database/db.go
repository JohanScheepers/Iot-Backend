package database

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"iot_backend/internal/config"
)

type DBHub struct {
	MySQL  *sql.DB
	Influx influxdb2.Client
}

func InitDBs(cfg *config.Config) *DBHub {
	// 1. Init MySQL
	sqlDB, err := sql.Open("mysql", cfg.MySQLDSN)
	if err != nil {
		log.Fatalf("Error opening MySQL connection: %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("MySQL unreachable: %v", err)
	}

	// 2. Init InfluxDB
	influxClient := influxdb2.NewClient(cfg.InfluxURL, cfg.InfluxToken)

	log.Println("⚡ Successfully connected to MySQL and InfluxDB")

	return &DBHub{
		MySQL:  sqlDB,
		Influx: influxClient,
	}
}