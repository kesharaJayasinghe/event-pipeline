#!/bin/bash
# Spams the API with random data
echo "Starting Load Test... Press Ctrl+C to stop."

while true; do
  # Generate random sensor ID (1-10) and random temp (20-30)
  ID=$((1 + $RANDOM % 10))
  TEMP=$((20 + $RANDOM % 10))
  TS=$(date +%s)

  curl -s -X POST http://localhost:8000/readings \
       -H "Content-Type: application/json" \
       -d "{\"sensor_id\": \"sensor-$ID\", \"value\": $TEMP, \"timestamp\": $TS}" > /dev/null &
  
  # Sleep to avoid crashing completely
  sleep 0.01
done