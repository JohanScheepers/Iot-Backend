package database

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

type DBHub struct {
	MySQL  *sql.DB
	Influx influxdb2.Client
}

func InitDBs() *DBHub {
	// 1. Init MySQL
	mysqlDSN := os.Getenv("MYSQL_DSN")
	sqlDB, err := sql.Open("mysql", mysqlDSN)
	if err != nil {
		log.Fatalf("Error opening MySQL connection: %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("MySQL unreachable: %v", err)
	}

	// 2. Init InfluxDB
	influxURL := os.Getenv("INFLUX_URL")
	influxToken := os.Getenv("INFLUX_TOKEN")
	influxClient := influxdb2.NewClient(influxURL, influxToken)

	log.Println("⚡ Successfully connected to MySQL and InfluxDB")

	return &DBHub{
		MySQL:  sqlDB,
		Influx: influxClient,
	}
}