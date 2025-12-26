#!/bin/bash
cd "$(dirname "$0")/server"
export PORT=8080
export DB_TYPE=postgres
export DATABASE_URL="postgres://omx_user:omx_password@localhost:5432/omx_db?sslmode=disable"
export CHROMA_URL="http://localhost:8000"
export JWT_SECRET="default-secret-change-in-production"
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5432
export POSTGRES_USER=omx_user
export POSTGRES_PASSWORD=omx_password
export POSTGRES_DB=omx_db
export DEFAULT_ADMIN_TENANT_ID=OMX26
export DEFAULT_ADMIN_EMAIL=OMX2026@gmail.com
export DEFAULT_ADMIN_PASSWORD=OMX@2026

echo "Starting server with new code..."
./bin/api
