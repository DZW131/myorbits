#!/bin/sh

set -eu

script_dir=$(CDPATH= cd -- "$(dirname "$0")" && pwd)
target_script="$script_dir/build_binaries.sh"

tmpdir="$(mktemp -d)"
cleanup() {
  rm -rf "$tmpdir"
}
trap cleanup EXIT INT TERM

assert_file_exists() {
  path=$1

  if [ ! -f "$path" ]; then
    echo "expected file to exist: $path" >&2
    exit 1
  fi
}

assert_executable() {
  path=$1

  if [ ! -x "$path" ]; then
    echo "expected executable file: $path" >&2
    exit 1
  fi
}

assert_contains() {
  file=$1
  pattern=$2

  if ! grep -Fq "$pattern" "$file"; then
    echo "expected output to contain: $pattern" >&2
    cat "$file" >&2
    exit 1
  fi
}

output_dir="$tmpdir/bin"
sh "$target_script" "$output_dir" >"$tmpdir/build.txt" 2>&1

assert_file_exists "$output_dir/orbit"
assert_file_exists "$output_dir/harness"
assert_executable "$output_dir/orbit"
assert_executable "$output_dir/harness"
assert_contains "$tmpdir/build.txt" "built orbit:"
assert_contains "$tmpdir/build.txt" "built harness:"

"$output_dir/orbit" --help >"$tmpdir/orbit-help.txt" 2>&1
"$output_dir/harness" --help >"$tmpdir/harness-help.txt" 2>&1
"$output_dir/orbit" completion bash >"$tmpdir/orbit-completion.txt" 2>&1
"$output_dir/harness" completion bash >"$tmpdir/harness-completion.txt" 2>&1

assert_contains "$tmpdir/orbit-help.txt" "orbit"
assert_contains "$tmpdir/orbit-help.txt" "template"
assert_contains "$tmpdir/harness-help.txt" "harness"
assert_contains "$tmpdir/harness-help.txt" "install"
assert_contains "$tmpdir/orbit-completion.txt" "orbit"
assert_contains "$tmpdir/harness-completion.txt" "harness"

echo "build_binaries.sh tests passed"
