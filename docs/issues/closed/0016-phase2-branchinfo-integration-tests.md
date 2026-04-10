# ISSUE-0016 Phase 2 Branch Info Integration Tests

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-21
- Updated: 2026-03-21

## Summary

Add CLI integration coverage for the complete Phase 2B-2 branch info surface so `branch status`, `branch inspect`, and `branch list` are protected end to end and keep reusing the shared classifier contract.

## Scope

- Extend the CLI integration suite with isolated temp-repo scenarios for:
  - current branch status on template / runtime / plain branches
  - inspect on template / runtime / plain revisions
  - list on mixed local branches
  - template branch recognition independent of branch naming
- Assert command output stays stable on the key fields required by the spec:
  - branch kind
  - reason / summary
  - manifest summary for templates
  - install/runtime summary for runtime branches
- If `--json` lands naturally during Phase 2B-2, cover the JSON shape here as well; otherwise keep this issue focused on stable human-readable output.

## Done When

- CLI integration tests exercise the real Cobra command path for all branch info commands.
- Mixed template / runtime / plain repositories are covered with realistic temp-repo setup.
- Tests make it hard for future command implementations to drift away from `branchinfo.ClassifyRevision`.
- The suite gives direct regression coverage for Phase 2B-2 before remote-source work begins.

## Notes

- Supports `docs/context/orbit_phase2_development_plan.md` Phase 2B-2 as the phase-level regression net.
- Keep test assertions anchored to stable behavior, not internal helper call order.
- 2026-03-21: 已完成 Phase 2B-2 的 CLI 集成回归网，覆盖 `cmd/orbit/cli/branch_status_integration_test.go`、`cmd/orbit/cli/branch_inspect_integration_test.go` 与 `cmd/orbit/cli/branch_list_integration_test.go`；三条命令都通过真实 Cobra 路径复用共享 classifier，mixed template/runtime/plain 与 branch-name-independent 场景已纳入回归保护。
