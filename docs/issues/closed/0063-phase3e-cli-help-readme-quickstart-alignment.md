# ISSUE-0063 Phase 3E CLI Help README Quickstart Alignment

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-26
- Updated: 2026-03-26

## Summary

统一 CLI help、README、quickstart 和主文档中的命令示例与产品口径，确保正式主路径已经切到 `harness install` 和 `harness template save`，并明确 `orbit` 与 `harness` 的对象边界。

## Scope

- 更新 README 中的产品描述、初始化与安装主路径。
- 更新 `docs/quickstart.md` 的推荐流程：
  - `harness create/init`
  - `harness install`
  - `harness check`
  - `harness template save`
- 更新 CLI help 文案与示例：
  - `harness`
  - `orbit`
- 保持 `orbit` 负责 definition / projection / orbit template；
- 保持 `harness` 负责 runtime / install / members / template export。

## Done When

- README 和 quickstart 不再把 `orbit template apply` 当正式主路径。
- `harness --help` 与 `orbit --help` 的示例命令符合 v0.3 口径。
- 文档中对 `.orbit/` 与 `.harness/` 的职责描述一致且无冲突。

## Notes

- 对应 `docs/harness_centric_runtime_development_plan.md` 的 12.3 / 任务 1。
- 相关技术合同见 `docs/harness_centric_runtime_prd.md` 的 11.1、11.2、12。
- Completed:
  - `orbit --help` 与 `harness --help` 现在都带有符合 v0.3 口径的根示例。
  - `harness install --help` 已补正式安装路径示例，明确 `--dry-run` 只是预览而非主路径。
  - `README.md` 已切换到双二进制与 harness-centric runtime 主叙事。
  - `docs/quickstart.md` 已改为 `harness create/init -> harness install -> harness check -> harness template save` 的正式 quickstart。
