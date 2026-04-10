#!/bin/sh

set -eu

script_dir=$(CDPATH= cd -- "$(dirname "$0")" && pwd)
target_script="$script_dir/acceptance_quickstart.sh"

tmpdir="$(mktemp -d)"
cleanup() {
  rm -rf "$tmpdir"
}
trap cleanup EXIT INT TERM

assert_file_equals() {
  file=$1
  expected=$2

  actual=$(cat "$file")
  if [ "$actual" != "$expected" ]; then
    echo "expected $file to equal: $expected" >&2
    echo "actual: $actual" >&2
    exit 1
  fi
}

assert_contains() {
  file=$1
  expected=$2

  if ! grep -Fq "$expected" "$file"; then
    echo "expected $file to contain: $expected" >&2
    cat "$file" >&2
    exit 1
  fi
}

test_preserves_host_go_caches_when_home_isolated() {
  test_root="$tmpdir/cache-preserve"
  mkdir -p "$test_root/scripts"

  cp "$target_script" "$test_root/scripts/acceptance_quickstart.sh"

  expected_gomodcache=$(go env GOMODCACHE)
  expected_gocache=$(go env GOCACHE)

  cat <<'EOF' >"$test_root/scripts/build_binaries.sh"
#!/bin/sh
set -eu

output_dir=$1
mkdir -p "$output_dir"

printf '%s' "${GOMODCACHE:-}" >"$TEST_ROOT/gomodcache.txt"
printf '%s' "${GOCACHE:-}" >"$TEST_ROOT/gocache.txt"

cat <<'BIN' >"$output_dir/orbit"
#!/bin/sh
exit 1
BIN
chmod +x "$output_dir/orbit"

cat <<'BIN' >"$output_dir/harness"
#!/bin/sh
exit 1
BIN
chmod +x "$output_dir/harness"
EOF
  chmod +x "$test_root/scripts/build_binaries.sh"

  set +e
  (
    unset GOMODCACHE GOCACHE
    TEST_ROOT="$test_root" sh "$test_root/scripts/acceptance_quickstart.sh" >/dev/null 2>&1
  )
  status=$?
  set -e

  if [ "$status" -eq 0 ]; then
    echo "expected acceptance quickstart fixture run to stop after stub binaries are used" >&2
    exit 1
  fi

  assert_file_equals "$test_root/gomodcache.txt" "$expected_gomodcache"
  assert_file_equals "$test_root/gocache.txt" "$expected_gocache"
}

test_acceptance_includes_migrated_runtime_writeback() {
  test_root="$tmpdir/migrated-runtime"
  output_file="$test_root/output.txt"
  mkdir -p "$test_root"

  (
    cd "$script_dir/.."
    sh ./scripts/acceptance_quickstart.sh
  ) >"$output_file" 2>&1

  assert_contains "$output_file" "[quickstart-acceptance] verifying migrated runtime writeback"
  assert_contains "$output_file" "[quickstart-acceptance] verifying source branch taxonomy"
}

test_preserves_host_go_caches_when_home_isolated

test_acceptance_includes_migrated_runtime_writeback

echo "acceptance_quickstart.sh tests passed"
