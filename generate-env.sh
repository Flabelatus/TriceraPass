#!/bin/bash

SETTINGS_FILE="./settings.yml"
POSTGRES_HOST="localhost"

# Install yq based on the operating system
install_yq() {
  if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    # Install yq on Linux
    echo "Installing yq for Linux..."
    sudo wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/download/v4.16.1/yq_linux_amd64
    sudo chmod +x /usr/local/bin/yq
  elif [[ "$OSTYPE" == "darwin"* ]]; then
    # Install yq on macOS using Homebrew
    echo "Installing yq for macOS..."
    brew install yq
  elif [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "mingw"* ]]; then
    # Install yq on Windows (MSYS/MinGW environment)
    echo "Installing yq for Windows..."
    YQ_VERSION="v4.16.1"
    YQ_BINARY="yq_windows_amd64.exe"
    wget https://github.com/mikefarah/yq/releases/download/${YQ_VERSION}/${YQ_BINARY} -O yq.exe
    chmod +x yq.exe
    mv yq.exe /usr/local/bin/yq.exe
  else
    echo "Unsupported OS. Please install yq manually: https://github.com/mikefarah/yq"
    exit 1
  fi
}

# Check if yq is installed, install it if not found
if ! command -v yq &> /dev/null; then
  echo "yq not found. Installing yq..."
  install_yq
else
  echo "yq is already installed."
fi

# Check if the --docker flag is passed
if [[ "$1" == "--docker" ]]; then
  POSTGRES_HOST="postgres"
fi

# Extracting other settings
# POSTGRES_HOST=$POSTGRES_HOST
POSTGRES_USER=$(yq e '.database.user' $SETTINGS_FILE)
POSTGRES_PASSWORD=$(yq e '.database.password' $SETTINGS_FILE)
POSTGRES_DB=$(yq e '.database.dbname' $SETTINGS_FILE)

# Extract mail server credentials
MAIL_SERVER_NAME=$(yq e '.email_server.server_name' $SETTINGS_FILE)
MAIL_SERVER_API_KEY=$(yq e '.email_server.api_key' $SETTINGS_FILE)
MAIL_SERVER_DOMAIN=$(yq e '.email_server.domain' $SETTINGS_FILE)

# Extract the server port from settings.yml
AUTH_SERVICE_PORT=$(yq e '.server.port' $SETTINGS_FILE)

# Extract allowed_origins and join them into a single line separated by commas
CORS=$(yq e '.api.allowed_origins[]' $SETTINGS_FILE | paste -sd "," -)

# Generate the .env file
cat > .env <<EOL
POSTGRES_HOST=$POSTGRES_HOST
POSTGRES_USER=$POSTGRES_USER
POSTGRES_PASSWORD=$POSTGRES_PASSWORD
POSTGRES_DB=$POSTGRES_DB
CONFIG_FILE=./settings.yml
AUTH_SERVICE_PORT=$AUTH_SERVICE_PORT
MAIL_SERVER_NAME=$MAIL_SERVER_NAME
MAIL_SERVER_API_KEY=$MAIL_SERVER_API_KEY
MAIL_SERVER_DOMAIN=$MAIL_SERVER_DOMAIN
CORS=$CORS
EOL

echo ".env file has been generated from settings.yml"