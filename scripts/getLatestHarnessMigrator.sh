#!/bin/bash
# Copyright 2025 Harness Inc. All rights reserved.
# Use of this source code is governed by the PolyForm Free Trial 1.0.0 license
# that can be found in the licenses directory at the root of this repository, also available at
# https://polyformproject.org/wp-content/uploads/2020/05/PolyForm-Free-Trial-1.0.0.txt.

#########################
# Check that GITHUBTOK exists in users' environment
# and if not, check to see if the token has been passed
# into script
#########################
if [ -z "$GITHUBTOK" ]; then
  if [ -n "$1" ]; then
    GITHUBTOK="$1"
  else
    echo "GITHUBTOK is not environment and no arguments passed in"
    echo "Make sure that the variable GITHUBTOK is in your environment, or"
    echo "pass your github token into this script as follows:"
    echo "$0 ghp_xxxxxxxxxxx"
    exit 1
  fi
fi

# Make sure that jq is installed
if command -v "jq" &>/dev/null; then
  # We can continue
  jqExists=True
else
  # jq is not installed. Error out
  echo "Requirements not met - must have jq in path"
  echo "Please install jq"
  exit 2
fi

# This gets the latest revision tag from github
echo "Getting latest release information from github"
latestRevTag=$(curl -s -L -H "Accept: application/vnd.github+json" -H "Authorization: Bearer $GITHUBTOK" -H "X-GitHub-Api-Version: 2022-11-28" https://api.github.com/repos/harness/harness-migrate/releases?per_page=1 | jq -r '.[].tag_name')
# The tag is "vX.X.X" - Strip v off the front
latestRev="${latestRevTag#?}"

# Check to see if it's already installed
if command -v "harness-migrate" &>/dev/null; then
  # Compare to current revision already installed
  curRev=$(harness-migrate --version)
  if [[ "$curRev" == "$latestRev" ]]; then
    echo "No need to update - you're already on the latest revision ($latestRev)"
    exit 0
  fi
fi

echo "Downloading and installing latest harness-migrate ($latestRev)"

pushd /tmp > /dev/null

#########################
# Figure out which distribution to download/install
# "uname -s" will return Linux or Darwin
# "uname -m" will return arm64 or x86_64
#########################
osNameCap=$(uname -s)

# Check for CYGWIN (Windows)
if [[ $osNameCap == *"CYGWIN"* ]]; then
  osName="windows"
else
  # lowercase first letter of osName: Linux->linux, Darwin->darwin
  osName="${osNameCap,}"
fi

arch=$(uname -m)
if [[ "$arch" == "x86_64" ]]; then
  arch="amd64"
fi

fileToDownload="harness-migrate-$osName-$arch.tar.gz"
echo -e "Architecture = [$arch], osName = [$osName]\nDownloading $fileToDownload"
curl -s -L https://github.com/harness/harness-migrate/releases/latest/download/$fileToDownload | tar zx

if [[ "$osName" == "windows" ]]; then
  userBinariesPath="/usr/local/bin"
  fName="harness-migrate.exe"
else
  # Use systemd-path to get the user-binaries path
  userBinariesPath=$(systemd-path user-binaries)
  fName="harness-migrate"
fi

cd "$userBinariesPath"

# Remove harness-migrate if it already exists
if test -f "$fName"; then
  rm "$fName"
fi
mv "/tmp/$fName" .
popd > /dev/null

# Confirm new version is now installed
newVer=$(harness-migrate --version)
echo "harness-migrate version: $newVer"
