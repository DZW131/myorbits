#!/bin/sh

set -eu

script_dir=$(CDPATH= cd -- "$(dirname "$0")" && pwd)
repo_root=$(CDPATH= cd -- "$script_dir/.." && pwd)

assert_contains() {
  file=$1
  expected=$2

  if ! grep -Fq "$expected" "$file"; then
    echo "expected $file to contain: $expected" >&2
    cat "$file" >&2
    exit 1
  fi
}

assert_not_contains() {
  file=$1
  unexpected=$2

  if grep -Fq "$unexpected" "$file"; then
    echo "expected $file to not contain: $unexpected" >&2
    cat "$file" >&2
    exit 1
  fi
}

quickstart_doc="$repo_root/docs/quickstart.md"
release_doc="$repo_root/docs/release.md"
testing_doc="$repo_root/docs/testing-strategy.md"

assert_not_contains "$quickstart_doc" "这份文档给 v0.3 的双二进制主路径一个最短可执行示例"
assert_contains "$quickstart_doc" '这里的 `HEAD` 应被识别为 `kind=source`。'

assert_not_contains "$release_doc" "git tag -a v0.3.0 -m \"v0.3.0\""
assert_contains "$release_doc" "git tag -a v0.4.0 -m \"v0.4.0\""

assert_not_contains "$testing_doc" "当前 v0.3 / harness-centric 主线在此基础上再加一层单控制面约束："
assert_not_contains "$testing_doc" "作用：把发布中的 v0.3 quickstart 主路径固化成一条 doc-derived smoke，覆盖双二进制与三类 branch identity。"
assert_contains "$testing_doc" "作用：把当前 v0.4 quickstart 主路径固化成一条 doc-derived smoke，覆盖双二进制与四类 branch identity。"

echo "active v0.4 docs tests passed"
