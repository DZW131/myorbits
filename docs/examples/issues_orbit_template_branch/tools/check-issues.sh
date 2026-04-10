#!/usr/bin/env bash

set -euo pipefail

if [ "$#" -gt 0 ]; then
  files=("$@")
else
  files=(docs/issues/open/*.md docs/issues/closed/*.md)
fi

checked=0
failed=0

for path in "${files[@]}"; do
  [ -f "$path" ] || continue
  [ "$(basename "$path")" = "README.md" ] && continue

  checked=$((checked + 1))
  missing=()

  grep -q '^Status:' "$path" || missing+=("Status")
  grep -q '^Outcome:' "$path" || missing+=("Outcome")
  grep -q '^## Evidence$' "$path" || missing+=("Evidence")

  if grep -q '^Status: closed$' "$path"; then
    grep -Eq '^Outcome: (success|failed)$' "$path" || missing+=("closed Outcome")
    grep -q '^## Closure Note$' "$path" || missing+=("Closure Note")
  fi

  if [ "${#missing[@]}" -eq 0 ]; then
    echo "ready: $path"
  else
    echo "not_ready: $path missing ${missing[*]}"
    failed=1
  fi
done

if [ "$checked" -eq 0 ]; then
  echo "no_issue_files_checked"
fi

exit "$failed"
