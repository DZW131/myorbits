# ISSUE-0056 Phase 3C Harness Check Drift Diagnostics

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-26
- Updated: 2026-03-26

## Summary

把 `Phase 3B-2` 已落地的 replay / drift primitives 接到 `harness check`，提供正式的 drift 诊断输出。

## Scope

- `harness check` 集成 drift 诊断：
  - `definition_drift`
  - `runtime_file_drift`
  - `provenance_unresolvable`
- 针对 install-backed member：
  - 调用共享 replay primitive
  - 调用共享 drift comparison primitive
- 保持 drift 是诊断结果：
  - 不自动修复
  - 不阻断 projection
- 输出应包含可操作的 orbit id / path / drift kind 信息

## Done When

- `harness check` drift 分类与 `Phase 3B-2` 的 primitive 结果一致。
- zero-member runtime 不会误报 drift。
- 有测试覆盖 definition drift、runtime file drift、provenance unresolvable 三类结果。

## Notes

- 对应 `docs/harness_centric_runtime_development_plan.md` 的 10.3 / 任务 3。
- 相关技术合同见 `docs/harness_centric_runtime_technical_spec.md` 的 8.6、9.4。
- Completed:
  - `harness check` 已接入 install replay / drift primitive。
  - `definition_drift`、`runtime_file_drift`、`provenance_unresolvable` 三类结果现在都会出现在正式诊断输出中。
  - zero-member runtime 不会误报 drift，相关回归已在 harness CLI 集成测试中覆盖。
