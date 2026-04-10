#!/usr/bin/env sh

set -eu

script_dir=$(CDPATH= cd -- "$(dirname "$0")" && pwd)
repo_root=$(CDPATH= cd -- "$script_dir/.." && pwd)
output_dir="${1:-$repo_root/.dist/bin}"

mkdir -p "$output_dir"

run_go() {
  if command -v go >/dev/null 2>&1; then
    go "$@"
    return
  fi

  if command -v mise >/dev/null 2>&1; then
    mise exec -- go "$@"
    return
  fi

  echo "go toolchain not found in PATH and mise is unavailable" >&2
  exit 127
}

(
  cd "$repo_root"
  run_go build -o "$output_dir/orbit" ./cmd/orbit
  run_go build -o "$output_dir/harness" ./cmd/harness
)

printf 'built orbit: %s\n' "$output_dir/orbit"
printf 'built harness: %s\n' "$output_dir/harness"
