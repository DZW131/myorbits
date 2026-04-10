# ISSUE-0005 Implement Bindings Merge Engine

- Status: closed
- Priority: medium
- Owner:
- Created: 2026-03-21
- Updated: 2026-03-21

## Summary

实现 bindings merge engine，把模板 manifest 变量声明、`--bindings` 文件、当前仓库 `.orbit/vars.yaml` 以及后续交互补填结果合成为一份 resolved bindings，并输出缺失项和来源摘要。

## Scope

- 接收模板变量声明作为输入基线。
- 按 `--bindings` > `.orbit/vars.yaml` > 交互/编辑器补填 的优先级合并值。
- 区分 resolved 和 unresolved 变量。
- 为每个变量记录来源摘要，便于 dry-run、提示和调试。
- 为缺失变量、覆盖优先级和 description 保留增加测试。

## Done When

- 合并结果能稳定输出 resolved bindings。
- unresolved 集合完整且顺序稳定。
- 高优先级来源会覆盖低优先级来源。
- 未声明变量不会被悄悄混入结果。
- 单测覆盖优先级、缺失项和来源摘要。

## Notes

- 对应 `docs/context/orbit_phase2_development_plan.md` 的 Phase 2A-0 / 任务 3。
- 主要依据：`docs/context/orbit_phase2_technical_spec.md` 的 8.4、10.3。
- 2026-03-21: 已实现 `bindings.Merge`，支持 `--bindings` > `.orbit/vars.yaml` > interactive/editor 的优先级，输出 resolved/unresolved 结果并记录来源摘要，相关单测已补齐。
