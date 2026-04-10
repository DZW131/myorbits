# ISSUE-0057 Phase 3D Harness Template Save

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-26
- Updated: 2026-03-26

## Summary

推进 v0.3 开发计划的 `Phase 3D`，完成多 member runtime -> harness template 的正式导出闭环，让 `harness template save` 具备稳定的候选构建、合并、冲突分析和根目录 `AGENTS.md` whole-file 处理能力。

## Scope

- 协调本阶段的 4 个执行子项：
  - member candidate builder
  - candidate merge and conflict analysis
  - root `AGENTS.md` whole-file lane
  - `harness template save` command and integration
- 明确本阶段只交付 harness template 导出闭环，不提前进入 `Phase 3E` 的文档/发布收尾。
- 以 `docs/harness_centric_runtime_development_plan.md` 的 `Phase 3D` 作为验收基线。

## Done When

- `harness template save` 能从当前 harness runtime 稳定导出 harness template branch。
- member candidate、merge/conflict、root `AGENTS.md` lane 都有稳定测试覆盖。
- 导出结果写 `.harness/template.yaml`，不生成 `.orbit/template.yaml`。

## Notes

- 对应 `docs/harness_centric_runtime_development_plan.md` 的 `Phase 3D`。
- 相关技术合同见 `docs/harness_centric_runtime_technical_spec.md` 的 8.7、9.4。
- 建议实现分支：`codex/v0.3-harness-template-save`。
- Completed:
  - member candidate builder、candidate merge engine、root `AGENTS.md` whole-file lane 与正式 `harness template save` 命令已全部落地。
  - `harness template save` 现在会写 `.harness/template.yaml`，不再生成 `.orbit/template.yaml`。
  - `Phase 3D` 的 unit / integration coverage 已补齐，并通过 `fmt`、`lint`、`test:ci`。
