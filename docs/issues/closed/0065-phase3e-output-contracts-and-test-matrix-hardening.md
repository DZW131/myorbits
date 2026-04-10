# ISSUE-0065 Phase 3E Output Contracts And Test Matrix Hardening

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-26
- Updated: 2026-03-26

## Summary

补齐 `orbit` / `harness` 的 JSON 与 text 输出契约测试，统一对照 v0.3 的最低覆盖矩阵做 hardening，避免在发布收尾阶段留下未冻结的输出面。

## Scope

- 复查并补齐以下命令的 JSON / text 契约测试：
  - `harness inspect`
  - `harness check`
  - `harness install`
  - `harness template save`
  - `orbit branch status`
  - `orbit branch list`
  - `orbit branch inspect`
- 对照 `docs/testing-strategy.md` 与 v0.3 development plan 补最低覆盖矩阵。
- 为关键输出面补 golden tests 或等价稳定断言。
- 以 `mise run fmt`、`mise run lint`、`mise run test:ci` 作为阶段收口基线。

## Done When

- 关键 JSON / text 输出面已有稳定回归测试。
- `Phase 3A` 到 `Phase 3D` 的新命令面不再存在明显未覆盖的契约缺口。
- 阶段基线校验可以稳定跑通。

## Notes

- 对应 `docs/harness_centric_runtime_development_plan.md` 的 12.3 / 任务 3。
- 相关技术合同见 `docs/harness_centric_runtime_technical_spec.md` 的 9.2、10.1、10.2、10.3。
- Completed:
  - `harness check` 现在有稳定的 text output 契约测试，zero-member runtime 的文本输出已冻结。
  - `harness install` 补了正式 text output 契约测试，覆盖 success text 输出的关键字段。
  - `harness template save` 补了正式 text output 契约测试，覆盖 member count 与 `includes_root_agents`。
  - `orbit branch inspect/status/list` 的 golden / JSON 契约测试与已有 harness-aware 集成测试一并构成当前 v0.3 输出基线。
  - 阶段基线 `mise run fmt`、`mise run lint`、`mise run test:ci` 已通过。
