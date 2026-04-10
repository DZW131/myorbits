# ISSUE-0107 Harness Single Control Plane Docs And Acceptance Hardening

- Status: closed
- Priority: medium
- Owner:
- Created: 2026-04-05
- Updated: 2026-04-06

## Summary

在单目录控制面主链路稳定后，统一收口 help、quickstart、testing docs 与 acceptance matrix，并评估是否需要显式迁移工具。

## Scope

- 更新：
  - `docs/quickstart.md`
  - `docs/testing-strategy.md`
  - help / examples / release 文档
- 删除或替换旧 `.orbit/*` 控制面示例。
- 增加三类 branch 的 acceptance matrix：
  - runtime
  - orbit_template
  - harness_template
- 评估是否需要显式迁移工具：
  - `harness migrate-control-plane`

## Done When

- 文档与实现口径一致。
- 新控制面有覆盖三类 branch 的 acceptance smoke。
- 若决定不做迁移工具，也有明确记录说明。

## Notes

- 对应 `docs/context/harness_single_control_plane_development_plan.md` 的 Phase 5。
- 依赖 ISSUE-0106。
- Completed:
  - `docs/quickstart.md` 已切到单控制面主路径，删除 runtime 顶层 `.orbit/*` 示例，并明确三类 branch inspect、runtime first commit、以及无独立 `harness migrate-control-plane` 的当前结论。
  - `docs/testing-strategy.md` 已补 v0.3 单控制面测试目标、runtime/orbit_template/harness_template acceptance matrix、以及 `mise run acceptance:quickstart` 的文档驱动 smoke 规则。
  - CLI help/examples 已补齐 `harness init` 与 `orbit init` legacy compatibility 说明。
  - quickstart acceptance smoke 已实际接入并通过，覆盖 runtime、orbit_template、harness_template 三类 branch。
