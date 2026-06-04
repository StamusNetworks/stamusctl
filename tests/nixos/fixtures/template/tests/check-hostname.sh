#!/bin/bash
# Verify that the hostname parameter was injected into configuration.nix
set -euo pipefail

CONFIG="$STAMUSCTL_CONFIG_PATH/configuration.nix"

if ! grep -q 'networking.hostName' "$CONFIG"; then
  echo "FAIL: networking.hostName not found in $CONFIG"
  exit 1
fi

echo "OK: hostname is configured"
