# ISSUE-0122 v0.4 Brief Lane Drift Diagnostics

- Status: closed
- Priority: medium
- Owner:
- Created: 2026-04-07
- Updated: 2026-04-08

## Summary

为 brief lane 增加轻量 drift diagnostics，让用户在 `materialize / backfill / publish` 之前就能看清当前 orbit brief 处于 `structured_only`、`materialized_in_sync`、`materialized_drifted`、`invalid_container` 还是 `missing_truth`，避免在多状态模型下继续靠猜。

## Scope

- 固化 brief lane 的五种状态模型
- 提供最小可用的 `--check` 或 status/check 诊断出口
- 为 text / json 输出定义稳定字段
- 明确哪些状态允许 `materialize`，哪些状态允许 `backfill`
- 补齐状态矩阵与诊断测试

## Done When

- 用户可在命令前稳定得知当前 brief lane 状态
- 诊断结果能区分 in-sync、drifted、invalid 与 missing truth
- materialize/backfill 的允许条件与 diagnostics 保持一致
- 状态矩阵有完整单测或 integration tests

## Notes

- 对应 `docs/orbit_brief_lane_v0_4_technical_spec.md` 的第 5、7 节
- 对应 `docs/orbit_v0_4_development_plan.md` 的 Phase 4
- 依赖 ISSUE-0120、ISSUE-0121

## Resolution

- brief lane 的五种状态模型已经在代码中固化：`structured_only`、`materialized_in_sync`、`materialized_drifted`、`invalid_container`、`missing_truth`。
- `orbit brief materialize --check` 与 `orbit brief backfill --check` 都已提供稳定的 text/json 诊断出口。
- diagnostics 与 materialize/backfill 的允许条件已在实现和测试中保持一致，drift/status 矩阵已有专门集成测试。
