#!/bin/sh

set -eu

is_remote_schema_network_failure() {
  output_file=$1

  if ! grep -Eqi '(https?://[^[:space:]]+|golangci-lint\.run|jsonschema/.+\.json|raw\.githubusercontent\.com/.+jsonschema)' "$output_file"; then
    return 1
  fi

  if grep -Eqi '(no such host|i/o timeout|dial tcp|context deadline exceeded|client\.timeout exceeded|tls handshake timeout|temporary failure in name resolution|connection timed out|network is unreachable|connection reset by peer)' "$output_file"; then
    return 0
  fi

  return 1
}

if [ ! -f go.mod ]; then
  echo "No go.mod yet; skipping golangci-lint"
  exit 0
fi

export GOPROXY="${GOPROXY:-direct}"
export GOSUMDB="${GOSUMDB:-off}"

golangci-lint version

verify_output_file="$(mktemp)"
cleanup() {
  rm -f "$verify_output_file"
}
trap cleanup EXIT INT TERM

if golangci-lint config verify >"$verify_output_file" 2>&1; then
  cat "$verify_output_file"
else
  verify_status=$?
  cat "$verify_output_file"
  if is_remote_schema_network_failure "$verify_output_file"; then
    echo "Skipping golangci-lint config verify because remote schema lookup is unavailable"
  else
    exit "$verify_status"
  fi
fi

golangci-lint run --timeout=10m ./...
