# ISSUE-0080 Orbit Scope Visibility Layering

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-31
- Updated: 2026-03-31

## Summary

按 `docs/orbit_scope_visibility_technical_spec.md` 收口 Orbit scope visibility 分层：把 `always_visible` 迁到 `shared_scope`，新增 `projection_visible`，并把 projection scope 与 scoped-operation scope 正式拆开。

## Scope

- 协调本批次的执行子项：
  - `shared_scope` / `projection_visible` schema 与解析层落地
  - projection/status 与 scoped operations scope 分层
  - template save/publish 与 harness template candidate 改用 owned scope
- 保持 harness runtime host、install entry、source-branch / publish 工作流边界不变。
- 以 `docs/orbit_scope_visibility_technical_spec.md` 为本批次唯一技术基线。

## Done When

- `shared_scope` / `projection_visible` 的产品合同和代码行为一致。
- `projection_visible` 只影响 projection / status，不进入 template save/publish 与 scoped ops。
- 相关测试矩阵和主文档已收口到新术语与新优先级。

## Notes

- 建议实现顺序：`0081 -> 0082 -> 0083`。
- 这是 Orbit kernel enhancement，不应顺手扩大 harness runtime 功能面。
- 已完成：
  - `shared_scope` / `projection_visible` schema 与解析层落地
  - projection scope 与 scoped-operation scope 已拆开
  - template save/publish 与 harness template candidate 已统一到 owned scope
- 已完成追加收口：
  - 默认 `shared_scope` 已改为空，与 visibility spec 对齐
  - 旧 MVP / context 背景文档已同步到 `shared_scope` / `projection_visible` 术语与新默认值
