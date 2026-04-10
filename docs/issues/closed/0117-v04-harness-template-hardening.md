# ISSUE-0117 v0.4 Harness Template Hardening

- Status: closed
- Priority: medium
- Owner:
- Created: 2026-04-07
- Updated: 2026-04-09

## Summary

把 harness template 的保存、bundle provenance 与安装语义按 v0.4 正式硬化，让“组合好的 runtime 打包为 harness template”成为建立在稳定基础能力之上的正式工作流。

## Scope

- 硬化 `harness template save`
- 统一 bundle provenance 与 runtime member source 记录
- 补 harness template save / install 的 acceptance smoke
- 对接已有 open issues 中与 conflict provenance、install variable policy 相关的边界

## Done When

- harness template payload、bundle record 与 runtime member source 保持一致
- save / install 行为在文档、CLI、tests 中口径一致
- 复杂 bundle 场景有最小 acceptance coverage

## Notes

- 对应 `docs/orbit_v0_4_technical_spec.md` 的第 6、9 节
- 对应 `docs/orbit_v0_4_development_plan.md` 的 Phase 6
- 与 `docs/issues/open/0071-harness-template-save-conflict-provenance.md`、`docs/issues/open/0087-install-variable-conflict-compatibility-policy.md` 有交叉，实施时应统一收口而不是再次分叉

## Progress

- `harness template save` 已补齐 `--default`、`--dry-run`、`--overwrite`、`--edit-template`，并完成 branch-manifest 校验与 default-template 收口。
- harness template save/install 的 manifest preflight、dry-run JSON、ambiguity contributors、save failure JSON contract 已显著硬化。
- conflict provenance 已从纯文本错误推进到 dry-run/save JSON 可消费合同，当前模板保存链路已有较完整测试覆盖。
- bundle provenance 与 runtime member source 现在已形成稳定主链：`install_bundle` 已进入 runtime manifest contract，bundle record、same-bundle overwrite、bundle-backed member 替换与 `harness check` 的 bundle mismatch diagnostics 都已落地。
- harness template save / install 的 quickstart acceptance、dry-run/save JSON、branch-manifest contract 与 overwrite 行为已在文档、CLI help 与 integration tests 中对齐。

## Resolution

- `harness template save` 的正式工作流已经收口：`--default`、`--dry-run`、`--overwrite`、`--edit-template`、manifest preflight、preview/save JSON contract、conflict provenance 与 branch-manifest validation 都已进入主链。
- `harness install` 的 harness-template lane 也已完成基本硬化：local/remote source resolution 以 branch manifest 为准，bundle record 会稳定落盘，runtime member source 会写成 `install_bundle`，same-bundle overwrite 与 bundle-owned cleanup 也已有最小 acceptance coverage。
- 因此 “组合好的 runtime 打包成 harness template 并重新安装” 现在已经是 v0.4 中可用、口径一致、且有 acceptance smoke 支撑的正式工作流。
- 与 `0071` 的交叉项现在只剩更细的 diagnostics/contract 延展；与 `0087` 的交叉项则已经属于更广义的 install variable policy 设计，不再是 harness-template hardening 主线的阻塞条件。
- 剩余风险已转入 `docs/technical-debt.md` 与独立 follow-up issue 跟踪，例如 mixed-install variable compatibility policy、bundle diagnostics payload 语义、same-bundle cleanup replay 保守性，以及 harness-template install 的非事务性写路径。

因此本 issue 按 v0.4 Phase 6 的主链完成标准关闭。
