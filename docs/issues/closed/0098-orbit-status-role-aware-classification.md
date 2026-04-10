# ISSUE-0098 Orbit Status Role-Aware Classification

- Status: open
- Priority: high
- Owner:
- Created: 2026-04-05
- Updated: 2026-04-05

## Summary

把当前 `orbit status` 的 path classification 从单一的 in-scope / out-of-scope 视角升级为 role-aware 结果，让 tracked 与 untracked 路径都能稳定输出 member role 和 scope flags。

## Scope

- 基于 `ProjectionPlan` 扩展 status classification：
  - tracked changes 标记 role 与 scope flags
  - untracked paths 标记 role 与 scope flags
- 保持现有命令面和基础 stdout 合同尽量稳定，只增加新的结构化分类信息。
- 为后续 `show` / `state` / 调试观察面复用 classification 结果。
- 不在本 issue 内切换 commit / restore / template save 的消费面。

## Done When

- `status` 不再只依赖单个 `InScope bool`。
- tracked / untracked 都能稳定归属 `meta / subject / rule / process / outside`。
- 相关单测与集成测试覆盖主要变更场景。

## Notes

- 对应 `docs/orbit_member_runtime_development_plan.md` 的 Phase 3 任务 3-4。
- 依赖 ISSUE-0097 提供稳定 `ProjectionPlan`。
