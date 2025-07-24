#!/bin/bash

# Start PostgreSQL database using Docker Compose
echo "Starting PostgreSQL database..."

cd docker
docker-compose up -d postgres

echo "Waiting for database to be ready..."
sleep 10

echo "Database is ready!"
echo "Connection details:"
echo "  Host: localhost"
echo "  Port: 5432"
echo "  Database: seta_training"
echo "  Username: postgres"
echo "  Password: password"
echo ""
echo "To stop the database, run: docker-compose down"
