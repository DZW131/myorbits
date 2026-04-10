# ISSUE-0069 V0.3 Quickstart Acceptance Smoke

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-26
- Updated: 2026-04-06

## Summary

把当前发布中的 v0.3 quickstart 主路径固化成一条自动化 acceptance smoke，降低 README / quickstart 与真实 CLI 行为重新漂移的风险。

## Scope

- 基于 `docs/quickstart.md` 的正式主路径实现 temp-repo acceptance script：
  - build `orbit` / `harness`
  - `harness init` 或 `harness create`
  - `orbit add`
  - `orbit template save`
  - `harness install`
  - `harness check`
  - `orbit enter` / `status`
  - `harness template save`
- 把脚本接入现有 `mise` task 或等价 CI smoke 入口。
- 保持脚本输出可定位失败步骤，不依赖交互输入。

## Done When

- 文档主路径有一条自动化 smoke flow 覆盖。
- CI / 本地校验可以稳定运行这条 acceptance。
- quickstart 的关键命令序列出现 drift 时能被自动发现。

## Notes

- 对应此前 `docs/technical-debt.md` 中关于 quickstart acceptance smoke 的残余项；该残余项已随本 issue 完成一起移除。
- 这项应优先复用现有 shell script 与 temp-repo harness，不重复发明新的 acceptance 基建。
- Completed:
  - 新增 `scripts/acceptance_quickstart.sh`，把 `docs/quickstart.md` 的正式主路径固化成一条 temp-repo acceptance smoke。
  - `scripts/acceptance_mvp.sh` 现在保留为 legacy alias，直接复用 quickstart smoke。
  - `mise.toml` 新增 `acceptance:quickstart`，并把 `test:scripts` 接到这条文档驱动 smoke 上。
  - smoke 现已覆盖双二进制构建、`harness init/create`、`orbit template save`、`harness install/check`、`orbit enter/status/diff/leave`、以及 `harness template save`。
