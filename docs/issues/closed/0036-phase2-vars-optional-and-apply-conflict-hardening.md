# ISSUE-0036 Phase 2 Vars Optionality And Apply Conflict Hardening

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-23
- Updated: 2026-03-23

## Summary

Retire the Phase 2 gap where template save/apply treated `.orbit/vars.yaml` as effectively mandatory and where apply could materialize or replace that file more aggressively than the product/spec contract allows.

## Scope

- Treat `.orbit/vars.yaml` as optional runtime config when no persisted bindings are needed.
- Allow `orbit template save` to proceed when the repo has no vars file.
- Prevent `orbit template apply` from creating an empty `.orbit/vars.yaml` for no-variable templates.
- Only persist `.orbit/vars.yaml` during apply when new or updated bindings actually need to be written.
- Include `.orbit/vars.yaml` in apply conflict analysis so replacing an existing vars file is fail-closed unless the caller opts into overwrite.
- Add focused unit/integration coverage for the no-vars and vars-conflict paths.

## Done When

- `orbit template save` succeeds for templates that declare no variables even when `.orbit/vars.yaml` is absent.
- `orbit template apply` does not create `.orbit/vars.yaml` for templates that do not require persisted bindings.
- Reusing existing repo vars does not trigger unnecessary vars-file rewrites.
- Applying bindings that would replace an existing `.orbit/vars.yaml` is surfaced in preview conflict output and blocked without `--overwrite-existing`.

## Notes

- 2026-03-23: `template save` now loads repo vars through an optional path, treating a missing vars file as an empty binding set.
- 2026-03-23: `template apply` now plans vars-file writes only when resolved bindings introduce a real persisted change; no-variable templates no longer materialize an empty `.orbit/vars.yaml`.
- 2026-03-23: `.orbit/vars.yaml` is now part of apply conflict analysis and is gated by the existing `--overwrite-existing` behavior.
- 2026-03-23: Deferred on purpose: interactive resolution for same-name/different-value variable conflicts and any namespaced alias strategy such as `<orbit-id>-<var>`; that needs a schema-backed persistence design rather than an implicit apply-time rename.
