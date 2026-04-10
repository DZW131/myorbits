# ISSUE-0100 Orbit State Ledger Models And FSStore Layout

- Status: open
- Priority: high
- Owner:
- Created: 2026-04-05
- Updated: 2026-04-05

## Summary

为 Phase 4 引入新的 orbit-local ledger 文件模型和 `FSStore` 布局，让 `.git/orbit/state/orbits/<orbit-id>/` 下的 `file_inventory.json`、`runtime_state.json`、`git_state.json` 具备稳定的 typed read/write API，但暂不切换命令写入链路。

## Scope

- 在 `cmd/orbit/cli/state` 中新增：
  - `FileInventorySnapshot`
  - `RuntimeStateSnapshot`
  - `GitStateSnapshot`
- 为 `FSStore` 增加稳定路径布局：
  - `orbits/<orbit-id>/file_inventory.json`
  - `orbits/<orbit-id>/runtime_state.json`
  - `orbits/<orbit-id>/git_state.json`
- 增加读写 API、缺失文件错误、原子写保障。
- 保持现有：
  - `current_orbit.json`
  - `resolved_scope/*.txt`
  - `warnings.json`
  - `last_status.json`
- 不在本 issue 内接入 `enter/status/commit/restore` 的实际写入调用。

## Done When

- 新 ledger 文件有稳定的数据结构和 `FSStore` 访问 API。
- 读写测试、缺失文件测试、原子写保护测试通过。
- 与现有 state 文件并存，不破坏旧路径。

## Notes

- 对应 `docs/orbit_member_runtime_development_plan.md` 的 Phase 4 起始任务。
- 为 ISSUE-0101 与 ISSUE-0102 提供基础设施。
