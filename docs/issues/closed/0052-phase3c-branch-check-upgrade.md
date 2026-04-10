# ISSUE-0052 Phase 3C Branch Check Upgrade

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-26
- Updated: 2026-03-26

## Summary

推进 v0.3 开发计划的 `Phase 3C`，完成 branch taxonomy、inspect contract 和 `harness check` 的正式升级，让 harness runtime / harness template 在 CLI 输出和诊断层稳定可见。

## Scope

- 协调本阶段的 4 个执行子项：
  - branch classifier taxonomy upgrade
  - branch inspect contract hardening
  - `harness check` schema / membership diagnostics
  - `harness check` drift diagnostics
- 明确本阶段只交付完整可观察性与诊断能力，不提前进入 `harness template save`。
- 以 `docs/harness_centric_runtime_development_plan.md` 的 `Phase 3C` 作为验收基线。

## Done When

- branch classifier 能稳定区分 orbit template / harness template / harness runtime / plain。
- `orbit branch inspect` 的 text / json 契约冻结并有 golden tests。
- `harness check` 能报告 schema、membership 和 drift 诊断，且 zero-member runtime 成功返回。

## Notes

- 对应 `docs/harness_centric_runtime_development_plan.md` 的 `Phase 3C`。
- 相关技术合同见 `docs/harness_centric_runtime_technical_spec.md` 的 7.3、8.5、8.6、9.4。
- 建议实现分支：`feature/v0.3-branch-check-upgrade`。
- Completed:
  - branch classifier 已稳定区分 orbit template / harness template / harness runtime / plain。
  - `orbit branch inspect` 已升级到 harness-aware counts/ids 合同，并用 text/json golden tests 冻结。
  - `harness check` 已覆盖 schema、membership、drift 诊断，Phase 3C 交付目标完成。
