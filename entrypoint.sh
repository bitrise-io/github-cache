#!/bin/sh
set -e

# Run the bitrise-cache binary
# Phase detection (restore vs save) is handled internally via state variables
exec /bitrise-cache
