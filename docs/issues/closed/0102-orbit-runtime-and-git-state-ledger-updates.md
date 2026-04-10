# ISSUE-0102 Orbit Runtime And Git State Ledger Updates

- Status: closed
- Priority: high
- Owner:
- Created: 2026-04-05
- Updated: 2026-04-05

## Summary

在新的 ledger 文件模型稳定后，把 `runtime_state.json` 和 `git_state.json` 接到当前 Orbit 运行链路，让 `enter/leave/status/commit/restore` 能持续落盘 orbit-local runtime 与 Git 观察信息。

## Scope

- 定义并写入 `runtime_state.json`：
  - `orbit`
  - `running`
  - `phase`
  - `entered_at`
  - `updated_at`
  - `plan_hash`
- 定义并写入 `git_state.json` 的最小观察面：
  - orbit projection state
  - orbit stage state
  - orbit commit state
  - global stage state
  - global commit state
- 在合适的命令链路更新 ledger：
  - `enter`
  - `leave`
  - `status`
  - `commit`
  - `restore`
- 保持旧 state 文件继续可读写，不删除兼容文件。

## Done When

- `runtime_state.json` 与 `git_state.json` 能随 Orbit 运行稳定更新。
- 缺失/损坏文件恢复测试通过。
- 不破坏当前命令输出和旧 state 文件读取链路。

## Notes

- 依赖 ISSUE-0100，建议在 ISSUE-0101 之后接入。
- 对应 `docs/orbit_member_runtime_development_plan.md` 的 Phase 4 收尾部分。
