# ISSUE-0020 Phase 2 Remote Template Apply Command

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-21
- Updated: 2026-03-21

## Summary

Implement `orbit template apply <git-url>` so Phase 2C-1 can resolve a remote template source and then reuse the existing local-apply pipeline for rendering, conflict handling, install records, and bindings behavior.

## Scope

- Extend `orbit template apply` to accept a Git URL as the source input.
- Connect remote source enumeration, default selection, and temp-ref reads to the existing local apply backend.
- Support the Phase 2C-1 flag surface:
  - `--ref`
  - `--bindings`
  - `--overwrite-existing`
  - `--interactive`
  - `--dry-run`
  - `--json`
- Preserve current apply guarantees:
  - no implicit `enter`
  - install record written in `.orbit/installs/<orbit-id>.yaml`
  - runtime control-plane writes stay in `.orbit/`, not `.git/orbit/state/`
- Ensure remote installs record `source_kind: remote_git`.

## Done When

- `orbit template apply <git-url>` can resolve and apply a valid remote template source end to end.
- Ambiguous, missing-template, and remote-read failure cases surface clear errors without leaving local ref pollution behind.
- Successful remote apply writes the expected runtime files and an install record with remote source metadata.
- Existing local-branch apply behavior is not regressed.

## Notes

- Corresponds to `docs/context/orbit_phase2_development_plan.md` Phase 2C-1 / task 4.
- Primary spec references: `docs/context/orbit_phase2_technical_spec.md` sections 7.5, 12.2, 12.3, 15, and 17.4.
- 2026-03-21: 已实现 `orbit template apply <git-url>` 的第一轮落地，命令层会在本地 branch 与 remote Git URL 间分流，并把 remote source selection + temp-ref snapshot 接到既有 apply 后半段；补齐 `--ref` 与 `--json` 支持，并确认成功应用后 install record 写入 `source_kind: remote_git` 且不会自动 `enter`。
