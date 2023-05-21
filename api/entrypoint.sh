#!/bin/sh -eu

# Run the migrations
sleep 10
echo "Running migrations..."
migrate -database "${POSTGRESQL_URL}" -path db/migrations up

# Run the server
echo "Starting server..."
/root/server
