# ISSUE-0114 v0.4 Orbit Template Direct Publish

- Status: closed
- Priority: high
- Owner:
- Created: 2026-04-07
- Updated: 2026-04-08

## Summary

支持在 `orbit_template` revision 中直接完成校验、brief 收口与 publish，使 direct template authoring 成为 v0.4 的正式一等工作流。

## Scope

- 定义 `orbit_template` revision 下 `orbit template publish` 的正式合同
- 在 publish 前校验 installable payload 不应包含根 `AGENTS.md`
- 若 materialized brief 与结构化 truth 不一致，给出显式 backfill 流程
- 补 direct template authoring 的 help、docs 与 integration tests

## Done When

- 开发者可直接在 `orbit_template` revision 中完成 publish
- publish path 会拦截不合法的根 `AGENTS.md` payload
- direct template authoring 有独立 acceptance smoke

## Notes

- 对应 `docs/orbit_v0_4_prd.md` 的第 6、7 节
- 对应 `docs/orbit_v0_4_development_plan.md` 的 Phase 5
- 依赖 ISSUE-0113 与 ISSUE-0122

## Resolution

- `orbit template publish` 已支持在 `orbit_template` revision 中直接执行校验型 publish。
- direct template publish 已拦截根 `AGENTS.md` payload，并能区分普通 legacy payload 与 drifted materialized brief。
- publish diagnostics 已优先跟随 hosted brief truth，并在 drift 时明确引导 `orbit brief backfill --orbit <id>`。
- direct template publish 的集成测试已覆盖 validate/no-op/push/default-template mismatch 等主链路行为。
