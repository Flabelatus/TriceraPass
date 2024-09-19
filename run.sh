#!/bin/bash

# Function to display usage
usage(){
  echo "Usage: ./run.sh [OPTIONS]"
  echo ""
  echo "Options:"
  echo "  --containerize       Run the services all within one docker container"
  echo "  --help               Display this help message"
  echo ""
  echo "The --containerize flag runs the auth-service, data-service and your custom application
  inside a Docker container, setting up the network and environment needed for your application."
}

CONTAINERIZE=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    --containerize)
      CONTAINERIZE=true
      shift
      ;;
    --help)
      usage
      exit 0
      ;;
    *)
      echo "Error: Unknown flag '$1'. Use --help for usage information."
      exit 1
      ;;
  esac
done

# Check if curl is installed
if ! command -v curl >/dev/null 2>&1; then
  echo "Curl is not installed. Installing curl..."
  apk --no-cache add curl || { echo "Failed to install curl. Exiting."; exit 1; }
fi

# Create the .env file 
if [[ "$CONTAINERIZE" == true ]]; then
  echo "Generating .env file from settings.yml for docker instance... "
  ./generate-env.sh --docker
else
  echo "Generating .env file from settings.yml for local instance... "
  ./generate-env.sh 
fi


if [ $? -eq 0 ]; then
  echo ".env file generated successfully."
else
  echo "Failed to generate .env file. Exiting."
  exit 1
fi

# Ensure the .env file was generated successfully
if [ ! -f .env ]; then
    echo ".env file is missing, please check the generation process."
    exit 1
fi

echo ".env file generated successfully."
cat .env

# Download wait-for-it.sh if it doesn't exist
if [ ! -f ./wait-for-it.sh ]; then
    echo "wait-for-it.sh not found, downloading..."
    curl -o wait-for-it.sh https://raw.githubusercontent.com/vishnubob/wait-for-it/master/wait-for-it.sh || { echo "Failed to download wait-for-it.sh. Exiting."; exit 1; }
    chmod +x wait-for-it.sh
fi

# Check for flags
if [[ "$CONTAINERIZE" == true ]]; then
    # Run with Docker
    echo "Running both services in Docker..."
    docker-compose down
    docker-compose -f docker-compose.docker.yml up --build -d

    # Check if docker-compose was successful
    if [ $? -eq 0 ]; then
      echo "Docker containers started successfully."
    else
      echo "Failed to start Docker containers. Exiting."
      exit 1
    fi
else
    # Run locally
    echo "Running postgres in Docker and auth-service locally..."
    docker-compose up -d postgres
    go run ./cmd/api

    if [ $? -eq 0 ]; then
      echo "API started successfully."
    else
      echo "Failed to start the API. Exiting."
      exit 1
    fi
fi