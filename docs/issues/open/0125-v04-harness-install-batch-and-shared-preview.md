# ISSUE-0125 v0.4 Harness Install Batch And Shared Preview

- Status: open
- Priority: medium
- Owner:
- Created: 2026-04-09
- Updated: 2026-04-09

## Summary

为 harness 作者提供基于同一份共享 bindings 的批量 install / preview 路径，减少逐个 `harness install` 的重复操作，并让“多 orbit runtime 组合”更自然地呈现为一次 harness 级动作。

## Scope

- 设计 `harness install batch <template...>` 或等价命令
- 基于同一份 `.harness/vars.yaml` 对多个 orbit template 做统一 preview
- 汇总每个安装项的写入路径、冲突、警告和 overwrite 需求
- 在 preview 通过后执行批量安装
- 保持现有单个 `harness install` 能力不退化

## Done When

- harness 作者可以对多个模板做一次 shared preview
- preview 输出能清晰区分每个模板的结果与整体 runtime 影响
- 实际安装不会绕过当前 conflict / overwrite 约束
- 批量路径只是现有 install primitive 的高意图封装，而不是第二套安装引擎

## Notes

- 当前能力已经支持：
  - 单 template `--dry-run`
  - bindings file / repo vars / interactive / editor merge
- 这张 issue 的目标是把多次单装组合成更自然的 harness 作者入口
- 依赖 ISSUE-0123，建议与 ISSUE-0124 配套推进
