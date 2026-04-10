# ISSUE-0051 Phase 3B-2 Drift Replay Primitives

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-26
- Updated: 2026-03-26

## Summary

把 overwrite 路径中可复用的 source replay 与期望输出重建逻辑下沉成共享原语，为后续 `harness check` 的 drift 诊断提供稳定输入。

## Scope

- 抽取 install source replay primitive：
  - 按 install record 重新解析模板 source
  - 在当前 runtime repo 上重建期望 definition / rendered files / shared `AGENTS.md` 输出
- 抽取 drift comparison primitive，服务后续：
  - `definition_drift`
  - `runtime_file_drift`
  - `provenance_unresolvable`
- 让 overwrite 路径直接复用这些原语，而不是复制第二套逻辑

## Done When

- overwrite path 与后续 check path 共用一套 source replay / expected-output reconstruction primitive。
- drift replay primitive 有 focused unit tests。
- 无法安全 replay install source 时能稳定产出 `provenance_unresolvable` 所需的错误边界。

## Notes

- 对应 `docs/harness_centric_runtime_development_plan.md` 的 9.3 / 任务 3。
- 相关技术合同见 `docs/harness_centric_runtime_technical_spec.md` 的 9.4。
- Completed:
  - 已新增 install replay primitive，按 install record 的 `template_commit` 重建期望 definition / rendered files / shared `AGENTS.md` payload。
  - 已新增 drift comparison primitive，可稳定产出 `definition_drift`、`runtime_file_drift`、`provenance_unresolvable`。
  - 单测已覆盖 replay 使用 recorded commit、definition/runtime drift、以及 provenance 不可解析三条路径。
