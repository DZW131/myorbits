# ISSUE-0126 v0.4 Harness Agents Compose For Runtime Orchestration

- Status: open
- Priority: medium
- Owner:
- Created: 2026-04-09
- Updated: 2026-04-09

## Summary

为 harness 作者提供一个显式的根 `AGENTS.md` 组合入口，把 harness 级宏观编排和已安装 orbit 的 brief block 更自然地合成到 runtime 容器中，减少手工拼装负担。

## Scope

- 设计 `harness agents compose` 或等价命令
- 消费已安装 orbit 的 brief truth / materialized block
- 保持根 `AGENTS.md` 的容器语义与 block-preserving patch 语义
- 允许 harness 级宏观内容与 orbit brief block 明确分层
- 不把根 `AGENTS.md` 重新定义为 authored truth

## Done When

- harness 作者可以显式生成或更新当前 runtime 根 `AGENTS.md`
- orbit brief block 与 harness 级宏观 prose 的边界清晰
- 组合过程不会误覆盖其他 orbit block 或容器外文本
- 该命令与 `orbit brief materialize` / `orbit brief backfill` 的职责边界清楚

## Notes

- 当前已有 primitive：
  - `orbit brief materialize`
  - `orbit brief backfill`
  - runtime `AGENTS.md` block container patch
- 这张 issue 只负责 harness 级 orchestration 组合，不负责 export / publish
- 依赖 ISSUE-0123 的主流程定义
