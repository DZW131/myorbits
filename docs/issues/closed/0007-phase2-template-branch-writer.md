# ISSUE-0007 Phase 2 Template Branch Writer

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-21
- Updated: 2026-03-21

## Summary

Implement the Git-side writer for Phase 2A-1 so `orbit template save` can write a complete template tree to a target branch without switching branches or mutating the current worktree. This must follow the documented temp-index + `git commit-tree` + `git update-ref` flow.

## Scope

- Add `git` helpers for the A-1 write path:
  - resolve whether a target branch already exists
  - build a tree from in-memory template files through a temp index
  - create a commit with `git commit-tree`
  - create or update the target ref with `git update-ref`
- Keep the implementation worktree-free and shell-free.
- Define the pure input/output contract for writing a template branch from:
  - candidate template files
  - generated `.orbit/template.yaml`
  - target branch name
  - overwrite policy
- Preserve current branch, current HEAD checkout, and current worktree contents.
- Add focused unit tests around branch existence checks, temp-index tree creation, overwrite gating, and ref update behavior.

## Done When

- A caller can provide the final template file tree and manifest bytes and receive a new template commit/ref without checking out another branch.
- Existing target branch fails closed unless `--overwrite` is explicitly enabled.
- The resulting branch contains only the files passed into the writer.
- Tests prove the writer does not require worktree mutation and does not silently overwrite an existing branch.
- The implementation stays inside the documented storage boundary and does not depend on `.git/orbit/state/*`.

## Notes

- This issue directly absorbs the main technical risk surfaced by `docs/technical-debt.md`: the new Phase 2 flows now start doing real multi-file writes, so Git-level writer tests need to be stronger than the indirect schema tests.
- Keep the writer reusable by later `template apply` and branch-info flows where possible, but do not broaden scope beyond A-1.
- 2026-03-21: 已实现 Git-side template branch writer，采用 temp index + `git commit-tree` + `git update-ref` 路径，在不切换 branch、不改动当前 worktree 的前提下写入模板分支，并补齐 overwrite gating 与 worktree-free 单测。
