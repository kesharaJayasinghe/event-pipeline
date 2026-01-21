package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	// "time"
	
	"github.com/twmb/franz-go/pkg/kgo"
)

// Global producer client
var client *kgo.Client

func main() {
	// Initialize Kafka client
	// Point to localhost:19092 (external port used in Docker)
	var err error
	client, err = kgo.NewClient(
		kgo.SeedBrokers("localhost:19092"),
		kgo.AllowAutoTopicCreation(), // Create topic if missing (dev only)
	)
	if err != nil {
		log.Fatalf("Failed to create Kafka client: %v", err)
	}
	defer client.Close()

	// Setup HTTP Handler
	http.HandleFunc("POST /readings", handleReadings)

	log.Println("Firehose API listening on :8000")
	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// Data packet structure
type Reading struct {
	SensorID string `json:"sensor_id"`
	Value float64 `json:"value"`
	Timestamp int64 `json:"timestamp"` // Unix epoch time
}

func handleReadings(w http.ResponseWriter, r *http.Request) {
	// Parse Request
	var reading Reading
	if err := json.NewDecoder(r.Body).Decode(&reading); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Serialize for Kafka
	payload, _ := json.Marshal(reading)

	// Async Produce
	record := &kgo.Record{
		Topic: "sensor-readings",
		Key:   []byte(reading.SensorID),
		Value: payload,
	}

	client.Produce(context.Background(), record, nil)

	w.WriteHeader(http.StatusAccepted)
}