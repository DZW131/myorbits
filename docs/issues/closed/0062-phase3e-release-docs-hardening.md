# ISSUE-0062 Phase 3E Release Docs Hardening

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-26
- Updated: 2026-03-26

## Summary

推进 v0.3 开发计划的 `Phase 3E`，完成文档、帮助、发布面、测试矩阵与遗留入口的统一收口，让 `orbit` / `harness` 双前缀模型在对外口径和工程交付上完全一致。

## Scope

- 协调本阶段的 4 个执行子项：
  - CLI help / README / quickstart 对齐
  - dual-binary build / release / completion 收口
  - JSON / text 输出契约与测试矩阵 hardening
  - 遗留入口与旧 runtime host 叙事收口
- 明确本阶段不再新增新的 runtime 功能，只做发布与对外一致性收尾。
- 以 `docs/harness_centric_runtime_development_plan.md` 的 `Phase 3E` 作为验收基线。

## Done When

- README / quickstart / help / PRD / technical spec / development plan 口径一致。
- 双二进制 build / release 入口稳定。
- JSON / text 输出契约和回归矩阵补齐。
- 遗留入口不再在正式帮助和文档中暴露旧 runtime host 概念。

## Notes

- 对应 `docs/harness_centric_runtime_development_plan.md` 的 12。
- 相关技术合同见 `docs/harness_centric_runtime_technical_spec.md` 的 9.2、10.2、10.3。
- Completed:
  - `ISSUE-0063` 已完成：README、quickstart、`orbit --help`、`harness --help`、`harness install --help` 已统一到 v0.3 双前缀主路径。
  - `ISSUE-0064` 已完成：新增 dual-binary build script、completion smoke tests 与 `mise run build` 发布入口。
  - `ISSUE-0065` 已完成：`harness inspect/check/install/template save` 与 branch outputs 的 text / JSON 契约已经补齐关键回归。
  - `ISSUE-0066` 已完成：legacy `orbit template apply` wrapper 与旧 runtime host 叙事已收口到最小兼容面。
  - 阶段基线 `mise run fmt`、`mise run lint`、`mise run test:ci` 已通过。
