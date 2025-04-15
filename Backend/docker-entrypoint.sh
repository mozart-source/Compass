#!/bin/bash
set -e

# Wait for MongoDB to be ready
echo "Waiting for MongoDB to be ready..."
MAX_ATTEMPTS=30
COUNTER=0

until mongosh --host $MONGODB_HOST --port $MONGODB_PORT --eval "db.adminCommand('ping')" &>/dev/null; do
  COUNTER=$((COUNTER+1))
  if [ $COUNTER -gt $MAX_ATTEMPTS ]; then
    echo "Failed to connect to MongoDB after $MAX_ATTEMPTS attempts - continuing anyway"
    break
  fi
  echo "Waiting for MongoDB... ($COUNTER/$MAX_ATTEMPTS)"
  sleep 2
done

if [ $COUNTER -lt $MAX_ATTEMPTS ]; then
  echo "MongoDB is ready!"
fi

# Run the MongoDB verification script
echo "Verifying MongoDB setup..."
python scripts/test_mongodb.py || true

# Start the application
echo "Starting the application..."
exec uvicorn main:app --host 0.0.0.0 --port 8001 