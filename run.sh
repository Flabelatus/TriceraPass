#!/bin/bash

# Step 1: Run the generate-env.sh script to create the .env file
echo "Generating .env file from settings.yml..."
./generate-env.sh

# Check if the .env file was successfully created
if [ $? -eq 0 ]; then
  echo ".env file generated successfully."
else
  echo "Failed to generate .env file. Exiting."
  exit 1
fi

# Step 2: Build and start the Docker containers using docker-compose for the Postgres server in development
echo "Building and starting Docker containers..."
docker-compose up -d

# Step 3: Run the api via go command in development 
echo "Starting the API..."
go run ./cmd/api

# Check if docker-compose was successful
if [ $? -eq 0 ]; then
  echo "Docker containers started successfully."
else
  echo "Failed to start Docker containers. Exiting."
  exit 1
fi