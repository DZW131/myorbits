# ISSUE-0013 Phase 2 Branch Status Command

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-21
- Updated: 2026-03-21

## Summary

Implement `orbit branch status` so Phase 2B-2 can classify the current branch as `template`, `runtime`, or `plain`, and explain the result using schema-backed file contracts rather than branch naming.

## Scope

- Add the `branch` command tree and the `status` subcommand to the Cobra root.
- Wire the command to reuse `branchinfo.ClassifyRevision` for current-branch classification instead of duplicating classification logic in command code.
- Emit stable human-readable output for:
  - branch kind
  - short reason summary
- Fail closed for:
  - non-Git working directory
  - repository resolution failure
  - unreadable or invalid revision inputs surfaced from the classifier

## Done When

- `orbit branch status` reports `template`, `runtime`, or `plain` for the current branch.
- Output does not depend on branch name heuristics.
- Command files stay thin and the classification contract remains centralized in `branchinfo`.
- The implementation closes the current `technical-debt.md` risk around command-level drift from `branchinfo.ClassifyRevision`.

## Notes

- Corresponds to `docs/context/orbit_phase2_development_plan.md` Phase 2B-2 / task 1.
- Primary spec references: `docs/context/orbit_phase2_technical_spec.md` sections 14.1 and 14.4.
- 2026-03-21: 已实现 `orbit branch status`，通过 `git.CurrentBranch` 获取当前分支并直接复用 `branchinfo.ClassifyRevision` 做分类，补齐 template/runtime/plain 与非 Git 目录的 CLI 集成测试，命令层未引入重复分类逻辑。
