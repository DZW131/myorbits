# ISSUE-0094 Orbit Member Object Model And Projection Plan

- Status: closed
- Priority: high
- Owner:
- Created: 2026-04-04
- Updated: 2026-04-04

## Summary

把 Orbit 从当前偏 path-list 的内部心智，先收口成稳定的成员化对象模型，在代码里引入 `OrbitMemberRole`、`OrbitMember`、`ProjectionPlan` 等核心结构，但暂不改变 CLI 对外行为。

## Scope

- 在内部模型中引入固定角色：
  - `meta`
  - `subject`
  - `rule`
  - `process`
- 引入最小可用的 `ProjectionPlan`，至少能表达：
  - role 归属
  - `ProjectionPaths`
  - `OrbitWritePaths`
  - `ExportPaths`
  - `OrchestrationPaths`
- 保持现有命令外壳不变，不在本 issue 内切换 `orbit diff/log/commit/restore` 的最终消费面。
- 为新对象模型补基础单测与 YAML codec 单测。

## Done When

- 代码内已有稳定的成员角色枚举和成员结构体。
- 代码内已有可被后续解析层消费的 `ProjectionPlan`。
- 未使用新 schema 字段时，现有解析与命令行为不回归。
- 基础对象模型和 codec 测试已覆盖。

## Notes

- 对应 `docs/orbit_member_runtime_development_plan.md` 的 Phase 1。
- 这是后续 schema 兼容和 role-aware classification 的前置项。
- Completed:
  - 新增 `OrbitSpec` / `OrbitMeta` / `OrbitMemberRole` / `OrbitMember` / `ProjectionPlan` 等 Phase 1 内部对象模型。
  - 新增 `ParseOrbitSpecData`、`OrbitSpecFromDefinition`、`LegacyDefinition` 等兼容辅助，但没有切换现有 `Definition` 命令消费面。
  - 新增角色枚举、ProjectionPlan role 访问、member-model YAML codec、legacy parser 回归测试。
  - `go test ./cmd/orbit/cli/orbit` 与 `go test ./cmd/orbit/cli/...` 通过，`mise run lint` 通过。
