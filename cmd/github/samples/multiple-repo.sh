#!/bin/bash

# Define your inputs
ORG_NAME="<your-org-name>"
USERNAME="<your-username>"
ZIP_FOLDER_PATH="<your-zip-folder-path>"

# List of repository names
REPO_NAMES=("repo1" "repo2" "repo3")  # Add your repo names here

# Iterate over each repo name and start a migrator process
for REPO_NAME in "${REPO_NAMES[@]}"; do
  echo "Starting migration for repository: $REPO_NAME"
  ./migrator github git-export --org "$ORG_NAME" --repository "$REPO_NAME" --username "$USERNAME" "$ZIP_FOLDER_PATH/$REPO_NAME.zip"
    echo "Migrator completed for $REPO_NAME"
done

echo "All migrations completed."
