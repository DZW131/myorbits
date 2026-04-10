# ISSUE-0087 Harness Mixed Install Batch

- Status: closed
- Priority: high
- Owner:
- Created: 2026-04-01
- Updated: 2026-04-01

## Summary

推进 `harness install` 的 mixed install 增强：支持 harness template install unit，引入 disjoint mixed install，并为后续 same install unit replace 打基础。

## Scope

- 协调本批次的执行子项：
  - harness template install source 识别与 preview 基线
  - bundle-level provenance / runtime member source 扩展
  - mixed disjoint conflict set
  - harness template `AGENTS.md` bundle block lane
  - `harness template save` runtime marker normalize / strip
  - same install unit replace 的第二阶段设计与实现
- 当前批次不处理 source repo 直装。
- 当前批次不开放危险冲突覆盖。

## Done When

- 本批次子 issue 已完成或明确推迟。
- `harness install` 至少能安全处理 orbit template 与 harness template 的 disjoint mixed install。
- 第二阶段 same install unit replace 的依赖已清晰冻结。

## Notes

- 对应 `docs/harness_mixed_install_technical_spec.md`。
- source repo alias 相关工作由 `0084`-`0086` 单独处理，不纳入本批次。
