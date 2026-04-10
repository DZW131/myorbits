# ISSUE-0028 Phase 2 Apply Overwrite Preview Summary

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-21
- Updated: 2026-03-21

## Summary

Retire the current apply preview debt where `--dry-run --overwrite-existing` stops showing which paths or records would be replaced.

## Scope

- Separate conflict collection from conflict blocking in the apply preview/analyzer path.
- Preserve explicit overwrite semantics for real apply.
- Keep dry-run informative even when overwrite intent is enabled.
- Cover both local and remote template apply paths if they share the same analyzer.

## Done When

- `orbit template apply --dry-run --overwrite-existing` still returns a full overwrite/conflict summary.
- Real apply continues to require explicit overwrite intent before replacing existing runtime content.
- Focused tests cover the preview behavior for the shared apply pipeline.

## Notes

- Tracks the “Local Template Apply Preview” technical-debt entry.
- This issue should be treated as part of Phase 2C-2 dry-run hardening, not a separate post-phase cleanup.
- 2026-03-21: 预览路径已将“收集冲突”和“是否阻断 apply”解耦；`orbit template apply --dry-run --overwrite-existing` 现在会继续输出完整冲突摘要，而真实 apply 仍要求显式 overwrite 才会继续写入。
