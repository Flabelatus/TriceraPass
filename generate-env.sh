#!/bin/bash

SETTINGS_FILE="./settings.yml"

# Extracting other settings
POSTGRES_USER=$(grep 'user:' $SETTINGS_FILE | sed 's/.*: //')
POSTGRES_PASSWORD=$(grep 'password:' $SETTINGS_FILE | sed 's/.*: //')
POSTGRES_DB=$(grep 'dbname:' $SETTINGS_FILE | sed 's/.*: //')

# Extract allowedOrigins block and join them into a single line separated by commas
CORS=$(sed -n '/allowedOrigins:/,/^[^ ]/p' $SETTINGS_FILE | grep -E 'http' | sed 's/^- //g' | paste -sd "," -)

# Generate the .env file
cat > .env <<EOL
POSTGRES_USER=$POSTGRES_USER
POSTGRES_PASSWORD=$POSTGRES_PASSWORD
POSTGRES_DB=$POSTGRES_DB
CONFIG_FILE=./settings.yml
CORS=$CORS
EOL

echo ".env file has been generated from settings.yml"