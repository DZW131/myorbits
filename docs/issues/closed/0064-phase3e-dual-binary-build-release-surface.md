# ISSUE-0064 Phase 3E Dual Binary Build Release Surface

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-26
- Updated: 2026-03-26

## Summary

把 `harness` 纳入正式 build / release / CI 产物，确保双二进制不会在发布链路、命令发现入口或 shell completion 中丢失其中一个。

## Scope

- 复查并补齐 build / release 配置中的双二进制产物：
  - `orbit`
  - `harness`
- 补齐安装说明、命令发现入口或等价的用户入口文档。
- 复查 shell completion、帮助入口和打包脚本是否已经覆盖 `harness`。
- 在 CI 或等价校验里保证双二进制都能构建。

## Done When

- 正式构建链路稳定产出 `orbit` 和 `harness`。
- 发布或安装说明中能明确发现 `harness` 二进制。
- 不存在只验证 `orbit`、遗漏 `harness` 的发布链路缺口。

## Notes

- 对应 `docs/harness_centric_runtime_development_plan.md` 的 12.3 / 任务 2。
- 对应 `docs/harness_centric_runtime_technical_spec.md` 的双二进制实现方向与 `docs/harness_cli_full_recommendations.md` 的命令前缀原则。
- Completed:
  - 新增 `scripts/build_binaries.sh`，现在有一条明确的双二进制构建链路可同时产出 `orbit` 与 `harness`。
  - `mise run build` 已接入这条正式构建链路，默认输出到 `./.dist/bin`。
  - 新增 `scripts/test_build_binaries.sh`，直接校验两个二进制都能构建、跑 `--help`，并暴露 `completion` 入口。
  - `test:scripts` 已把双二进制构建校验纳入仓库基线。
  - README 已补本地双二进制构建与 shell completion 的发现入口。
