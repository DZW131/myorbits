# ISSUE-0017 Phase 2 Remote Template Source Enumeration

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-21
- Updated: 2026-03-21

## Summary

Implement remote template source enumeration for `orbit template apply <git-url>` so Phase 2C-1 can discover which remote heads are actually valid template branches using Git only.

## Scope

- Add Git-backed remote head enumeration for one Git URL.
- Collect candidate heads from `git ls-remote --heads`.
- Filter candidates to heads that expose a valid `.orbit/template.yaml` once the remote revision-read primitive is available.
- Keep branch naming as a hint only; template identity must still come from manifest validation.
- Keep GitHub API and hosting-specific assumptions out of scope.

## Done When

- Remote head discovery runs with Git only.
- Non-template heads are excluded from the candidate set.
- Candidate output is stable enough for later default-resolution and apply orchestration.
- The implementation does not require persistent auxiliary refs for correctness.

## Notes

- Corresponds to `docs/context/orbit_phase2_development_plan.md` Phase 2C-1 / task 1.
- Primary spec references: `docs/context/orbit_phase2_technical_spec.md` sections 7.4, 10.5, and 15.
- 2026-03-21: 已实现远程模板源枚举的第一轮落地，使用 `git ls-remote --heads` 枚举远程 heads，并通过临时 shallow fetch + manifest 校验过滤出合法 template branches；补齐 Git adapter 与 template domain 测试，确认非模板分支会被排除且 temp refs 不会残留。
