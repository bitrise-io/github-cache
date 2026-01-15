#!/bin/sh
set -e

# If no arguments are provided (post step), default to "save"
if [ $# -eq 0 ]; then
    exec /bitrise-cache save
fi

# Run the bitrise-cache binary with the provided command
exec /bitrise-cache "$@"
