# ISSUE-0027 Phase 2 Editor Command Parsing Hardening

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-21
- Updated: 2026-03-21

## Summary

Retire the current `EDITOR` parsing debt so both template-save editing and apply `--editor` mode can launch real editor commands safely and predictably.

## Scope

- Replace the current `strings.Fields`-based editor parsing with a more robust argv strategy.
- Support:
  - quoted arguments
  - executable paths containing spaces
  - consistent environment-driven editor resolution for both save/edit and apply/editor flows
- Keep command execution explicit and fail-closed on malformed editor configuration.
- Add tests that reproduce the currently broken quoting/space cases.

## Done When

- Editor-backed flows no longer depend on brittle whitespace splitting.
- The same editor invocation contract is reused by both template save and template apply.
- Regression tests cover quoted arguments and editor executable paths with spaces.

## Notes

- Tracks technical-debt entry `0010`.
- This issue should land alongside ISSUE-0023, not after it.
- 2026-03-21: 已用共享的 shell-like argv 解析替换 `strings.Fields`，`template save --edit-template` 和 `template apply --editor` 现在都支持带引号参数与包含空格的 editor 可执行路径，并在 malformed `EDITOR` 下 fail-closed。
