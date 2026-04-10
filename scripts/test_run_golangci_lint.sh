#!/bin/sh

set -eu

script_dir=$(CDPATH= cd -- "$(dirname "$0")" && pwd)
target_script="$script_dir/run_golangci_lint.sh"

tmpdir="$(mktemp -d)"
cleanup() {
  rm -rf "$tmpdir"
}
trap cleanup EXIT INT TERM

assert_contains() {
  file=$1
  pattern=$2

  if ! grep -Fq "$pattern" "$file"; then
    echo "expected output to contain: $pattern" >&2
    cat "$file" >&2
    exit 1
  fi
}

assert_not_contains() {
  file=$1
  pattern=$2

  if grep -Fq "$pattern" "$file"; then
    echo "expected output to not contain: $pattern" >&2
    cat "$file" >&2
    exit 1
  fi
}

run_timeout_fallback_test() {
  test_root="$tmpdir/timeout"
  mkdir -p "$test_root/bin" "$test_root/repo"

  cat <<'EOF' >"$test_root/bin/golangci-lint"
#!/bin/sh
set -eu

case "${1:-}" in
  version)
    echo "golangci-lint has version 2.10.1"
    exit 0
    ;;
  config)
    cat <<'ERR' >&2
Get "https://golangci-lint.run/jsonschema/golangci.jsonschema.json": context deadline exceeded (Client.Timeout exceeded while awaiting headers)
ERR
    exit 1
    ;;
  run)
    echo "lint ok"
    exit 0
    ;;
esac

echo "unexpected args: $*" >&2
exit 99
EOF
  chmod +x "$test_root/bin/golangci-lint"

  cat <<'EOF' >"$test_root/repo/go.mod"
module example.com/test

go 1.26.0
EOF

  (
    cd "$test_root/repo"
    PATH="$test_root/bin:$PATH" sh "$target_script"
  ) >"$test_root/output.txt" 2>&1

  assert_contains "$test_root/output.txt" "Skipping golangci-lint config verify because remote schema lookup is unavailable"
  assert_contains "$test_root/output.txt" "lint ok"
}

run_schema_error_test() {
  test_root="$tmpdir/schema"
  mkdir -p "$test_root/bin" "$test_root/repo"

  cat <<'EOF' >"$test_root/bin/golangci-lint"
#!/bin/sh
set -eu

case "${1:-}" in
  version)
    echo "golangci-lint has version 2.10.1"
    exit 0
    ;;
  config)
    cat <<'ERR' >&2
jsonschema: "/linters/default" does not validate with "/properties/linters/properties/default/type": expected string or array, but got number
ERR
    exit 1
    ;;
  run)
    echo "lint should not run" >&2
    exit 91
    ;;
esac

echo "unexpected args: $*" >&2
exit 99
EOF
  chmod +x "$test_root/bin/golangci-lint"

  cat <<'EOF' >"$test_root/repo/go.mod"
module example.com/test

go 1.26.0
EOF

  set +e
  (
    cd "$test_root/repo"
    PATH="$test_root/bin:$PATH" sh "$target_script"
  ) >"$test_root/output.txt" 2>&1
  status=$?
  set -e

  if [ "$status" -eq 0 ]; then
    echo "expected schema validation failure to stop the script" >&2
    cat "$test_root/output.txt" >&2
    exit 1
  fi

  assert_contains "$test_root/output.txt" "jsonschema: \"/linters/default\" does not validate"
  assert_not_contains "$test_root/output.txt" "Skipping golangci-lint config verify because remote schema lookup is unavailable"
}

run_timeout_fallback_test
run_schema_error_test

echo "run_golangci_lint.sh tests passed"
