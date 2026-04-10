# ISSUE-0067 Post V0.3 Follow-Ups

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-26
- Updated: 2026-04-09

## Summary

在 `Phase 3E` 收口之后，继续处理已经明确记录但未进入主线实现的 post-v0.3 follow-up，优先覆盖遗留兼容输出、文档 acceptance smoke、以及关键诊断面的可操作性补强。

## Scope

- 协调本批次的执行子项：
  - legacy `orbit template apply` wrapper runtime output 对齐
  - v0.3 quickstart acceptance smoke
  - `harness check` install-record 诊断细化
  - `harness template save` conflict diagnostics provenance
- 只收口当前已知 residual risk，不扩展新的 runtime 功能面。
- 以 `docs/technical-debt.md` 中的 v0.3 残余项为输入基线。

## Done When

- 本批次子 issue 已完成或明确降级为后续批次。
- 当前已接受的 v0.3 residual risk 至少完成一轮可执行收口。
- 对外主路径和调试体验没有新的明显空档。

## Notes

- 对应 `docs/technical-debt.md` 的 10、12、13、14。
- 这是 post-v0.3 follow-up 批次，不回滚已冻结的 clean-break CLI 边界。
- orbit-level `AGENTS.md` 真相源切换与 `orbit brief backfill` 已单独拆到 `0108`，不混入本批 residual-risk 收口。

## Resolution

- 该 umbrella 下原定的几条主收口线已经各自落地：
  - legacy `orbit template apply` wrapper 已切成明确的 hidden compatibility wrapper，并与 `harness install` 对齐；
  - quickstart acceptance smoke 已完成并扩展到 v0.4 的四类 revision kind；
  - `harness check` install-record diagnostics 已由 `install_record_invalid`、`provenance_unresolvable` 等 finding 覆盖；
  - `harness template save` conflict provenance 已由 ISSUE-0071 与后续 `0117` 硬化工作完成。
- 因此这张 post-v0.3 umbrella issue 已不再承载独立未完成实现；剩余风险已转入具体技术债或后续独立 issue 跟踪，不再继续保留一个宽泛的 open umbrella。
