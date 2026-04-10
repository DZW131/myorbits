# ISSUE-0015 Phase 2 Branch List Command

- Status: closed
- Priority: medium
- Owner:
- Created: 2026-03-21
- Updated: 2026-03-21

## Summary

Implement `orbit branch list` so Phase 2B-2 can enumerate local branches and show the classification result for each branch using the same file-contract rules as `branch status` and `branch inspect`.

## Scope

- Add the `list` subcommand under `orbit branch`.
- In the first round, only enumerate local branches.
- Reuse the `branchinfo` classifier for each listed branch.
- Emit stable output that includes:
  - branch name
  - branch kind
  - short reason or summary signal
- Keep remote branch enumeration out of scope for this issue.

## Done When

- `orbit branch list` classifies local branches as `template`, `runtime`, or `plain`.
- The command does not depend on branch naming conventions.
- Output order is stable enough for CLI assertions and later JSON shaping.
- Remote branch listing remains deferred to the remote-source phase.

## Notes

- Corresponds to `docs/context/orbit_phase2_development_plan.md` Phase 2B-2 / task 3.
- Primary spec references: `docs/context/orbit_phase2_technical_spec.md` sections 14.1 and 14.3.
- 2026-03-21: 已实现 `orbit branch list`，第一轮只枚举本地 branches，输出稳定的人类可读分类结果，并自然补齐 `--json` 形态；命令层通过 `branchinfo.ListLocalBranches` 统一复用 `branchinfo.ClassifyRevision`，覆盖 mixed local branches、非 Git 目录和 JSON 输出的 CLI 集成测试。
