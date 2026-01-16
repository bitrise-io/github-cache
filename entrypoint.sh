#!/bin/sh
set -e

# Create .envstore.yml if it doesn't exist (required by envman)
if [ ! -f "${GITHUB_WORKSPACE}/.envstore.yml" ]; then
    touch "${GITHUB_WORKSPACE}/.envstore.yml"
fi

# Set ENVMAN_ENVSTORE_PATH so envman knows where to find the envstore
export ENVMAN_ENVSTORE_PATH="${GITHUB_WORKSPACE}/.envstore.yml"

# Run the bitrise-cache binary
# Phase detection (restore vs save) is handled internally via state variables
exec /bitrise-cache
