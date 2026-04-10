# ISSUE-0090 Harness Install Phase Progress Wiring

- Status: closed
- Priority: high
- Owner:
- Created: 2026-04-01
- Updated: 2026-04-01

## Summary

把 progress emitter 接到 `harness install` 的关键阶段，让本地 orbit install、远程 orbit install、source alias 解析、harness template install 都具备稳定阶段反馈。

## Scope

- 至少接入以下阶段：
  - `resolving install source`
  - `resolving remote template candidates`
  - `source branch detected; resolving published template`
  - `fetching selected template`
  - `resolving bindings`
  - `checking conflicts`
  - `writing files`
  - `updating runtime metadata`
  - `install complete`
- harness template 路径可增加：
  - `loading harness template manifest`
  - `validating harness template members`
- 保持错误链路 fail-closed，不吞掉原始错误
- 失败时不要求结构化阶段对象，但必须保证最后输出的阶段对排查有意义

## Done When

- 本地 orbit template install 有阶段输出
- 远程 orbit template install 有阶段输出
- source alias 解析时会出现专用阶段
- harness template install 有阶段输出
- dry-run 也有合适的阶段反馈

## Notes

- 依赖 `ISSUE-0089`
- 第一版不要求底层 Git transport 透传原生 fetch 进度
- 优先保持命令文件薄，复杂逻辑留在可复用 helper 中
