# ISSUE-0035 AGENTS Integration And Doc Sync

- Status: closed
- Priority: medium
- Owner:
- Created: 2026-03-23
- Updated: 2026-03-23

## Summary

在 AGENTS 专用 lane 跑通后，补齐跨 save / apply / validate 的集成保护，并同步 help / docs / technical debt 状态。

## Scope

- 增加端到端回归测试，覆盖：
  - runtime -> template save -> template apply
  - local apply 与 remote apply
  - malformed runtime `AGENTS.md` 的 validate / apply 阻断
- 同步命令帮助与 `docs/quickstart.md` 中相关入口。
- 更新 `docs/context/orbit_agents_md_development.md` 为“已实现合同”状态。
- 清理或调整相关 `technical-debt` 条目。

## Done When

- AGENTS lane 的关键行为都有跨链路测试保护。
- docs / help 与最终实现口径一致。
- 已接受风险仍有明确记录，不会在文档同步时丢失。

## Notes

- 这是收尾 issue，不应在核心 save/apply/validate 行为尚未稳定前抢跑。
- 若实现阶段新增明显跨链路风险，可在本 issue 中记录并收口。
- Completed:
  - 补齐了 save / apply / validate 的 AGENTS 相关集成覆盖。
  - 同步更新了 `docs/context/orbit_agents_md_development.md` 的实现状态。
  - 把真正残留的 AGENTS 风险改写进 `docs/technical-debt.md`，并移除了“planned but not implemented”口径。
