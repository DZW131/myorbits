#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
TMP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/orbit-quickstart.XXXXXX")"
BIN_DIR="$TMP_DIR/bin"
ORBIT_BIN="$BIN_DIR/orbit"
HARNESS_BIN="$BIN_DIR/harness"
TEMPLATE_REPO="$TMP_DIR/template-source"
SOURCE_REPO="$TMP_DIR/source-authoring"
RUNTIME_REPO="$TMP_DIR/runtime-repo"
BINDINGS_FILE="$TMP_DIR/install-bindings.yaml"
HOME_DIR="$TMP_DIR/home"
LAST_STDOUT=""
LAST_STDERR=""
HOST_GOMODCACHE="${GOMODCACHE:-}"
HOST_GOCACHE="${GOCACHE:-}"

cleanup() {
  chmod -R u+w "$TMP_DIR" 2>/dev/null || true
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT

info() {
  printf '[quickstart-acceptance] %s\n' "$*"
}

fail() {
  printf '[quickstart-acceptance] ERROR: %s\n' "$*" >&2
  exit 1
}

assert_contains() {
  local haystack="$1"
  local needle="$2"
  local label="$3"

  case "$haystack" in
    *"$needle"*) ;;
    *)
      fail "$label: expected to find [$needle] in output: $haystack"
      ;;
  esac
}

assert_not_contains() {
  local haystack="$1"
  local needle="$2"
  local label="$3"

  case "$haystack" in
    *"$needle"*)
      fail "$label: expected to avoid [$needle] in output: $haystack"
      ;;
    *)
      ;;
  esac
}

assert_file_exists() {
  local path="$1"

  [ -e "$path" ] || fail "expected file to exist: $path"
}

assert_file_missing() {
  local path="$1"

  [ ! -e "$path" ] || fail "expected file to be absent: $path"
}

assert_file_contains() {
  local path="$1"
  local needle="$2"
  local label="$3"

  [ -e "$path" ] || fail "$label: expected file to exist: $path"
  local content
  content="$(cat "$path")"
  assert_contains "$content" "$needle" "$label"
}

run_cli() {
  local working_dir="$1"
  shift

  local stdout_file="$TMP_DIR/command.stdout"
  local stderr_file="$TMP_DIR/command.stderr"
  local status=0

  if (
    cd "$working_dir"
    "$@" >"$stdout_file" 2>"$stderr_file"
  ); then
    status=0
  else
    status=$?
  fi

  LAST_STDOUT="$(cat "$stdout_file")"
  LAST_STDERR="$(cat "$stderr_file")"
  return "$status"
}

run_orbit() {
  local working_dir="$1"
  shift

  run_cli "$working_dir" "$ORBIT_BIN" "$@"
}

run_harness() {
  local working_dir="$1"
  shift

  run_cli "$working_dir" "$HARNESS_BIN" "$@"
}

git_in_repo() {
  local repo_root="$1"
  shift

  git -C "$repo_root" "$@"
}

write_file() {
  local path="$1"
  local content="$2"

  mkdir -p "$(dirname "$path")"
  printf '%b' "$content" >"$path"
}

capture_host_go_cache_defaults() {
  if [ -z "$HOST_GOMODCACHE" ] && command -v go >/dev/null 2>&1; then
    HOST_GOMODCACHE="$(go env GOMODCACHE 2>/dev/null || true)"
  fi

  if [ -z "$HOST_GOCACHE" ] && command -v go >/dev/null 2>&1; then
    HOST_GOCACHE="$(go env GOCACHE 2>/dev/null || true)"
  fi
}

setup_git_environment() {
  mkdir -p "$HOME_DIR"
  export HOME="$HOME_DIR"
  export XDG_CONFIG_HOME="$HOME_DIR/.config"
  export GIT_CONFIG_NOSYSTEM=1
  export GIT_CONFIG_GLOBAL="$HOME_DIR/.gitconfig"
  if [ -n "$HOST_GOMODCACHE" ]; then
    export GOMODCACHE="$HOST_GOMODCACHE"
  fi
  if [ -n "$HOST_GOCACHE" ]; then
    export GOCACHE="$HOST_GOCACHE"
  fi
  : >"$GIT_CONFIG_GLOBAL"
}

build_binaries() {
  info "building orbit and harness binaries"
  sh "$REPO_ROOT/scripts/build_binaries.sh" "$BIN_DIR" >/dev/null
}

init_template_source_repo() {
  info "initializing template source repository"
  mkdir -p "$TEMPLATE_REPO"
  git init -b main "$TEMPLATE_REPO" >/dev/null
  git_in_repo "$TEMPLATE_REPO" config user.name "Orbit Acceptance"
  git_in_repo "$TEMPLATE_REPO" config user.email "orbit-acceptance@example.com"

  run_harness "$TEMPLATE_REPO" init || fail "harness init failed for template source: $LAST_STDERR"
  run_orbit "$TEMPLATE_REPO" add docs || fail "orbit add docs failed for template source: $LAST_STDERR"

  write_file "$TEMPLATE_REPO/.harness/vars.yaml" $'schema_version: 1\nvariables:\n  project_name:\n    value: Orbit\n    description: Product title\n'
  write_file "$TEMPLATE_REPO/docs/guide.md" "Orbit guide\n"

  git_in_repo "$TEMPLATE_REPO" add -- .harness docs
  git_in_repo "$TEMPLATE_REPO" commit -m "seed docs orbit" >/dev/null

  run_orbit "$TEMPLATE_REPO" template save docs --to orbit-template/docs || fail "orbit template save failed: $LAST_STDERR"
  assert_contains "$LAST_STDOUT" "saved template orbit docs to branch orbit-template/docs" "template save"

  run_orbit "$TEMPLATE_REPO" branch inspect orbit-template/docs --json || fail "orbit branch inspect failed for orbit template: $LAST_STDERR"
  assert_contains "$LAST_STDOUT" '"kind": "template"' "orbit template branch kind"
  assert_contains "$LAST_STDOUT" '"template_kind": "orbit"' "orbit template branch template kind"
}

init_runtime_repo() {
  info "initializing runtime repository"
  run_harness "$TMP_DIR" create "$RUNTIME_REPO" || fail "harness create failed: $LAST_STDERR"
  git_in_repo "$RUNTIME_REPO" config user.name "Orbit Acceptance"
  git_in_repo "$RUNTIME_REPO" config user.email "orbit-acceptance@example.com"
}

verify_source_branch_taxonomy() {
  info "verifying source branch taxonomy"
  mkdir -p "$SOURCE_REPO"
  git init -b main "$SOURCE_REPO" >/dev/null
  git_in_repo "$SOURCE_REPO" config user.name "Orbit Acceptance"
  git_in_repo "$SOURCE_REPO" config user.email "orbit-acceptance@example.com"

  write_file "$SOURCE_REPO/.orbit/config.yaml" $'version: 1\nshared_scope: []\nbehavior:\n  outside_changes_mode: warn\n  block_switch_if_hidden_dirty: true\n  commit_append_trailer: true\n  sparse_checkout_mode: no-cone\n'
  write_file "$SOURCE_REPO/.orbit/orbits/docs.yaml" $'id: docs\ndescription: Docs orbit\ninclude:\n  - docs/**\n'
  write_file "$SOURCE_REPO/docs/guide.md" "Source Orbit guide\n"
  git_in_repo "$SOURCE_REPO" add .orbit docs
  git_in_repo "$SOURCE_REPO" commit -m "seed source authoring repo" >/dev/null

  run_orbit "$SOURCE_REPO" template init-source || fail "orbit template init-source failed: $LAST_STDERR"
  assert_contains "$LAST_STDOUT" "initialized template source" "template init-source"
  git_in_repo "$SOURCE_REPO" add .harness/manifest.yaml .harness/orbits
  git_in_repo "$SOURCE_REPO" commit -m "initialize source branch" >/dev/null

  run_orbit "$SOURCE_REPO" branch inspect HEAD --json || fail "orbit branch inspect failed for source branch: $LAST_STDERR"
  assert_contains "$LAST_STDOUT" '"kind": "source"' "source branch kind"
  assert_contains "$LAST_STDOUT" '"publish_orbit_id": "docs"' "source branch publish orbit"
}

install_template_into_runtime() {
  info "installing template into runtime repository"
  write_file "$BINDINGS_FILE" $'schema_version: 1\nvariables:\n  project_name:\n    value: Installed Orbit\n'

  run_harness "$RUNTIME_REPO" install "$TEMPLATE_REPO" --ref orbit-template/docs --bindings "$BINDINGS_FILE" || fail "harness install failed: $LAST_STDERR"
  assert_contains "$LAST_STDOUT" "installed orbit docs into harness" "harness install"

  assert_file_exists "$RUNTIME_REPO/.harness/manifest.yaml"
  assert_file_exists "$RUNTIME_REPO/.harness/orbits/docs.yaml"
  assert_file_exists "$RUNTIME_REPO/.harness/installs/docs.yaml"
  assert_file_exists "$RUNTIME_REPO/.harness/vars.yaml"
  assert_file_exists "$RUNTIME_REPO/docs/guide.md"
  assert_file_missing "$RUNTIME_REPO/.orbit/config.yaml"
  assert_file_contains "$RUNTIME_REPO/.harness/manifest.yaml" "source: install_orbit" "runtime manifest install provenance"
  assert_file_contains "$RUNTIME_REPO/.harness/installs/docs.yaml" "source_ref: orbit-template/docs" "runtime install record source_ref"

  run_harness "$RUNTIME_REPO" inspect || fail "harness inspect failed: $LAST_STDERR"
  assert_contains "$LAST_STDOUT" "member_count: 1" "harness inspect member count"

  run_harness "$RUNTIME_REPO" check --json || fail "harness check failed: $LAST_STDERR"
  assert_contains "$LAST_STDOUT" '"ok": true' "harness check ok"

  git_in_repo "$RUNTIME_REPO" add -A
  git_in_repo "$RUNTIME_REPO" commit -m "commit installed runtime" >/dev/null
}

verify_runtime_template_writeback() {
  info "saving runtime orbit back to installed template branch"
  write_file "$RUNTIME_REPO/docs/guide.md" "Improved Installed Orbit guide\n"

  run_orbit "$RUNTIME_REPO" template save docs || fail "orbit template save from runtime failed: $LAST_STDERR"
  assert_contains "$LAST_STDOUT" "saved template orbit docs to branch orbit-template/docs" "runtime template save"

  template_guide="$(git_in_repo "$RUNTIME_REPO" show orbit-template/docs:docs/guide.md)"
  assert_contains "$template_guide" 'Improved $project_name guide' "runtime template writeback payload"
  assert_file_contains "$RUNTIME_REPO/.harness/manifest.yaml" "source: install_orbit" "runtime manifest install provenance after writeback"
  assert_file_contains "$RUNTIME_REPO/.harness/installs/docs.yaml" "source_ref: orbit-template/docs" "runtime install record source_ref after writeback"
}

verify_migrated_runtime_writeback() {
  info "verifying migrated runtime writeback"
  write_file "$RUNTIME_REPO/.orbit/config.yaml" $'version: 1\nshared_scope:\n  - README.md\nbehavior:\n  outside_changes_mode: warn\n  block_switch_if_hidden_dirty: true\n  commit_append_trailer: true\n  sparse_checkout_mode: no-cone\n'
  git_in_repo "$RUNTIME_REPO" add .orbit/config.yaml
  git_in_repo "$RUNTIME_REPO" commit -m "add legacy compatibility config" >/dev/null

  write_file "$RUNTIME_REPO/docs/guide.md" "Migrated Installed Orbit guide\n"

  run_orbit "$RUNTIME_REPO" template save docs --overwrite || fail "orbit template save from migrated runtime failed: $LAST_STDERR"
  assert_contains "$LAST_STDOUT" "saved template orbit docs to branch orbit-template/docs" "migrated runtime template save"

  template_guide="$(git_in_repo "$RUNTIME_REPO" show orbit-template/docs:docs/guide.md)"
  assert_contains "$template_guide" 'Migrated $project_name guide' "migrated runtime template writeback payload"

  template_files="$(git_in_repo "$RUNTIME_REPO" ls-tree -r --name-only orbit-template/docs)"
  assert_not_contains "$template_files" ".orbit/config.yaml" "migrated runtime template payload"
  assert_file_exists "$RUNTIME_REPO/.orbit/config.yaml"
  assert_file_contains "$RUNTIME_REPO/.harness/manifest.yaml" "source: install_orbit" "migrated runtime manifest install provenance"
  assert_file_contains "$RUNTIME_REPO/.harness/installs/docs.yaml" "source_ref: orbit-template/docs" "migrated runtime install record source_ref"
}

verify_runtime_projection_flow() {
  info "verifying runtime projection commands"
  run_orbit "$RUNTIME_REPO" branch inspect HEAD --json || fail "orbit branch inspect failed for runtime branch: $LAST_STDERR"
  assert_contains "$LAST_STDOUT" '"kind": "runtime"' "runtime branch kind"

  run_orbit "$RUNTIME_REPO" enter docs || fail "orbit enter failed in runtime repo: $LAST_STDERR"
  assert_contains "$LAST_STDOUT" "entered orbit docs" "orbit enter"

  run_orbit "$RUNTIME_REPO" current || fail "orbit current failed in runtime repo: $LAST_STDERR"
  assert_contains "$LAST_STDOUT" "docs" "orbit current"

  run_orbit "$RUNTIME_REPO" status || fail "orbit status failed in runtime repo: $LAST_STDERR"
  assert_contains "$LAST_STDOUT" "current: docs" "orbit status current"

  run_orbit "$RUNTIME_REPO" diff || fail "orbit diff failed in runtime repo: $LAST_STDERR"

  run_orbit "$RUNTIME_REPO" leave || fail "orbit leave failed in runtime repo: $LAST_STDERR"
  assert_contains "$LAST_STDOUT" "left orbit docs" "orbit leave"
}

verify_harness_template_branch() {
  info "saving and inspecting harness template branch"
  run_harness "$RUNTIME_REPO" template save --to harness-template/workspace || fail "harness template save failed: $LAST_STDERR"
  assert_contains "$LAST_STDOUT" "saved harness template" "harness template save"

  run_orbit "$RUNTIME_REPO" branch inspect harness-template/workspace --json || fail "orbit branch inspect failed for harness template branch: $LAST_STDERR"
  assert_contains "$LAST_STDOUT" '"kind": "template"' "harness template branch kind"
  assert_contains "$LAST_STDOUT" '"template_kind": "harness"' "harness template branch template kind"
}

main() {
  capture_host_go_cache_defaults
  setup_git_environment
  build_binaries
  init_template_source_repo
  verify_source_branch_taxonomy
  init_runtime_repo
  install_template_into_runtime
  verify_runtime_template_writeback
  verify_migrated_runtime_writeback
  verify_runtime_projection_flow
  verify_harness_template_branch
  info "quickstart acceptance passed"
}

main "$@"
