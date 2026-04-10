#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
exec sh "$SCRIPT_DIR/acceptance_quickstart.sh" "$@"
