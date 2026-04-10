# ISSUE-0115 v0.4 Source Publish Pipeline

- Status: closed
- Priority: medium
- Owner:
- Created: 2026-04-07
- Updated: 2026-04-09

## Summary

把 `source` revision 的作者工作流收口为正式的 source publish pipeline，使复杂 orbit 能在保留 author-only 文件与开发辅助工具的同时，稳定发布到 installable `orbit_template` revision。

## Scope

- 定义 `kind=source` 的 manifest contract 与 bootstrap 行为
- 硬化 `source -> orbit_template` 的 `orbit template publish` 主链路
- 明确 source-only 文件、publish metadata、author tooling 的保留边界
- 补 source authoring 文档与 integration tests

## Done When

- `source` revision 能被稳定识别并支持正式 publish
- source publish 会生成 installable orbit template payload，而不是把 source 工作目录直接复制出去
- source authoring 与 direct template authoring 的工作流边界已写清楚并有测试覆盖

## Notes

- 对应 `docs/orbit_v0_4_technical_spec.md` 的第 4、9 节
- 对应 `docs/orbit_v0_4_development_plan.md` 的 Phase 5
- 依赖 ISSUE-0109、ISSUE-0113 与 ISSUE-0122

## Progress

- `kind=source` 的 manifest contract、branch classify/inspect、`template init-source` 与 `template publish` 主链路已经落地。
- source authoring 已优先使用 hosted `.harness/orbits/*.yaml`，`init-source` 还会把 legacy definition host 迁移到 hosted control。
- source publish 已 fail-closed 拒绝 legacy-only 或 stray legacy definitions，不再把 source 工作目录直接复制成 template payload。
- source `orbit template publish` 不再被遗留 `.orbit/template.yaml` marker 阻断；source publish 的 no-op 判断也已经收口到 branch manifest + exported files。
- source authoring 文档、help、quickstart smoke 与四类 revision kind 验收已在 ISSUE-0118 中完成收口，不再是 source publish pipeline 的未完成项。

## Resolution

- `source` revision 现在已经能稳定被识别、初始化并发布到 installable `orbit_template` revision；正式 source publish 主链不再依赖 legacy source/template marker。
- source authoring 已 hosted-first，并在 publish 入口上对 legacy-only / mixed legacy definitions fail-closed，确保 source branch 不会被误当成 template payload 直接发布。
- 与 source publish 配套的文档、help、quickstart smoke 和 revision-kind acceptance 已由 ISSUE-0118 收口；`0109` 与 `0110` 也已经完成，source pipeline 不再被 revision identity 或 hosted OrbitSpec cutover 阻塞。
- 剩余 source 兼容尾项已转入 `docs/technical-debt.md` 跟踪，例如 source-manifest codec duplication 与 `init-source` 迁移非事务性，不再作为 Phase 5 source publish pipeline 的阻塞项。

因此本 issue 按 v0.4 Phase 5 的主链完成标准关闭。
