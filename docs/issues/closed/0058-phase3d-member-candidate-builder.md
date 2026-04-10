# ISSUE-0058 Phase 3D Member Candidate Builder

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-26
- Updated: 2026-03-26

## Summary

实现 harness template save 的 member candidate builder，针对每个 runtime member 从当前 runtime repo 构建可合并的 template candidate。

## Scope

- 对每个 member：
  - 读取 `.orbit/orbits/<orbit-id>.yaml`
  - 用现有 scope resolver 得到 user scope
  - 从 runtime repo 读取当前文件内容
  - 用 `.harness/vars.yaml` 做 replacement
  - 自动把 `.orbit/orbits/<orbit-id>.yaml` 注入 candidate
- builder 输出应与后续 merge engine 对接，不直接写 branch。
- 缺失 definition、无法解析 scope、无法读取 candidate file 时 fail-closed。

## Done When

- 能从 manual member / install-backed member 构建稳定 candidate。
- candidate builder 单测覆盖正常路径、definition 缺失、vars replacement、生效文件集。
- 构建结果可直接供后续 merge engine 使用。

## Notes

- 对应 `docs/harness_centric_runtime_development_plan.md` 的 11.3 / 任务 1。
- 相关技术合同见 `docs/harness_centric_runtime_technical_spec.md` 的 8.7。
- Completed:
  - `BuildTemplateMemberCandidate` 已落地，复用现有 repo config、scope resolver、content builder、replacement 原语。
  - manual member 与 install-backed member 现在都能构建稳定 candidate。
  - 单测已覆盖 definition 缺失 fail-closed、vars replacement、生效文件集和 companion definition 注入。
