# ISSUE-0021 Phase 2 Remote Template Apply Integration Tests

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-21
- Updated: 2026-03-21

## Summary

Add remote-source integration coverage for Phase 2C-1 so the Git URL apply flow is protected end to end before Phase 2C-2 hardening begins.

## Scope

- Use a local bare repo to simulate a remote template source.
- Cover:
  - `git ls-remote --heads`
  - valid template candidate discovery
  - default-template resolution
  - temp-ref fetch and cleanup
  - `orbit template apply <git-url>`
  - no-template / multi-template / unique-default scenarios
- Assert that remote apply does not auto-enter an orbit and does not rely on `.git/orbit/state/*` for template install correctness.

## Done When

- The remote apply path has CLI integration coverage through the real Cobra command path.
- Bare-repo fixtures cover the documented selection and failure modes.
- Tests make it hard to regress temp-ref cleanup, remote source selection, or remote install record metadata.
- The suite gives direct regression coverage for Phase 2C-1 before Phase 2C-2 polish work starts.

## Notes

- Corresponds to `docs/context/orbit_phase2_development_plan.md` Phase 2C-1 / task 5.
- Primary spec references: `docs/context/orbit_phase2_technical_spec.md` sections 10.5, 15, 17.3, and 17.4.
- Added end-to-end Cobra coverage for:
  - explicit remote apply with `--ref`
  - `--dry-run --json`
  - unique default-template auto-selection
  - no-template failure
  - ambiguous multi-template failure
  - temp remote ref cleanup on the real command path
- The remote apply integration suite now also asserts that install metadata stays under `.orbit/installs/` and does not move into `.git/orbit/state/*`.
