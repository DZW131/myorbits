# ISSUE-0091 Harness Install Progress Regression Matrix

- Status: closed
- Priority: medium
- Owner:
- Created: 2026-04-01
- Updated: 2026-04-01

## Summary

用集成测试冻结 `harness install` 的进度输出矩阵，确保新进度能力不打破 text/json 结果契约。

## Scope

- 覆盖以下场景：
  - 本地 orbit template install
  - 远程 orbit template install with `--ref`
  - 远程 orbit template install 自动选择
  - source repo alias install
  - harness template install
  - dry-run
- 覆盖以下输出契约：
  - `stderr` 包含阶段进度
  - `stdout` text 结果保持不变
  - `--json` 的 stdout 仍为纯 JSON
  - `--progress quiet` 静默
- help 测试补充 `--progress`

## Done When

- 相关 CLI 集成测试稳定通过
- 现有 install text/json 合同未被回归破坏
- progress 默认行为与 `plain` / `quiet` 差异有测试锁定

## Notes

- 依赖 `ISSUE-0089` 与 `ISSUE-0090`
- 若 `auto` 的 TTY 判定难以在现有 harness 中稳定测试，可优先锁 `plain` / `quiet` 并补最小 `auto` 行为单测
