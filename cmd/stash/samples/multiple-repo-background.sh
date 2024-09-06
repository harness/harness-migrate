#!/bin/bash

# Define your inputs
PROJECT_NAME="<your-project-name>"
HOST_URL="<your-host-url>"
USERNAME="<your-stash-username>"
TOKEN="<your-token>"
ZIP_FOLDER_PATH="<your-zip-folder-path>"

# List of repository names
REPO_NAMES=("repo1" "repo2" "repo3")  # Add your repo names here

# Iterate over each repo name and start a separate migrator process
for REPO_NAME in "${REPO_NAMES[@]}"; do
  echo "Starting migration for repository: $REPO_NAME"

  # Start the migrator as a background process
  ./migrator stash git-export --project "$PROJECT_NAME" --repository "$REPO_NAME" --host "$HOST_URL" --username "$USERNAME" --token "$TOKEN" "$ZIP_FOLDER_PATH/$REPO_NAME.zip" &

  # Capture the process ID (PID) and display it
  PID=$!
  echo "Migrator started for $REPO_NAME with PID $PID"
done

# Wait for all background processes to finish
wait

echo "All migrations completed."