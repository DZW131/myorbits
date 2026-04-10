# ISSUE-0041 Phase 3B-1 Vars And Install Host Switch

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-26
- Updated: 2026-03-26

## Summary

完成 runtime vars / install record 的正式宿主切换，让 `orbit template save`、`orbit template apply` 与 `orbit bindings init` 对齐 `.harness/*` 合同。

## Scope

- `orbit template save` 默认从 `.harness/vars.yaml` 读取 runtime bindings。
- `orbit template apply` / shared apply preview 写入与冲突检测改用 `.harness/vars.yaml`、`.harness/installs/*`。
- template content builder 不再允许 `.harness/*` runtime metadata 进入模板内容。
- `orbit bindings init` 保持 stdout 默认，同时把帮助与测试口径更新为推荐 `--out .harness/vars.yaml`。

## Done When

- runtime vars 不再默认读写 `.orbit/vars.yaml`。
- install record 不再默认读写 `.orbit/installs/*`。
- 相关单测与集成测试切换到 `.harness/*` 契约。

## Notes

- 对应 `docs/harness_centric_runtime_development_plan.md` 的 8.3 / 任务 1。
- 相关技术合同见 `docs/harness_centric_runtime_technical_spec.md` 的 8.1、8.2、5.3、5.4。
- Completed:
  - `orbit template save` 默认 bindings 来源已切到 `.harness/vars.yaml`。
  - `orbit template apply` / shared apply preview 已改用 `.harness/vars.yaml` 与 `.harness/installs/*`。
  - `template content builder` 已排除 `.harness/*` runtime metadata，不把 runtime host 文件写进模板内容。
  - `orbit bindings init` 继续保持 stdout 默认，并把帮助与测试示例收敛到 `--out .harness/vars.yaml`。
