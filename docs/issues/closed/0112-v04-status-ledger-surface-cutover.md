# ISSUE-0112 v0.4 Status Ledger Surface Cutover

- Status: closed
- Priority: high
- Owner:
- Created: 2026-04-07
- Updated: 2026-04-09

## Summary

让 `orbit status`、path classification 与 `.git/orbit/state/*` ledger 正式切到 four-surface 语义，避免 projection、orbit_write、export、orchestration 在观察面上继续混写。

## Scope

- 升级 `path_classification.go` 与 status output，显式展示 role / surface 信息
- 更新 tracked / untracked 分类逻辑
- 更新 file inventory、runtime state、git state ledger 的 surface-aware 语义
- 确保 `orbit diff/log/commit/restore` 只消费 `orbit_write`
- 保持普通 Git 行为不被 Orbit scoped 命令面重定义

## Done When

- `orbit status` 可清楚区分 projection-visible 与 orbit_write/export/orchestration 归属
- `.git/orbit/state/*` ledger 能稳定表达 surface-aware state
- `diff/log/commit/restore` 的 regression tests 全部切到 `orbit_write` surface

## Notes

- 对应 `docs/orbit_v0_4_technical_spec.md` 的第 7、8 节
- 对应 `docs/orbit_v0_4_development_plan.md` 的 Phase 3
- 依赖 ISSUE-0111

## Progress

- `orbit status`、path classification、file inventory ledger 与 scoped command regression 已切到 four-surface 语义。
- `git_state.json` 现在已补齐 `orbit_export_state` 与 `orbit_orchestration_state`，不再只保留 projection / commit 两类 orbit-local bucket。
- `diff/log/commit/restore` 的 member-schema regression 已明确使用 `orbit_write` surface，避免 process / subject path 再误入 scoped write 行为。
- 当 active runtime ledger 的 `plan_hash` 与当前 revision 重算结果不匹配时，`status` 与 scoped read/write 命令现在会 fail-closed，并提示重新 `orbit enter <id>` 刷新投影。
- `orbit current` 现在也能识别上述 stale ledger 场景，并在 `--json` 下输出稳定的 `stale_reason`，不再只有 scoped 命令失败时才能感知该状态。

## Resolution

- `orbit status`、path classification、file inventory ledger、git ledger、`diff/log/commit/restore` 已统一切到 four-surface 语义，`projection / orbit_write / export / orchestration` 不再混写。
- `git_state.json` 已补齐 `orbit_export_state` 与 `orbit_orchestration_state`，orbit-local ledger 现在能表达完整四面 bucket，而不是只保留 projection / commit。
- active runtime ledger 的 `plan_hash` 现已成为正式一致性门槛：`status` 与 scoped read/write 命令在 plan 过期时会 fail-closed，并提示重新 `orbit enter <id>`。
- `orbit current` 也已补齐只读 stale diagnostics，并在 `--json` 下提供稳定的 `stale_reason`。
- `files` 等其他观察命令若后续需要进一步统一 stale classification，应另开 follow-up issue；这已不再阻塞 v0.4 的 status/ledger surface cutover 主收口。
