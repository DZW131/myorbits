# ISSUE-0011 Phase 2 Template Save Integration Tests

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-21
- Updated: 2026-03-21

## Summary

Add CLI integration coverage for the complete Phase 2A-1 `orbit template save` flow so the new template-branch export path is protected end to end.

## Scope

- Extend the main CLI integration suite with isolated temp-repo scenarios for:
  - normal save
  - `--dry-run`
  - `--edit-template`
  - `--default`
  - target branch exists without `--overwrite`
  - ambiguity fail-closed
  - hidden tracked scope files recovered from `HEAD`
- Assert the saved template branch contains only:
  - `.orbit/template.yaml`
  - `.orbit/orbits/<orbit-id>.yaml`
  - templated user files
- Assert excluded files never enter the template branch:
  - `.orbit/config.yaml`
  - `.orbit/vars.yaml`
  - `.orbit/installs/*`
- Verify save does not switch the current branch or pollute the current worktree.

## Done When

- CLI integration tests exercise the real command path from Cobra entry to Git branch output.
- The saved branch is recognized as `template` by the existing branch classifier.
- Failure cases prove the command is fail-closed instead of silently degrading.
- The suite gives direct regression coverage for the A-1 write path and reduces the current technical debt around Phase 2 write flows.

## Notes

- Favor realistic temp-repo flows over excessive mocking because this stage’s core risk is the Git integration path.
- Keep tests aligned to stable user-visible behavior, not internal helper call order.
- 2026-03-21: 已补齐 `template save` 端到端 CLI 集成测试，覆盖正常保存、`--dry-run`、`--edit-template`、`--default`、branch 已存在、ambiguity fail-closed 和 hidden tracked files from `HEAD` 等关键路径。
