# ISSUE-0008 Phase 2 Template Save Dry-Run

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-21
- Updated: 2026-03-21

## Summary

Implement the `orbit template save --dry-run` preview pipeline for the minimal Phase 2A-1 loop. The dry-run should exercise the same runtime-to-template build path as real save, but stop before branch writing and emit a stable summary for review.

## Scope

- Introduce the save-preview service that composes existing A-0 primitives:
  - orbit definition loading
  - tracked scope resolution
  - runtime bindings loading from `.orbit/vars.yaml`
  - template content building
  - ambiguity collection
  - manifest generation summary
- Define the manifest-building step for non-edit save:
  - template metadata
  - declared variables
  - default flag
- Produce stable preview output for:
  - template file list
  - replacement summary
  - ambiguity summary
  - generated `.orbit/template.yaml` summary
- Add tests that prove dry-run does not write refs or mutate the worktree.

## Done When

- The preview path reuses the same candidate and manifest logic that real save will use.
- Dry-run returns a stable, reviewable summary without creating commits or refs.
- Ambiguity stays fail-closed and is surfaced in preview output rather than ignored.
- `.orbit/config.yaml`, `.orbit/vars.yaml`, `.orbit/installs/*`, and `.git/orbit/state/*` never appear in the previewed template tree.
- Tests cover both a normal preview and an ambiguity/failure case.

## Notes

- The replacement engine debt about global ambiguity scope becomes visible here; do not silently narrow that behavior in dry-run unless the product docs are updated first.
- Keep the output contract tight enough for later CLI rendering, but avoid prematurely designing the full Phase 2C dry-run enhancement surface.
- 2026-03-21: 已实现 `template save` preview service 与 CLI dry-run 输出，复用真实 save 的候选模板和 manifest 生成路径，保持 ambiguity fail-closed，并验证 dry-run 不写 refs、不修改 worktree。
