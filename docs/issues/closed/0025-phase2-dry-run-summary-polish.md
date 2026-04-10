# ISSUE-0025 Phase 2 Dry-Run Summary Polish

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-21
- Updated: 2026-03-21

## Summary

Improve Phase 2 dry-run outputs so save/apply previews are detailed enough for safe review, automation, and regression testing before Phase 2 is declared complete.

## Scope

- Audit current dry-run behavior for:
  - `orbit template save`
  - `orbit template apply`
- Ensure save dry-run surfaces:
  - file list
  - replacement summary
  - ambiguity summary
  - manifest summary
- Ensure apply dry-run surfaces:
  - source summary
  - manifest summary
  - bindings source summary
  - file write list
  - conflict / overwrite summary
- Add or update JSON coverage when dry-run output is part of the machine contract.

## Done When

- Save/apply dry-run output is detailed enough to review the intended write set and source metadata.
- `--dry-run --overwrite-existing` still reports what would be replaced instead of hiding overwrite details.
- The polished summaries are covered by focused tests in both human-readable and JSON modes where applicable.

## Notes

- Corresponds to `docs/context/orbit_phase2_development_plan.md` Phase 2C-2 / task 4.
- Primary spec references: `docs/context/orbit_phase2_development_plan.md` section 10.2 and `docs/context/orbit_phase2_technical_spec.md` sections 10.6 and 17.4.
- This stage task absorbs technical-debt entry “Local Template Apply Preview”; the debt-specific tracking is split out in ISSUE-0028.
- 2026-03-21: 已补齐 `template save` dry-run 的 manifest provenance 摘要，补强 save/apply dry-run 的人类可读与 JSON 合同测试，并让 apply dry-run 在 `--overwrite-existing` 下仍保留完整冲突摘要。
