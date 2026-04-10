# ISSUE-0119 v0.4 Brief Lane Revision Gating And Target Resolution

- Status: closed
- Priority: high
- Owner:
- Created: 2026-04-07
- Updated: 2026-04-08

## Summary

为 `orbit brief materialize` / `orbit brief backfill` 建立统一的 revision gating 与 orbit targeting 合同，确保它们只在 `runtime / source / orbit_template` 三态下运行，并在 current orbit 与 `--orbit <id>` 两种目标解析路径上给出稳定、fail-closed 的行为。

## Scope

- 为 brief lane 命令接入当前 revision kind 判断
- 允许 `runtime / source / orbit_template`
- 拒绝 `plain / harness_template`
- 统一 current orbit 默认解析与 `--orbit <id>` 覆盖语义
- 为允许态与拒绝态定义稳定 text / json 诊断
- 补齐 revision matrix integration tests

## Done When

- `materialize` / `backfill` 在三种允许态下都能稳定解析目标 orbit
- `plain` 与 `harness_template` 会被 fail-closed 拒绝
- current orbit 缺失、无效 orbit id、非法 revision kind 都有稳定错误输出
- revision gating 与 target resolution 的测试覆盖完整

## Notes

- 对应 `docs/orbit_brief_lane_v0_4_technical_spec.md` 的第 3、6、7 节
- 对应 `docs/orbit_v0_4_development_plan.md` 的 Phase 4
- 依赖 ISSUE-0109

## Resolution

- brief lane 命令已接入 revision gating：允许 `runtime / source / orbit_template`，拒绝 `plain / harness_template`。
- `current orbit` 与 `--orbit <id>` 的目标解析已统一进 brief lane 命令实现与测试。
- text / json 诊断及 revision-matrix 集成测试已覆盖 materialize/backfill 的允许态与拒绝态。
