# ISSUE-0055 Phase 3C Harness Check Schema And Membership

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-26
- Updated: 2026-03-26

## Summary

实现 `harness check` 的基础诊断层，先覆盖 runtime schema、member/definition/install 一致性，以及 zero-member runtime 的成功返回路径。

## Scope

- `harness check` 校验：
  - `.harness/runtime.yaml` schema 合法性
  - duplicate member
  - member 指向的 orbit definition 缺失
  - install record 与 members 不一致
  - install record 路径 / `orbit_id` 不一致
- zero-member runtime：
  - 只要 schema 合法就成功返回
  - 不额外制造 warning
- 输出为后续 drift 诊断保留可扩展结构

## Done When

- `harness check` 在 zero-member runtime 下成功返回。
- 缺失 definition / member-install mismatch / install path mismatch 有明确诊断。
- 有 temp-repo 集成测试覆盖 zero-member runtime 和至少两类 membership 失败场景。

## Notes

- 对应 `docs/harness_centric_runtime_development_plan.md` 的 10.3 / 任务 3。
- 相关技术合同见 `docs/harness_centric_runtime_technical_spec.md` 的 8.6。
- Completed:
  - `harness check` 已作为正式命令接入 `harness` root CLI。
  - zero-member runtime 现在返回 `ok=true`、`finding_count=0`。
  - runtime schema invalid、missing definition、install-member mismatch、install path mismatch 都已有稳定 JSON 诊断和集成测试覆盖。
