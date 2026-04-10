# ISSUE-0010 Phase 2 Template Save Edit Loop

- Status: closed
- Priority: medium
- Owner:
- Created: 2026-03-21
- Updated: 2026-03-21

## Summary

Implement `orbit template save --edit-template` so the user can edit the generated template file tree in a temp directory, then save the edited result as a template branch without mutating the runtime repository.

## Scope

- Materialize the candidate template tree in a temp dir.
- Launch the configured editor only against the temp dir payload.
- Re-read the edited files after the editor exits.
- Re-scan variable references from the edited template files.
- Rebuild and validate `.orbit/template.yaml` from the edited result.
- Fail closed when edited files reference undeclared variables.
- Add unit or focused service tests around:
  - temp-dir isolation
  - manifest regeneration after edit
  - new variable discovery
  - undeclared-variable failure

## Done When

- Editing never mutates the runtime worktree.
- Newly introduced variables from edited template files are reflected in the regenerated manifest.
- Edited templates with undeclared variables fail before branch writing.
- The final branch write still uses the same temp-index writer path as normal save.
- The edit loop remains a narrow Phase 2A-1 feature and does not implement the deferred `AGENTS.md` fragment model.

## Notes

- `docs/context/orbit_agents_md_development.md` explicitly says `AGENTS.md` support is a deferred extension and must not reshape `template save` semantics here.
- Keep the editor abstraction minimal; a richer editor UX belongs to later phases.
- 2026-03-21: 已实现 `template save --edit-template`，在 temp dir materialize 模板候选树、调用编辑器、回读编辑结果并重新生成 manifest；编辑流程不改动 runtime worktree，仍沿用同一条 template branch writer 路径落盘。
