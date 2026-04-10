# ISSUE-0068 Legacy Apply Wrapper Output Alignment

- Status: closed
- Priority: medium
- Owner:
- Created: 2026-03-26
- Updated: 2026-04-06

## Summary

把 hidden `orbit template apply` compatibility wrapper 的运行时 text 输出也收口到 legacy bridge 口径，避免 preview / success summary 继续把旧命令呈现成正式主路径。

## Scope

- 调整 `orbit template apply` 的 text preview / success 输出文案：
  - 明确这是 legacy compatibility wrapper
  - 明确正式安装入口是 `harness install`
- 保持现有共享 apply pipeline、exit code、flag 语义与 JSON payload 不变。
- 为本地 apply 的 dry-run / success 路径补回归测试。

## Done When

- `orbit template apply --help` 与实际 text 输出口径一致。
- wrapper 的 text 输出不再把 `template apply` 呈现成正式主路径。
- 兼容入口仍可正常驱动现有 local / remote apply 行为。

## Notes

- 对应此前 `docs/technical-debt.md` 中关于 legacy apply wrapper runtime output 的残余项；该残余项已随本 issue 完成一起移除。
- 这项只收口 text output，不扩大 wrapper 的公开命令面。
- Completed:
  - `orbit template apply` 的 text dry-run / success 输出现在都显式标记为 legacy wrapper。
  - text 输出新增 `preferred_command: harness install`，把正式安装入口与兼容入口清晰分开。
  - JSON payload、exit code、flag 语义与共享 apply pipeline 保持不变。
  - 本地 dry-run / success 与 remote success 的回归测试已补齐并通过。
