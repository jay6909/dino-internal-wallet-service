#!/bin/bash
set -e
docker compose build --no-cache
echo "Starting database & seeding data..."

docker-compose up -d db

echo "🌱 Running Go seeder..."
docker-compose run --rm seed

echo "✅ Setup complete"
