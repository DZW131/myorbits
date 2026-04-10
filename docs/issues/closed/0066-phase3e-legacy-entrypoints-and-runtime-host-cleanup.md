# ISSUE-0066 Phase 3E Legacy Entrypoints And Runtime Host Cleanup

- Status: closed
- Priority: medium
- Owner:
- Created: 2026-03-26
- Updated: 2026-03-26

## Summary

收口遗留入口和旧 runtime host 叙事，确保兼容 wrapper 只保留必要的 hidden/internal 行为，不再在正式帮助、示例或用户输出里继续放大旧世界观。

## Scope

- 复查 `orbit template apply` 的保留方式：
  - 允许 hidden/internal wrapper 存在
  - 不在正式 help / README / quickstart 中公开推荐
- 清理正式输出与帮助里的旧 runtime host 叙事：
  - 不继续把 `.orbit/vars.yaml` / `.orbit/installs/*` 描述为正式 runtime host
  - 不让 Orbit 重新长成 runtime owner
- 复查过渡别名、旧示例和帮助文案是否仍有外露。

## Done When

- 遗留 wrapper 不再进入正式帮助与主文档。
- 正式输出与文档不再残留旧 runtime host 主叙事。
- 对兼容入口保留最小必要回归，避免无意删断现有内部桥接。

## Notes

- 对应 `docs/harness_centric_runtime_development_plan.md` 的 12.3 / 任务 4。
- 相关技术合同见 `docs/harness_centric_runtime_prd.md` 的 clean-break CLI 口径与 `docs/harness_centric_runtime_technical_spec.md` 的 runtime host 切换原则。
- Completed:
  - `orbit template apply --help` 现在明确标记为 legacy compatibility wrapper，并把正式示例切到 `harness install`。
  - `orbit template --help` 继续隐藏 `apply`，公开 `template` 命令树不再把安装型入口作为正式子命令暴露。
  - `AGENTS.md` 已改为当前 v0.3 source-of-truth 与 `.harness/*` runtime host 口径，不再把 `.orbit/vars.yaml` / `.orbit/installs/*` 写成正式运行态宿主。
  - `mise run fmt`、`mise run lint`、`mise run test:ci` 已通过。
