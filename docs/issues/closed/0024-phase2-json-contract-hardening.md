# ISSUE-0024 Phase 2 JSON Contract Hardening

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-21
- Updated: 2026-03-21

## Summary

Unify and harden the Phase 2 command JSON contracts so the already-shipped machine-readable surfaces stay stable and the missing ones are filled in consistently.

## Scope

- Inventory current JSON behavior across:
  - `orbit template save`
  - `orbit template apply`
  - `orbit bindings init`
  - `orbit branch status`
  - `orbit branch inspect`
  - `orbit branch list`
- Preserve already-shipped JSON fields unless a spec-backed change is required.
- Add missing `--json` support where the command surface is still human-only.
- Standardize stdout/stderr responsibilities for JSON mode.
- Add focused contract tests so later formatting changes do not silently drift the JSON shape.

## Done When

- Every Phase 2 command documented with `--json` has a stable JSON success contract.
- Existing JSON outputs remain backward-compatible unless docs/spec are updated first.
- Human-readable output and JSON mode have clearly separated responsibilities.
- Contract tests protect the finalized JSON payloads.

## Notes

- Corresponds to `docs/context/orbit_phase2_development_plan.md` Phase 2C-2 / task 3.
- Primary spec references: `docs/context/orbit_phase2_technical_spec.md` sections 7.5 and 17.4.
- 2026-03-21: 已补齐 `orbit template save`、`orbit branch status`、`orbit branch inspect` 的 `--json`，并为 `template save` dry-run/result、`template apply` result、`bindings init`、`branch status`、`branch inspect`、`branch list` 增加或保留集成合同测试，确保 Phase 2 已公开的 JSON 结构有稳定回归网。
