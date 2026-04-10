# ISSUE-0059 Phase 3D Candidate Merge And Conflict Analysis

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-26
- Updated: 2026-03-26

## Summary

实现 harness template save 的候选合并与冲突分析层，负责多个 member candidate 的文件集合、变量声明和 manifest-level 元数据合并。

## Scope

- 合并 candidate files
- 合并变量声明
- 路径冲突规则：
  - 相同内容允许
  - 不同内容 fail-closed
- 变量冲突规则：
  - 描述一致允许
  - 一个空一个非空取非空
  - 两个非空且不同 fail-closed
  - `required` 按 OR 合并
- 产出：
  - 合并后的 template files
  - 合并后的 variables
  - `.harness/template.yaml` 所需 metadata

## Done When

- candidate merge 单测覆盖正常合并、path collision、variable collision。
- 合并结果稳定排序，适合 command 输出和集成测试。
- merge engine 可供 `harness template save` 直接复用。

## Notes

- 对应 `docs/harness_centric_runtime_development_plan.md` 的 11.3 / 任务 2。
- 相关技术合同见 `docs/harness_centric_runtime_technical_spec.md` 的 8.7。
- Completed:
  - `MergeTemplateMemberCandidates` 已落地，负责稳定合并 member candidates 的 files、variables、members。
  - 相同路径相同内容现在允许合并；相同路径不同内容 fail-closed。
  - 变量描述冲突、描述空值收敛、`required` OR 合并都已有单测覆盖。
