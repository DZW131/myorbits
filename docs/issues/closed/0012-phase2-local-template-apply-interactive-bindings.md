# ISSUE-0012 Phase 2 Local Template Apply Interactive Bindings

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-21
- Updated: 2026-03-21

## Summary

Complete the remaining Phase 2B-1 interactive bindings path so `orbit template apply <local-branch> --interactive` can prompt for missing required variables instead of failing with a placeholder “not implemented” error.

## Scope

- Extend local template apply bindings resolution with an interactive fallback for unresolved required variables.
- Prompt using manifest-backed variable names and descriptions after applying the documented precedence:
  `--bindings` > current repo `.orbit/vars.yaml` > interactive fill.
- Keep non-interactive mode fail-closed when required variables are still missing.
- Persist interactively provided values through the existing `.orbit/vars.yaml` update path.
- Add focused tests for:
  - missing required variables resolved by interactive input
  - no prompt when repo vars or `--bindings` already satisfy all requirements
  - apply still does not auto-enter the orbit

## Done When

- `orbit template apply <local-branch> --interactive` succeeds for templates that are otherwise only blocked by missing required bindings.
- Prompt text is driven by manifest declarations instead of ad hoc strings.
- The final rendered output and install record stay stable for the same template source + resolved bindings.
- Non-interactive apply continues to fail closed on missing required variables.
- CLI and service tests cover both the interactive success path and the non-interactive failure path.

## Notes

- Corresponds to `docs/context/orbit_phase2_development_plan.md` Phase 2B-1 / task 5.
- As of commit `a95ca39`, the command surface exists but `--interactive` returns an explicit not-implemented error.
- 2026-03-21: 已实现 local template apply 的 interactive bindings fallback，按 `--bindings` > repo vars > interactive 的优先级补齐缺失变量，prompt 文案复用 manifest description，非交互路径继续 fail-closed，并补齐服务层与 CLI 集成测试。
