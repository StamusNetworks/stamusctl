#!/bin/bash
# Verify that configuration.nix was rendered (no Go template markers remain)
set -euo pipefail

CONFIG="$STAMUSCTL_CONFIG_PATH/configuration.nix"

if [ ! -f "$CONFIG" ]; then
  echo "FAIL: $CONFIG not found"
  exit 1
fi

if grep -q '{{ "{{" }}' "$CONFIG"; then
  echo "FAIL: unrendered Go template markers found in $CONFIG"
  exit 1
fi

echo "OK: configuration.nix is fully rendered"
