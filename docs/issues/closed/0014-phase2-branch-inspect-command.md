# ISSUE-0014 Phase 2 Branch Inspect Command

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-21
- Updated: 2026-03-21

## Summary

Implement `orbit branch inspect <branch>` so Phase 2B-2 can explain one revision in more detail, including template manifest details for template branches and runtime metadata for runtime branches.

## Scope

- Add the `inspect` subcommand under `orbit branch`.
- Accept one branch or revision argument and classify it using the existing `branchinfo` domain primitive.
- For template branches, surface at least:
  - branch kind
  - orbit id
  - default template marker
  - template manifest summary
- For runtime branches, surface at least:
  - branch kind
  - orbit id when install records or definitions make it available
  - install record summary
  - runtime control-plane summary
- For plain branches, return a clear classification result instead of failing only because the branch is plain.

## Done When

- `orbit branch inspect <branch>` produces stable output for template, runtime, and plain branches.
- Template summaries are driven by `.orbit/template.yaml`, not branch naming conventions.
- Runtime summaries are driven by `.orbit/config.yaml`, `.orbit/orbits/*.yaml`, and `.orbit/installs/*.yaml` when present.
- The command stays inside the current storage boundary and does not introduce `.git/orbit/state/*` dependencies for correctness.

## Notes

- Corresponds to `docs/context/orbit_phase2_development_plan.md` Phase 2B-2 / task 2.
- Primary spec references: `docs/context/orbit_phase2_technical_spec.md` sections 14.1 and 14.2.
- 2026-03-21: 已实现 `orbit branch inspect <branch-or-revision>`，命令层通过 `branchinfo.InspectRevision` 复用既有分类原语，并在 template/runtime/plain 三类 revision 上输出稳定摘要；补齐 template manifest、runtime control-plane、install record 的 CLI 集成测试。
