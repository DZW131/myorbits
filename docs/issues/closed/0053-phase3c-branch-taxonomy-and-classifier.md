# ISSUE-0053 Phase 3C Branch Taxonomy And Classifier

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-26
- Updated: 2026-03-26

## Summary

升级 branch classifier，让分支层能稳定识别 orbit template、harness template、harness runtime 以及冲突场景，并对 zero-member runtime 保持正确分类。

## Scope

- classifier 识别：
  - orbit template
  - harness template
  - harness runtime
  - plain branch
- 处理 invalid conflict：
  - `.orbit/template.yaml` 与 `.harness/template.yaml` 同时存在
- runtime 分支识别继续要求：
  - 合法 `.harness/runtime.yaml`
  - 合法 `.orbit/config.yaml`
- zero-member runtime 必须被识别为有效 runtime，而不是 plain 或 invalid

## Done When

- branch classifier 单测覆盖 orbit template / harness template / harness runtime / invalid conflict / zero-member runtime。
- `orbit branch status` 与 `orbit branch list` 的分类输出和新 taxonomy 保持一致。

## Notes

- 对应 `docs/harness_centric_runtime_development_plan.md` 的 10.3 / 任务 1。
- 相关技术合同见 `docs/harness_centric_runtime_technical_spec.md` 的 7.3、8.5、8.6。
- Completed:
  - classifier 现已区分 `template_kind=orbit|harness`，并保持 runtime / plain / zero-member runtime 合同。
  - 合法 `.harness/template.yaml` 与合法 `.orbit/template.yaml` 同时存在时，分支会按 conflict reason 落回 plain。
  - `orbit branch status --json` 与 `orbit branch list --json` 已稳定暴露 `template_kind`。
