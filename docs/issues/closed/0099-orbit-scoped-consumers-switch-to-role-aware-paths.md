# ISSUE-0099 Orbit Scoped Consumers Switch To Role-Aware Paths

- Status: open
- Priority: high
- Owner:
- Created: 2026-04-05
- Updated: 2026-04-05

## Summary

在 `ProjectionPlan` 和 role-aware classification 稳定后，把当前 Orbit 命令消费面切换到新的行为路径：`enter/files/status` 继续消费 projection，而 `diff/log/commit/restore` 改消费 `orbit_write`，`template save` 改消费 `export`。

## Scope

- `enter/files/status`
  - 继续消费 `ProjectionPaths`
- `diff/log/commit/restore`
  - 改消费 `OrbitWritePaths`
- `orbit template save`
  - 改消费 `ExportPaths`
- 补回归测试，明确：
  - `subject` 默认不进入 `orbit_write/export`
  - `rule` 默认进入 `orbit_write/export`
  - `process` 默认只进入 projection/orchestration

## Done When

- orbit-local 行为不再把所有 in-scope 文件视作同一种 scoped path。
- `subject` 与 `rule/process` 的行为差异在命令面正式生效。
- 现有 legacy 仓库与 member-schema 仓库都通过回归测试。

## Notes

- 对应 `docs/orbit_member_runtime_development_plan.md` 的 Phase 3 收尾部分。
- 依赖 ISSUE-0097 与 ISSUE-0098。
