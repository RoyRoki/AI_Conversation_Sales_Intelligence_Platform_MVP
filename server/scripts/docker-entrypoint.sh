#!/bin/sh
set -e

echo "Starting OMX AI Conversation Platform..."

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL..."
POSTGRES_HOST="${POSTGRES_HOST:-postgres}"
POSTGRES_PORT="${POSTGRES_PORT:-5432}"
POSTGRES_USER="${POSTGRES_USER:-omx_user}"

until pg_isready -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" 2>/dev/null; do
  echo "PostgreSQL is unavailable - sleeping"
  sleep 2
done
echo "PostgreSQL is up!"

# Wait for ChromaDB to be ready (with timeout)
echo "Waiting for ChromaDB..."
CHROMA_URL="${CHROMA_URL:-http://chromadb:8000}"
CHROMA_TIMEOUT=30
CHROMA_ELAPSED=0
while [ $CHROMA_ELAPSED -lt $CHROMA_TIMEOUT ]; do
  if curl -f -s "$CHROMA_URL/api/v2/heartbeat" >/dev/null 2>&1; then
    echo "ChromaDB is up!"
    break
  fi
  echo "ChromaDB is unavailable - sleeping"
  sleep 2
  CHROMA_ELAPSED=$((CHROMA_ELAPSED + 2))
done
if [ $CHROMA_ELAPSED -ge $CHROMA_TIMEOUT ]; then
  echo "Warning: ChromaDB not ready after ${CHROMA_TIMEOUT}s, continuing anyway (graceful fallback enabled)"
fi

# Run migrations
echo "Running database migrations..."
./migrate -direction=up
echo "Migrations completed!"

# Start the API server
echo "Starting API server..."
exec ./api

