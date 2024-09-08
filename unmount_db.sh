#!/bin/bash

# Stop the containers and remove the Postgres service
echo "Stopping Docker containers with docker-compose down..."
docker-compose down -v
docker-compose -f docker-compose.docker.yml down -v

# Check if the 'database' directory exists and delete it
if [ -d "database" ]; then
  echo "Deleting the 'database' directory..."
  rm -rf database
else
  echo "'database' directory does not exist."
fi

# Check if the 'postgres-data' directory exists and delete it
if [ -d "postgres-data" ]; then
  echo "Deleting the 'postgres-data' directory..."
  rm -rf postgres-data
else
  echo "'postgres-data' directory does not exist."
fi

echo "Cleanup complete."