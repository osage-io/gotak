#!/bin/bash

echo "Testing GoTAK Server with structured logging..."
echo "Starting server (will auto-terminate after 3 seconds)..."

# Start server in background
./bin/gotak-server -config config/test.yaml &
SERVER_PID=$!

# Wait for 3 seconds
sleep 3

# Terminate the server
kill -TERM $SERVER_PID 2>/dev/null

# Wait a bit for graceful shutdown
sleep 1

echo "Test complete!"
