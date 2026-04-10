# ISSUE-0118 v0.4 Docs Help And Acceptance Cleanup

- Status: closed
- Priority: medium
- Owner:
- Created: 2026-04-07
- Updated: 2026-04-09

## Summary

在 v0.4 主链路落地后，统一清理旧兼容入口、更新 docs/help/examples/fixtures，并补齐文档驱动 acceptance，确保仓库对外入口与实际实现一致。

## Scope

- 清理或彻底隐藏旧兼容入口
- 更新 quickstart、release、CLI help、examples 与 acceptance fixtures
- 明确是否保留一跳迁移工具，以及若保留时的使用口径
- 记录 v0.4 发布说明与剩余技术债

## Done When

- 活跃文档与 CLI help 不再把旧控制文件或兼容 wrapper 当正式主路径
- acceptance smoke 覆盖 `runtime / source / orbit_template / harness_template` 四类 revision kind
- clean repo 与 migrated repo 至少各有一条基础验证路径

## Notes

- 对应 `docs/orbit_v0_4_development_plan.md` 的 Phase 7
- 应在前置主链路 issue 基本完成后执行

## Progress

- v0.4 主链路的 local issue 追踪已开始按代码现状收口，`0111/0112/0113/0114/0119/0120/0121/0122` 已归档。
- `orbit template save --help` 现已补齐 install-record-driven writeback 示例与说明，命令入口开始反映 runtime -> template 的正式主路径，而不是只强调显式 `--to`。
- `template publish`、`brief materialize/backfill`、`harness template save` 等命令帮助已在前几轮开发中跟随 v0.4 主链路更新。
- `scripts/acceptance_quickstart.sh` 现在会在隔离 `HOME` 的 acceptance 环境里保留宿主 `GOMODCACHE/GOCACHE`，避免 quickstart smoke 因临时 HOME 下的依赖冷启动而被外部网络波动放大。
- quickstart 文档与 acceptance 现在都已显式覆盖 clean runtime 与 migrated runtime 两条 runtime writeback 基本路径。
- quickstart acceptance 现已补上独立 source authoring repo，四类 revision kind (`runtime / source / orbit_template / harness_template`) 都有基础 smoke 覆盖。
- 活跃 `quickstart` / `release` / `testing-strategy` 文档现已补上 v0.4 主线口径，并通过独立 docs smoke 防止再次回退到旧 `v0.3` 表述。
- 本 issue 的完成条件已满足：活跃 docs/help 已切到 v0.4 主路径，acceptance smoke 已覆盖四类 revision kind，clean/migrated runtime 也都有基础 writeback 验证路径。

## Remaining Work

- 后续若继续有 examples、release/upgrade 周边说明或兼容 wrapper 收敛工作，可作为新的 follow-up issue 独立推进。
