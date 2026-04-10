# ISSUE-0026 Phase 2 Docs, Help, and Examples Sync

- Status: closed
- Priority: medium
- Owner:
- Created: 2026-03-21
- Updated: 2026-03-21

## Summary

Sync command help, examples, and Phase 2 docs after the hardening work lands so the final V0.2 command surface matches the implemented behavior.

## Scope

- Update CLI help text for:
  - `template save`
  - `template apply`
  - `bindings init`
  - `branch status`
  - `branch inspect`
  - `branch list`
- Refresh examples and docs entry links for:
  - local apply
  - remote apply
  - bindings skeleton generation
  - editor mode
  - dry-run and JSON usage
- Tighten user-facing error/help wording where the current output no longer matches the documented contract.

## Done When

- Help text, examples, and docs all reflect the finalized Phase 2 command behavior.
- No documented flag or output mode contradicts the actual implementation.
- Docs updates are reviewed together with the hardening changes instead of as a later cleanup.

## Notes

- Corresponds to `docs/context/orbit_phase2_development_plan.md` Phase 2C-2 / task 5.
- This issue should be closed near the end of the stage, after the command surface is stable.
- 2026-03-21: 已为 `template save`、`template apply`、`bindings init`、`branch status`、`branch inspect`、`branch list` 补齐 help examples / long help，并在 `docs/quickstart.md` 增加 Phase 2 命令示例与文档入口链接；新增 focused CLI help 测试，防止后续文案与当前命令面漂移。
