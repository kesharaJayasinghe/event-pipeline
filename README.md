# High-Throughput IoT Ingestion Pipeline
A high-performance event streaming architecture designed to ingest, buffer, and store massive volumes of IoT sensor data. This project demonstrates the "Firehose" pattern, decoupling high-velocity ingestion from storage writes to ensure system stability under load.

## System Architecture
1. Ingestion Layer (Go): A lightweight, non-blocking REST API that accepts JSON payloads and asynchronously pushes them to the message bus.

2. Buffering Layer (Redpanda): A Kafka-compatible streaming platform that acts as a shock absorber, handling backpressure when the database is under load.

3. Processing Layer (Go): A consumer service that implements Batch Processing. It pulls records from Redpanda and performs bulk inserts (COPY protocol) into the database to maximize write throughput.

4. Storage Layer (TimescaleDB): A PostgreSQL extension optimized for time-series data, partitioned by time for fast analytics queries.

## Key Features
* Backpressure Handling: By decoupling the API from the Database using Redpanda, the API can accept 10k+ req/sec even if the database slows down. The queue simply grows until the consumer catches up.

* Zero-Allocation Ingestion: The Producer utilizes asynchronous dispatch to minimize API latency (p99 < 5ms).

* Smart Batching: The Consumer dynamically adjusts batch sizes based on load. Under low traffic, it processes messages instantly (Real-time). Under high traffic, it groups messages (Batching) to reduce database round-trips.

* Infrastructure as Code: Fully dockerized environment including Redpanda Console for observability.

## Tech Stack
- Language: Go (Golang) 1.21+

- Message Broker: Redpanda (Kafka Protocol)

- Database: TimescaleDB (PostgreSQL 16)

- Drivers: franz-go (Kafka), pgx/v5 (Postgres)

## Quick Start
1. Infrastructure
```
docker compose up -d
# Wait 30s for Redpanda to initialize
```

2. Start the Consumer (Processor)
```
go run cmd/processor/main.go
```

3. Start the Producer (API)
```
go run cmd/api/main.go
```

4. Load Test
```
# Run multiple instances of this to stress the batching logic
./load_test.sh
```

## Analytics Sample
Run the following SQL to see the aggregated sensor data:
```
SELECT time_bucket('1 minute', time) AS bucket, sensor_id, avg(value) 
FROM readings 
GROUP BY bucket, sensor_id 
ORDER BY bucket DESC;
```