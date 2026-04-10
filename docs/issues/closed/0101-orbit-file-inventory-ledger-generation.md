# ISSUE-0101 Orbit File Inventory Ledger Generation

- Status: open
- Priority: high
- Owner:
- Created: 2026-04-05
- Updated: 2026-04-05

## Summary

基于 Phase 3 的 `ProjectionPlan` 和 role-aware path classification，生成并落盘 `file_inventory.json`，为观察面和调试链路提供每个 path 的稳定角色与 scope flags 说明。

## Scope

- 基于 `OrbitSpec + ProjectionPlan` 生成 `file_inventory.json` 内容：
  - `path`
  - `member_key`
  - `role`
  - `projection`
  - `orbit_write`
  - `export`
  - `orchestration`
- 覆盖 tracked 路径与 meta companion path。
- member-schema 和 legacy schema 都要有稳定行为。
- 补单测与必要的集成测试，证明：
  - `subject` 默认不进入 `orbit_write/export`
  - `rule` 默认进入 `orbit_write/export`
  - `process` 默认只进入 `projection/orchestration`
- 不在本 issue 内落盘 runtime/git state。

## Done When

- `file_inventory.json` 可以稳定表达当前 orbit 的完整文件说明。
- member-schema 与 legacy schema 都通过测试。
- 生成逻辑可被后续命令和调试观察面复用。

## Notes

- 依赖 ISSUE-0100。
- 对应 `docs/orbit_member_runtime_technical_spec.md` 的 `7.3 file_inventory.json`。
