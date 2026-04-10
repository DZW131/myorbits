# ISSUE-0022 Phase 2 Bindings Init Command

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-21
- Updated: 2026-03-21

## Summary

Implement `orbit bindings init <template-source>` so Phase 2C-2 can generate a fillable bindings skeleton from a template manifest without modifying the current runtime repository.

## Scope

- Add the `orbit bindings init` command on the documented command surface.
- Resolve the template source through the existing local/remote source resolvers:
  - local template branch
  - remote Git URL using the existing remote-selection rules
- Read the template manifest and emit a bindings skeleton with:
  - `schema_version: 1`
  - declared variable names
  - empty string `value`
  - description text when available
- Support:
  - `--out`
  - `--json`
- Keep the command non-mutating for the current repository state.

## Done When

- `orbit bindings init <template-source>` produces a stable skeleton from a valid template source.
- The command does not write `.orbit/` or `.git/orbit/state/*` in the current repo.
- Local and remote template sources both work through the existing manifest/source contracts.
- Human-readable output and `--json` output are covered by focused tests.

## Notes

- Corresponds to `docs/context/orbit_phase2_development_plan.md` Phase 2C-2 / task 1.
- Primary spec references: `docs/context/orbit_phase2_technical_spec.md` sections 7.5, 10.5, 13, and 17.4.
- 2026-03-21: 已实现 `orbit bindings init <template-source>`，支持本地 template branch 与 remote Git URL，默认把 YAML skeleton 输出到 stdout，支持 `--out` 与 `--json`，并通过现有模板源解析逻辑生成带 description 的空值 bindings skeleton。
