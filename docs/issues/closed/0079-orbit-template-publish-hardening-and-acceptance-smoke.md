# ISSUE-0079 Orbit Template Publish Hardening And Acceptance Smoke

- Status: closed
- Priority: medium
- Owner:
- Created: 2026-03-31
- Updated: 2026-03-31

## Summary

在 source branch 合同收口后，继续补齐 `orbit template publish` 的剩余硬化与端到端验收，避免主路径可用但边界场景未被冻结。

## Scope

- 增加 `--push --remote <name>` 成功路径测试
- 增加 no-op publish 与 `--push` 组合路径测试
- 改善把 source branch 当 install source 时的诊断信息
- 增加 `source branch -> publish -> harness install -> harness check` 的 acceptance smoke

## Done When

- publish 的关键 push/no-op 组合场景有稳定测试
- source branch 被误当 install source 时提示更明确
- 仓库内有一条可执行的 publish/install/check smoke 路径

## Notes

- 这一项以硬化和可观测性为主，不改变 fixed ref naming 或 push 边界
