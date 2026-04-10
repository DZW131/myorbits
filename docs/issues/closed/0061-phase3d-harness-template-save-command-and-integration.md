# ISSUE-0061 Phase 3D Harness Template Save Command And Integration

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-26
- Updated: 2026-03-26

## Summary

把 member candidate builder、merge/conflict 分析和 root `AGENTS.md` lane 接到正式的 `harness template save` 命令面，并补齐集成测试。

## Scope

- 新增正式命令：
  - `harness template save --to <template-branch> [--path <dir>]`
- command 行为：
  - 读取当前 harness runtime
  - 构建 member candidates
  - 执行 merge/conflict analysis
  - 产出 `.harness/template.yaml`
  - 写 branch
- 失败时保持 fail-closed，不写半成品 branch。
- 不生成 `.orbit/template.yaml`，不复用 `orbit template save` 的 branch contract。

## Done When

- `harness template save` 集成测试覆盖多 member runtime、冲突失败、根 `AGENTS.md` 导出。
- 导出后的 branch 可被 `orbit branch inspect` 识别为 `template_kind=harness`。
- CLI help / JSON 输出与 command 合同一致。

## Notes

- 对应 `docs/harness_centric_runtime_development_plan.md` 的 11.2、11.3、11.4。
- 相关技术合同见 `docs/harness_centric_runtime_technical_spec.md` 的 6.2、8.7、9.1。
- Completed:
  - 新增正式命令 `harness template save --to <template-branch> [--path <dir>] [--json]`。
  - member candidate、merge 引擎和 root `AGENTS.md` whole-file lane 已接入正式 branch save 流程。
  - branch writer 现已支持显式 manifest path，同一 Git 写 branch 内核可同时服务 orbit template 与 harness template。
  - 集成测试覆盖多 member runtime 导出、root `AGENTS.md` 导出、replacement ambiguity fail-closed、`orbit branch inspect` 对 `template_kind=harness` 的识别。
