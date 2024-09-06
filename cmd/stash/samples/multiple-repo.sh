#!/bin/bash

# Define your inputs
PROJECT_NAME="<your-project-name>"
HOST_URL="<your-host-url>"
USERNAME="<your-stash-username>"
TOKEN="<your-token>"
ZIP_FOLDER_PATH="<your-zip-folder-path>"

# List of repository names
REPO_NAMES=("repo1" "repo2" "repo3")  # Add your repo names here

# Iterate over each repo name and start a migrator process
for REPO_NAME in "${REPO_NAMES[@]}"; do
  echo "Starting migration for repository: $REPO_NAME"

    ./migrator stash git-export --project "$PROJECT_NAME" --repository "$REPO_NAME" --host "$HOST_URL" --username "$USERNAME" --token "$TOKEN" "$ZIP_FOLDER_PATH/$REPO_NAME.zip"
    echo "Migration completed for $REPO_NAME"
done

echo "All migrations completed."
