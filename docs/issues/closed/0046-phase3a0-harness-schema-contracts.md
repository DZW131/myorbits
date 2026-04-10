# ISSUE-0046 Phase 3A-0 Harness Runtime And Template Schema Contracts

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-25
- Updated: 2026-03-26

## Summary

实现并冻结 `Phase 3A-0` 所需的 `.harness/runtime.yaml` 与 `.harness/template.yaml` schema contracts，包括 YAML codec、validation、稳定写回与 zero-member runtime 合同。

## Scope

- 定义 `.harness/runtime.yaml` 的 Go model、YAML codec 和 validation。
- 定义 `.harness/template.yaml` 的 Go model、YAML codec 和 validation。
- 明确 `schema_version`、`kind`、必填字段、枚举值和稳定排序行为。
- 实现 zero-member runtime 的合法读写行为。
- 补齐 `created_at` / `updated_at` 等稳定时间字段的写回策略。
- 为 schema 的合法样例、非法样例和 zero-member runtime 增加单测。

## Done When

- `.harness/runtime.yaml` 与 `.harness/template.yaml` 都能稳定完成读写和校验。
- 缺少必要字段、字段类型错误或非法枚举时会 fail-closed。
- zero-member runtime 被视为合法 schema，而不是异常输入。
- 单测锁定合法样例、非法样例、稳定排序与 zero-member 约束。

## Notes

- 对应 `docs/harness_centric_runtime_development_plan.md` 的 6.3 / 任务 2。
- 主要依据：`docs/harness_centric_runtime_technical_spec.md` 的 5.2、5.6、5.7。
- Completed:
  - 已实现 `.harness/runtime.yaml` 与 `.harness/template.yaml` 的 model、YAML codec、validation 与稳定写回。
  - zero-member runtime 已作为合法 schema 输入固定下来。
  - 已补齐合法样例、非法样例、稳定排序与 round-trip 单测。
