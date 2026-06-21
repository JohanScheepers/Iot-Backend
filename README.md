# IoT Central Gateway Backend

This is the core, high-performance intermediate API layer for the IoT ecosystem. Built in **Go**, this backend bridges mobile client interactions (secured via Firebase) with an industrial-grade operational and time-series data layer.

It is designed to be highly concurrent, stateless, and optimized for low-latency I/O operations.

---

## 🛠️ Tech Stack & Role Architecture

* **Language:** Go (Golang) 1.21+
* **Authentication:** Firebase Admin SDK (JWT Validation & Session Security)
* **Operational Database:** MySQL (Relational business logic, metadata, and hardware profiles)
* **Time-Series Engine:** InfluxDB v2 (High-throughput historical telemetry streams)
* **Ingestion Protocol:** MQTT (Device message broker handling persistent connections)

---

## 📐 System Architecture & Data Flow

To ensure security and transactional integrity, mobile clients and hardware devices do not talk directly to the underlying databases. The Go backend acts as the unified gatekeeper.

### 1. Client Authentication & Access Control
1. The mobile client authenticates with **Firebase Auth** and receives a JWT.
2. The client passes this token via the `Authorization: Bearer <JWT>` header on every request.
3. The Go backend intercepts the request, decodes and verifies the token using the **Firebase Admin SDK**, and extracts the `firebase_uid`.
4. The backend references the `firebase_uid` against **MySQL** to perform multi-tenant permission validation (e.g., verifying if the requesting user owns or has access to the target `device_id`).

### 2. Isolated Storage Engine Model

#### 🧠 MySQL (The Operational Brain)
Manages relational business logic, configurations, and network parameters.

```sql
-- Core Schema Blueprints
CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    firebase_uid VARCHAR(128) UNIQUE NOT NULL,
    email VARCHAR(255) NOT NULL,
    company_id INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE devices (
    id VARCHAR(64) PRIMARY KEY,
    mac_address VARCHAR(17) UNIQUE NOT NULL,
    hardware_type VARCHAR(50) NOT NULL,
    firmware_version VARCHAR(20),
    installation_date TIMESTAMP,
    status VARCHAR(20) DEFAULT 'inactive'
);

CREATE TABLE gateways_or_networks (
    id INT AUTO_INCREMENT PRIMARY KEY,
    provision_key VARCHAR(100),
    routing_key VARCHAR(100),
    regional_profile VARCHAR(20) -- e.g., AS923, AU915, EU868
);
```

## 📊 InfluxDB (The Historical Stream)
An append-only time-series engine optimized for rapid ingestion of raw sensor data.

* **Measurement:** sensor_telemetry

* **Tags (Indexed):** device_id, hardware_type

* **Fields (Unindexed values):** temperature, humidity, battery_voltage, rssi

## 🌐 Application REST API (For Flutter Client)

The backend exposes an authorized REST interface specifically optimized to feed data directly into reactive state components and charting packages (like fl_chart) inside the Flutter application.

**Get Device Historical Telemetry**
Returns a flat, lightweight JSON array containing server-side downsampled telemetry data.

* **Endpoint:** GET /api/v1/devices/{id}/history

* **Headers:** Authorization: Bearer <Firebase_JWT>

* **Query Parameters:**

    * **metric (string, optional):** Target field name (e.g., temperature, humidity). Defaults to temperature.

    * **range (string, optional):** Time lookup window (e.g., 24h, 7d, 30d). Defaults to 7d.

    * **window (string, optional):** Server-side downsampling bucket intervals (e.g., 15m, 1h, 12h). Defaults to 1h.

**Sample Response** (application/json)
```
JSON
[
  { "timestamp": "2026-06-21T10:00:00Z", "value": 22.4 },
  { "timestamp": "2026-06-21T10:15:00Z", "value": 22.6 },
  { "timestamp": "2026-06-21T10:30:00Z", "value": 23.1 }
]
```


### Critical Implementation Protocols
1. Centralized InfluxDB Ingestion
Devices submit packets directly to the Go API or via the MQTT Broker hook. The backend executes a fast-path MySQL check to ensure the device_id status is active before writing the multi-field metric payload to InfluxDB. This prevents rogue or unprovisioned hardware from polluting the time-series buckets.

2. Server-Side Downsampling
To reduce mobile client network overhead and optimize chart rendering performance, the backend uses downsampling. It windows raw telemetry server-side before returning data to the client.

```
SQL
-- InfluxDB InfluxQL/SQL Downsampling Example
SELECT MEAN("temperature") 
FROM "sensor_telemetry" 
WHERE "device_id" = 'dev_123' AND time > now() - 30d 
GROUP BY time(1h)
```

3. Lightweight Client State
Firebase is strictly utilized for Authentication. To retrieve live dashboard statuses, the client executes lightweight API requests to the Go backend, which queries the latest records from InfluxDB/MySQL. This keeps the client ecosystem simple and avoids dual-write sync lags

## Getting Started

**Prerequisites**
* Go 1.21 or higher

* Access to a MySQL Instance

* Access to an InfluxDB v2 Instance

* A Firebase Service Account JSON file (firebase-service-account.json)

```
PORT=8080
ENV=development

# Firebase
FIREBASE_CREDENTIALS_PATH=./configs/firebase-service-account.json

# MySQL
MYSQL_DSN="user:password@tcp(127.0.0.1:3306)/iot_db?parseTime=true"

# InfluxDB
INFLUX_URL=http://localhost:8086
INFLUX_TOKEN=your-super-secret-auth-token
INFLUX_ORG=your-org
INFLUX_BUCKET=telemetry_bucket
```

**Installation & Execution**
1. Clone the repository:
```
Bash
git clone [https://github.com/your-repo/iot-backend.git](https://github.com/your-repo/iot-backend.git)
cd iot-backend
```
2. Download Go dependencies:
```
Bash
go mod download
```
3. Run the service locally:
```
go run cmd/api/main.go
```

## Project Structure
```
├── cmd/
│   └── api/                  # Application entry point
├── internal/
│   ├── auth/                 # Firebase JWT verification middleware
│   ├── config/               # Environment & configuration parsing
│   ├── database/             # MySQL & InfluxDB pool initializations
│   ├── handler/              # HTTP / MQTT route handlers
│   └── repository/           # Data access objects (SQL statements, Influx writes)
├── configs/
│   └── firebase-service-account.json # Git-ignored service keys
├── Go.mod
└── README.md
```
