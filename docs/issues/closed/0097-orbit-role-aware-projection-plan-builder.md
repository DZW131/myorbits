# ISSUE-0097 Orbit Role-Aware ProjectionPlan Builder

- Status: closed
- Priority: high
- Owner:
- Created: 2026-04-05
- Updated: 2026-04-05

## Summary

基于已经落地的 member schema 与兼容解析层，正式构建 role-aware `ProjectionPlan`，让 Orbit 内核不再只依赖 path-list 风格的 `ScopeSet`，但先保持现有命令外壳可继续消费兼容投影。

## Scope

- 基于 `OrbitSpec` 与 tracked files 生成稳定的 `ProjectionPlan`：
  - `ControlPaths`
  - `MetaPaths`
  - `SubjectPaths`
  - `RulePaths`
  - `ProcessPaths`
  - `ProjectionPaths`
  - `OrbitWritePaths`
  - `ExportPaths`
  - `OrchestrationPaths`
- 支持 legacy schema 的兼容映射：
  - legacy `include/exclude` 自动进入兼容 `ProjectionPlan`
- 提供当前 `ScopeSet` 到 `ProjectionPlan` 的过渡适配，避免一次性重写全部 consumers。
- 不在本 issue 内切换 `status` 输出或 scoped 命令行为。

## Done When

- 代码内可以稳定生成 role-aware `ProjectionPlan`。
- `subject`、`rule`、`process` 在默认 role-scope 语义下能正确分流。
- legacy orbit 仍能生成兼容 plan，不打断现有仓库。
- `ProjectionPlan` builder 的单测已覆盖主要 role / scope 组合。

## Notes

- 对应 `docs/orbit_member_runtime_development_plan.md` 的 Phase 3 第一部分。
- 这是 `status` role-aware classification 和 state ledger 升级的前置项。
- Completed:
  - 新增 `ResolveProjectionPlan`，支持 member schema 的 role-aware path 分流与 legacy schema 的兼容 plan。
  - 新增 `ScopeSetFromProjectionPlan` 与 `ResolveScopeSetForSpec` 作为过渡适配层。
  - 当前 `ResolveScopeSet` 已切到新的 compatibility path，但不改变现有命令消费面。
  - 新增 ProjectionPlan builder 测试矩阵，`go test ./cmd/orbit/cli/...` 与 `mise run lint` 通过。
