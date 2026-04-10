# ISSUE-0050 Phase 3B-2 Owned File Reconstruction

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-26
- Updated: 2026-03-26

## Summary

为 overwrite 路径实现 old owned file reconstruction，确保系统只在能够安全重建旧模板输出时删除已不再属于新模板的 runtime 文件。

## Scope

- 基于现有 install record 重新解析旧模板快照
- 重建旧 install-owned file set，包括：
  - materialized runtime files
  - shared `AGENTS.md` payload 的 install-owned block
  - 当前 `.orbit/orbits/<orbit-id>.yaml` 期望 definition
- 比较旧 owned file set 与新渲染结果
- 只在安全确认后删除“旧模板拥有、但新模板已不再拥有”的路径
- 旧来源不可安全解析、不可读取或无法稳定重建时 fail-closed

## Done When

- overwrite 路径会先重建旧 owned file set，再执行删除或替换。
- 无法安全重建旧 owned file set 时明确阻断 overwrite。
- 有单测覆盖 owned file reconstruction 的正常路径和 fail-closed 路径。

## Notes

- 对应 `docs/harness_centric_runtime_development_plan.md` 的 9.3 / 任务 2。
- 相关技术合同见 `docs/harness_centric_runtime_technical_spec.md` 的 8.2、9.4。
- Completed:
  - overwrite 前会基于 install record replay 旧模板输出，构建 stale install-owned cleanup plan。
  - 普通 runtime files 和 shared `AGENTS.md` block 都已接入安全确认后删除。
  - 旧 source pin 无法 replay 时会 fail-closed，阻断 overwrite。
