# ISSUE-0048 Phase 3B-2 Overwrite Reinstall Drift Foundations

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-26
- Updated: 2026-03-26

## Summary

推进 v0.3 开发计划的 `Phase 3B-2`，把 `harness install` 从 basic path 收紧到可安全覆盖更新的 install 主路径，并为后续 `harness check` 提供 drift replay 基础原语。

## Scope

- 协调本阶段的 3 个执行子项：
  - overwrite / reinstall contract
  - old owned file reconstruction
  - drift replay primitives
- 明确本阶段只交付 install overwrite foundations，不提前进入完整 `harness check` 或 branch/check UI 收口。
- 以 `docs/harness_centric_runtime_development_plan.md` 的 `Phase 3B-2` 作为验收基线。

## Done When

- `harness install --overwrite-existing` 稳定可用。
- remove 后同一 `orbit-id` 的 reinstall 明确走 overwrite 路径。
- old owned file reconstruction 与 drift replay primitives 有单测和集成测试保护。

## Notes

- 对应 `docs/harness_centric_runtime_development_plan.md` 的 `Phase 3B-2`。
- 相关技术合同见 `docs/harness_centric_runtime_technical_spec.md` 的 8.2、8.4、9.4。
- 建议实现分支：`feature/v0.3-install-overwrite-drift-foundations`。
- Completed:
  - `harness install --overwrite-existing` 已落地，覆盖 install-backed member 和 remove 后 reinstall 两条路径。
  - old owned file reconstruction 已支持 stale runtime files 与 shared `AGENTS.md` block 的安全删除计划。
  - replay / drift primitive 已补齐，并由 overwrite 路径直接复用。
  - `mise run fmt`、`mise run lint`、`mise run test:ci` 已通过。
