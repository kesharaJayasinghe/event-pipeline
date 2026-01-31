package main

import (
	"context"
	"encoding/json"
	// "fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/twmb/franz-go/pkg/kgo"
)

type Reading struct {
	SensorID  string  `json:"sensor_id"`
	Value     float64 `json:"value"`
	Timestamp int64   `json:"timestamp"`
}

func main() {
	// Connect to db (TimescaleDB)
	ctx := context.Background()
	connStr := "postgres://user:password@localhost:5434/firehose"
	db, err := pgx.Connect(ctx, connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close(ctx)
	log.Println("Connected to TimescaleDB")

	// Connect to Kafka (consumer mode)
	// Consumer group allows running 10 copies of this processor (will automatically share load)
	client, err := kgo.NewClient(
		kgo.SeedBrokers("localhost:19092"),
		kgo.ConsumerGroup("processor-group"),
		kgo.ConsumeTopics("sensor-readings"),
	)
	if err != nil {
		log.Fatalf("Failed to create Kafka client: %v", err)
	}
	defer client.Close()
	log.Println("Connected to Redpanda Kafka")

	// Processing loop
	for {
		// Fetch a batch (up to 100 messages or wait 1 second)
		// This is "Backpressure" handling implicitly
		fetches := client.PollRecords(ctx, 1000)
		if fetches.IsClientClosed() {
			break
		}

		if errs := fetches.Errors(); len(errs) > 0 {
			log.Printf("Kafka errors: %v", errs)
			continue
		}

		// If no data, loop again
		if fetches.Empty() {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// Prepare batch for Postgres
		batch := &pgx.Batch{}
		recordCount := 0

		fetches.EachRecord(func(rec *kgo.Record) {
			var reading Reading
			if err := json.Unmarshal(rec.Value, &reading); err != nil {
				log.Printf("Skipping bad JSON: %v", err)
				return
			}

			// Convert Unix timestamp to Go Time
			ts := time.Unix(reading.Timestamp, 0)

			// Add to batch queue
			batch.Queue("INSERT INTO readings (time, sensor_id, value) VALUES ($1, $2, $3)", ts, reading.SensorID, reading.Value)
			recordCount++
		})

		// Execute bulk insert (the "Transaction")
		if recordCount > 0 {
			br := db.SendBatch(ctx, batch)
			if _, err := br.Exec(); err != nil {
				log.Printf("Failed to insert batch: %v", err)
				br.Close()
				continue // Don't commit offsets if DB fails
			}
			br.Close()
			log.Printf("Processed batch of %d records", recordCount)
		}
		// Commit offsets after successful processing
		client.CommitUncommittedOffsets(ctx)
	}
}