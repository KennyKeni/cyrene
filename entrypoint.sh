#!/bin/sh
set -e

if [ -n "$DB_HOST" ]; then
    echo "Running database migrations..."
    goose -dir /app/migrations postgres "postgresql://${DB_USERNAME}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_DATABASE}?sslmode=disable" up
fi

echo "Starting application..."
exec ./main
