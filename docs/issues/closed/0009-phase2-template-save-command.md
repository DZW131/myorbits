# ISSUE-0009 Phase 2 Template Save Command

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-21
- Updated: 2026-03-21

## Summary

Wire the minimal `orbit template save <orbit-id> --to <template-branch>` command into the CLI so Phase 2A-1 can produce a valid template branch from the current runtime repository.

## Scope

- Add the `template` command tree and the `save` subcommand to the Cobra root.
- Implement the non-edit save state machine:
  - load target orbit definition
  - resolve tracked user scope from repo state, not current sparse visibility
  - load runtime bindings from `.orbit/vars.yaml`
  - build template candidates
  - review replacement ambiguities
  - build manifest
  - call the template branch writer
- Support the A-1 minimum flags:
  - `--to`
  - `--dry-run`
  - `--overwrite`
  - `--default`
- Return fail-closed errors for:
  - invalid or missing orbit
  - invalid runtime vars file
  - ambiguity
  - target branch already exists without overwrite
  - branch write failure
- Emit stable human-readable output for successful save and dry-run.

## Done When

- `orbit template save <orbit-id> --to <template-branch>` creates a valid template branch without switching the current branch.
- `--default` only affects manifest metadata on the target template branch.
- `--dry-run` goes through the same orchestration path but stops before writing.
- The command does not read `.git/orbit/state/resolved_scope/*` or depend on `current_orbit.json`.
- The implementation keeps command files thin and leaves business orchestration in `template` and `git` packages.

## Notes

- Keep `--json` out of scope unless the implementation falls out naturally from the command response model. The development plan lists broader JSON output under later experience enhancements, so A-1 should not stall on that surface.
- Reuse `branchinfo.ClassifyRevision` later for branch-info commands; do not couple this save command to branch-status work.
- 2026-03-21: 已接入 `orbit template save <orbit-id> --to <template-branch>`，支持 `--to`、`--dry-run`、`--overwrite`、`--default`，命令层保持薄，核心编排落在 `template` / `git` 包中。
