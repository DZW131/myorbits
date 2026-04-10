# ISSUE-0018 Phase 2 Remote Default Template Resolution

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-21
- Updated: 2026-03-21

## Summary

Implement remote template selection rules so Phase 2C-1 can deterministically choose one remote template branch from the discovered candidates, or fail closed with a clear ambiguity result.

## Scope

- Resolve one remote template source from:
  - Git URL
  - optional `--ref`
  - discovered valid template candidates
- Enforce the documented precedence:
  - explicit `--ref`
  - exactly one valid template branch
  - exactly one `default_template: true`
  - otherwise return ambiguity candidates
- Treat multiple defaults as ambiguity, not auto-selection.
- Return clear errors when no valid template branches exist.

## Done When

- Explicit `--ref` takes priority when it resolves to a valid template branch.
- A unique valid template branch auto-resolves without user input.
- A unique default template auto-resolves when multiple candidates exist.
- Multi-default, no-default-multi-candidate, and no-template cases fail closed with stable candidate/error output.

## Notes

- Corresponds to `docs/context/orbit_phase2_development_plan.md` Phase 2C-1 / task 2.
- Primary spec references: `docs/context/orbit_phase2_technical_spec.md` sections 10.5 and 11.5.
- 2026-03-21: 已实现远程模板选择规则的第一轮落地，补齐显式 `--ref`、唯一候选、唯一 default、无模板、无唯一 default 与多 default 的选择/失败闭环；返回稳定的 not-found / ambiguity error 形态，供后续 CLI 输出候选列表复用。
